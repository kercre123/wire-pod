package main

import (
	"context"
	"fmt"

	sdk_wrapper "github.com/kercre123/vector-go-sdk/pkg/sdk-wrapper"
)

// test of SDK implementation

var Utterances = []string{"hello world"}
var Name = "SDK Plugin Test"

func Action(transcribedText string, botSerial string) string {
	fmt.Println("hello world plugin test")
	phrase := "hello world"
	sdk_wrapper.InitSDK(botSerial)
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
