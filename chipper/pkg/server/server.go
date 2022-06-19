package server

import (
	"github.com/digital-dream-labs/chipper/pkg/vtt"
//	log "github.com/digital-dream-labs/hugh/log"
)

type intentProcessor interface {
	ProcessIntent(*vtt.IntentRequest) (*vtt.IntentResponse, error)
	ProcessKnowledgeGraph(*vtt.KnowledgeGraphRequest) (*vtt.KnowledgeGraphResponse, error)
}

// Server defines the service used.
type Server struct {
	intent intentProcessor
	kg     intentProcessor
}

// New accepts a list of args and returns the service
func New(opts ...Option) (*Server, error) {
	cfg := options{
//		log: log.Base(),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	s := Server{
		intent: cfg.intent,
		kg:     cfg.kg,
	}

	return &s, nil

}
