//go:build inbuiltble
// +build inbuiltble

package botsetup

import (
	"archive/tar"
	"compress/bzip2"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/digital-dream-labs/vector-bluetooth/ble"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/mdnshandler"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// need JSONable type
type VectorsBle struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type WifiNetwork struct {
	SSID     string `json:"ssid"`
	AuthType int    `json:"authtype"`
}

var BleClient *ble.VectorBLE
var BleStatusChan chan ble.StatusChannel
var BleInited bool
var OtaStatus string
var LogStatus string

func doOTAStatus() {
	for {
		r := <-BleStatusChan
		if r.OTAStatus.Error != "" {
			OtaStatus = "Error downloading OTA: " + r.OTAStatus.Error
			return
		}
		if r.OTAStatus.PacketNumber == 0 {
			OtaStatus = "OTA download progress: 0%"
		} else {
			percent := (float64(r.OTAStatus.PacketNumber) / float64(r.OTAStatus.PacketTotal)) * 100
			OtaStatus = "OTA download progress: " + fmt.Sprint(roundFloat(float64(percent), 3)) + "%"
			if r.OTAStatus.PacketNumber == r.OTAStatus.PacketTotal {
				OtaStatus = "OTA download is complete!"
				return
			}
		}
	}
}

func doLogStatus() {
	for {
		r := <-BleStatusChan
		if r.LogStatus.Error != "" {
			LogStatus = "Error downloading logs: " + r.LogStatus.Error
			return
		}
		if r.LogStatus.PacketNumber == 0 {
			LogStatus = "Log download progress: 0%"
		} else {
			percent := (float64(r.LogStatus.PacketNumber) / float64(r.LogStatus.PacketTotal)) * 100
			OtaStatus = "Log download progress: " + fmt.Sprint(roundFloat(float64(percent), 3)) + "%"
			if r.LogStatus.PacketNumber == r.LogStatus.PacketTotal {
				OtaStatus = "Log download is complete!"
				return
			}
		}
	}
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func IsDevRobot(esn, firmware string) bool {
	if strings.Contains(firmware, "ankidev") {
		return true
	}
	if strings.HasPrefix(esn, "00e") {
		return true
	}
	return false
}

func InitBle() (*ble.VectorBLE, error) {
	BleStatusChan = nil
	BleStatusChan = make(chan ble.StatusChannel)
	done := make(chan bool)
	var client *ble.VectorBLE
	var err error

	go func() {
		client, err = ble.New(
			ble.WithStatusChan(BleStatusChan),
			ble.WithLogDirectory(os.TempDir()),
		)
		done <- true
	}()

	select {
	case <-done:
		if err != nil {
			if strings.Contains(err.Error(), "hci0: can't down device") || strings.Contains(err.Error(), "hci0: can't up device") {
				FixBLEDriver()
				client, err = ble.New(
					ble.WithStatusChan(BleStatusChan),
					ble.WithLogDirectory(os.TempDir()),
				)
				return client, err
			}
		}
		return client, err
	case <-time.After(5 * time.Second):
		done2 := make(chan bool)
		FixBLEDriver()
		go func() {
			client, err = ble.New(
				ble.WithStatusChan(BleStatusChan),
				ble.WithLogDirectory(os.TempDir()),
			)
			done2 <- true
		}()

		select {
		case <-done2:
			return client, err
		case <-time.After(5 * time.Second):
			return nil, errors.New("error: took more than 5 seconds")
		}
	}
}

func FixBLEDriver() {
	logger.Println("BLE driver has broken. Removing then inserting bluetooth kernel modules")
	rmmodList := []string{"btusb", "btrtl", "btmtk", "btintel", "btbcm"}
	modprobeList := []string{"btrtl", "btmtk", "btintel", "btbcm", "btusb"}
	for _, mod := range rmmodList {
		exec.Command("/bin/rmmod", mod).Run()
	}
	time.Sleep(time.Second / 2)
	for _, mod := range modprobeList {
		exec.Command("/bin/modprobe", mod).Run()
	}
	time.Sleep(time.Second / 2)
}

func ScanForVectors(client *ble.VectorBLE) ([]VectorsBle, error) {
	var returnDevices []VectorsBle
	resp, err := client.Scan()
	if err != nil {
		return nil, err
	}
	for _, device := range resp.Devices {
		var vectorble VectorsBle
		vectorble.Address = device.Address
		vectorble.ID = device.ID
		vectorble.Name = device.Name
		returnDevices = append(returnDevices, vectorble)
	}
	return returnDevices, nil
}

func ConnectVector(client *ble.VectorBLE, device int) error {
	done := make(chan bool)
	var err error
	go func() {
		err = client.Connect(device)
		done <- true
	}()

	select {
	case <-done:
		return err
	case <-time.After(5 * time.Second):
		FixBLEDriver()
		return errors.New("error: took more than 5 seconds")
	}
}

func SendPin(pin string, client *ble.VectorBLE) error {
	if len([]rune(pin)) != 6 {
		return fmt.Errorf("error: length of pin must be 6")
	}
	err := client.SendPin(pin)
	return err
}

func RobotStatus(client *ble.VectorBLE) string {
	var firmware string
	var esn string
	//v1.8.1.6051-453e582_os1.8.1.6051ep-1536e0d-202202282217
	//v0.9.0-12efb91_os0.9.0-3e8307e-201806191226

	status, _ := client.GetStatus()
	esn = status.ESN
	firmware = status.Version

	if IsDevRobot(esn, firmware) && strings.Contains(firmware, "0.9.0") {
		return "in_recovery_dev"
	}
	if strings.Contains(firmware, "0.9.0") {
		return "in_recovery_prod"
	}
	if IsDevRobot(esn, firmware) {
		return "in_firmware_dev"
	}
	if strings.Contains(firmware, "ep-") {
		return "in_firmware_ep"
	}
	return "in_firmware_nonep"
}

func AuthRobot(client *ble.VectorBLE) (bool, error) {
	resp, err := client.Auth("2vMhFgktH3Jrbemm2WHkfGN")
	if err != nil {
		return false, err
	}
	return resp.Success, nil
}

func BluetoothSetupAPI(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/api-ble/init":
		if BleInited {
			fmt.Fprint(w, "success (ble already initiated, disconnect to reinit)")
			return
		}
		var err error
		BleClient, err = InitBle()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		BleInited = true
		fmt.Fprint(w, "success")
		return
	case r.URL.Path == "/api-ble/scan":
		if !BleInited {
			fmt.Fprint(w, "error: init ble first")
			return
		}
		devices, err := ScanForVectors(BleClient)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		returnBytes, _ := json.Marshal(devices)
		w.Write(returnBytes)
		return
	case r.URL.Path == "/api-ble/connect":
		if !BleInited {
			fmt.Fprint(w, "error: init ble first")
			return
		}
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		err = ConnectVector(BleClient, id)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			if err.Error() == "error: took more than 5 seconds" {
				logger.Println("It took too long to connect to Vector. Quitting and letting systemd restart")
				os.Exit(1)
			}
		}
		fmt.Fprint(w, "success")
		return
	case r.URL.Path == "/api-ble/send_pin":
		if !BleInited {
			fmt.Fprint(w, "error: init ble first")
			return
		}
		pin := r.FormValue("pin")
		err := SendPin(pin, BleClient)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				logger.Println("Wrong BLE pin was entered (sendpin error = eof), reinitializing BLE client")
				BleClient.Close()
				time.Sleep(time.Second)
				BleClient, err = InitBle()
				if err != nil {
					fmt.Fprint(w, "error reinitializing ble: "+err.Error())
					return
				}
				fmt.Fprint(w, "incorrect pin")
			} else {
				fmt.Fprint(w, "error: "+err.Error())
			}
			return
		}
		fmt.Fprint(w, "success")
		return
	case r.URL.Path == "/api-ble/get_wifi_status":
		resp, err := BleClient.GetStatus()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, resp.WifiState)
		return
	case r.URL.Path == "/api-ble/get_firmware_version":
		//v1.8.1.6051-453e582_os1.8.1.6051ep-1536e0d-202202282217
		//v0.9.0-12efb91_os0.9.0-3e8307e-201806191226
		resp, err := BleClient.GetStatus()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		// split version to just get ankiversion
		splitOsAndRev := strings.Split(resp.Version, "_os")
		splitOs := strings.Split(splitOsAndRev[1], "-")[0]
		fmt.Fprint(w, splitOs)
		return
	case r.URL.Path == "/api-ble/get_ip_address":
		resp, err := BleClient.WifiIP()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, resp.IPv4)
		return
	case r.URL.Path == "/api-ble/start_ota":
		otaUrl := r.FormValue("url")
		if strings.TrimSpace(otaUrl) == "local" {
			logger.Println("Starting proxy download from archive.org")
			otaUrl = "http://" + vars.GetOutboundIP().String() + ":" + vars.WebPort + "/api/get_ota/vicos-2.0.1.6076ep.ota"
			logger.Println("(" + otaUrl + ")")
		}
		if strings.Contains(otaUrl, "https://") {
			fmt.Fprint(w, "error: ota URL must be http")
			return
		}
		go doOTAStatus()
		done := make(chan bool)
		var resp *ble.OTAStartResponse
		var err error

		go func() {
			resp, err = BleClient.OTAStart(otaUrl)
			done <- true
		}()

		select {
		case <-done:
			fmt.Println("done")
			if err != nil {
				fmt.Fprint(w, "error: "+err.Error())
				return
			}
			fmt.Fprint(w, resp.Status)
			return
		case <-time.After(15 * time.Second):
			fmt.Fprint(w, "likely success")
			return
		}
	case r.URL.Path == "/api-ble/get_ota_status":
		fmt.Fprint(w, OtaStatus)
		return
	case r.URL.Path == "/api-ble/stop_ota":
		resp, err := BleClient.OTACancel()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, "success: "+string(resp))
		return
	case r.URL.Path == "/api-ble/do_dev_setup":
		go doLogStatus()
		resp, err := BleClient.DownloadLogs()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		logger.Println(resp.Filename)
		zip, err := os.Open(resp.Filename)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		tarFile := bzip2.NewReader(zip)
		tarReader := tar.NewReader(tarFile)
		for {
			header, err := tarReader.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Fprint(w, "error: "+err.Error())
				return
			}
			name := header.FileInfo().Name()
			logger.Println(name)
		}
		fmt.Fprint(w, "done")
		return
	case r.URL.Path == "/api-ble/scan_wifi":
		var returnNetworks []WifiNetwork
		resp, err := BleClient.WifiScan()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		for _, network := range resp.Networks {
			var returnNetwork WifiNetwork
			returnNetwork.SSID = network.WifiSSID
			returnNetwork.AuthType = network.AuthType
			returnNetworks = append(returnNetworks, returnNetwork)
		}
		returnJson, _ := json.Marshal(returnNetworks)
		w.Write(returnJson)
		return
	case r.URL.Path == "/api-ble/connect_wifi":
		if r.FormValue("ssid") == "" || r.FormValue("password") == "" {
			fmt.Fprint(w, "error: ssid or password empty")
			return
		}
		authType, err := strconv.Atoi(r.FormValue("authType"))
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
		}
		resp, err := BleClient.WifiConnect(r.FormValue("ssid"), r.FormValue("password"), 15, authType)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, resp.Result)
	case r.URL.Path == "/api-ble/get_robot_status":
		if !BleInited {
			fmt.Fprint(w, "error: init ble first")
			return
		}
		fmt.Fprint(w, RobotStatus(BleClient))
	case r.URL.Path == "/api-ble/do_auth":
		if !BleInited {
			fmt.Fprint(w, "error: init ble first")
			return
		}
		for i := 0; i <= 3; i++ {
			resp, err := AuthRobot(BleClient)
			if err != nil {
				fmt.Fprint(w, "error: "+err.Error())
				return
			}
			if resp {
				time.Sleep(time.Second)
				fmt.Fprint(w, "success")
				return
			} else {
				if vars.APIConfig.Server.EPConfig {
					logger.Println("BLE authentication was not successful. Posting mDNS and trying again (" + fmt.Sprint(i) + "/3)...")
					mdnshandler.PostmDNSNow()
					time.Sleep(time.Second * 2)
					continue
				} else {
					logger.Println("BLE authentication was not successful")
					fmt.Fprint(w, "error authenticating")
					return
				}
			}
		}
		logger.Println("BLE authentication attempts exhausted")
		fmt.Fprint(w, "error authenticating")
		return
	case r.URL.Path == "/api-ble/reset_onboarding":
		BleClient.SDKProxy(
			&ble.SDKProxyRequest{
				URLPath: "/v1/send_onboarding_input",
				Body:    `{"onboarding_set_phase_request": {"phase": 2}}`,
			},
		)
	case r.URL.Path == "/api-ble/onboard":
		wAnim := r.FormValue("with_anim")
		if wAnim == "true" {
			BleClient.SDKProxy(
				&ble.SDKProxyRequest{
					URLPath: "/v1/send_onboarding_input",
					Body:    `{"onboarding_wake_up_request": {}}`,
				},
			)
			time.Sleep(time.Second * 21)
		}
		BleClient.SDKProxy(
			&ble.SDKProxyRequest{
				URLPath: "/v1/send_onboarding_input",
				Body:    `{"onboarding_mark_complete_and_exit": {}}`,
			},
		)
		fmt.Fprint(w, "done")
	case r.URL.Path == "/api-ble/disconnect":
		err := BleClient.Close()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		BleInited = false
		fmt.Fprint(w, "success")
		return
	}
}

func RegisterBLEAPI() {
	http.HandleFunc("/api-ble/", BluetoothSetupAPI)
}
