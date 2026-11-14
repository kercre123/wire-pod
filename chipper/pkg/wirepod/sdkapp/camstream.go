package sdkapp

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
)

const (
	// camControlTimeout bounds each EnableImageStreaming / IsImageStreamingEnabled call.
	camControlTimeout = 3 * time.Second
	// camIdleTimeout ends a stream that delivers no decodable frame for this long,
	// for example while the robot is asleep and its camera produces nothing.
	camIdleTimeout = 6 * time.Second
	camJPEGQuality = 50
	// camBoundary is the multipart boundary. The header names it without the
	// leading dashes, as RFC 2046 requires; each part starts with "--" + camBoundary.
	camBoundary = "frame"
)

// sdkCameraClient is the subset of the SDK the camera stream uses, so the
// handler can be tested without a robot.
type sdkCameraClient interface {
	IsImageStreamingEnabled(ctx context.Context) (bool, error)
	EnableImageStreaming(ctx context.Context, enable bool) error
	OpenCameraFeed(ctx context.Context) (func() ([]byte, error), error)
}

type liveSDKCameraClient struct {
	robotObj Robot
}

func (c liveSDKCameraClient) IsImageStreamingEnabled(ctx context.Context) (bool, error) {
	if c.robotObj.Vector == nil {
		return false, context.Canceled
	}
	cctx, cancel := context.WithTimeout(ctx, camControlTimeout)
	defer cancel()
	resp, err := c.robotObj.Vector.Conn.IsImageStreamingEnabled(cctx, &vectorpb.IsImageStreamingEnabledRequest{})
	if err != nil {
		return false, err
	}
	return resp.GetIsImageStreamingEnabled(), nil
}

func (c liveSDKCameraClient) EnableImageStreaming(ctx context.Context, enable bool) error {
	if c.robotObj.Vector == nil {
		return context.Canceled
	}
	cctx, cancel := context.WithTimeout(ctx, camControlTimeout)
	defer cancel()
	_, err := c.robotObj.Vector.Conn.EnableImageStreaming(cctx, &vectorpb.EnableImageStreamingRequest{
		Enable: enable,
	})
	return err
}

func (c liveSDKCameraClient) OpenCameraFeed(ctx context.Context) (func() ([]byte, error), error) {
	if c.robotObj.Vector == nil {
		return nil, context.Canceled
	}
	stream, err := c.robotObj.Vector.Conn.CameraFeed(ctx, &vectorpb.CameraFeedRequest{})
	if err != nil {
		return nil, err
	}
	return func() ([]byte, error) {
		resp, recvErr := stream.Recv()
		if recvErr != nil {
			return nil, recvErr
		}
		return resp.GetData(), nil
	}, nil
}

// One viewer per robot, as before. A new viewer replaces the current one.
// "owned" means this chain of viewers turned image streaming on and is
// therefore the one to turn it off; it passes to a replacing viewer.
type camViewer struct {
	id     uint64
	cancel context.CancelFunc
	owned  bool
}

var (
	camViewerMu  sync.Mutex
	camViewerSeq uint64
	camViewers   = map[string]camViewer{}
)

func registerCamViewer(esn string, cancel context.CancelFunc, enabledHere bool) uint64 {
	camViewerMu.Lock()
	defer camViewerMu.Unlock()
	camViewerSeq++
	owned := enabledHere
	if prev, ok := camViewers[esn]; ok {
		owned = owned || prev.owned
		prev.cancel()
	}
	camViewers[esn] = camViewer{id: camViewerSeq, cancel: cancel, owned: owned}
	return camViewerSeq
}

// unregisterCamViewer reports whether this viewer must turn streaming off:
// only when it was not replaced and its chain turned streaming on.
func unregisterCamViewer(esn string, id uint64) (disable bool) {
	camViewerMu.Lock()
	defer camViewerMu.Unlock()
	cur, ok := camViewers[esn]
	if !ok || cur.id != id {
		return false
	}
	delete(camViewers, esn)
	return cur.owned
}

// stopCamViewer ends the current stream for a robot, if any.
func stopCamViewer(esn string) {
	camViewerMu.Lock()
	defer camViewerMu.Unlock()
	if cur, ok := camViewers[normalizeESN(esn)]; ok {
		cur.cancel()
	}
}

