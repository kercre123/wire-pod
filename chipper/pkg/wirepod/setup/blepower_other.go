//go:build !linux && !darwin

package botsetup

import (
	"errors"

	"github.com/kercre123/vector-bluetooth/ble"
)

func checkAdapterReady(client *ble.VectorBLE) error {
	if _, err := client.Scan(); err != nil {
		return errors.New("bluetooth isn't usable: " + err.Error())
	}
	return nil
}
