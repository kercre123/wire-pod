//go:build !android

package botsetup

import (
	"archive/tar"
	"compress/bzip2"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/vector-bluetooth/ble"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/mdnshandler"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

type VectorsBle struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type WifiNetwork struct {
	SSID     string `json:"ssid"`
	AuthType int    `json:"authtype"`
}

var (
	bleMu         sync.Mutex
	bleClient     *ble.VectorBLE
	bleStatusStop chan struct{}

	bleStatusMu sync.Mutex
	otaStatus   string
)

func setOTAStatus(s string) {
	bleStatusMu.Lock()
	otaStatus = s
	bleStatusMu.Unlock()
}

func getOTAStatus() string {
	bleStatusMu.Lock()
	defer bleStatusMu.Unlock()
	return otaStatus
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func otaProgressString(s *ble.StatusCounter) string {
	if s.Error != "" {
		return "Error downloading OTA: " + s.Error
	}
	if s.PacketTotal != 0 && s.PacketNumber == s.PacketTotal {
		return "OTA download is complete!"
	}
	if s.PacketNumber == 0 || s.PacketTotal == 0 {
		return "OTA download progress: 0%"
	}
	percent := float64(s.PacketNumber) / float64(s.PacketTotal) * 100
	return "OTA download progress: " + fmt.Sprint(roundFloat(percent, 3)) + "%"
}

func watchBLEStatus(ch chan ble.StatusChannel, stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case r := <-ch:
			if r.OTAStatus != nil {
				setOTAStatus(otaProgressString(r.OTAStatus))
			}
		}
	}
}

func IsDevRobot(esn, firmware string) bool {
	return strings.Contains(firmware, "ankidev") || strings.Contains(firmware, "Unlock")
}

func closeBLEClient(c *ble.VectorBLE) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Println("BLE close panicked (likely never connected to a bot): ", r)
			err = nil
		}
	}()
	return c.Close()
}

func newBLEClient() (*ble.VectorBLE, chan ble.StatusChannel, error) {
	type result struct {
		client *ble.VectorBLE
		err    error
	}
	ch := make(chan ble.StatusChannel)
	res := make(chan result, 1)
	go func() {
		c, err := ble.New(
			ble.WithStatusChan(ch),
			ble.WithLogDirectory(os.TempDir()),
		)
		res <- result{c, err}
	}()
	select {
	case r := <-res:
		return r.client, ch, r.err
	case <-time.After(10 * time.Second):
		go func() {
			if r := <-res; r.client != nil {
				closeBLEClient(r.client)
			}
		}()
		return nil, nil, errors.New("error: took more than 10 seconds to init BLE")
	}
}

func initBLE() error {
	client, ch, err := newBLEClient()
	if err != nil {
		return err
	}
	if err := checkAdapterReady(client); err != nil {
		closeBLEClient(client)
		return err
	}
	bleClient = client
	bleStatusStop = make(chan struct{})
	go watchBLEStatus(ch, bleStatusStop)
	return nil
}

func resetBLE() error {
	if bleClient == nil {
		return nil
	}
	client := bleClient
	close(bleStatusStop)
	bleClient = nil
	bleStatusStop = nil
	return closeBLEClient(client)
}

func ScanForVectors(client *ble.VectorBLE) ([]VectorsBle, error) {
	resp, err := client.Scan()
	if err != nil {
		return nil, err
	}
	returnDevices := []VectorsBle{}
	for _, device := range resp.Devices {
		returnDevices = append(returnDevices, VectorsBle{
			ID:      device.ID,
			Name:    device.Name,
			Address: device.Address,
		})
	}
	return returnDevices, nil
}

func ConnectVector(client *ble.VectorBLE, device int) error {
	done := make(chan error, 1)
	go func() {
		done <- client.Connect(device)
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		return errors.New("error: took more than 5 seconds")
	}
}

func SendPin(pin string, client *ble.VectorBLE) error {
	if len([]rune(pin)) != 6 {
		return fmt.Errorf("error: length of pin must be 6")
	}
	return client.SendPin(pin)
}

