package wirepod_ttr

import (
	"context"
	"log"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

func sayText(robot *vector.Vector, text string) {
	controlRequest := &vectorpb.BehaviorControlRequest{
		RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
			ControlRequest: &vectorpb.ControlRequest{
				Priority: vectorpb.ControlRequest_OVERRIDE_BEHAVIORS,
			},
		},
	}
	go func() {
		start := make(chan bool)
		stop := make(chan bool)
		go func() {
			// * begin - modified from official vector-go-sdk
			r, err := robot.Conn.BehaviorControl(
				context.Background(),
			)
			if err != nil {
				log.Println(err)
				return
			}

			if err := r.Send(controlRequest); err != nil {
				log.Println(err)
				return
			}

			for {
				ctrlresp, err := r.Recv()
				if err != nil {
					log.Println(err)
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
						log.Println(err)
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
			robot.Conn.SayText(
				context.Background(),
				&vectorpb.SayTextRequest{
					Text:           text,
					UseVectorVoice: true,
					DurationScalar: 1.0,
				},
			)
			stop <- true
		}
	}()
}

func BControl(robot *vector.Vector, ctx context.Context, start, stop chan bool) {
	controlRequest := &vectorpb.BehaviorControlRequest{
		RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
			ControlRequest: &vectorpb.ControlRequest{
				Priority: vectorpb.ControlRequest_OVERRIDE_BEHAVIORS,
			},
		},
	}

	go func() {
		// * begin - modified from official vector-go-sdk
		r, err := robot.Conn.BehaviorControl(
			ctx,
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
				logger.Println("KGSim: releasing behavior control (interrupt)")
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
}
