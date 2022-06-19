package logcollector

import (
	"context"
	"fmt"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"
	"github.com/digital-dream-labs/vector-cloud/internal/log"
)

type cladHandler struct {
	collector *logCollector
}

func newCladHandler(opts *options) (*cladHandler, error) {
	collector, err := newLogCollector(opts)
	if err != nil {
		return nil, err
	}

	return &cladHandler{collector}, nil
}

func (c *cladHandler) handleRequest(ctx context.Context, req *cloud.LogCollectorRequest) (*cloud.LogCollectorResponse, error) {
	switch req.Tag() {
	case cloud.LogCollectorRequestTag_Upload:
		return c.uploadRequest(ctx, req.GetUpload())
	}
	err := fmt.Errorf("Major error: received unknown tag %d", req.Tag())
	log.Println(err)
	return nil, err
}

// This is a global variabe used for performance issues (not referenced outside this file)
var connectErrorResponse = cloud.NewLogCollectorResponseWithErr(&cloud.LogCollectorErrorResponse{cloud.LogCollectorError_ErrorConnecting})

func (c *cladHandler) uploadRequest(ctx context.Context, cladReq *cloud.UploadRequest) (*cloud.LogCollectorResponse, error) {
	url, err := c.collector.Upload(ctx, cladReq.LogFileName)
	if err != nil {
		return connectErrorResponse, err
	}
	return cloud.NewLogCollectorResponseWithUpload(&cloud.UploadResponse{url}), nil
}
