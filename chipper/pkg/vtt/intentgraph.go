package vtt

import (
	"time"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
)

// IntentGraphRequest is the necessary request type for VTT intent processors
type IntentGraphRequest struct {
	Time       time.Time
	Stream     pb.ChipperGrpc_StreamingIntentGraphServer
	Device     string
	Session    string
	LangString string
	FirstReq   *pb.StreamingIntentGraphRequest
	AudioCodec pb.AudioEncoding

	// KnowledgeGraph specific
	Mode pb.RobotMode
}

// IntentGraphResponse is the response type VTT intent processors
type IntentGraphResponse struct {
	Intent   *pb.IntentGraphResponse
	Params   string
	Duration *time.Duration
}
