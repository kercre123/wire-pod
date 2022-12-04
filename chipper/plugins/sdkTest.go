package main

import (
	"context"
	"fmt"

	sdk_wrapper "github.com/fforchino/vector-go-sdk/pkg/sdk-wrapper"
)

// test of SDK implementation

var Utterances = []string{"hello world"}
var Name = "SDK Plugin Test"

func Action(transcribedText string, botSerial string) string {
	fmt.Println("hello world plugin test")
	phrase := "hello world"
	sdk_wrapper.InitSDKForWirepod(botSerial)
	ctx := context.Background()
	start := make(chan bool)
	stop := make(chan bool)
	go func() {
		err := sdk_wrapper.Robot.BehaviorControl(ctx, start, stop)
		if err != nil {
			fmt.Println(err)
		}
	}()

	for {
		select {
		case <-start:
			sdk_wrapper.SayText(phrase)
			stop <- true
			return "intent_imperative_praise"
		}
	}
}
