package ble

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

// StatusResponse is the unified response for status messages
type StatusResponse struct {
	WifiSSID      string `json:"wifi_ssid"`
	Version       string `json:"version"`
	ESN           string `json:"esn"`
	WifiState     int    `json:"wifi_state"`
	AccessPoint   bool   `json:"access_point"`
	OtaInProgress bool   `json:"ota_in_progress"`
	HasOwner      bool   `json:"has_owner"`
	CloudAuthed   bool   `json:"cloud_authed"`
}

// Marshal converts a StatusResponse message to bytes
func (sr *StatusResponse) Marshal() ([]byte, error) {
	return json.Marshal(sr)
}

// Unmarshal converts a StatusResponse byte slice to a StatusResponse
func (sr *StatusResponse) Unmarshal(b []byte) error {
	return json.Unmarshal(b, sr)
}

// GetStatus sends a GetStatus message to the vector robot
func (v *VectorBLE) GetStatus() (*StatusResponse, error) {
	if !v.state.getAuth() {
		return nil, errors.New(errNotAuthorized)
	}

	msg, err := rts.BuildStatusMessage(v.ble.Version())
	if err != nil {
		return nil, err
	}

	if err := v.ble.Send(msg); err != nil {
		return nil, err
	}

	b, err := v.watch()

	resp := StatusResponse{}
	if err := resp.Unmarshal(b); err != nil {
		return nil, err
	}
	return &resp, err
}

func handleRSTStatusResponse(v *VectorBLE, msg interface{}) (data []byte, cont bool, err error) {
	var sr StatusResponse
	switch v.ble.Version() {
	case rtsV2:
		t, ok := msg.(*rts.RtsConnection_2)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		r := t.GetRtsStatusResponse2()
		ssid, _ := hex.DecodeString(r.WifiSsidHex)

		sr = StatusResponse{
			WifiSSID:      string(ssid),
			Version:       r.Version,
			WifiState:     int(r.WifiState),
			AccessPoint:   r.AccessPoint,
			OtaInProgress: r.OtaInProgress,
		}

	case rtsV3:
		t, ok := msg.(*rts.RtsConnection_3)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		r := t.GetRtsStatusResponse3()
		ssid, _ := hex.DecodeString(r.WifiSsidHex)

		sr = StatusResponse{
			WifiSSID:      string(ssid),
			Version:       r.Version,
			WifiState:     int(r.WifiState),
			AccessPoint:   r.AccessPoint,
			OtaInProgress: r.OtaInProgress,
		}

	case rtsV4:
		t, ok := msg.(*rts.RtsConnection_4)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		r := t.GetRtsStatusResponse4()
		ssid, _ := hex.DecodeString(r.WifiSsidHex)

		sr = StatusResponse{
			WifiSSID:      string(ssid),
			Version:       r.Version,
			WifiState:     int(r.WifiState),
			AccessPoint:   r.AccessPoint,
			OtaInProgress: r.OtaInProgress,
		}

	case rtsV5:
		t, ok := msg.(*rts.RtsConnection_5)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		r := t.GetRtsStatusResponse5()
		ssid, _ := hex.DecodeString(r.WifiSsidHex)

		sr = StatusResponse{
			WifiSSID:      string(ssid),
			Version:       r.Version,
			WifiState:     int(r.WifiState),
			AccessPoint:   r.AccessPoint,
			OtaInProgress: r.OtaInProgress,
			HasOwner:      r.HasOwner,
			CloudAuthed:   r.IsCloudAuthed,
		}

	default:
		return handlerUnsupportedVersionError()
	}

	b, err := sr.Marshal()
	return b, false, err
}
