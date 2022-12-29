package server

import (
	"context"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TextIntent handles text-based request/responses from the device
func (s *Server) TextIntent(ctx context.Context, req *pb.TextRequest) (*pb.IntentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "")
}
