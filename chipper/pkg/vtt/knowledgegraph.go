package vtt

import (
	"time"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
)

// KnowledgeGraphRequest is the necessary request type for VTT knowledge graph processors
type KnowledgeGraphRequest struct {
	Time       time.Time
	Stream     pb.ChipperGrpc_StreamingKnowledgeGraphServer
	Device     string
	Session    string
	LangString string
	FirstReq   *pb.StreamingKnowledgeGraphRequest
	Mode       pb.RobotMode
	AudioCodec pb.AudioEncoding
}

// KnowledgeGraphResponse is the response type VTT knowledge graph processors
type KnowledgeGraphResponse struct {
	Intent   *pb.KnowledgeGraphResponse
	Params   string
	Duration *time.Duration
}