func normalizeESN(esn string) string {
	return strings.TrimSpace(strings.ToLower(esn))
}

func handleCamStream(w http.ResponseWriter, r *http.Request) {
	esn := normalizeESN(r.FormValue("serial"))
	if esn == "" {
		http.Error(w, "error: missing serial", http.StatusBadRequest)
		return
	}
	robotObj, _, err := getRobot(esn)
	if err != nil {
		http.Error(w, "error: "+err.Error(), http.StatusBadGateway)
		return
	}
	streamCamera(w, r, esn, liveSDKCameraClient{robotObj: robotObj}, camIdleTimeout)
}

// streamCamera serves an MJPEG stream bound to the request: it ends when the
// client leaves, when no frame arrives for idleTimeout, or when replaced.
func streamCamera(w http.ResponseWriter, r *http.Request, esn string, client sdkCameraClient, idleTimeout time.Duration) {
	esn = normalizeESN(esn)
	if esn == "" {
		http.Error(w, "error: missing serial", http.StatusBadRequest)
		return
	}
	if idleTimeout <= 0 {
		idleTimeout = camIdleTimeout
	}
	// If the state cannot be read, enable anyway but do not claim it: never
	// turn off streaming that someone else may have turned on.
	enabled, checkErr := client.IsImageStreamingEnabled(r.Context())
	enabledHere := false
	if checkErr != nil || !enabled {
		if err := client.EnableImageStreaming(r.Context(), true); err != nil {
			http.Error(w, "error: "+err.Error(), http.StatusBadGateway)
			return
		}
		enabledHere = checkErr == nil
	}
	streamCtx, cancel := context.WithCancel(r.Context())
	viewerID := registerCamViewer(esn, cancel, enabledHere)
	defer func() {
		cancel()
		if unregisterCamViewer(esn, viewerID) {
			sctx, scancel := context.WithTimeout(context.Background(), camControlTimeout)
			_ = client.EnableImageStreaming(sctx, false) // best effort; the robot also stops on stream exit
			scancel()
		}
	}()

	var timedOut atomic.Bool
	watchdog := time.AfterFunc(idleTimeout, func() {
		timedOut.Store(true)
		cancel()
	})
	defer watchdog.Stop()

	recv, err := client.OpenCameraFeed(streamCtx)
	if err != nil {
		writeCamError(w, r, err, timedOut.Load())
		return
	}
	sendCamFrames(w, r, recv, watchdog, idleTimeout, &timedOut)
}

func sendCamFrames(w http.ResponseWriter, r *http.Request, recv func() ([]byte, error),
	watchdog *time.Timer, idleTimeout time.Duration, timedOut *atomic.Bool) {
	sentHeaders := false
	flusher, canFlush := w.(http.Flusher)
	for {
		frameBytes, err := recv()
		if err != nil {
			if !sentHeaders {
				writeCamError(w, r, err, timedOut.Load())
			}
			return
		}
		img, _, decErr := image.Decode(bytes.NewReader(frameBytes))
		if decErr != nil || img == nil {
			continue
		}
		var jpgBuf bytes.Buffer
		if encErr := jpeg.Encode(&jpgBuf, img, &jpeg.Options{Quality: camJPEGQuality}); encErr != nil {
			continue
		}
		watchdog.Reset(idleTimeout)
		if !sentHeaders {
			w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary="+camBoundary)
			sentHeaders = true
		}
		part := "--" + camBoundary + "\r\nContent-Type: image/jpeg\r\nContent-Length: " +
			strconv.Itoa(jpgBuf.Len()) + "\r\n\r\n"
		if _, werr := w.Write([]byte(part)); werr != nil {
			return
		}
		if _, werr := w.Write(append(jpgBuf.Bytes(), '\r', '\n')); werr != nil {
			return
		}
		if canFlush {
			flusher.Flush()
		}
	}
}

func writeCamError(w http.ResponseWriter, r *http.Request, err error, timedOut bool) {
	if r.Context().Err() != nil {
		return
	}
	if timedOut {
		http.Error(w, "error: camera_frame_timeout", http.StatusGatewayTimeout)
		return
	}
	http.Error(w, "error: "+err.Error(), http.StatusBadGateway)
}
