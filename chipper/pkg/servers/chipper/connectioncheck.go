package server

import (
	"context"
	"strconv"
	"time"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

const (
	connectionCheckTimeout = 15 * time.Second
	check                  = "check"
)

// StreamingConnectionCheck is used by the end device to make sure it can successfully communicate
func (s *Server) StreamingConnectionCheck(stream pb.ChipperGrpc_StreamingConnectionCheckServer) error {
	req, err := stream.Recv()
	logger.Println("Incoming connection check from " + req.DeviceId)
	if err != nil {
		logger.Println("Connection check unexpected error")
		logger.Println(err)
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
			logger.Println("Connection check expiration. Frames Recieved: " + strconv.Itoa(int(frames)))
			toSend.Status = "Timeout"
			break receiveLoop
		default:
			req, suberr := stream.Recv()

			if suberr != nil || req == nil {
				err = suberr
				logger.Println("Connection check unexpected error. Frames Recieved: " + strconv.Itoa(int(frames)))
				logger.Println(err)

				toSend.Status = "Error"
				break receiveLoop
			}

			frames++
			toSend.FramesReceived = frames
			if frames >= framesPerRequest {
				logger.Println("Connection check success")
				toSend.Status = "Success"
				break receiveLoop
			}
		}
	}
	senderr := stream.Send(&toSend)
	if senderr != nil {
		logger.Println("Failed to send connection check response to client")
		logger.Println(err)
		return senderr
	}
	return err

}
