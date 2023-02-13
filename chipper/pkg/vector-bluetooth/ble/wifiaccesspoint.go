package ble

import (
	"encoding/json"
	"errors"

	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

// WifiAPResponse is the unified response for wifi access point messages
type WifiAPResponse struct {
	Enabled  bool
	SSID     string
	Password string
}

// Marshal converts a WifiAPResponse message to bytes
func (sr *WifiAPResponse) Marshal() ([]byte, error) {
	return json.Marshal(sr)
}

// Unmarshal converts a WifiAPResponse byte slice to a WifiAPResponse
func (sr *WifiAPResponse) Unmarshal(b []byte) error {
	return json.Unmarshal(b, sr)
}

// WifiAccessPoint sends a WifiAccessPoint message to the vector robot
func (v *VectorBLE) WifiAccessPoint(enabled bool) (*WifiIPResponse, error) {
	if !v.state.getAuth() {
		return nil, errors.New(errNotAuthorized)
	}

	msg, err := rts.BuildWifiAccesspointMessage(v.ble.Version(), enabled)
	if err != nil {
		return nil, err
	}

	if err := v.ble.Send(msg); err != nil {
		return nil, err
	}

	b, err := v.watch()

	resp := WifiIPResponse{}
	if err := resp.Unmarshal(b); err != nil {
		return nil, err
	}

	return &resp, err
}

func handleRSTWifiAccessPointResponse(v *VectorBLE, msg interface{}) (data []byte, cont bool, err error) {
	var sr *rts.RtsWifiAccessPointResponse

	switch v.ble.Version() {
	case rtsV2:
		t, ok := msg.(*rts.RtsConnection_2)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsWifiAccessPointResponse()

	case rtsV3:
		t, ok := msg.(*rts.RtsConnection_3)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsWifiAccessPointResponse()

	case rtsV4:
		t, ok := msg.(*rts.RtsConnection_4)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsWifiAccessPointResponse()

	case rtsV5:
		t, ok := msg.(*rts.RtsConnection_5)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsWifiAccessPointResponse()

	default:
		return handlerUnsupportedVersionError()
	}

	resp := WifiAPResponse{
		Enabled:  sr.Enabled,
		SSID:     sr.Ssid,
		Password: sr.Password,
	}
	b, err := resp.Marshal()
	return b, false, err
}
