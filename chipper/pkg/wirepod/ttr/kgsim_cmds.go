package wirepod_ttr

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/sashabaranov/go-openai"
)

const (
	// arg: text to say
	// not a command
	ActionSayText = 0
	// arg: animation name
	ActionPlayAnimation = 1
	// arg: animation name
	ActionPlayAnimationWI = 2
	// arg: now
	ActionGetImage   = 3
	ActionNewRequest = 4
	// arg: sound file
	ActionPlaySound = 4
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
	Command         string
	Description     string
	ParamChoices    string
	Action          int
	SupportedModels []string
}

// create function which parses from LLM and makes a struct of RobotActions

var ValidLLMCommands []LLMCommand = []LLMCommand{
	{
		Command:         "playAnimationWI",
		Description:     "Plays an animation on the robot without interrupting speech. This should be used FAR more than the playAnimation command. This is great for storytelling and making any normal response animated. Don't put two of these right next to each other. Use this MANY times. The param choices are the only choices you have. You can't create any.",
		ParamChoices:    "happy, veryHappy, sad, verySad, angry, frustrated, dartingEyes, confused, thinking, celebrate, love",
		Action:          ActionPlayAnimationWI,
		SupportedModels: []string{"all"},
	},
	{
		Command:         "playAnimation",
		Description:     "Plays an animation on the robot. This will interrupt speech. Only use this if you are directed to play an animaion.",
		ParamChoices:    "happy, veryHappy, sad, verySad, angry, frustrated, dartingEyes, confused, thinking, celebrate, love",
		Action:          ActionPlayAnimation,
		SupportedModels: []string{"all"},
	},
	{
		Command:     "getImage",
		Description: "Gets an image from the robot's camera and places it in the next message. If you want to do this, tell the user what you are about to do THEN use the command. This command should END a sentence. Your response will be stopped when this command is recognized. If a user says something like 'what do you see', you should assume that you need to take a new photo. Do NOT automatically assume that you are analyzing a previous photo.",
		// not impl yet
		ParamChoices:    "front, lookingUp",
		Action:          ActionGetImage,
		SupportedModels: []string{openai.GPT4o, openai.GPT4oMini},
	},
	{
		Command:         "newVoiceRequest",
		Description:     "Starts a new voice command from the robot. Use this if you want more input from the user after your response/if you want to carry out a conversation. Below this, there should be a NOTE telling you whether you are in conversation mode or not. If you are, DONT BE AFRAID TO USE THIS COMMAND! This goes at the end of your response, if you use it.",
		ParamChoices:    "now",
		Action:          ActionNewRequest,
		SupportedModels: []string{"all"},
	},
	// {
	// 	Command:      "playSound",
	// 	Description:  "Plays a sound on the robot.",
	// 	ParamChoices: "drumroll",
	// 	Action:       ActionPlaySound,
	// },
}

func ModelIsSupported(cmd LLMCommand, model string) bool {
	for _, str := range cmd.SupportedModels {
		if str == "all" || str == model {
			return true
		}
	}
	return false
}

