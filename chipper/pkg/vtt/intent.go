package vtt

import (
	"time"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
)

// IntentRequest is the necessary request type for VTT intent processors
type IntentRequest struct {
	Time       time.Time
	Stream     pb.ChipperGrpc_StreamingIntentServer
	Device     string
	Session    string
	LangString string
	FirstReq   *pb.StreamingIntentRequest
	AudioCodec pb.AudioEncoding
}

// IntentResponse is the response type VTT intent processors
type IntentResponse struct {
	Intent   *pb.IntentResponse
	Params   string
	Duration *time.Duration
}
