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
	resp, err := client.Auth("dontneedakey")
	if err != nil {
		return false, err
	}
	return resp.Success, nil
}

func BluetoothSetupAPI(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/api-ble/init":
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
				time.Sleep(time.Second / 3)
				BleClient, err = InitBle()
				if err != nil {
					fmt.Fprint(w, "error reinitializing ble: "+err.Error())
					return
				}
				fmt.Fprint(w, "incorrect pin or other unexplained error")
			}
			return
		}
		fmt.Fprint(w, "success")
		return
	case r.URL.Path == "/api-ble/get_firmware":
		if !BleInited {
			fmt.Fprint(w, "error: init ble first")
			return
		}
		resp, err := BleClient.GetStatus()
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, resp.Version)
		return
	case r.URL.Path == "/api-ble/auth":
		if !BleInited {
			fmt.Fprint(w, "error: init ble first")
			return
		}
		success, err := AuthRobot(BleClient)
		if err != nil {
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		fmt.Fprint(w, success)
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
