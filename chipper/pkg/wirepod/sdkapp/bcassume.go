package sdkapp

import (
	"context"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

func assumeBehaviorControl(robot *Robot, priority string) {
	prio := vectorpb.ControlRequest_DEFAULT
	if priority == "high" {
		prio = vectorpb.ControlRequest_OVERRIDE_BEHAVIORS
	}
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
	err = stream.Send(&vectorpb.BehaviorControlRequest{
		RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
			ControlRequest: &vectorpb.ControlRequest{
				Priority: prio,
			},
		},
	})
	if err != nil {
		logger.Println("Behavior control error: ", err)
		return
	}
	go func() {
		for {
			if _, err := stream.Recv(); err != nil {
				return
			}
		}
	}()
	<-release.Done()
	err = stream.Send(&vectorpb.BehaviorControlRequest{
		RequestType: &vectorpb.BehaviorControlRequest_ControlRelease{
			ControlRelease: &vectorpb.ControlRelease{},
		},
	})
	if err != nil && streamCtx.Err() == nil {
		logger.Println("Behavior control release error: ", err)
	}
}
