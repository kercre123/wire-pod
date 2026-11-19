package sdkapp

import (
	"context"
	"strings"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// Behavior control for the SDK web app.
//
// Control is requested at DEFAULT, which runs below the robot's mandatory
// physical reactions: a cliff, a fall or being picked up still interrupts
// whatever the web app is doing. OVERRIDE_BEHAVIORS suppresses those
// reactions, so a robot driven by hand drives off a table; it is used only
// when the caller asks for it by name.
//
// When the robot takes control away (to back off a cliff, for example),
// control is requested again after bcRegainDelay while the hold is still
// wanted, so the controls keep working once the reaction is over.

const bcPriorityOverride = "override_behaviors"

var bcRegainDelay = time.Second

// bcPriority maps the web app's priority parameter to a control priority.
// Anything other than the explicit override name, including "high", is
// DEFAULT.
func bcPriority(p string) vectorpb.ControlRequest_Priority {
	if strings.EqualFold(strings.TrimSpace(p), bcPriorityOverride) {
		return vectorpb.ControlRequest_OVERRIDE_BEHAVIORS
	}
	return vectorpb.ControlRequest_DEFAULT
}

func assumeBehaviorControl(robot *Robot, priority string) {
	prio := bcPriority(priority)
	release, done, _ := robot.bc.start(robot.Ctx)
	go holdBehaviorControl(robot, prio, release, done)
}

func releaseBehaviorControl(robot *Robot) {
	robot.bc.stop()
}

func holdBehaviorControl(robot *Robot, prio vectorpb.ControlRequest_Priority, release context.Context, done func()) {
	defer done()
	streamCtx, cancel := context.WithCancel(robot.Ctx)
	defer cancel()
	stream, err := robot.Vector.Conn.BehaviorControl(streamCtx)
	if err != nil {
		logger.Println("Behavior control error: ", err)
		return
	}
	if err := runControlHold(release, stream, prio, bcRegainDelay); err != nil && streamCtx.Err() == nil {
		logger.Println("Behavior control error: ", err)
	}
}

// bcControlStream is the part of the BehaviorControl stream a hold uses.
type bcControlStream interface {
	Send(*vectorpb.BehaviorControlRequest) error
	Recv() (*vectorpb.BehaviorControlResponse, error)
}

func controlRequest(prio vectorpb.ControlRequest_Priority) *vectorpb.BehaviorControlRequest {
	return &vectorpb.BehaviorControlRequest{
		RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
			ControlRequest: &vectorpb.ControlRequest{Priority: prio},
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

// runControlHold requests control, requests it again regainDelay after each
// loss, and sends a release once release is done. All sends happen on the
// calling goroutine. It returns when control is released or the stream ends.
func runControlHold(release context.Context, stream bcControlStream, prio vectorpb.ControlRequest_Priority, regainDelay time.Duration) error {
	if err := stream.Send(controlRequest(prio)); err != nil {
		return err
	}
	lost := make(chan struct{}, 1)
	recvErr := make(chan error, 1)
	go func() {
		for {
			resp, err := stream.Recv()
			if err != nil {
				recvErr <- err
				return
			}
			if resp.GetControlLostEvent() != nil {
				select {
				case lost <- struct{}{}:
				default: // a re-request is already pending
				}
			}
		}
	}()
	var regain <-chan time.Time
	for {
		select {
		case <-release.Done():
			return stream.Send(controlRelease())
		case err := <-recvErr:
			return err
		case <-lost:
			if regain == nil {
				regain = time.After(regainDelay)
			}
		case <-regain:
			regain = nil
			if release.Err() != nil {
				continue // the release case runs next
			}
			if err := stream.Send(controlRequest(prio)); err != nil {
				return err
			}
		}
	}
}
