package botsetup

import (
	"encoding/json"
	"fmt"
	"net/http"
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
var BleInited bool

func InitBle() (*ble.VectorBLE, error) {
	client, err := ble.New()
	return client, err
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
				time.Sleep(time.Second / 2)
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
