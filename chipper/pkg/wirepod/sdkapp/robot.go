package sdkapp

import (
	"context"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/botconn"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

const robotIdleTimeout = 5 * time.Minute

type Robot struct {
	ESN    string
	GUID   string
	Target string
	Vector *vector.Vector
	Ctx    context.Context

	cancel    context.CancelFunc
	lastUsed  atomic.Int64
	stimState atomic.Uint32

	bc     robotTask
	cam    robotTask
	events robotTask
}

type robotTask struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	gen    int
}

func (t *robotTask) start(parent context.Context) (context.Context, func(), bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	replaced := t.cancel != nil
	if replaced {
		t.cancel()
	}
	ctx, cancel := context.WithCancel(parent)
	t.gen++
	gen := t.gen
	t.cancel = cancel
	done := func() {
		cancel()
		t.mu.Lock()
		if t.gen == gen {
			t.cancel = nil
		}
		t.mu.Unlock()
	}
	return ctx, done, replaced
}

func (t *robotTask) stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}
}

func (t *robotTask) running() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.cancel != nil
}

func (r *Robot) touch() {
	r.lastUsed.Store(time.Now().UnixNano())
}

func (r *Robot) StimState() float32 {
	return math.Float32frombits(r.stimState.Load())
}

func (r *Robot) setStimState(v float32) {
	r.stimState.Store(math.Float32bits(v))
}

func (r *Robot) close() {
	r.bc.stop()
	r.cam.stop()
	r.events.stop()
	r.cancel()
}

func (r *Robot) watchIdle() {
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-r.Ctx.Done():
			return
		case <-t.C:
			if time.Since(time.Unix(0, r.lastUsed.Load())) >= robotIdleTimeout {
				logger.Println("Closing SDK connection for " + r.ESN + ", source: idle timer")
				removeRobot(r)
				return
			}
		}
	}
}

var (
	robotsMu      sync.Mutex
	robots        = map[string]*Robot{}
	robotsPending = map[string]chan struct{}{}
)

func newRobot(esn string) (*Robot, error) {
	info, err := botconn.Lookup(esn)
	if err != nil {
		return nil, err
	}
	logger.Println("Connecting to " + esn + " with GUID " + info.GUID)
	vec, err := botconn.Get(esn)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	checkCtx, checkCancel := context.WithTimeout(ctx, 10*time.Second)
	defer checkCancel()
	if _, err := vec.Conn.BatteryState(checkCtx, &vectorpb.BatteryStateRequest{}); err != nil {
		cancel()
		return nil, err
	}
	r := &Robot{
		ESN:    esn,
		GUID:   info.GUID,
		Target: info.Target,
		Vector: vec,
		Ctx:    ctx,
		cancel: cancel,
	}
	r.touch()
	return r, nil
}

func getRobot(serial string) (*Robot, error) {
	esn := strings.ToLower(strings.TrimSpace(serial))
	for {
		robotsMu.Lock()
		if r, ok := robots[esn]; ok {
			robotsMu.Unlock()
			r.touch()
			return r, nil
		}
		if wait, ok := robotsPending[esn]; ok {
			robotsMu.Unlock()
			<-wait
			continue
		}
		done := make(chan struct{})
		robotsPending[esn] = done
		robotsMu.Unlock()

		r, err := newRobot(esn)

		robotsMu.Lock()
		delete(robotsPending, esn)
		if err == nil {
			robots[esn] = r
		}
		robotsMu.Unlock()
		close(done)

		if err != nil {
			return nil, err
		}
		go r.watchIdle()
		return r, nil
	}
}

func removeRobot(r *Robot) {
	robotsMu.Lock()
	if robots[r.ESN] == r {
		delete(robots, r.ESN)
	}
	robotsMu.Unlock()
	r.close()
}
