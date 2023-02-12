package ble

import (
	"encoding/json"
	"errors"

	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

// SDKProxyRequest is the type required for SDK proxy messages
type SDKProxyRequest struct {
	URLPath string
	Body    string
}

// SDKProxyResponse is the response type of an SDKProxy request
type SDKProxyResponse struct {
	MessageID    string `json:"message_id,omitempty"`
	StatusCode   uint16 `json:"status_code,omitempty"`
	ResponseType string `json:"response_type,omitempty"`
	ResponseBody string `json:"response_body,omitempty"`
}

// Marshal converts a SDKProxyResponse message to bytes
func (sr *SDKProxyResponse) Marshal() ([]byte, error) {
	return json.Marshal(sr)
}

// Unmarshal converts a SDKProxyResponse byte slice to a SDKProxyResponse
func (sr *SDKProxyResponse) Unmarshal(b []byte) error {
	return json.Unmarshal(b, sr)
}

// SDKProxy sends a BLE-tunneled SDK request
func (v *VectorBLE) SDKProxy(settings *SDKProxyRequest) (*SDKProxyResponse, error) {
	if !v.state.getAuth() || v.state.getClientGUID() == "" {
		return nil, errors.New(errNotAuthorized)
	}

	msg, err := rts.BuildSDKMessage(
		v.ble.Version(),
		v.state.getClientGUID(),
		"1",
		settings.URLPath,
		settings.Body,
	)
	if err != nil {
		return nil, err
	}

	if err := v.ble.Send(msg); err != nil {
		return nil, err
	}

	b, err := v.watch()

	resp := SDKProxyResponse{}
	if err := resp.Unmarshal(b); err != nil {
		return nil, err
	}

	return &resp, err
}

func handleRSTSDKProxyResponse(v *VectorBLE, msg interface{}) (data []byte, cont bool, err error) {
	var r *rts.RtsSdkProxyResponse

	switch v.ble.Version() {
	case rtsV5:
		t, ok := msg.(*rts.RtsConnection_5)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		r = t.GetRtsSdkProxyResponse()

	default:
		return handlerUnsupportedVersionError()
	}

	resp := SDKProxyResponse{
		MessageID:    r.MessageId,
		StatusCode:   r.StatusCode,
		ResponseType: r.ResponseType,
		ResponseBody: r.ResponseBody,
	}

	b, err := resp.Marshal()
	return b, false, err
}
