package ble

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

// WifiScanResponse is the unified response for wifi scan messages
type WifiScanResponse struct {
	Networks []*WifiNetwork `json:"networks,omitempty"`
}

// WifiNetwork is an entry for one network
type WifiNetwork struct {
	WifiSSID       string `json:"wifi_ssid,omitempty"`
	SignalStrength int    `json:"signal_strength,omitempty"`
	AuthType       int    `json:"auth_type,omitempty"`
	Hidden         bool   `json:"hidden,omitempty"`
}

// Marshal converts a WifiScanResponse message to bytes
func (sr *WifiScanResponse) Marshal() ([]byte, error) {
	return json.Marshal(sr)
}

// Unmarshal converts a WifiScanResponse byte slice to a WifiScanResponse
func (sr *WifiScanResponse) Unmarshal(b []byte) error {
	return json.Unmarshal(b, sr)
}

// WifiScan sends a WifiScan message to the vector robot
func (v *VectorBLE) WifiScan() (*WifiScanResponse, error) {
	if !v.state.getAuth() {
		return nil, errors.New(errNotAuthorized)
	}

	msg, err := rts.BuildWifiScanMessage(v.ble.Version())
	if err != nil {
		return nil, err
	}

	if err := v.ble.Send(msg); err != nil {
		return nil, err
	}

	b, err := v.watch()

	resp := WifiScanResponse{}
	if err := resp.Unmarshal(b); err != nil {
		return nil, err
	}

	return &resp, err
}

func handleRSTWifiScanResponse(v *VectorBLE, msg interface{}) (data []byte, cont bool, err error) {
	nw := []*WifiNetwork{}

	switch v.ble.Version() {
	case rtsV2:
		t, ok := msg.(*rts.RtsConnection_2)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		m := t.GetRtsWifiScanResponse2()

		for _, v := range m.ScanResult {
			ssid, _ := hex.DecodeString(v.WifiSsidHex)

			tn := WifiNetwork{
				WifiSSID:       string(ssid),
				SignalStrength: int(v.SignalStrength),
				Hidden:         v.Hidden,
				AuthType:       int(v.AuthType),
			}
			nw = append(nw, &tn)
		}

	case rtsV3:
		t, ok := msg.(*rts.RtsConnection_3)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		m := t.GetRtsWifiScanResponse3()

		for _, v := range m.ScanResult {
			ssid, _ := hex.DecodeString(v.WifiSsidHex)

			tn := WifiNetwork{
				WifiSSID:       string(ssid),
				SignalStrength: int(v.SignalStrength),
				Hidden:         v.Hidden,
				AuthType:       int(v.AuthType),
			}
			nw = append(nw, &tn)
		}

	case rtsV4:
		t, ok := msg.(*rts.RtsConnection_4)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		m := t.GetRtsWifiScanResponse3()

		for _, v := range m.ScanResult {
			ssid, _ := hex.DecodeString(v.WifiSsidHex)

			tn := WifiNetwork{
				WifiSSID:       string(ssid),
				SignalStrength: int(v.SignalStrength),
				Hidden:         v.Hidden,
				AuthType:       int(v.AuthType),
			}
			nw = append(nw, &tn)
		}

	case rtsV5:
		t, ok := msg.(*rts.RtsConnection_5)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		m := t.GetRtsWifiScanResponse3()

		for _, v := range m.ScanResult {
			ssid, _ := hex.DecodeString(v.WifiSsidHex)

			tn := WifiNetwork{
				WifiSSID:       string(ssid),
				SignalStrength: int(v.SignalStrength),
				Hidden:         v.Hidden,
				AuthType:       int(v.AuthType),
			}
			nw = append(nw, &tn)
		}

	default:
		return handlerUnsupportedVersionError()
	}

	resp := WifiScanResponse{
		Networks: nw,
	}

	b, err := resp.Marshal()
	return b, false, err
}
