//go:build !inbuiltble
// +build !inbuiltble

package botsetup

import "github.com/kercre123/wire-pod/chipper/pkg/logger"

func RegisterBLEAPI() {
	logger.Println("BLE API is unregistered")
}
