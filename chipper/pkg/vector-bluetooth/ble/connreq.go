package ble

import (
	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

func handleRtsConnRequest(v *VectorBLE, msg interface{}) (data []byte, cont bool, err error) {
	var cr *rts.RtsConnRequest

	switch v.ble.Version() {
	case rtsV2:
		t, ok := msg.(*rts.RtsConnection_2)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		cr = t.GetRtsConnRequest()

	case rtsV3:
		t, ok := msg.(*rts.RtsConnection_3)
		if !ok {
			return handlerUnsupportedTypeError()
		}

		cr = t.GetRtsConnRequest()

	case rtsV4:
		t, ok := msg.(*rts.RtsConnection_4)
		if !ok {
			return handlerUnsupportedTypeError()
		}

		cr = t.GetRtsConnRequest()

	case rtsV5:
		t, ok := msg.(*rts.RtsConnection_5)
		if !ok {
			return handlerUnsupportedTypeError()
		}

		cr = t.GetRtsConnRequest()

	default:
		return handlerUnsupportedVersionError()
	}

	if err := v.ble.SetRemotePublicKey(cr); err != nil {
		return nil, false, err
	}

	b, err := rts.GetConnResponse(v.ble.Version(), v.ble.GetRemotePublicKey())
	if err != nil {
		return nil, false, err
	}

	if err := v.ble.Send(b); err != nil {
		return nil, false, err
	}
	return nil, true, nil
}
