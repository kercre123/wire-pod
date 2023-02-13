package ble

import (
	"encoding/json"
	"errors"

	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

// AuthStatus defines the status of the response message
type AuthStatus int

const (
	UnknownError AuthStatus = iota
	ConnectionError
	WrongAccount
	InvalidSessionToken
	AuthorizedAsPrimary
	AuthorizedAsSecondary
	Reauthorized
)

// AuthResponse is the unified response for Auth  messages
type AuthResponse struct {
	Status          AuthStatus `json:"status,omitempty"`
	ClientTokenGUID string     `json:"client_token_guid,omitempty"`
	Success         bool       `json:"success,omitempty"`
}

// Marshal converts a AuthResponse message to bytes
func (sr *AuthResponse) Marshal() ([]byte, error) {
	return json.Marshal(sr)
}

// Unmarshal converts a AuthResponse byte slice to a AuthResponse
func (sr *AuthResponse) Unmarshal(b []byte) error {
	return json.Unmarshal(b, sr)
}

// Auth sends a Auth message to the vector robot
func (v *VectorBLE) Auth(key string) (*AuthResponse, error) {
	msg, err := rts.BuildAuthMessage(v.ble.Version(), key)
	if err != nil {
		return nil, err
	}

	if err := v.ble.Send(msg); err != nil {
		return nil, err
	}

	b, err := v.watch()

	resp := AuthResponse{}
	if err := resp.Unmarshal(b); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, errors.New("authorization failed")
	}

	v.state.setClientGUID(resp.ClientTokenGUID)

	return &resp, err
}

func handleRSTCloudSessionResponse(v *VectorBLE, msg interface{}) (data []byte, cont bool, err error) {
	var m *rts.RtsCloudSessionResponse
	switch v.ble.Version() {
	case rtsV3:
		t, ok := msg.(*rts.RtsConnection_3)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		m = t.GetRtsCloudSessionResponse()

	case rtsV4:
		t, ok := msg.(*rts.RtsConnection_4)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		m = t.GetRtsCloudSessionResponse()

	case rtsV5:
		t, ok := msg.(*rts.RtsConnection_5)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		m = t.GetRtsCloudSessionResponse()

	default:
		return handlerUnsupportedVersionError()
	}

	resp := AuthResponse{
		Status:          AuthStatus(m.StatusCode),
		ClientTokenGUID: m.ClientTokenGuid,
		Success:         m.Success,
	}

	b, err := resp.Marshal()
	return b, false, err
}