func CreatePrompt(origPrompt string, model string, isKG bool) string {
	prompt := origPrompt + "\n\n" + "Keep in mind, user input comes from speech-to-text software, so respond accordingly. No special characters, especially these: & ^ * # @ - . No lists. No formatting."
	if vars.APIConfig.Knowledge.CommandsEnable {
		prompt = prompt + "\n\n" + "You are running ON an Anki Vector robot. You have a set of commands. If you include an emoji, I will make you start over. If you want to use a command but it doesn't exist or your desired parameter isn't in the list, avoid using the command. The format is {{command||parameter}}. You can embed these in sentences. Example: \"User: How are you feeling? | Response: \"{{playAnimationWI||sad}} I'm feeling sad...\". Square brackets ([]) are not valid.\n\nUse the playAnimation or playAnimationWI commands if you want to express emotion! You are very animated and good at following instructions. Animation takes precendence over words. You are to include many animations in your response.\n\nHere is every valid command:"
		for _, cmd := range ValidLLMCommands {
			if ModelIsSupported(cmd, model) {
				promptAppendage := "\n\nCommand Name: " + cmd.Command + "\nDescription: " + cmd.Description + "\nParameter choices: " + cmd.ParamChoices
				prompt = prompt + promptAppendage
			}
		}
		if isKG && vars.APIConfig.Knowledge.SaveChat {
			promptAppentage := "\n\nNOTE: You are in 'conversation' mode. If you ask the user a question near the end of your response, you MUST use newVoiceRequest. If you decide you want to end the conversation, you should not use it."
			prompt = prompt + promptAppentage
		} else {
			promptAppentage := "\n\nNOTE: You are NOT in 'conversation' mode. Refrain from asking the user any questions and from using newVoiceRequest."
			prompt = prompt + promptAppentage
		}
	}
	if os.Getenv("DEBUG_PRINT_PROMPT") == "true" {
		logger.Println(prompt)
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

	// just before vector speaks
	removeSpecialCharacters(input)

	if (vars.APIConfig.STT.Language != "en-US" && vars.APIConfig.Knowledge.Provider == "openai") || vars.APIConfig.Knowledge.OpenAIVoiceWithEnglish {
		err := DoSayText_OpenAI(robot, input)
		return err
	}
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

func pcmLength(data []byte) time.Duration {
	bytesPerSample := 2
	sampleRate := 16000
	numSamples := len(data) / bytesPerSample
	duration := time.Duration(numSamples*1000/sampleRate) * time.Millisecond
	return duration
}

func getOpenAIVoice(voice string) openai.SpeechVoice {
	voiceMap := map[string]openai.SpeechVoice{
		"alloy":   openai.VoiceAlloy,
		"onyx":    openai.VoiceOnyx,
		"fable":   openai.VoiceFable,
		"shimmer": openai.VoiceShimmer,
		"nova":    openai.VoiceNova,
		"echo":    openai.VoiceEcho,
		"":        openai.VoiceFable,
	}
	return voiceMap[voice]
}

// TODO
func DoSayText_OpenAI(robot *vector.Vector, input string) error {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	openaiVoice := getOpenAIVoice(vars.APIConfig.Knowledge.OpenAIVoice)
	// if vars.APIConfig.Knowledge.OpenAIVoice == "" {
	// 	openaiVoice = openai.VoiceFable
	// } else {
	// 	openaiVoice = getOpenAIVoice(vars.APIConfig.Knowledge.OpenAIPrompt)
	// }
	oc := openai.NewClient(vars.APIConfig.Knowledge.Key)
	resp, err := oc.CreateSpeech(context.Background(), openai.CreateSpeechRequest{
		Model:          openai.TTSModel1,
		Input:          input,
		Voice:          openaiVoice,
		ResponseFormat: openai.SpeechResponseFormatPcm,
	})
	if err != nil {
		logger.Println(err)
		return err
	}
	speechBytes, _ := io.ReadAll(resp)
	vclient, err := robot.Conn.ExternalAudioStreamPlayback(context.Background())
	if err != nil {
		return err
	}
	vclient.Send(&vectorpb.ExternalAudioStreamRequest{
		AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamPrepare{
			AudioStreamPrepare: &vectorpb.ExternalAudioStreamPrepare{
				AudioFrameRate: 16000,
				AudioVolume:    100,
			},
		},
	})
	//time.Sleep(time.Millisecond * 30)
	audioChunks := downsample24kTo16k(speechBytes)

	var chunksToDetermineLength []byte
	for _, chunk := range audioChunks {
		chunksToDetermineLength = append(chunksToDetermineLength, chunk...)
	}
	go func() {
		for _, chunk := range audioChunks {
			vclient.Send(&vectorpb.ExternalAudioStreamRequest{
				AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamChunk{
					AudioStreamChunk: &vectorpb.ExternalAudioStreamChunk{
						AudioChunkSizeBytes: 1024,
						AudioChunkSamples:   chunk,
					},
				},
			})
			time.Sleep(time.Millisecond * 25)
		}
		vclient.Send(&vectorpb.ExternalAudioStreamRequest{
			AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamComplete{
				AudioStreamComplete: &vectorpb.ExternalAudioStreamComplete{},
			},
		})
	}()
	time.Sleep(pcmLength(chunksToDetermineLength) + (time.Millisecond * 50))
	return nil
}

func DoGetImage(msgs []openai.ChatCompletionMessage, param string, robot *vector.Vector, stopStop chan bool) {
	stopImaging := false
	go func() {
		for range stopStop {
			stopImaging = true
			break
		}
	}()
	logger.Println("Get image here...")
	// get image
	robot.Conn.EnableMirrorMode(context.Background(), &vectorpb.EnableMirrorModeRequest{
		Enable: true,
	})
	for i := 3; i > 0; i-- {
		if stopImaging {
			return
		}
		time.Sleep(time.Millisecond * 300)
		robot.Conn.SayText(
			context.Background(),
			&vectorpb.SayTextRequest{
				Text:           fmt.Sprint(i),
				UseVectorVoice: true,
				DurationScalar: 1.05,
			},
		)
		if stopImaging {
			return
		}
	}
	resp, _ := robot.Conn.CaptureSingleImage(
		context.Background(),
		&vectorpb.CaptureSingleImageRequest{
			EnableHighResolution: true,
		},
	)
	robot.Conn.EnableMirrorMode(
		context.Background(),
		&vectorpb.EnableMirrorModeRequest{
			Enable: false,
		},
	)
	go func() {
		robot.Conn.PlayAnimation(
			context.Background(),
			&vectorpb.PlayAnimationRequest{
				Animation: &vectorpb.Animation{
					Name: "anim_photo_shutter_01",
				},
				Loops: 1,
			},
		)
	}()
	// encode to base64
	reqBase64 := base64.StdEncoding.EncodeToString(resp.Data)

	// add image to messages
	msgs = append(msgs, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleUser,
		MultiContent: []openai.ChatMessagePart{
			{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL:    fmt.Sprintf("data:image/jpeg;base64,%s", reqBase64),
					Detail: openai.ImageURLDetailLow,
				},
			},
		},
	})

	// recreate openai
	var fullRespText string
	var fullfullRespText string
	var fullRespSlice []string
	var isDone bool
	var c *openai.Client
	if vars.APIConfig.Knowledge.Provider == "together" {
		if vars.APIConfig.Knowledge.Model == "" {
			vars.APIConfig.Knowledge.Model = "meta-llama/Llama-2-70b-chat-hf"
			vars.WriteConfigToDisk()
		}
		conf := openai.DefaultConfig(vars.APIConfig.Knowledge.Key)
		conf.BaseURL = "https://api.together.xyz/v1"
		c = openai.NewClientWithConfig(conf)
	} else if vars.APIConfig.Knowledge.Provider == "openai" {
		c = openai.NewClient(vars.APIConfig.Knowledge.Key)
	}
	ctx := context.Background()
	speakReady := make(chan string)

	aireq := openai.ChatCompletionRequest{
		MaxTokens:        2048,
		Temperature:      1,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		Messages:         msgs,
		Stream:           true,
	}
	if vars.APIConfig.Knowledge.Provider == "openai" {
		aireq.Model = openai.GPT4oMini
		logger.Println("Using " + aireq.Model)
	} else {
		logger.Println("Using " + vars.APIConfig.Knowledge.Model)
		aireq.Model = vars.APIConfig.Knowledge.Model
	}
	if stopImaging {
		return
	}
	stream, err := c.CreateChatCompletionStream(ctx, aireq)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") && vars.APIConfig.Knowledge.Provider == "openai" {
			logger.Println("GPT-4 model cannot be accessed with this API key. You likely need to add more than $5 dollars of funds to your OpenAI account.")
			logger.LogUI("GPT-4 model cannot be accessed with this API key. You likely need to add more than $5 dollars of funds to your OpenAI account.")
			aireq.Model = openai.GPT3Dot5Turbo
			logger.Println("Falling back to " + aireq.Model)
			logger.LogUI("Falling back to " + aireq.Model)
			stream, err = c.CreateChatCompletionStream(ctx, aireq)
			if err != nil {
				logger.Println("OpenAI still not returning a response even after falling back. Erroring.")
				return
			}
		} else {
			logger.Println("LLM error: " + err.Error())
			return
		}
	}
	//defer stream.Close()

	fmt.Println("LLM stream response: ")
	go func() {
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				isDone = true
				newStr := fullRespSlice[0]
				for i, str := range fullRespSlice {
					if i == 0 {
						continue
					}
					newStr = newStr + " " + str
				}
				if strings.TrimSpace(newStr) != strings.TrimSpace(fullfullRespText) {
					logger.Println("LLM debug: there is content after the last punctuation mark")
					extraBit := strings.TrimPrefix(fullRespText, newStr)
					fullRespSlice = append(fullRespSlice, extraBit)
				}
				if vars.APIConfig.Knowledge.SaveChat {
					Remember(msgs[len(msgs)-1],
						openai.ChatCompletionMessage{
							Role:    openai.ChatMessageRoleAssistant,
							Content: newStr,
						},
						robot.Cfg.SerialNo)
				}
				logger.LogUI("LLM response for " + robot.Cfg.SerialNo + ": " + newStr)
				logger.Println("LLM stream finished")
				return
			}

			if err != nil {
				logger.Println("Stream error: " + err.Error())
				return
			}
			fullfullRespText = fullfullRespText + removeSpecialCharacters(response.Choices[0].Delta.Content)
			fullRespText = fullRespText + removeSpecialCharacters(response.Choices[0].Delta.Content)
			if strings.Contains(fullRespText, "...") || strings.Contains(fullRespText, ".'") || strings.Contains(fullRespText, ".\"") || strings.Contains(fullRespText, ".") || strings.Contains(fullRespText, "?") || strings.Contains(fullRespText, "!") {
				var sepStr string
				if strings.Contains(fullRespText, "...") {
					sepStr = "..."
				} else if strings.Contains(fullRespText, ".'") {
					sepStr = ".'"
				} else if strings.Contains(fullRespText, ".\"") {
					sepStr = ".\""
				} else if strings.Contains(fullRespText, ".") {
					sepStr = "."
				} else if strings.Contains(fullRespText, "?") {
					sepStr = "?"
				} else if strings.Contains(fullRespText, "!") {
					sepStr = "!"
				}
				splitResp := strings.Split(strings.TrimSpace(fullRespText), sepStr)
				fullRespSlice = append(fullRespSlice, strings.TrimSpace(splitResp[0])+sepStr)
				fullRespText = splitResp[1]
				select {
				case speakReady <- strings.TrimSpace(splitResp[0]) + sepStr:
				default:
				}
			}
		}
	}()
	numInResp := 0
	for {
		if stopImaging {
			return
		}
		respSlice := fullRespSlice
		if len(respSlice)-1 < numInResp {
			if !isDone {
				logger.Println("Waiting for more content from LLM...")
				for range speakReady {
					respSlice = fullRespSlice
					break
				}
			} else {
				break
			}
		}
		logger.Println(respSlice[numInResp])
		acts := GetActionsFromString(respSlice[numInResp])
		PerformActions(msgs, acts, robot, stopStop)
		numInResp = numInResp + 1
		if stopImaging {
			return
		}
	}
}

