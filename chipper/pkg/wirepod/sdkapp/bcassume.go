package sdkapp

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
)

// Behavior control for the SDK web app.
//
// Control is requested at DEFAULT priority, which runs below the robot's
// mandatory physical reactions: a cliff, a fall or being picked up still
// interrupts whatever the web app is doing. OVERRIDE_BEHAVIORS suppresses
// those reactions, so a robot driven by hand will drive off a table; it is
// used only when the caller asks for it by name.
//
// While control is held, it is requested again whenever the robot takes it
// away (for example to back away from a cliff), so the controls keep working
// once the reaction is over. Releasing control in the app ends the hold.
var (
	bcRegainDelay = 1 * time.Second        // wait after control is lost before asking again
	bcPollEvery   = 250 * time.Millisecond // how often the hold checks it is still wanted
)

const bcPriorityOverride = "override_behaviors"

// bcPriority maps the web app's priority parameter to a control priority.
// Anything other than the explicit override name, including the old "high",
// gets DEFAULT.
func bcPriority(p string) vectorpb.ControlRequest_Priority {
	if strings.EqualFold(strings.TrimSpace(p), bcPriorityOverride) {
		return vectorpb.ControlRequest_OVERRIDE_BEHAVIORS
	}
	return vectorpb.ControlRequest_DEFAULT
}

// bcControlStream is the part of the BehaviorControl stream the hold uses.
type bcControlStream interface {
	Send(*vectorpb.BehaviorControlRequest) error
	Recv() (*vectorpb.BehaviorControlResponse, error)
}

var (
	bcHoldsMu sync.Mutex
	bcHolds   = map[string]uint64{}
	bcHoldSeq uint64
)

// claimBcHold makes a new hold the only one for esn; an older hold for the
// same robot sees it is no longer current and releases.
func claimBcHold(esn string) uint64 {
	bcHoldsMu.Lock()
	defer bcHoldsMu.Unlock()
	bcHoldSeq++
	bcHolds[esn] = bcHoldSeq
	return bcHoldSeq
}

func bcHoldCurrent(esn string, id uint64) bool {
	bcHoldsMu.Lock()
	defer bcHoldsMu.Unlock()
	return bcHolds[esn] == id
}

func dropBcHold(esn string, id uint64) {
	bcHoldsMu.Lock()
	defer bcHoldsMu.Unlock()
	if bcHolds[esn] == id {
		delete(bcHolds, esn)
	}
}

// bcAssumed reports whether the app still wants control of esn. The robot
// list can change while a hold runs, so it is looked up by ESN each time.
func bcAssumed(esn string) bool {
	for i := range robots {
		if strings.EqualFold(robots[i].ESN, esn) {
			return robots[i].BcAssumption
		}
	}
	return false
}

func setBcAssumed(esn string, v bool) {
	for i := range robots {
		if strings.EqualFold(robots[i].ESN, esn) {
			robots[i].BcAssumption = v
		}
	}
}

func controlRequest(p vectorpb.ControlRequest_Priority) *vectorpb.BehaviorControlRequest {
	return &vectorpb.BehaviorControlRequest{
		RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
			ControlRequest: &vectorpb.ControlRequest{Priority: p},
		},
	}
}

func controlRelease() *vectorpb.BehaviorControlRequest {
	return &vectorpb.BehaviorControlRequest{
		RequestType: &vectorpb.BehaviorControlRequest_ControlRelease{
			ControlRelease: &vectorpb.ControlRelease{},
		},
	}
}

func assumeBehaviorControl(robot Robot, robotIndex int, priority string) {
	_ = robotIndex // the robot list can change under a hold; it is looked up by ESN
	esn := robot.ESN
	if robot.Vector == nil {
		log.Println("behavior control: no connection to " + esn)
		return
	}
	id := claimBcHold(esn)
	setBcAssumed(esn, true)
	prio := bcPriority(priority)
	go func() {
		defer dropBcHold(esn, id)
		parent := robot.Ctx
		if parent == nil {
			parent = context.Background()
		}
		ctx, cancel := context.WithCancel(parent)
		defer cancel()
		stream, err := robot.Vector.Conn.BehaviorControl(ctx)
		if err != nil {
			log.Println("behavior control: " + err.Error())
			return
		}
		wanted := func() bool { return bcHoldCurrent(esn, id) && bcAssumed(esn) }
		reason := holdBehaviorControl(ctx, stream, prio, wanted, bcRegainDelay, bcPollEvery)
		log.Println("behavior control for " + esn + " ended: " + reason)
	}()
}

// holdBehaviorControl requests control, asks again each time the robot
// takes it away, and releases it once wanted() turns false. It returns why
// the hold ended.
func holdBehaviorControl(ctx context.Context, stream bcControlStream, prio vectorpb.ControlRequest_Priority,
	wanted func() bool, regainDelay, pollEvery time.Duration) string {
	if err := stream.Send(controlRequest(prio)); err != nil {
		return "request failed: " + err.Error()
	}
	responses := make(chan *vectorpb.BehaviorControlResponse, 1)
	recvErr := make(chan error, 1)
	go func() {
		for {
			resp, err := stream.Recv()
			if err != nil {
				recvErr <- err
				return
			}
			select {
			case responses <- resp:
			case <-ctx.Done():
				return
			}
		}
	}()
	poll := time.NewTicker(pollEvery)
	defer poll.Stop()
	var regain <-chan time.Time
	for {
		select {
		case <-ctx.Done():
			return "connection closed"
		case err := <-recvErr:
			return "stream ended: " + err.Error()
		case <-poll.C:
			if !wanted() {
				if err := stream.Send(controlRelease()); err != nil {
					return "released (send failed: " + err.Error() + ")"
				}
				return "released"
			}
		case resp := <-responses:
			if resp.GetControlLostEvent() != nil && regain == nil {
				regain = time.After(regainDelay)
			}
		case <-regain:
			regain = nil
			if !wanted() {
				continue
			}
			if err := stream.Send(controlRequest(prio)); err != nil {
				return "re-request failed: " + err.Error()
			}
		}
	}
}
