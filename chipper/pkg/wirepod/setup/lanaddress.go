package botsetup

import (
	"strconv"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// RecordRobotLANAddress stores the address a robot reported for itself during
// BLE adoption, if it is usable. Reports whether one was recorded.
//
// Deliberately in an untagged file, unlike its only caller: ble.go needs the
// inbuiltble tag and libsodium to compile, so anything left in there is
// invisible to the test suite and absent from every shipped image. This half
// needs neither, so it is the half that can be tested.
//
// The robot supplies this address over BLE, which is why it goes through the
// same validation as an admin-entered one. A malformed value would be
// concatenated with ":" + port when wire-pod dials, and a robot that reports
// the proxy's own address would make wire-pod dial itself.
func RecordRobotLANAddress(esn, addr string) bool {
	esn = strings.TrimSpace(esn)
	addr = strings.TrimSpace(addr)

	if esn == "" {
		logger.Println("Proxy mode: robot reported an empty ESN, not capturing an address")
		return false
	}
	if addr == "" {
		logger.Println("Proxy mode: robot " + esn + " reported no Wi-Fi address, not capturing one")
		return false
	}
	if err := vars.ValidateRobotFacingAddress(addr); err != nil {
		logger.Println("Proxy mode: robot " + esn + " reported an unusable Wi-Fi address " +
			strconv.Quote(addr) + " (" + err.Error() + "), not capturing it")
		return false
	}

	vars.SetRobotLANAddress(esn, addr)
	vars.WriteConfigToDisk()
	logger.Println("Proxy mode: captured LAN address " + addr + " for robot " + esn)
	return true
}
