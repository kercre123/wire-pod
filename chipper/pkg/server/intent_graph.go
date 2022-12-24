package server

import (
	"time"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/chipper/pkg/vtt"
	"github.com/digital-dream-labs/hugh/log"
)

// StreamingIntentGraph handles intent graph request streams
func (s *Server) StreamingIntentGraph(stream pb.ChipperGrpc_StreamingIntentGraphServer) error {
	recvTime := time.Now()

	req, err := stream.Recv()
	if err != nil {
		log.WithFields(log.Fields{
			"action": "recv_first_request",
			"status": "error",
		}).Info(err)

		return err
	}

	if _, err = s.intentGraph.ProcessIntentGraph(
		&vtt.IntentGraphRequest{
			Time:       recvTime,
			Stream:     stream,
			Device:     req.DeviceId,
			Session:    req.Session,
			LangString: req.LanguageCode.String(),
			FirstReq:   req,
			AudioCodec: req.AudioEncoding,
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
