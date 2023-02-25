package botsetup

import (
	"archive/tar"
	"compress/bzip2"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/digital-dream-labs/vector-bluetooth/ble"
	"github.com/kercre123/chipper/pkg/logger"
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

func doOTAStatus() {
	for {
		r := <-BleStatusChan
		if r.OTAStatus.Error != "" {
			OtaStatus = "Error downloading OTA: " + r.OTAStatus.Error
			return
		}
		if r.OTAStatus.PacketNumber == 0 || r.OTAStatus.PacketTotal == 0 {
			OtaStatus = "OTA download progress: 0%"
		} else {
			percent := r.OTAStatus.PacketNumber / r.OTAStatus.PacketTotal * 100
			OtaStatus = "OTA download progress: " + fmt.Sprint(float64(percent)) + "%"
			if float64(percent) == 100 || float64(percent) == 99 {
				OtaStatus = "OTA download is complete!"
				return
			}
		}
	}
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
		return client, err
	case <-time.After(5 * time.Second):
		return nil, errors.New("took more than 5 seconds to create client")
	}
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
	err := client.Connect(device)
	return err
}

func SendPin(pin string, client *ble.VectorBLE) error {
	if len([]rune(pin)) != 6 {
		return fmt.Errorf("error: length of pin must be 6")
	}
	err := client.SendPin(pin)
	return err
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
			return
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
	case r.URL.Path == "/api-ble/get_ssh_key":
		resp, err := BleClient.DownloadLogs()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
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
	case r.URL.Path == "/api-ble/do_auth":
		if !BleInited {
			fmt.Fprint(w, "error: init ble first")
			return
		}
		resp, err := AuthRobot(BleClient)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		if resp {
			BleClient.SDKProxy(&ble.SDKProxyRequest{
				URLPath: "/v1/send_onboarding_input",
				Body:    `{"onboarding_mark_complete_and_exit": {}}`,
			},
			)
		}
		fmt.Fprint(w, resp)
		return
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
