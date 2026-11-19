//go:build android

package botsetup

import (
	"fmt"
	"net/http"
)

func BluetoothSetupAPI(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "error: BLE setup isn't supported on Android")
}

func RegisterBLEAPI() {
	http.HandleFunc("/api-ble/", BluetoothSetupAPI)
}
