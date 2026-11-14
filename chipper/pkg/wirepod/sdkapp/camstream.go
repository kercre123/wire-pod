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
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
)

// The robot's gateway turns image streaming on when a CameraFeed subscriber
// connects and off when ANY subscriber leaves; the flag is robot-wide. A
// viewer that opens and closes its own feed therefore cuts off every other
// camera user, including SDK programs. So /cam-stream shares one feed per
// robot among all viewers, keeps it open for camLinger after the last viewer
// leaves, never turns streaming off itself, and turns it back on when someone
// else's exit has switched it off.
var (
	camControlTimeout = 3 * time.Second  // bound on each EnableImageStreaming call
	camIdleTimeout    = 6 * time.Second  // a viewer that gets no frame for this long ends
	camReenableAfter  = 2 * time.Second  // silence before streaming is switched back on
	camLinger         = 60 * time.Second // feed kept open after the last viewer leaves
	camJPEGQuality    = 50
	camBoundary       = "frame"
)

const camSubscriberBuffer = 2

// sdkCameraClient is the subset of the SDK the camera feed uses, so the feed
// can be tested without a robot.
type sdkCameraClient interface {
	EnableImageStreaming(ctx context.Context, enable bool) error
	OpenCameraFeed(ctx context.Context) (func() ([]byte, error), error)
}

type liveSDKCameraClient struct {
	robotObj Robot
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

// camFeed is the one shared CameraFeed of a robot.
type camFeed struct {
	esn       string
	client    sdkCameraClient
	cancel    context.CancelFunc
	mu        sync.Mutex
	subs      map[uint64]chan []byte
	nextID    uint64
	lastFrame time.Time
	linger    *time.Timer
	closed    bool
	// timings captured when the feed starts, so its goroutines never read
	// the package variables again
	reenableAfter time.Duration
	lingerFor     time.Duration
}

var (
	camFeedsMu sync.Mutex
	camFeeds   = map[string]*camFeed{}
)

func normalizeESN(esn string) string {
	return strings.TrimSpace(strings.ToLower(esn))
}

// subscribeCamFeed attaches a viewer to the robot's shared feed, starting it
// if needed. The returned channel is closed when the feed ends.
func subscribeCamFeed(esn string, client sdkCameraClient) (*camFeed, uint64, chan []byte) {
	camFeedsMu.Lock()
	defer camFeedsMu.Unlock()
	feed, ok := camFeeds[esn]
	if !ok {
		ctx, cancel := context.WithCancel(context.Background())
		feed = &camFeed{esn: esn, client: client, cancel: cancel, subs: map[uint64]chan []byte{},
			lastFrame: time.Now(), reenableAfter: camReenableAfter, lingerFor: camLinger}
		camFeeds[esn] = feed
		go feed.run(ctx)
	}
	feed.mu.Lock()
	defer feed.mu.Unlock()
	if feed.linger != nil {
		feed.linger.Stop()
		feed.linger = nil
	}
	feed.nextID++
	ch := make(chan []byte, camSubscriberBuffer)
	feed.subs[feed.nextID] = ch
	return feed, feed.nextID, ch
}

// unsubscribe detaches a viewer; the feed lingers before it closes.
func (f *camFeed) unsubscribe(id uint64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.subs[id]; !ok {
		return
	}
	delete(f.subs, id)
	if len(f.subs) == 0 && !f.closed {
		f.linger = time.AfterFunc(f.lingerFor, f.stopIfIdle)
	}
}

func (f *camFeed) stopIfIdle() {
	f.mu.Lock()
	idle := len(f.subs) == 0
	f.mu.Unlock()
	if idle {
		f.stop()
	}
}

// stop ends the feed and every viewer on it.
func (f *camFeed) stop() {
	camFeedsMu.Lock()
	if camFeeds[f.esn] == f {
		delete(camFeeds, f.esn)
	}
	camFeedsMu.Unlock()
	f.cancel()
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return
	}
	f.closed = true
	for id, ch := range f.subs {
		close(ch)
		delete(f.subs, id)
	}
}

