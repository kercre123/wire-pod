package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kercre123/vector-go-sdk/pkg/vector"
	"github.com/kercre123/vector-go-sdk/pkg/vectorpb"
)

// test of SDK implementation

var Utterances = []string{"hello world"}
var Name = "SDK Plugin Test"

func Action(transcribedText string, botSerial string) string {
	fmt.Println("hello world plugin test")
	phrase := "hello world"
	tokenBytes, _ := os.ReadFile("./robotGUID")
	token := string(tokenBytes)

	robot, _ := vector.New(
		vector.WithTarget("192.168.0.199:443"),
		vector.WithToken(token),
	)
	ctx := context.Background()
	start := make(chan bool)
	stop := make(chan bool)
	go func() {
		err := robot.BehaviorControl(ctx, start, stop)
		if err != nil {
			fmt.Println(err)
		}
	}()

	for {
		select {
		case <-start:
			robot.Conn.SayText(
				ctx,
				&vectorpb.SayTextRequest{
					Text:           phrase,
					UseVectorVoice: true,
					DurationScalar: 1.0,
				},
			)
			stop <- true
			return "intent_imperative_praise"
		}
	}
	return "intent_imperative_praise"
}
