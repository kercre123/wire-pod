package server

import (
	"time"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/kercre123/chipper/pkg/vtt"
	"github.com/digital-dream-labs/hugh/log"
)

// StreamingKnowledgeGraph is used for knowledge graph request/responses
func (s *Server) StreamingKnowledgeGraph(stream pb.ChipperGrpc_StreamingKnowledgeGraphServer) error {
	recvTime := time.Now()
	req, err := stream.Recv()
	if err != nil {
		log.WithFields(log.Fields{
			"action": "recv_first_request",
			"status": "error",
		}).Info(err)

		return err
	}

	if _, err = s.kg.ProcessKnowledgeGraph(
		&vtt.KnowledgeGraphRequest{
			Time:       recvTime,
			Stream:     stream,
			Device:     req.DeviceId,
			Session:    req.Session,
			LangString: req.LanguageCode.String(),
			FirstReq:   req,
			AudioCodec: req.AudioEncoding,
			// Why is this not passed
			// Mode:
		},
	); err != nil {
		log.WithFields(log.Fields{
			"action": "get_intent",
			"status": "failed",
		}).Info(err)
		return err
	}

	return nil
}
