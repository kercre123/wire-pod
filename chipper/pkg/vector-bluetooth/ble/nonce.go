package ble

import (
	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

func handleRTSNonceRequest(v *VectorBLE, msg interface{}) (data []byte, cont bool, err error) {
	var nr *rts.RtsNonceMessage

	switch v.ble.Version() {
	case rtsV2:
		t, ok := msg.(*rts.RtsConnection_2)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		nr = t.GetRtsNonceMessage()
	case rtsV3:
		t, ok := msg.(*rts.RtsConnection_3)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		nr = t.GetRtsNonceMessage()
	case rtsV4:
		t, ok := msg.(*rts.RtsConnection_4)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		nr = t.GetRtsNonceMessage()
	case rtsV5:
		t, ok := msg.(*rts.RtsConnection_5)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		nr = t.GetRtsNonceMessage()

	default:
		return handlerUnsupportedVersionError()
	}

	if err := v.ble.SetNonces(nr); err != nil {
		return nil, false, err
	}

	b, err := rts.BuildNonceResponse(v.ble.Version())
	if err != nil {
		return nil, false, err
	}

	v.state.setNonce(b)

	return nil, false, nil
}
