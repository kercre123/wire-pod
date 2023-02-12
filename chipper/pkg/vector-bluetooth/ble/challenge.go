package ble

import (
	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

// nolint
func handleRTSChallengeMessage(v *VectorBLE, msg interface{}) ([]byte, bool, error) {
	var m uint32

	switch v.ble.Version() {
	case rtsV2:
		t, ok := msg.(*rts.RtsConnection_2)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		m = t.GetRtsChallengeMessage().Number

	case rtsV3:
		t, ok := msg.(*rts.RtsConnection_3)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		m = t.GetRtsChallengeMessage().Number

	case rtsV4:
		t, ok := msg.(*rts.RtsConnection_4)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		m = t.GetRtsChallengeMessage().Number

	case rtsV5:
		t, ok := msg.(*rts.RtsConnection_5)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		m = t.GetRtsChallengeMessage().Number

	default:
		return handlerUnsupportedVersionError()
	}

	b, err := rts.BuildChallengeResponse(v.ble.Version(), m)
	if err != nil {
		return nil, false, err
	}

	if err := v.ble.Send(b); err != nil {
		return nil, false, err
	}
	return nil, true, nil
}

func handleRTSChallengeSuccessMessage(v *VectorBLE, msg interface{}) (data []byte, cont bool, err error) {
	v.state.setAuth(true)
	return nil, false, nil
}