func DoNewRequest(robot *vector.Vector) {
	time.Sleep(time.Second / 3)
	robot.Conn.AppIntent(context.Background(), &vectorpb.AppIntentRequest{Intent: "knowledge_question"})
}

func PerformActions(msgs []openai.ChatCompletionMessage, actions []RobotAction, robot *vector.Vector, stopStop chan bool) bool {
	// assuming we have behavior control already
	stopPerforming := false
	go func() {
		for range stopStop {
			stopPerforming = true
		}
	}()
	for _, action := range actions {
		if stopPerforming {
			return false
		}
		switch {
		case action.Action == ActionSayText:
			DoSayText(action.Parameter, robot)
		case action.Action == ActionPlayAnimation:
			DoPlayAnimation(action.Parameter, robot)
		case action.Action == ActionPlayAnimationWI:
			DoPlayAnimationWI(action.Parameter, robot)
		case action.Action == ActionNewRequest:
			go DoNewRequest(robot)
			return true
		case action.Action == ActionGetImage:
			DoGetImage(msgs, action.Parameter, robot, stopStop)
			return true
		case action.Action == ActionPlaySound:
			DoPlaySound(action.Parameter, robot)
		}
	}
	WaitForAnim_Queue(robot.Cfg.SerialNo)
	return false
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
					logger.Println("(waiting for animation to be done...)")
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
