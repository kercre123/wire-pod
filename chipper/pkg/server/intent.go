package server

import (
	"time"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/chipper/pkg/vtt"
	"github.com/digital-dream-labs/hugh/log"
)

// StreamingIntent handles voice streams
func (s *Server) StreamingIntent(stream pb.ChipperGrpc_StreamingIntentServer) error {
	recvTime := time.Now()

	req, err := stream.Recv()
	if err != nil {
		log.WithFields(log.Fields{
			"action": "recv_first_request",
			"status": "error",
		}).Info(err)

		return err
	}

	if _, err = s.intent.ProcessIntent(
		&vtt.IntentRequest{
			Time:       recvTime,
			Stream:     stream,
			Device:     req.DeviceId,
			Session:    req.Session,
			LangString: req.LanguageCode.String(),
			FirstReq:   req,
			AudioCodec: req.AudioEncoding,
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
