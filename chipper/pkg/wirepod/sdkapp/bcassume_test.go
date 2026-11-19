package sdkapp

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
)

// fakeControlStream stands in for the robot's BehaviorControl stream.
type fakeControlStream struct {
	ctx   context.Context
	mu    sync.Mutex
	sent  []*vectorpb.BehaviorControlRequest
	resps chan *vectorpb.BehaviorControlResponse
	fail  chan error
}

func newFakeControlStream(ctx context.Context) *fakeControlStream {
	return &fakeControlStream{ctx: ctx, resps: make(chan *vectorpb.BehaviorControlResponse, 8), fail: make(chan error, 1)}
}

func (f *fakeControlStream) Send(r *vectorpb.BehaviorControlRequest) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, r)
	return nil
}

func (f *fakeControlStream) Recv() (*vectorpb.BehaviorControlResponse, error) {
	select {
	case r := <-f.resps:
		return r, nil
	case err := <-f.fail:
		return nil, err
	case <-f.ctx.Done():
		return nil, f.ctx.Err()
	}
}

// log returns what was sent: "req:<priority>" or "release".
func (f *fakeControlStream) log() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []string
	for _, r := range f.sent {
		if cr := r.GetControlRequest(); cr != nil {
			out = append(out, "req:"+cr.Priority.String())
		} else if r.GetControlRelease() != nil {
			out = append(out, "release")
		}
	}
	return out
}

func lostEvent() *vectorpb.BehaviorControlResponse {
	return &vectorpb.BehaviorControlResponse{
		ResponseType: &vectorpb.BehaviorControlResponse_ControlLostEvent{ControlLostEvent: &vectorpb.ControlLostResponse{}},
	}
}

func grantedEvent() *vectorpb.BehaviorControlResponse {
	return &vectorpb.BehaviorControlResponse{
		ResponseType: &vectorpb.BehaviorControlResponse_ControlGrantedResponse{ControlGrantedResponse: &vectorpb.ControlGrantedResponse{}},
	}
}

func bcWaitUntil(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

type holdRun struct {
	stream  *fakeControlStream
	release context.CancelFunc
	done    chan error
}

// startHold runs runControlHold against a fake stream. The stream ends with
// the test, the hold when release is called.
func startHold(t *testing.T, prio vectorpb.ControlRequest_Priority, regain time.Duration) *holdRun {
	t.Helper()
	streamCtx, endStream := context.WithCancel(context.Background())
	releaseCtx, release := context.WithCancel(context.Background())
	h := &holdRun{stream: newFakeControlStream(streamCtx), release: release, done: make(chan error, 1)}
	go func() { h.done <- runControlHold(releaseCtx, h.stream, prio, regain) }()
	t.Cleanup(func() {
		release()
		endStream()
		select {
		case <-h.done:
		case <-time.After(2 * time.Second):
			t.Errorf("hold did not end")
		}
	})
	return h
}

func (h *holdRun) result(t *testing.T) error {
	t.Helper()
	select {
	case err := <-h.done:
		h.done <- err
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("hold did not end")
		return nil
	}
}

func TestPriorityIsDefaultUnlessOverrideNamed(t *testing.T) {
	cases := map[string]vectorpb.ControlRequest_Priority{
		"":                     vectorpb.ControlRequest_DEFAULT,
		"default":              vectorpb.ControlRequest_DEFAULT,
		"high":                 vectorpb.ControlRequest_DEFAULT,
		"override":             vectorpb.ControlRequest_DEFAULT,
		"override_behaviors":   vectorpb.ControlRequest_OVERRIDE_BEHAVIORS,
		" OVERRIDE_BEHAVIORS ": vectorpb.ControlRequest_OVERRIDE_BEHAVIORS,
	}
	for in, want := range cases {
		if got := bcPriority(in); got != want {
			t.Errorf("bcPriority(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestHoldRequestsControlAtGivenPriority(t *testing.T) {
	h := startHold(t, vectorpb.ControlRequest_DEFAULT, time.Hour)
	bcWaitUntil(t, "first request", func() bool { return len(h.stream.log()) == 1 })
	if got := h.stream.log()[0]; got != "req:DEFAULT" {
		t.Fatalf("first message %q, want req:DEFAULT", got)
	}
}

func TestHoldAsksAgainAfterControlIsLost(t *testing.T) {
	h := startHold(t, vectorpb.ControlRequest_DEFAULT, 30*time.Millisecond)
	h.stream.resps <- grantedEvent()
	h.stream.resps <- lostEvent()
	bcWaitUntil(t, "re-request", func() bool { return len(h.stream.log()) == 2 })
	if got := h.stream.log(); got[1] != "req:DEFAULT" {
		t.Fatalf("after losing control sent %v, want a second req:DEFAULT", got)
	}
}

func TestLostEventsWhileWaitingCauseOneRequest(t *testing.T) {
	h := startHold(t, vectorpb.ControlRequest_DEFAULT, 60*time.Millisecond)
	for i := 0; i < 4; i++ {
		h.stream.resps <- lostEvent()
	}
	bcWaitUntil(t, "re-request", func() bool { return len(h.stream.log()) >= 2 })
	time.Sleep(150 * time.Millisecond)
	if n := len(h.stream.log()); n != 2 {
		t.Fatalf("four lost events in one wait produced %d messages, want 2: %v", n, h.stream.log())
	}
}

func TestRepeatedLossStillGetsControlBack(t *testing.T) {
	h := startHold(t, vectorpb.ControlRequest_DEFAULT, 60*time.Millisecond)
	stop := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(stop) {
		h.stream.resps <- lostEvent()
		time.Sleep(20 * time.Millisecond)
	}
	n := len(h.stream.log()) - 1
	if n < 1 {
		t.Fatalf("control lost every 20 ms for 300 ms and was never asked for again: %v", h.stream.log())
	}
	if n > 6 {
		t.Fatalf("asked for control %d times in 300 ms with a 60 ms delay: %v", n, h.stream.log())
	}
}

func TestNoRequestAfterRelease(t *testing.T) {
	h := startHold(t, vectorpb.ControlRequest_DEFAULT, 80*time.Millisecond)
	bcWaitUntil(t, "first request", func() bool { return len(h.stream.log()) == 1 })
	h.stream.resps <- lostEvent()
	time.Sleep(20 * time.Millisecond)
	h.release()
	h.result(t)
	time.Sleep(150 * time.Millisecond)
	for _, m := range h.stream.log()[1:] {
		if strings.HasPrefix(m, "req:") {
			t.Fatalf("asked for control again after release: %v", h.stream.log())
		}
	}
}

func TestReleaseEndsTheHold(t *testing.T) {
	h := startHold(t, vectorpb.ControlRequest_DEFAULT, time.Hour)
	bcWaitUntil(t, "first request", func() bool { return len(h.stream.log()) == 1 })
	h.release()
	if err := h.result(t); err != nil {
		t.Fatalf("release returned %v", err)
	}
	log := h.stream.log()
	if log[len(log)-1] != "release" {
		t.Fatalf("last message %q, want release: %v", log[len(log)-1], log)
	}
}

func TestStreamErrorEndsTheHold(t *testing.T) {
	h := startHold(t, vectorpb.ControlRequest_DEFAULT, time.Hour)
	h.stream.fail <- errors.New("robot gone")
	if err := h.result(t); err == nil || !strings.Contains(err.Error(), "robot gone") {
		t.Fatalf("hold ended with %v, want the stream error", err)
	}
}
