package scripting

import (
	"context"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	lua "github.com/yuin/gopher-lua"
)

func SetBControlFunctions(L *lua.LState) {
	start := make(chan bool)
	stop := make(chan bool)
	currentlyAssumed := false
	L.SetGlobal("assumeBehaviorControl", L.NewFunction(func(*lua.LState) int {
		robot := gRfLS(L)
		priority := vectorpb.ControlRequest_OVERRIDE_BEHAVIORS
		if priority != 0 && priority != 10 && priority != 20 && priority != 30 {
			logger.Println("LUA: Behavior control priority was not valid. Valid choices are 10, 20, and 30. Assuming 10.")
		} else {
			priority = vectorpb.ControlRequest_Priority(L.ToInt(1))
		}
		controlRequest := &vectorpb.BehaviorControlRequest{
			RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
				ControlRequest: &vectorpb.ControlRequest{
					Priority: priority,
				},
			},
		}
		go func() {
			// * begin - modified from official vector-go-sdk
			r, err := robot.Conn.BehaviorControl(
				context.Background(),
			)
			if err != nil {
				logger.Println(err)
				return
			}

			if err := r.Send(controlRequest); err != nil {
				logger.Println(err)
				return
			}

			for {
				ctrlresp, err := r.Recv()
				if err != nil {
					logger.Println(err)
					return
				}
				if ctrlresp.GetControlGrantedResponse() != nil {
					start <- true
					break
				}
			}

			for {
				select {
				case <-stop:
					if err := r.Send(
						&vectorpb.BehaviorControlRequest{
							RequestType: &vectorpb.BehaviorControlRequest_ControlRelease{
								ControlRelease: &vectorpb.ControlRelease{},
							},
						},
					); err != nil {
						logger.Println(err)
						return
					}
					return
				default:
					continue
				}
			}
			// * end - modified from official vector-go-sdk
		}()
		for range start {
			break
		}
		currentlyAssumed = true
		return 0
	}))
	L.SetGlobal("releaseBehaviorControl", L.NewFunction(func(*lua.LState) int {
		if currentlyAssumed {
			stop <- true
			currentlyAssumed = false
			return 0
		}
		return 1
	}))

}