func RobotStatus(client *ble.VectorBLE) (string, error) {
	status, err := client.GetStatus()
	if err != nil {
		return "", err
	}
	esn := status.ESN
	firmware := status.Version
	switch {
	case IsDevRobot(esn, firmware) && (strings.Contains(firmware, "0.9.0") || strings.Contains(firmware, "0.9.3")):
		return "in_recovery_dev", nil
	case (strings.Contains(firmware, "0.9.0") || strings.Contains(firmware, "0.9.3")):
		return "in_recovery_prod", nil
	case IsDevRobot(esn, firmware):
		if ip, err := bleRobotIP(client); err == nil && sshDoneFor(ip) {
			return "in_firmware_dev_ready", nil
		}
		return "in_firmware_dev", nil
	case strings.Contains(firmware, "ep-"):
		return "in_firmware_ep", nil
	}
	return "in_firmware_nonep", nil
}

func bleRobotIP(client *ble.VectorBLE) (string, error) {
	resp, err := client.WifiIP()
	if err != nil {
		return "", err
	}
	ip := net.ParseIP(strings.TrimSpace(resp.IPv4))
	if ip == nil || ip.IsUnspecified() {
		return "", errors.New("robot doesn't have an IP address, is it connected to Wi-Fi?")
	}
	return ip.String(), nil
}

func AuthRobot(client *ble.VectorBLE) (bool, error) {
	resp, err := client.Auth("2vMhFgktH3Jrbemm2WHkfGN")
	if err != nil {
		return false, err
	}
	return resp.Success, nil
}

func sendOnboardingInput(body string) error {
	_, err := bleClient.SDKProxy(&ble.SDKProxyRequest{
		URLPath: "/v1/send_onboarding_input",
		Body:    body,
	})
	return err
}

