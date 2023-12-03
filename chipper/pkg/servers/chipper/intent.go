package server

import (
	"time"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vtt"
)

// StreamingIntent handles voice streams
func (s *Server) StreamingIntent(stream pb.ChipperGrpc_StreamingIntentServer) error {
	recvTime := time.Now()

	req, err := stream.Recv()
	if err != nil {
		logger.Println("Intent error")
		logger.Println(err)

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
			// Mode:
		},
	); err != nil {
		logger.Println("Intent error")
		logger.Println(err)
		return err
	}

	return nil
}
