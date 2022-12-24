package server

import (
	"context"
	"time"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/hugh/log"
)

const (
	connectionCheckTimeout = 15 * time.Second
	check                  = "check"
)

// StreamingConnectionCheck is used by the end device to make sure it can successfully communicate
func (s *Server) StreamingConnectionCheck(stream pb.ChipperGrpc_StreamingConnectionCheckServer) error {

	req, err := stream.Recv()
	if err != nil {
		log.WithFields(log.Fields{
			"action": "recv_first_request",
			"status": "unexpected_error",
			"svc":    check,
			"error":  err,
			"msg":    "connection first request fail",
		}).Error(err)
		return err
	}

	ctx, cancel := context.WithTimeout(stream.Context(), connectionCheckTimeout)
	defer cancel()

	framesPerRequest := req.TotalAudioMs / req.AudioPerRequest

	var toSend pb.ConnectionCheckResponse

	// count frames, we already pulled the first one
	frames := uint32(1)
	toSend.FramesReceived = frames
receiveLoop:
	for {
		select {
		case <-ctx.Done():
			log.WithFields(log.Fields{
				"action":             "connection_check_expired",
				"svc":                check,
				"status":             "server_timeout",
				"frames_received":    frames,
				"frames_per_request": framesPerRequest,
				"device":             req.DeviceId,
				"session":            req.Session,
			}).Info()

			toSend.Status = "Timeout"
			break receiveLoop
		default:
			req, suberr := stream.Recv()

			if suberr != nil || req == nil {
				err = suberr
				log.WithFields(log.Fields{
					"action":             "recv_audio",
					"status":             "unexpected_error",
					"svc":                check,
					"error":              err,
					"frames_received":    frames,
					"frames_per_request": framesPerRequest,
				}).Warn(err)

				toSend.Status = "Error"
				break receiveLoop
			}

			frames++
			toSend.FramesReceived = frames
			if frames >= framesPerRequest {
				log.WithFields(log.Fields{
					"action":             "connection_check",
					"status":             "success",
					"svc":                check,
					"frames_received":    frames,
					"frames_per_request": framesPerRequest,
					"device":             req.DeviceId,
					"session":            req.Session,
				}).Info()

				toSend.Status = "Success"
				break receiveLoop
			}
		}
	}
	senderr := stream.Send(&toSend)
	if senderr != nil {
		log.WithFields(log.Fields{
			"action":  "send_client",
			"status":  "error",
			"error":   senderr,
			"svc":     check,
			"device":  req.DeviceId,
			"session": req.Session,
			"msg":     "fail to send connection check response to client",
		}).Warn()
		return senderr
	}
	return err

}
