package ble

import (
	"errors"

	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

// OTACancel sends a OTACancel message to the vector robot
func (v *VectorBLE) OTACancel() ([]byte, error) {
	if !v.state.getAuth() {
		return nil, errors.New(errNotAuthorized)
	}
	msg, err := rts.BuildOTACancelMessage(v.ble.Version())
	if err != nil {
		return nil, err
	}

	if err := v.ble.Send(msg); err != nil {
		return nil, err
	}

	_, err = v.watch()

	return nil, err
}
