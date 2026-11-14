package sdkapp

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// fakeCamera behaves like the robot gateway: frames flow while streaming is
// on; in "stall" mode it goes silent after two frames (as when another
// subscriber leaves and the gateway turns streaming off) until streaming is
// switched back on.
type fakeCamera struct {
	mu      sync.Mutex
	mode    string // "silent", "stream", "stall", "finite"
	frame   []byte
	finite  [][]byte
	sent    int
	stalled bool
	opens   int
	enables []bool
	closed  bool
}

func (f *fakeCamera) EnableImageStreaming(ctx context.Context, enable bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.enables = append(f.enables, enable)
	if enable {
		f.stalled = false
	}
	return nil
}

func (f *fakeCamera) OpenCameraFeed(ctx context.Context) (func() ([]byte, error), error) {
	f.mu.Lock()
	f.opens++
	f.mu.Unlock()
	go func() { // like a gRPC stream, the feed closes when its context ends
		<-ctx.Done()
		f.markClosed()
	}()
	return func() ([]byte, error) {
		for i := 0; i < 2000; i++ {
			if ctx.Err() != nil {
				f.markClosed()
				return nil, ctx.Err()
			}
			if frame, err, ok := f.next(ctx); ok {
				return frame, err
			}
			select {
			case <-ctx.Done():
				f.markClosed()
				return nil, ctx.Err()
			case <-time.After(5 * time.Millisecond):
			}
		}
		return nil, io.EOF
	}, nil
}

func (f *fakeCamera) next(ctx context.Context) ([]byte, error, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	switch f.mode {
	case "stream":
		f.sent++
		return f.frame, nil, true
	case "stall":
		if f.sent == 2 && !f.stalled && len(f.enables) == 0 {
			f.stalled = true
		}
		if f.stalled {
			return nil, nil, false
		}
		f.sent++
		return f.frame, nil, true
	case "finite":
		if len(f.finite) == 0 {
			return nil, io.EOF, true
		}
		fr := f.finite[0]
		f.finite = f.finite[1:]
		return fr, nil, true
	}
	return nil, nil, false
}

func (f *fakeCamera) markClosed() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
}

func (f *fakeCamera) snapshot() (opens int, enables []bool, closed bool, sent int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.opens, append([]bool(nil), f.enables...), f.closed, f.sent
}

func tinyJPEG(t *testing.T, shade uint8) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 16, 12))
	for y := 0; y < 12; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.RGBA{R: shade, G: uint8(y * 15), B: 90, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("jpeg encode: %v", err)
	}
	return buf.Bytes()
}

func fastTimings(t *testing.T, idle, reenable, linger time.Duration) {
	t.Helper()
	oldIdle, oldRe, oldLinger := camIdleTimeout, camReenableAfter, camLinger
	camIdleTimeout, camReenableAfter, camLinger = idle, reenable, linger
	t.Cleanup(func() {
		stopAllCamFeeds()
		camIdleTimeout, camReenableAfter, camLinger = oldIdle, oldRe, oldLinger
	})
}

// stopAllCamFeeds ends every feed and waits for its goroutines to see it.
func stopAllCamFeeds() {
	camFeedsMu.Lock()
	feeds := make([]*camFeed, 0, len(camFeeds))
	for _, f := range camFeeds {
		feeds = append(feeds, f)
	}
	camFeedsMu.Unlock()
	for _, f := range feeds {
		f.stop()
	}
	time.Sleep(20 * time.Millisecond)
}

// safeRecorder is an http.ResponseWriter the test can read while the
// handler is still writing.
type safeRecorder struct {
	mu     sync.Mutex
	header http.Header
	code   int
	body   bytes.Buffer
}

func newSafeRecorder() *safeRecorder {
	return &safeRecorder{header: http.Header{}, code: http.StatusOK}
}

func (r *safeRecorder) Header() http.Header { return r.header }
func (r *safeRecorder) WriteHeader(code int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.code = code
}
func (r *safeRecorder) Write(b []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.body.Write(b)
}
func (r *safeRecorder) Flush() {}
func (r *safeRecorder) Bytes() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]byte(nil), r.body.Bytes()...)
}
func (r *safeRecorder) Code() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.code
}

// view runs one viewer until ctx ends or the handler returns.
func view(ctx context.Context, esn string, cam sdkCameraClient) (*safeRecorder, chan struct{}) {
	rec := newSafeRecorder()
	req := httptest.NewRequest(http.MethodGet, "/cam-stream?serial="+esn, nil).WithContext(ctx)
	done := make(chan struct{})
	go func() {
		streamCamera(rec, req, esn, cam)
		close(done)
	}()
	return rec, done
}