func BluetoothSetupAPI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api-ble/get_ota_status" {
		fmt.Fprint(w, getOTAStatus())
		return
	}

	bleMu.Lock()
	defer bleMu.Unlock()

	switch r.URL.Path {
	case "/api-ble/init":
		if bleClient != nil {
			if runtime.GOOS != "linux" || checkAdapterReady(bleClient) == nil {
				fmt.Fprint(w, "success (ble already initiated, disconnect to reinit)")
				return
			}
			resetBLE()
		}
		if err := initBLE(); err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, "success")
		return
	case "/api-ble/disconnect":
		if err := resetBLE(); err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, "success")
		return
	}

	if bleClient == nil {
		fmt.Fprint(w, "error: init ble first")
		return
	}

	switch r.URL.Path {
	case "/api-ble/scan":
		devices, err := ScanForVectors(bleClient)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		returnBytes, _ := json.Marshal(devices)
		w.Write(returnBytes)
	case "/api-ble/connect":
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		if err := ConnectVector(bleClient, id); err != nil {
			logger.Println("BLE connect error, resetting BLE client: " + err.Error())
			resetBLE()
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, "success")
	case "/api-ble/send_pin":
		err := SendPin(r.FormValue("pin"), bleClient)
		if err == nil {
			fmt.Fprint(w, "success")
			return
		}
		if !strings.Contains(err.Error(), "EOF") {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		logger.Println("Wrong BLE pin was entered (sendpin error = eof), reinitializing BLE client")
		resetBLE()
		time.Sleep(time.Second)
		if err := initBLE(); err != nil {
			fmt.Fprint(w, "error reinitializing ble: "+err.Error())
			return
		}
		fmt.Fprint(w, "incorrect pin")
	case "/api-ble/get_wifi_status":
		resp, err := bleClient.GetStatus()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, resp.WifiState)
	case "/api-ble/get_firmware_version":
		resp, err := bleClient.GetStatus()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		_, osVer, found := strings.Cut(resp.Version, "_os")
		if !found {
			fmt.Fprint(w, "error: unexpected version format: "+resp.Version)
			return
		}
		fmt.Fprint(w, strings.Split(osVer, "-")[0])
	case "/api-ble/get_ip_address":
		resp, err := bleClient.WifiIP()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, resp.IPv4)
	case "/api-ble/start_ota":
		otaUrl := strings.TrimSpace(r.FormValue("url"))
		if otaUrl == "local" {
			logger.Println("Starting proxy download from archive.org")
			otaUrl = "http://" + vars.GetOutboundIP().String() + ":" + vars.WebPort + "/api/get_ota/vicos-2.0.1.6076ep.ota"
			logger.Println("(" + otaUrl + ")")
		}
		if strings.Contains(otaUrl, "https://") {
			fmt.Fprint(w, "error: ota URL must be http")
			return
		}
		setOTAStatus("OTA download progress: 0%")
		type result struct {
			resp *ble.OTAStartResponse
			err  error
		}
		done := make(chan result, 1)
		client := bleClient
		go func() {
			resp, err := client.OTAStart(otaUrl)
			done <- result{resp, err}
		}()
		select {
		case res := <-done:
			if res.err != nil {
				fmt.Fprint(w, "error: "+res.err.Error())
				return
			}
			fmt.Fprint(w, res.resp.Status)
		case <-time.After(15 * time.Second):
			fmt.Fprint(w, "likely success")
		}
	case "/api-ble/stop_ota":
		resp, err := bleClient.OTACancel()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, "success: "+string(resp))
	case "/api-ble/do_dev_setup":
		resp, err := bleClient.DownloadLogs()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		logger.Println(resp.Filename)
		f, err := os.Open(resp.Filename)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		defer f.Close()
		tarReader := tar.NewReader(bzip2.NewReader(f))
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Fprint(w, "error: "+err.Error())
				return
			}
			logger.Println(header.FileInfo().Name())
		}
		fmt.Fprint(w, "done")
	case "/api-ble/scan_wifi":
		resp, err := bleClient.WifiScan()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		returnNetworks := []WifiNetwork{}
		for _, network := range resp.Networks {
			returnNetworks = append(returnNetworks, WifiNetwork{
				SSID:     network.WifiSSID,
				AuthType: network.AuthType,
			})
		}
		returnJson, _ := json.Marshal(returnNetworks)
		w.Write(returnJson)
	case "/api-ble/connect_wifi":
		authType, err := strconv.Atoi(r.FormValue("authType"))
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		if r.FormValue("ssid") == "" || (r.FormValue("password") == "" && authType != 0) {
			fmt.Fprint(w, "error: ssid or password empty")
			return
		}
		resp, err := bleClient.WifiConnect(r.FormValue("ssid"), r.FormValue("password"), 15, authType)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, resp.Result)
	case "/api-ble/get_robot_status":
		status, err := RobotStatus(bleClient)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, status)
	case "/api-ble/do_auth":
		for i := 0; i <= 3; i++ {
			ok, err := AuthRobot(bleClient)
			if err != nil {
				fmt.Fprint(w, "error: "+err.Error())
				return
			}
			if ok {
				time.Sleep(time.Second)
				fmt.Fprint(w, "success")
				return
			}
			if !vars.Config().Server.EPConfig {
				logger.Println("BLE authentication was not successful")
				fmt.Fprint(w, "error authenticating")
				return
			}
			logger.Println("BLE authentication was not successful. Posting mDNS and trying again (" + fmt.Sprint(i) + "/3)...")
			mdnshandler.PostmDNSNow()
			time.Sleep(time.Second * 2)
		}
		logger.Println("BLE authentication attempts exhausted")
		fmt.Fprint(w, "error authenticating")
	case "/api-ble/dev_ssh_setup":
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		keyBytes, err := readOptionalKey(r)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		if err := checkSSHReady(keyBytes); err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		ip, err := bleRobotIP(bleClient)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		if err := resetBLE(); err != nil {
			logger.Println("Error closing BLE before SSH setup: " + err.Error())
		}
		logger.Println("BLE disconnected, starting SSH setup for " + ip)
		if err := StartSSHSetup(ip, keyBytes); err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, "running "+ip)
	case "/api-ble/reset_onboarding":
		if err := sendOnboardingInput(`{"onboarding_set_phase_request": {"phase": 2}}`); err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, "done")
	case "/api-ble/onboard":
		if r.FormValue("with_anim") == "true" {
			if err := sendOnboardingInput(`{"onboarding_wake_up_request": {}}`); err != nil {
				fmt.Fprint(w, "error: "+err.Error())
				return
			}
			time.Sleep(time.Second * 21)
		}
		err := sendOnboardingInput(`{"onboarding_mark_complete_and_exit": {}}`)
		var syntaxErr *json.SyntaxError
		if err != nil && !errors.As(err, &syntaxErr) {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		if err := resetBLE(); err != nil {
			logger.Println("Error disconnecting BLE after onboarding: " + err.Error())
		}
		fmt.Fprint(w, "done")
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func RegisterBLEAPI() {
	http.HandleFunc("/api-ble/", BluetoothSetupAPI)
}
