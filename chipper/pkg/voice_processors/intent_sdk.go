package wirepod

import (
	"context"
	"fmt"
	sdk_wrapper "github.com/fforchino/vector-go-sdk/pkg/sdk-wrapper"
	"image/color"
	"math/rand"
	"strings"
	"time"
)

func handleSdkIntent(req interface{}, intent string, speechText string, intentParams map[string]string, isParam bool, justThisBotNum int, serial string) string {
	returnIntent := "intent_greeting_hello"

	sdk_wrapper.InitSDK(serial)

	ctx := context.Background()
	start := make(chan bool)
	stop := make(chan bool)

	go func() {
		_ = sdk_wrapper.Robot.BehaviorControl(ctx, start, stop)
	}()

	for {
		select {
		case <-start:
			if strings.Contains(intent, "intent_sdk_set_robot_name") {
				if intentParams["username"] != "" {
					sdk_wrapper.SetRobotName(intentParams["username"])
					returnIntent = "intent_imperative_affirmative"
					sdk_wrapper.SayText("Ok. My name is " + intentParams["username"])
				} else {
					returnIntent = "intent_imperative_negative"
				}
			} else if strings.Contains(intent, "intent_sdk_get_robot_name") {
				robotName := sdk_wrapper.GetRobotName()
				sdk_wrapper.SayText("My name is " + robotName)
			} else if strings.Contains(intent, "intent_sdk_roll_a_die") {
				rollADie()
			}
		}
	}

	return returnIntent
}

func rollADie() {
	sdk_wrapper.MoveHead(3.0)
	sdk_wrapper.SetBackgroundColor(color.RGBA{0, 0, 0, 0})

	sdk_wrapper.UseVectorEyeColorInImages(true)
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	die := r1.Intn(6) + 1
	dieImage := fmt.Sprintf("images/dice/%d.png", die)
	dieImage = sdk_wrapper.GetDataPath(dieImage)

	sdk_wrapper.DisplayAnimatedGif(sdk_wrapper.GetDataPath("images/dice/roll-the-dice.gif"), sdk_wrapper.ANIMATED_GIF_SPEED_FASTEST, 1, false)
	sdk_wrapper.DisplayImage(dieImage, 100, false)
	sdk_wrapper.PlaySound(sdk_wrapper.SYSTEMSOUND_WIN)
	sdk_wrapper.SayText(fmt.Sprintf("You rolled a %d", die))
	sdk_wrapper.DisplayImageWithTransition(dieImage, 1000, sdk_wrapper.IMAGE_TRANSITION_FADE_OUT, 10)
}
