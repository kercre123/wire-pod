package main

import (
	"context"
	"fmt"

	"github.com/kercre123/vector-go-sdk/pkg/vector"
)

// test of SDK implementation

var Utterances = []string{"hello world"}
var Name = "SDK Plugin Test"

func Action(transcribedText string, botSerial string) string {
	fmt.Println("hello world plugin test")
	phrase := "hello world"
	vector.New(
		vector.WithTarget(""),
		vector.WithToken(""),
	)
	ctx := context.Background()
	start := make(chan bool)
	stop := make(chan bool)
	go func() {
		_ = sdk_wrapper.Robot.BehaviorControl(ctx, start, stop)
	}()

	for {
		select {
		case <-start:
			sdk_wrapper.SayText(phrase)
			stop <- true
			return "intent_imperative_praise"
		}
	}
	return "intent_imperative_praise"
}
