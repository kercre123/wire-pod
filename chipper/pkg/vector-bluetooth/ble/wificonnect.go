package ble

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

// WifiConnectResponse is the unified response for wifi connect messages
type WifiConnectResponse struct {
	WifiSSID string `json:"wifi_ssid,omitempty"`
	State    int    `json:"state,omitempty"`
	Result   int    `json:"result,omitempty"`
}

// Marshal converts a WifiConnectResponse message to bytes
func (sr *WifiConnectResponse) Marshal() ([]byte, error) {
	return json.Marshal(sr)
}

// Unmarshal converts a WifiConnectResponse byte slice to a WifiConnectResponse
func (sr *WifiConnectResponse) Unmarshal(b []byte) error {
	return json.Unmarshal(b, sr)
}

// WifiConnect sends a wifi connect message to the robot
func (v *VectorBLE) WifiConnect(ssid, password string, timeout, authtype int) (*WifiConnectResponse, error) {
	if !v.state.getAuth() {
		return nil, errors.New(errNotAuthorized)
	}

	msg, err := rts.BuildWifiConnectMessage(v.ble.Version(), ssid, password, timeout, authtype)
	if err != nil {
		return nil, err
	}

	if err := v.ble.Send(msg); err != nil {
		return nil, err
	}

	b, err := v.watch()

	resp := WifiConnectResponse{}
	if err := resp.Unmarshal(b); err != nil {
		return nil, err
	}

	return &resp, err
}

func handleRSTWifiConnectionResponse(v *VectorBLE, msg interface{}) (data []byte, cont bool, err error) {
	var resp WifiConnectResponse
	switch v.ble.Version() {
	case rtsV2:
		t, ok := msg.(*rts.RtsConnection_2)
		if !ok {
			return handlerUnsupportedTypeError()
		}

		m := t.GetRtsWifiConnectResponse()

		ssid, _ := hex.DecodeString(m.WifiSsidHex)

		resp = WifiConnectResponse{
			WifiSSID: string(ssid),
			State:    int(m.WifiState),
		}

	case rtsV3:
		t, ok := msg.(*rts.RtsConnection_3)
		if !ok {
			return handlerUnsupportedTypeError()
		}

		m := t.GetRtsWifiConnectResponse3()

		ssid, _ := hex.DecodeString(m.WifiSsidHex)

		resp = WifiConnectResponse{
			WifiSSID: string(ssid),
			State:    int(m.WifiState),
			Result:   int(m.ConnectResult),
		}

	case rtsV4:
		t, ok := msg.(*rts.RtsConnection_4)
		if !ok {
			return handlerUnsupportedTypeError()
		}

		m := t.GetRtsWifiConnectResponse3()

		ssid, _ := hex.DecodeString(m.WifiSsidHex)

		resp = WifiConnectResponse{
			WifiSSID: string(ssid),
			State:    int(m.WifiState),
			Result:   int(m.ConnectResult),
		}

	case rtsV5:
		t, ok := msg.(*rts.RtsConnection_5)
		if !ok {
			return handlerUnsupportedTypeError()
		}

		m := t.GetRtsWifiConnectResponse3()

		ssid, _ := hex.DecodeString(m.WifiSsidHex)

		resp = WifiConnectResponse{
			WifiSSID: string(ssid),
			State:    int(m.WifiState),
			Result:   int(m.ConnectResult),
		}

	default:
		return handlerUnsupportedVersionError()
	}

	b, err := resp.Marshal()
	return b, false, err
}
