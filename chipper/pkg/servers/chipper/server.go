package server

import (
	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/kercre123/wire-pod/chipper/pkg/vtt"
)

type intentProcessor interface {
	ProcessIntent(*vtt.IntentRequest) (*vtt.IntentResponse, error)
}

type kgProcessor interface {
	ProcessKnowledgeGraph(*vtt.KnowledgeGraphRequest) (*vtt.KnowledgeGraphResponse, error)
}

type intentGraphProcessor interface {
	ProcessIntentGraph(*vtt.IntentGraphRequest) (*vtt.IntentGraphResponse, error)
}

// Server defines the service used.
type Server struct {
	intent      intentProcessor
	kg          kgProcessor
	intentGraph intentGraphProcessor

	pb.UnimplementedChipperGrpcServer
}

// New accepts a list of args and returns the service
func New(opts ...Option) (*Server, error) {
	cfg := options{
		//log: log.Base(),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	s := Server{
		intent:      cfg.intent,
		kg:          cfg.kg,
		intentGraph: cfg.intentGraph,
	}

	return &s, nil

}
