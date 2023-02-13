package ble

import (
	"encoding/json"
	"errors"

	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

// WifiForgetResponse is the unified response for wifi fprget messages
type WifiForgetResponse struct {
	Status bool
}

// Marshal converts a WifiForgetResponse message to bytes
func (sr *WifiForgetResponse) Marshal() ([]byte, error) {
	return json.Marshal(sr)
}

// Unmarshal converts a WifiForgetResponse byte slice to a WifiForgetResponse
func (sr *WifiForgetResponse) Unmarshal(b []byte) error {
	return json.Unmarshal(b, sr)
}

// WifiForget sends a WifiForget message to the vector robot
func (v *VectorBLE) WifiForget(ssid string) (*WifiIPResponse, error) {
	if !v.state.getAuth() {
		return nil, errors.New(errNotAuthorized)
	}

	msg, err := rts.BuildWifiForgetMessage(v.ble.Version(), ssid, false)
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

func handleRSTWifiForgetResponse(v *VectorBLE, msg interface{}) (data []byte, cont bool, err error) {
	var sr *rts.RtsWifiForgetResponse

	switch v.ble.Version() {
	case rtsV3:
		t, ok := msg.(*rts.RtsConnection_3)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsWifiForgetResponse()

	case rtsV4:
		t, ok := msg.(*rts.RtsConnection_4)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsWifiForgetResponse()

	case rtsV5:
		t, ok := msg.(*rts.RtsConnection_5)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsWifiForgetResponse()

	default:
		return handlerUnsupportedVersionError()
	}

	resp := WifiForgetResponse{
		Status: sr.DidDelete,
	}
	b, err := resp.Marshal()
	return b, false, err
}
