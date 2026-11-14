package sdkapp

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeCamera struct {
	mu          sync.Mutex
	enabled     bool
	checkErr    error
	enableCalls []bool
	frames      chan []byte // nil: the feed stays silent until cancelled
}

func (f *fakeCamera) IsImageStreamingEnabled(ctx context.Context) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.enabled, f.checkErr
}

func (f *fakeCamera) EnableImageStreaming(ctx context.Context, enable bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.enableCalls = append(f.enableCalls, enable)
	f.enabled = enable
	return nil
}

func (f *fakeCamera) OpenCameraFeed(ctx context.Context) (func() ([]byte, error), error) {
	return func() ([]byte, error) {
		if f.frames == nil {
			<-ctx.Done()
			return nil, ctx.Err()
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case frame, ok := <-f.frames:
			if !ok {
				return nil, io.EOF
			}
			return frame, nil
		}
	}, nil
}

func (f *fakeCamera) calls() []bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]bool(nil), f.enableCalls...)
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

func get() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/cam-stream?serial=00e20145", nil)
}

func equalCalls(got []bool, want ...bool) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestCamStreamMissingSerial(t *testing.T) {
	rec := httptest.NewRecorder()
	streamCamera(rec, get(), " ", &fakeCamera{}, 50*time.Millisecond)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestCamStreamSilentRobotTimesOutAndRestores(t *testing.T) {
	cam := &fakeCamera{}
	rec := httptest.NewRecorder()
	start := time.Now()
	streamCamera(rec, get(), "00e20145", cam, 50*time.Millisecond)
	if time.Since(start) > time.Second {
		t.Fatalf("handler did not give up on a silent feed")
	}
	if rec.Code != http.StatusGatewayTimeout || !strings.Contains(rec.Body.String(), "camera_frame_timeout") {
		t.Fatalf("want 504 camera_frame_timeout, got %d %q", rec.Code, rec.Body.String())
	}
	if c := cam.calls(); !equalCalls(c, true, false) {
		t.Fatalf("want enable then disable, got %v", c)
	}
}

func TestCamStreamIsValidMultipartAndLeavesExistingStreamingOn(t *testing.T) {
	frames := make(chan []byte, 2)
	frames <- tinyJPEG(t, 10)
	frames <- tinyJPEG(t, 200)
	close(frames)
	cam := &fakeCamera{enabled: true, frames: frames}
	rec := httptest.NewRecorder()
	streamCamera(rec, get(), "00e20145", cam, time.Second)

	mediaType, params, err := mime.ParseMediaType(rec.Header().Get("Content-Type"))
	if err != nil || mediaType != "multipart/x-mixed-replace" {
		t.Fatalf("bad content type %q: %v", rec.Header().Get("Content-Type"), err)
	}
	reader := multipart.NewReader(rec.Body, params["boundary"])
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
		t.Fatalf("want 2 parseable parts, got %d", parts)
	}
	if c := cam.calls(); len(c) != 0 {
		t.Fatalf("streaming was already on; want no toggles, got %v", c)
	}
}

func TestCamStreamUnknownStateIsNotTurnedOff(t *testing.T) {
	cam := &fakeCamera{checkErr: errors.New("deadline exceeded")}
	rec := httptest.NewRecorder()
	streamCamera(rec, get(), "00e20145", cam, 50*time.Millisecond)
	if c := cam.calls(); !equalCalls(c, true) {
		t.Fatalf("state unknown: want enable only, got %v", c)
	}
}

func TestCamStreamReplacementHandsOverOwnership(t *testing.T) {
	cam := &fakeCamera{}
	first := make(chan struct{})
	go func() {
		streamCamera(httptest.NewRecorder(), get(), "00e20145", cam, 2*time.Second)
		close(first)
	}()
	time.Sleep(50 * time.Millisecond)
	second := make(chan struct{})
	go func() {
		streamCamera(httptest.NewRecorder(), get(), "00e20145", cam, 150*time.Millisecond)
		close(second)
	}()
	select {
	case <-first:
	case <-time.After(time.Second):
		t.Fatalf("first viewer was not ended by the second")
	}
	if c := cam.calls(); !equalCalls(c, true) {
		t.Fatalf("replacement must not turn streaming off, got %v", c)
	}
	select {
	case <-second:
	case <-time.After(2 * time.Second):
		t.Fatalf("second viewer did not time out")
	}
	if c := cam.calls(); !equalCalls(c, true, false) {
		t.Fatalf("last viewer must turn streaming off once, got %v", c)
	}
}

func TestStopCamViewerEndsStream(t *testing.T) {
	cam := &fakeCamera{}
	done := make(chan struct{})
	go func() {
		streamCamera(httptest.NewRecorder(), get(), "00e20145", cam, 5*time.Second)
		close(done)
	}()
	time.Sleep(50 * time.Millisecond)
	stopCamViewer(" 00E20145 ")
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("stopCamViewer did not end the stream")
	}
	if c := cam.calls(); !equalCalls(c, true, false) {
		t.Fatalf("want enable then disable, got %v", c)
	}
}

func TestClientDisconnectEndsStream(t *testing.T) {
	cam := &fakeCamera{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		streamCamera(httptest.NewRecorder(), get().WithContext(ctx), "00e20145", cam, 5*time.Second)
		close(done)
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("handler outlived its client")
	}
	if c := cam.calls(); !equalCalls(c, true, false) {
		t.Fatalf("want enable then disable, got %v", c)
	}
}
