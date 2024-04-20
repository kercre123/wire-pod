//go:build windows || darwin || android || ios
// +build windows darwin android ios

package botsetup

import "github.com/kercre123/wire-pod/chipper/pkg/logger"

func RegisterBLEAPI() {
	logger.Println("Not registering BLE API because target isn't linux")
}