func waitFor(t *testing.T, what string, limit time.Duration, cond func() bool) {
	t.Helper()
	for i := 0; i < int(limit/(5*time.Millisecond)); i++ {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !cond() {
		t.Fatalf("timed out waiting for %s", what)
	}
}

func assertNeverDisabled(t *testing.T, cam *fakeCamera) {
	t.Helper()
	_, enables, _, _ := cam.snapshot()
	for _, e := range enables {
		if !e {
			t.Fatalf("image streaming was turned off: %v", enables)
		}
	}
}

func TestCamStreamMissingSerial(t *testing.T) {
	rec := httptest.NewRecorder()
	streamCamera(rec, httptest.NewRequest(http.MethodGet, "/cam-stream", nil), " ", &fakeCamera{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestSilentRobotTimesOutAndFeedCloses(t *testing.T) {
	fastTimings(t, 60*time.Millisecond, 20*time.Millisecond, 80*time.Millisecond)
	cam := &fakeCamera{mode: "silent"}
	rec, done := view(context.Background(), "e0000001", cam)
	waitFor(t, "viewer to give up", time.Second, func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	})
	if rec.Code() != http.StatusGatewayTimeout {
		t.Fatalf("want 504, got %d", rec.Code())
	}
	waitFor(t, "feed to close after linger", time.Second, func() bool { _, _, closed, _ := cam.snapshot(); return closed })
	assertNeverDisabled(t, cam)
}

func TestViewersShareOneFeed(t *testing.T) {
	fastTimings(t, time.Second, time.Second, 50*time.Millisecond)
	cam := &fakeCamera{mode: "stream", frame: tinyJPEG(t, 10)}
	ctx, cancel := context.WithCancel(context.Background())
	recA, doneA := view(ctx, "e0000002", cam)
	recB, doneB := view(ctx, "e0000002", cam)
	waitFor(t, "frames to both viewers", time.Second, func() bool {
		_, _, _, sent := cam.snapshot()
		return sent > 6
	})
	cancel()
	<-doneA
	<-doneB
	if opens, _, _, _ := cam.snapshot(); opens != 1 {
		t.Fatalf("want one shared feed, robot saw %d", opens)
	}
	for name, rec := range map[string]*safeRecorder{"A": recA, "B": recB} {
		if !bytes.Contains(rec.Bytes(), []byte{0xff, 0xd8}) {
			t.Fatalf("viewer %s got no frame", name)
		}
	}
	assertNeverDisabled(t, cam)
}

func TestLeavingViewerDoesNotStopOthers(t *testing.T) {
	fastTimings(t, time.Second, time.Second, 50*time.Millisecond)
	cam := &fakeCamera{mode: "stream", frame: tinyJPEG(t, 20)}
	ctxA, cancelA := context.WithCancel(context.Background())
	ctxB, cancelB := context.WithCancel(context.Background())
	_, doneA := view(ctxA, "e0000003", cam)
	recB, doneB := view(ctxB, "e0000003", cam)
	waitFor(t, "first frames", time.Second, func() bool { _, _, _, s := cam.snapshot(); return s > 3 })
	cancelA()
	<-doneA
	before := len(recB.Bytes())
	waitFor(t, "B to keep receiving", time.Second, func() bool { return len(recB.Bytes()) > before+100 })
	cancelB()
	<-doneB
	if opens, _, closed, _ := cam.snapshot(); opens != 1 {
		t.Fatalf("want one feed, got %d (closed=%v)", opens, closed)
	}
	assertNeverDisabled(t, cam)
}

func TestFeedLingersForTheNextViewer(t *testing.T) {
	fastTimings(t, time.Second, time.Second, 300*time.Millisecond)
	cam := &fakeCamera{mode: "stream", frame: tinyJPEG(t, 30)}
	ctx1, cancel1 := context.WithCancel(context.Background())
	_, done1 := view(ctx1, "e0000004", cam)
	waitFor(t, "first frames", time.Second, func() bool { _, _, _, s := cam.snapshot(); return s > 2 })
	cancel1()
	<-done1
	time.Sleep(50 * time.Millisecond)
	ctx2, cancel2 := context.WithCancel(context.Background())
	rec2, done2 := view(ctx2, "e0000004", cam)
	waitFor(t, "second viewer frames", time.Second, func() bool { return bytes.Contains(rec2.Bytes(), []byte{0xff, 0xd8}) })
	cancel2()
	<-done2
	if opens, _, _, _ := cam.snapshot(); opens != 1 {
		t.Fatalf("a viewer within the linger should reuse the feed; robot saw %d opens", opens)
	}
	waitFor(t, "feed to close after linger", 2*time.Second, func() bool { _, _, closed, _ := cam.snapshot(); return closed })
	assertNeverDisabled(t, cam)
}

func TestSilenceIsHealedByReenabling(t *testing.T) {
	fastTimings(t, time.Second, 30*time.Millisecond, 50*time.Millisecond)
	cam := &fakeCamera{mode: "stall", frame: tinyJPEG(t, 40)}
	ctx, cancel := context.WithCancel(context.Background())
	_, done := view(ctx, "e0000005", cam)
	waitFor(t, "frames after the stall", 2*time.Second, func() bool { _, _, _, s := cam.snapshot(); return s > 4 })
	cancel()
	<-done
	_, enables, _, _ := cam.snapshot()
	if len(enables) == 0 {
		t.Fatalf("streaming was never switched back on")
	}
	assertNeverDisabled(t, cam)
}

func TestBodyIsValidMultipart(t *testing.T) {
	fastTimings(t, time.Second, time.Second, 50*time.Millisecond)
	cam := &fakeCamera{mode: "finite", finite: [][]byte{tinyJPEG(t, 50), []byte("not a jpeg"), tinyJPEG(t, 60)}}
	rec, done := view(context.Background(), "e0000006", cam)
	<-done
	mediaType, params, err := mime.ParseMediaType(rec.Header().Get("Content-Type"))
	if err != nil || mediaType != "multipart/x-mixed-replace" {
		t.Fatalf("bad content type %q: %v", rec.Header().Get("Content-Type"), err)
	}
	reader := multipart.NewReader(bytes.NewReader(rec.Bytes()), params["boundary"])
	parts := 0
	for i := 0; i < 4; i++ {
		part, perr := reader.NextPart()
		if perr != nil {
			break
		}
		data, _ := io.ReadAll(part)
		if len(data) < 2 || data[0] != 0xff || data[1] != 0xd8 {
			t.Fatalf("part %d is not a JPEG", i)
		}
		parts++
	}
	if parts != 2 {
		t.Fatalf("want 2 parseable parts (undecodable frame skipped), got %d", parts)
	}
	assertNeverDisabled(t, cam)
}
