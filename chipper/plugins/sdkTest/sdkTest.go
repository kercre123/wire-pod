package main

import (
	"context"
	"log"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// test of SDK implementation

var Utterances = []string{"hello world"}
var Name = "SDK Plugin Test"

func behave(ctx context.Context, robot *vector.Vector, start chan bool, stop chan bool) {
	controlRequest := &vectorpb.BehaviorControlRequest{
		RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
			ControlRequest: &vectorpb.ControlRequest{
				Priority: vectorpb.ControlRequest_OVERRIDE_BEHAVIORS,
			},
		},
	}
	go func() {

		go func() {
			// * begin - modified from official vector-go-sdk
			r, err := robot.Conn.BehaviorControl(
				ctx,
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
					logger.Println("KGSim: releasing behavior control (interrupt)")
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
	}()
}

func Action(transcribedText string, botSerial string, guid string, target string) (string, string) {
	logger.Println("hello world plugin test")
	phrase := "hello world"
	robot, err := vector.New(
		vector.WithSerialNo(botSerial),
		vector.WithTarget(target),
		vector.WithToken(guid),
	)
	if err != nil {
		logger.Println(err)
		return "intent_imperative_praise", ""
	}
	ctx := context.Background()
	start := make(chan bool)
	stop := make(chan bool)
	go func() {
		behave(ctx, robot, start, stop)
	}()

	for {
		select {
		case <-start:
			robot.Conn.SayText(
				ctx,
				&vectorpb.SayTextRequest{
					Text:           phrase,
					UseVectorVoice: true,
					DurationScalar: 1,
				},
			)
			stop <- true
			return "intent_imperative_praise", ""
		}
	}
}
