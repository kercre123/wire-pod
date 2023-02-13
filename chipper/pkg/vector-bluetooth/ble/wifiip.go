package ble

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

// WifiIPResponse is the unified response for wifi ip messages
type WifiIPResponse struct {
	IPv4 string
	IPv6 string
}

// Marshal converts a WifiIPResponse message to bytes
func (sr *WifiIPResponse) Marshal() ([]byte, error) {
	return json.Marshal(sr)
}

// Unmarshal converts a WifiIPResponse byte slice to a WifiIPResponse
func (sr *WifiIPResponse) Unmarshal(b []byte) error {
	return json.Unmarshal(b, sr)
}

// WifiIP sends a WifiIP message to the vector robot
func (v *VectorBLE) WifiIP() (*WifiIPResponse, error) {
	if !v.state.getAuth() {
		return nil, errors.New(errNotAuthorized)
	}

	msg, err := rts.BuildWifiIPMessage(v.ble.Version())
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

func handleRSTWifiIPResponse(v *VectorBLE, msg interface{}) (data []byte, cont bool, err error) {
	var sr *rts.RtsWifiIpResponse

	switch v.ble.Version() {
	case rtsV2:
		t, ok := msg.(*rts.RtsConnection_2)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsWifiIpResponse()

	case rtsV3:
		t, ok := msg.(*rts.RtsConnection_3)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsWifiIpResponse()

	case rtsV4:
		t, ok := msg.(*rts.RtsConnection_4)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsWifiIpResponse()

	case rtsV5:
		t, ok := msg.(*rts.RtsConnection_5)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsWifiIpResponse()

	default:
		return handlerUnsupportedVersionError()
	}

	resp := WifiIPResponse{
		IPv4: fmt.Sprintf("%d.%d.%d.%d", sr.IpV4[0], sr.IpV4[1], sr.IpV4[2], sr.IpV4[3]),
		IPv6: fmt.Sprintf(
			"%x%x:%x%x:%x%x:%x%x:%x%x:%x%x:%x%x:%x%x",
			sr.IpV6[0], sr.IpV6[1], sr.IpV6[2], sr.IpV6[3],
			sr.IpV6[4], sr.IpV6[5], sr.IpV6[6], sr.IpV6[7],
			sr.IpV6[8], sr.IpV6[9], sr.IpV6[10], sr.IpV6[11],
			sr.IpV6[12], sr.IpV6[13], sr.IpV6[4], sr.IpV6[15],
		),
	}
	b, err := resp.Marshal()
	return b, false, err
}