func (f *camFeed) publish(jpg []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastFrame = time.Now()
	for _, ch := range f.subs {
		select {
		case ch <- jpg:
		default: // a slow viewer drops a frame rather than stalling the others
		}
	}
}

func (f *camFeed) silentFor() time.Duration {
	f.mu.Lock()
	defer f.mu.Unlock()
	return time.Since(f.lastFrame)
}

// run reads the robot's feed until it ends or is stopped. It never turns
// image streaming off; the gateway does that itself when this feed closes.
func (f *camFeed) run(ctx context.Context) {
	defer f.stop()
	recv, err := f.client.OpenCameraFeed(ctx)
	if err != nil {
		return
	}
	go f.keepStreaming(ctx)
	for ctx.Err() == nil {
		data, recvErr := recv()
		if recvErr != nil {
			return
		}
		if jpg, ok := reencodeJPEG(data); ok {
			f.publish(jpg)
		}
	}
}

// keepStreaming switches streaming back on when the feed has gone silent,
// which is what another subscriber leaving does. It does nothing while
// frames flow, and backs off (doubling up to camLinger) while the robot
// stays silent, e.g. asleep.
func (f *camFeed) keepStreaming(ctx context.Context) {
	ticker := time.NewTicker(f.reenableAfter)
	defer ticker.Stop()
	next := f.reenableAfter
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			silent := f.silentFor()
			if silent < f.reenableAfter {
				next = f.reenableAfter
				continue
			}
			if silent >= next {
				_ = f.client.EnableImageStreaming(ctx, true) // bounded by camControlTimeout
				next = minDuration(2*next, f.lingerFor)
				if next <= silent {
					next = silent + next
				}
			}
		}
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func reencodeJPEG(data []byte) ([]byte, bool) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil || img == nil {
		return nil, false
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: camJPEGQuality}); err != nil {
		return nil, false
	}
	return buf.Bytes(), true
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
	streamCamera(w, r, esn, liveSDKCameraClient{robotObj: robotObj})
}

// streamCamera serves one viewer from the shared feed. It ends when the
// client leaves, when no frame arrives for camIdleTimeout, or when the feed
// is stopped.
func streamCamera(w http.ResponseWriter, r *http.Request, esn string, client sdkCameraClient) {
	esn = normalizeESN(esn)
	if esn == "" {
		http.Error(w, "error: missing serial", http.StatusBadRequest)
		return
	}
	idleTimeout := camIdleTimeout
	feed, id, frames := subscribeCamFeed(esn, client)
	defer feed.unsubscribe(id)
	idle := time.NewTimer(idleTimeout)
	defer idle.Stop()
	flusher, canFlush := w.(http.Flusher)
	sent := 0
	for {
		select {
		case <-r.Context().Done():
			return
		case <-idle.C:
			if sent == 0 {
				http.Error(w, "error: camera_frame_timeout", http.StatusGatewayTimeout)
			}
			return
		case jpg, ok := <-frames:
			if !ok {
				if sent == 0 {
					http.Error(w, "error: camera feed ended", http.StatusBadGateway)
				}
				return
			}
			if !writeCamPart(w, jpg, sent == 0) {
				return
			}
			sent++
			if canFlush {
				flusher.Flush()
			}
			idle.Reset(idleTimeout)
		}
	}
}

func writeCamPart(w http.ResponseWriter, jpg []byte, first bool) bool {
	if first {
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary="+camBoundary)
	}
	head := "--" + camBoundary + "\r\nContent-Type: image/jpeg\r\nContent-Length: " +
		strconv.Itoa(len(jpg)) + "\r\n\r\n"
	if _, err := w.Write([]byte(head)); err != nil {
		return false
	}
	_, err := w.Write(append(append([]byte(nil), jpg...), '\r', '\n'))
	return err == nil
}
