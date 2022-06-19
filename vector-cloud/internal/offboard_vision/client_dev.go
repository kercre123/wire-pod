// +build !shipping

package offboard_vision

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/vision"

	"github.com/digital-dream-labs/vector-cloud/internal/config"
	"github.com/digital-dream-labs/vector-cloud/internal/ipc"
	"github.com/digital-dream-labs/vector-cloud/internal/log"
	"github.com/digital-dream-labs/vector-cloud/internal/util"

	pb "github.com/digital-dream-labs/vector-cloud/internal/proto/vision"

	"github.com/google/uuid"
	"github.com/gwatts/rootcerts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	deviceID = "vicos-build"
	// TODO: Actually read this in instead of hard coding it. (VIC-13955)
	modes = []string{"people", "faces"}
)

type client struct {
	ipc.Conn
}

var (
	defaultGroupName = "offboard_vision"
	defaultTLSCert   = credentials.NewClientTLSFromCert(rootcerts.ServerCertPool(), "")
	devURLReader     func(string) ([]byte, error, bool)
)

func (c *client) handleConn(ctx context.Context) {
	for {
		msgbuf := c.ReadBlock()
		if msgbuf == nil || len(msgbuf) == 0 {
			return
		}
		var msg vision.OffboardImageReady
		if err := msg.Unpack(bytes.NewBuffer(msgbuf)); err != nil {
			log.Println("Could not unpack offboard vision request:", err)
			continue
		}

		resp, err := c.handleRequest(ctx, &msg)
		if err != nil {
			log.Println("Error handling offboard vision request:", err)
			// Nothing left to do, we should just continue
			continue
		}

		var buf bytes.Buffer
		if err := resp.Pack(&buf); err != nil {
			log.Println("Error packing offboard vision response:", err)
		} else if n, err := c.Write(buf.Bytes()); n != buf.Len() || err != nil {
			log.Println("Error sending offboard vision response:", fmt.Sprintf("%d/%d,", n, buf.Len()), err)
		}
	}
}

func (c *client) handleRequest(ctx context.Context, msg *vision.OffboardImageReady) (*vision.OffboardResultReady, error) {
	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, util.CommonGRPC()...)
	dialOpts = append(dialOpts, grpc.WithInsecure())

	var wg sync.WaitGroup

	// Dial server and read file data in parallel.
	// It isn't clear this parallelism is entirely neccesary however
	// it's not causing any issues now and since this was copied over
	// from a demo branch (the box) just sticking with it for now
	// until there is a good motivation to change it.
	var rpcConn *grpc.ClientConn
	var rpcErr error
	rpcClose := func() error { return nil }

	var fileData []byte
	var fileErr error

	// Dial server, make it blocking
	launchProcess(&wg, func() {
		rpcConn, rpcErr = grpc.DialContext(ctx, *config.Env.OffboardVision, append(dialOpts, grpc.WithBlock())...)
		if rpcErr == nil {
			rpcClose = rpcConn.Close
		}
	})

	// Read file data
	launchProcess(&wg, func() {
		if devURLReader != nil {
			var handled bool
			if fileData, fileErr, handled = devURLReader(msg.Filename); handled {
				return
			}
		}
		fileData, fileErr = ioutil.ReadFile(msg.Filename)
	})

	// Wait for both routines above to finish
	wg.Wait()
	// If rpc connection didn't fail, this will be set to rpcConn.Close()
	defer rpcClose()

	err := util.NewErrors(rpcErr, fileErr).Error()
	if err != nil {
		return nil, err
	}

	sessionID := uuid.New().String()[:16]
	r := &pb.ImageRequest{
		Session:     sessionID,
		DeviceId:    deviceID,
		Lang:        "en",
		ImageData:   fileData,
		TimestampMs: msg.Timestamp,
		Modes:       modes,
	}
	// TODO don't hardcode this?
	r.Configs = &pb.ImageConfig{}
	r.Configs.GroupName = defaultGroupName

	client := pb.NewOffboardVisionGrpcClient(rpcConn)
	resp, err := client.AnalyzeImage(ctx, r)
	if err != nil {
		log.Println("image analysis error: ", err)
		return nil, err
	}
	log.Println("image analysis response: ", resp.String())

	var resultReady vision.OffboardResultReady
	resultReady.JsonResult = resp.RawResult
	resultReady.Timestamp = resp.TimestampMs

	return &resultReady, nil
}

func launchProcess(wg *sync.WaitGroup, launcher func()) {
	wg.Add(1)
	go func() {
		launcher()
		wg.Done()
	}()
}
