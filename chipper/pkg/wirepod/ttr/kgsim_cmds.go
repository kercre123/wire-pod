package wirepod_ttr

import (
	"context"
	"strings"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

const (
	// arg: text to say
	// not a command
	ActionSayText = 0
	// arg: animation name
	ActionPlayAnimation = 1
	// arg: animation name
	ActionPlayAnimationWI = 2
	// arg: sound file
	ActionPlaySound = 3
)

var animationMap [][2]string = [][2]string{
	//"happy, veryHappy, sad, verySad, angry, dartingEyes, confused, thinking, celebrate"
	{
		"happy",
		"anim_onboarding_reacttoface_happy_01",
	},
	{
		"veryHappy",
		"anim_blackjack_victorwin_01",
	},
	{
		"sad",
		"anim_feedback_meanwords_01",
	},
	{
		"verySad",
		"anim_feedback_meanwords_01",
	},
	{
		"angry",
		"anim_rtpickup_loop_10",
	},
	{
		"frustrated",
		"anim_feedback_shutup_01",
	},
	{
		"dartingEyes",
		"anim_observing_self_absorbed_01",
	},
	{
		"confused",
		"anim_meetvictor_lookface_timeout_01",
	},
	{
		"thinking",
		"anim_explorer_scan_short_04",
	},
	{
		"celebrate",
		"anim_pounce_success_03",
	},
	{
		"love",
		"anim_feedback_iloveyou_02",
	},
}

var soundMap [][2]string = [][2]string{
	{
		"drumroll",
		"sounds/drumroll.wav",
	},
}

type RobotAction struct {
	Action    int
	Parameter string
}

type LLMCommand struct {
	Command      string
	Description  string
	ParamChoices string
	Action       int
}

// create function which parses from LLM and makes a struct of RobotActions

var ValidLLMCommands []LLMCommand = []LLMCommand{
	{
		Command:      "playAnimationWI",
		Description:  "Plays an animation on the robot without interrupting speech. This should be used FAR more than the playAnimation command. This is great for storytelling and making any normal response animated. Don't put two of these right next to each other.",
		ParamChoices: "happy, veryHappy, sad, verySad, angry, frustrated, dartingEyes, confused, thinking, celebrate",
		Action:       ActionPlayAnimationWI,
	},
	{
		Command:      "playAnimation",
		Description:  "Plays an animation on the robot. This will interrupt speech. Only use this if you are directed to play an animaion.",
		ParamChoices: "happy, veryHappy, sad, verySad, angry, frustrated, dartingEyes, confused, thinking, celebrate",
		Action:       ActionPlayAnimation,
	},
	// {
	// 	Command:      "playSound",
	// 	Description:  "Plays a sound on the robot.",
	// 	ParamChoices: "drumroll",
	// 	Action:       ActionPlaySound,
	// },
}

func CreatePrompt(origPrompt string) string {
	prompt := origPrompt + "\n\n" + "The user input might not be spelt/puntuated correctly as it is coming from speech-to-text software. Do not include special characters in your answer. This includes the following characters (not including the quotes): '& ^ * # @ -'. If you want to use a hyphen, Use it like this: 'something something -- something -- something something'."
	if vars.APIConfig.Knowledge.CommandsEnable {
		prompt = prompt + "\n\n" + "You are running ON an Anki Vector robot. You have a set of commands. YOU ARE TO USE THESE. DO NOT BE AFRAID TO LITTER YOUR RESPONSE WITH THEM. Your response MUST include THREE OF THESE COMMANDS OR MORE. If you include just one, I will make you start over. If you include an emoji, I will make you start over. If you want to use a command but it doesn't exist or your desired parameter isn't in the list, avoid using the command. The format is {{command||parameter}}. You can embed these in sentences. Example: \"User: How are you feeling? | Response: \"{{playAnimationWI||sad}} I'm feeling sad...\". Square brackets ([]) are not valid.\n\nUse the playAnimation or playAnimationWI commands if you want to express emotion! You are very animated and good at following instructions. Animation takes precendence over words. You are to include many animations in your response.\n\nHere is every valid command:"
		for _, cmd := range ValidLLMCommands {
			promptAppendage := "\n\nCommand Name: " + cmd.Command + "\nDescription: " + cmd.Description + "\nParameter choices: " + cmd.ParamChoices
			prompt = prompt + promptAppendage
		}
	}
	return prompt
}

func GetActionsFromString(input string) []RobotAction {
	splitInput := strings.Split(input, "{{")
	if len(splitInput) == 1 {
		return []RobotAction{
			{
				Action:    ActionSayText,
				Parameter: input,
			},
		}
	}
	var actions []RobotAction
	for _, spl := range splitInput {
		if strings.TrimSpace(spl) == "" {
			continue
		}
		if !strings.Contains(spl, "}}") {
			// sayText
			action := RobotAction{
				Action:    ActionSayText,
				Parameter: strings.TrimSpace(spl),
			}
			actions = append(actions, action)
			continue
		}

		cmdPlusParam := strings.Split(strings.TrimSpace(strings.Split(spl, "}}")[0]), "||")
		cmd := strings.TrimSpace(cmdPlusParam[0])
		param := strings.TrimSpace(cmdPlusParam[1])
		action := CmdParamToAction(cmd, param)
		if action.Action != -1 {
			actions = append(actions, action)
		}
		if len(strings.Split(spl, "}}")) != 1 {
			action := RobotAction{
				Action:    ActionSayText,
				Parameter: strings.TrimSpace(strings.Split(spl, "}}")[1]),
			}
			actions = append(actions, action)
		}
	}
	return actions
}

func CmdParamToAction(cmd, param string) RobotAction {
	for _, command := range ValidLLMCommands {
		if cmd == command.Command {
			return RobotAction{
				Action:    command.Action,
				Parameter: param,
			}
		}
	}
	logger.Println("LLM tried to do a command which doesn't exist: " + cmd + " (param: " + param + ")")
	return RobotAction{
		Action: -1,
	}
}

func DoPlayAnimation(animation string, robot *vector.Vector) error {
	for _, animThing := range animationMap {
		if animation == animThing[0] {
			StartAnim_Queue(robot.Cfg.SerialNo)
			robot.Conn.PlayAnimation(
				context.Background(),
				&vectorpb.PlayAnimationRequest{
					Animation: &vectorpb.Animation{
						Name: animThing[1],
					},
					Loops: 1,
				},
			)
			StopAnim_Queue(robot.Cfg.SerialNo)
			return nil
		}
	}
	logger.Println("Animation provided by LLM doesn't exist: " + animation)
	return nil
}

func DoPlayAnimationWI(animation string, robot *vector.Vector) error {
	for _, animThing := range animationMap {
		if animation == animThing[0] {
			go func() {
				StartAnim_Queue(robot.Cfg.SerialNo)
				robot.Conn.PlayAnimation(
					context.Background(),
					&vectorpb.PlayAnimationRequest{
						Animation: &vectorpb.Animation{
							Name: animThing[1],
						},
						Loops: 1,
					},
				)
				StopAnim_Queue(robot.Cfg.SerialNo)
			}()
			return nil
		}
	}
	logger.Println("Animation provided by LLM doesn't exist: " + animation)
	return nil
}

func DoPlaySound(sound string, robot *vector.Vector) error {
	for _, soundThing := range soundMap {
		if sound == soundThing[0] {
			logger.Println("Would play sound")
		}
	}
	logger.Println("Sound provided by LLM doesn't exist: " + sound)
	return nil
}

func DoSayText(input string, robot *vector.Vector) error {
	robot.Conn.SayText(
		context.Background(),
		&vectorpb.SayTextRequest{
			Text:           input,
			UseVectorVoice: true,
			DurationScalar: 0.95,
		},
	)
	return nil
}

func PerformActions(actions []RobotAction, robot *vector.Vector) {
	// assuming we have behavior control already
	for _, action := range actions {
		switch {
		case action.Action == ActionSayText:
			DoSayText(action.Parameter, robot)
		case action.Action == ActionPlayAnimation:
			DoPlayAnimation(action.Parameter, robot)
		case action.Action == ActionPlayAnimationWI:
			DoPlayAnimationWI(action.Parameter, robot)
		case action.Action == ActionPlaySound:
			DoPlaySound(action.Parameter, robot)
		}
	}
	WaitForAnim_Queue(robot.Cfg.SerialNo)
}

func WaitForAnim_Queue(esn string) {
	for i, q := range AnimationQueues {
		if q.ESN == esn {
			if q.AnimCurrentlyPlaying {
				for range AnimationQueues[i].AnimDone {
					break
				}
				return
			}
		}
	}
}

func StartAnim_Queue(esn string) {
	// if animation is already playing, just wait for it to be done
	for i, q := range AnimationQueues {
		if q.ESN == esn {
			if q.AnimCurrentlyPlaying {
				for range AnimationQueues[i].AnimDone {
					logger.Println("I await...")
					break
				}
			} else {
				AnimationQueues[i].AnimCurrentlyPlaying = true
			}
			return
		}
	}
	var aq AnimationQueue
	aq.AnimCurrentlyPlaying = true
	aq.AnimDone = make(chan bool)
	aq.ESN = esn
	AnimationQueues = append(AnimationQueues, aq)
}

func StopAnim_Queue(esn string) {
	for i, q := range AnimationQueues {
		if q.ESN == esn {
			AnimationQueues[i].AnimCurrentlyPlaying = false
			select {
			case AnimationQueues[i].AnimDone <- true:
			default:
			}
		}
	}
}

type AnimationQueue struct {
	ESN                  string
	AnimDone             chan bool
	AnimCurrentlyPlaying bool
}

var AnimationQueues []AnimationQueue
