//go:build linux && !android

package botsetup

import (
	"errors"

	"github.com/godbus/dbus/v5"
	"github.com/kercre123/vector-bluetooth/ble"
)

// client is used in the "other" variant of this function
func checkAdapterReady(client *ble.VectorBLE) error {
	bus, err := dbus.SystemBus()
	if err != nil {
		return err
	}
	v, err := bus.Object("org.bluez", "/org/bluez/hci0").GetProperty("org.bluez.Adapter1.Powered")
	if err != nil {
		return errors.New("couldn't read bluetooth adapter state: " + err.Error())
	}
	if powered, ok := v.Value().(bool); !ok || !powered {
		return errors.New("bluetooth is turned off")
	}
	return nil
}
