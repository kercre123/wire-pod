package wirepod_ttr

//
// kgsim_cmds.go
//

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

//
// data section
//

const (
	ActionSayText      = 0
	ActionPlayAnimation = 1
	ActionPlayAnimationWI = 2
	ActionGetImage     = 3
	ActionNewRequest   = 4
	ActionPlaySound    = 5
)

var animationMap [][2]string = [][2]string{
	{"happy", "anim_onboarding_reacttoface_happy_01"},
	{"veryHappy", "anim_blackjack_victorwin_01"},
	{"sad", "anim_feedback_meanwords_01"},
	{"verySad", "anim_feedback_meanwords_01"},
	{"angry", "anim_rtpickup_loop_10"},
	{"frustrated", "anim_feedback_shutup_01"},
	{"dartingEyes", "anim_observing_self_absorbed_01"},
	{"confused", "anim_meetvictor_lookface_timeout_01"},
	{"thinking", "anim_explorer_scan_short_04"},
	{"celebrate", "anim_pounce_success_03"},
	{"love", "anim_feedback_iloveyou_02"},
}

var soundMap [][2]string = [][2]string{
	{"drumroll", "sounds/drumroll.wav"},
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

var (
	playAnimationWIDescription = "Enhances storytelling by playing an animation on the robot without interrupting speech. This should be used frequently to animate responses and engage the audience. Choose from parameters like happy, veryHappy, sad, verySad, angry, frustrated, dartingEyes, confused, thinking, celebrate, or love to complement the dialogue and maintain context."
	playAnimationDescription = "Interrupts speech to play an animation on the robot. Only use this when directed explicitly to play an animation, as it halts ongoing speech. Parameters include happy, veryHappy, sad, verySad, angry, frustrated, dartingEyes, confused, thinking, celebrate, or love for expressing emotions and reactions."
	getImageDescription = "Retrieves an image from the robot's camera and displays it in the next message. Use this command to conclude a sentence or response when prompted by the user or when describing visual content, such as what the robot sees, with options for front or lookingUp perspectives. If asked 'What do you see in front of you?' or similar, default to taking a photo. Inform the user of your action before using the command."
	newVoiceRequestDescription = "Starts a new voice command from the robot. Use this if you want more input from the user/if you want to carry out a conversation. You are the only one who can end it in this case. This goes at the end of your response, if you use it."
	// var playSoundDescription = "Plays a sound on the robot."
	LLMCommandsParamChoices = "happy, veryHappy, sad, verySad, angry, frustrated, dartingEyes, confused, thinking, celebrate, love"
)

var ValidLLMCommands []LLMCommand = []LLMCommand{
	{
		Command:         "playAnimationWI",
		Description:     playAnimationWIDescription,
		ParamChoices:    LLMCommandsParamChoices,
		Action:          ActionPlayAnimationWI,
		SupportedModels: []string{"all"},
	},
	{
		Command:         "playAnimation",
		Description:     playAnimationDescription,
		ParamChoices:    LLMCommandsParamChoices,
		Action:          ActionPlayAnimation,
		SupportedModels: []string{"all"},
	},
	{
		Command:         "getImage",
		Description:     getImageDescription,
		ParamChoices:    "front, lookingUp",
		Action:          ActionGetImage,
		SupportedModels: []string{openai.GPT4o},
	},
	{
		Command:         "newVoiceRequest",
		Description:     newVoiceRequestDescription,
		ParamChoices:    "now",
		Action:          ActionNewRequest,
		SupportedModels: []string{"all"},
	},
/*	{
	 	Command:      "playSound",
	 	Description:  playSoundDescription,
	 	ParamChoices: "drumroll",
	 	Action:       ActionPlaySound,
	},	*/
}

// is AI model supported
func ModelIsSupported(cmd LLMCommand, model string) bool {
	for _, str := range cmd.SupportedModels {
		if str == "all" || str == model {
			return true
		}
	}
	return false
}

var (
	promptForVector = " Format responses to communicate as an Anki Vector robot. User input may contain errors due to unreliable software. Evaluate for grammatical errors, missing words, or nonsensical phrases. Consider the overall context of the conversation to resolve ambiguities. If input is determined unintelligible or irrelevant, request clarification. Provide appropriate longer or shorter responses accordingly. "

	promptForVectorAI = " Integrate this format {{playAnimationWI||actionParameter}} into your responses, with the following command parameters: happy, veryHappy, sad, verySad, angry, frustrated, dartingEyes, confused, thinking, celebrate, and love. Add commas liberally in sentences to create pauses, for more natural-sounding speech breaks, and which allow for seamless embedding of commands. To enhance context, strategically place 'playAnimationWI' commands after commas, at the beginning of sentences, and one concluding command at the end of the finished LLM response. Select actions and parameters that amplify the expressiveness of the robot's phrases, aiming for engaging audience response. Examples: 1. Honestly, {{playAnimationWI||frustrated}} I never thought I'd ask, but, {{playAnimationWI||thinking}} where did I leave my keys? 2. You know, {{playAnimationWI||happy}} when everything goes right, {{playAnimationWI||veryHappy}} it feels like magic! 3. {{playAnimationWI||confused}} Sometimes, I think, {{playAnimationWI||thinking}} why do we even have Mondays? They're just like, {{playAnimationWI||frustrated}} a cruel joke, right? 4. Like, {{playAnimationWI||celebrate}} when I finish my tasks, {{playAnimationWI||excited}} I feel like celebrating, but, {{playAnimationWI||angry}} it's just Tuesday! 5. But overall, {{playAnimationWI||love}} I really do adore these odd days, {{playAnimationWI||excited}} they just add flavor to life! {{playAnimationWI||celebrate}} "

	conversationMode0 = "\n\nNOTE: You are NOT in 'conversation' mode. Avoid asking the user questions, and hence, from using 'newVoiceRequest'."

	conversationMode1 = "\n\nNOTE: You're now in 'conversation' mode. Near the end of your response, if you decide to ask the user a question, you MUST use 'newVoiceRequest'. Otherwise do not use 'newVoiceRequest' if you decide to end the conversation."
)

// add to orig prompt
func CreatePrompt(origPrompt string, model string, isKG bool) string {
	prompt := origPrompt + promptForVector  
	if vars.APIConfig.Knowledge.CommandsEnable {
		prompt += promptForVectorAI
		for _, cmd := range ValidLLMCommands {
			if ModelIsSupported(cmd, model) {
				prompt += fmt.Sprintf("\n\nCommand Name: %s\nDescription: %s\nParameter choices: %s", cmd.Command, cmd.Description, cmd.ParamChoices)
			}
		}
		if isKG && vars.APIConfig.Knowledge.SaveChat {
			promptAppentage := conversationMode1
			prompt = prompt + promptAppentage
		} else {
			promptAppentage := conversationMode0
			prompt = prompt + promptAppentage
		}
	}
	if os.Getenv("DEBUG_PRINT_PROMPT") == "true" {
		logger.Println(prompt)
	}
	return prompt
}

// parse actions from string
func GetActionsFromString(input string) []RobotAction {
	splitInput := strings.Split(input, "{{")
	if len(splitInput) == 1 {
		return []RobotAction{{Action: ActionSayText, Parameter: input}}
	}
	var actions []RobotAction
	for _, spl := range splitInput {
		if strings.TrimSpace(spl) == "" {
			continue
		}
		if !strings.Contains(spl, "}}") {
			actions = append(actions, RobotAction{Action: ActionSayText, Parameter: strings.TrimSpace(spl)})
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
			action := RobotAction{Action: ActionSayText, Parameter: strings.TrimSpace(strings.Split(spl, "}}")[1])}
			actions = append(actions, action)
		}
	}
	return actions
}

// commands to robot actions
func CmdParamToAction(cmd, param string) RobotAction {
	for _, command := range ValidLLMCommands {
		if cmd == command.Command {
			return RobotAction{Action: command.Action, Parameter: param}
		}
	}
	logger.Println("LLM tried to do a command which doesn't exist: " + cmd + " (param: " + param + ")")
	return RobotAction{Action: -1}
}


// Do 'Play Animation'
func DoPlayAnimation(animation string, esn string) error {
	// Find the robot instance associated with the given esn.
	var robot *vector.Vector
	for _, r := range vars.BotInfo.Robots {
		if r.Esn == esn {
			guid := r.GUID
			target := r.IPAddress + ":443"
			var err error
			robot, err = vector.New(vector.WithSerialNo(esn), vector.WithToken(guid), vector.WithTarget(target))
			if err != nil {
				return err
			}
			break
		}
	}

	if robot == nil {
		logger.Println("No robot found with ESN: " + esn)
		return errors.New("no robot found with the given ESN")
	}

	for _, animThing := range animationMap {
		if animation == animThing[0] {
			StartAnim_Queue(esn)
			robot.Conn.PlayAnimation(context.Background(), &vectorpb.PlayAnimationRequest{
				Animation: &vectorpb.Animation{Name: animThing[1]},
				Loops:     1,
			})
			StopAnim_Queue(esn)
			return nil
		}
	}

	logger.Println("Animation provided by LLM doesn't exist: " + animation)
	return nil
}


// Do 'Play AnimationWI'
func DoPlayAnimationWI(animation string, esn string) error {
	// Find the robot instance associated with the given esn.
	var robot *vector.Vector
	for _, r := range vars.BotInfo.Robots {
		if r.Esn == esn {
			guid := r.GUID
			target := r.IPAddress + ":443"
			var err error
			robot, err = vector.New(vector.WithSerialNo(esn), vector.WithToken(guid), vector.WithTarget(target))
			if err != nil {
				return err
			}
			break
		}
	}

	if robot == nil {
		logger.Println("No robot found with ESN: " + esn)
		return errors.New("no robot found with the given ESN")
	}

	for _, animThing := range animationMap {
		if animation == animThing[0] {
			go func() {
				StartAnim_Queue(esn)
				robot.Conn.PlayAnimation(context.Background(), &vectorpb.PlayAnimationRequest{
					Animation: &vectorpb.Animation{Name: animThing[1]},
					Loops:     1,
				})
				StopAnim_Queue(esn)
			}()
			return nil
		}
	}

	logger.Println("Animation provided by LLM doesn't exist: " + animation)
	return nil
}

// Do 'Play Sound'
func DoPlaySound(sound string, esn string) error {
	for _, soundThing := range soundMap {
		if sound == soundThing[0] {
			logger.Println("Would play sound")
		}
	}
	logger.Println("Sound provided by LLM doesn't exist: " + sound)
	return nil
}

// Do 'Say Text'
func DoSayText(input string, robot *vector.Vector) error {
	cleanedInput, err := removeSpecialCharacters(input)
	if err != nil {
		logger.Println("Error removing special characters:", err)
		return err
	}

	if (vars.APIConfig.STT.Language != "en-US" && vars.APIConfig.Knowledge.Provider == "openai") || vars.APIConfig.Knowledge.OpenAIVoiceWithEnglish {
		return DoSayText_OpenAI(robot, cleanedInput)
	}

	robot.Conn.SayText(context.Background(), &vectorpb.SayTextRequest{
		Text:           cleanedInput,
		UseVectorVoice: true,
		DurationScalar: 0.95,
	})
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

// Do 'Say Text for OpenAI'
func DoSayText_OpenAI(robot *vector.Vector, input string) error {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	openaiVoice := getOpenAIVoice(vars.APIConfig.Knowledge.OpenAIPrompt)
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


// Do 'Get Image'
func DoGetImage(msgs []openai.ChatCompletionMessage, param string, robot *vector.Vector, stopStop chan bool, esn string) {
	stopImaging := false
	go func() {
		for range stopStop {
			stopImaging = true
			break
		}
	}()
	logger.Println("Get image here...")
	robot.Conn.EnableMirrorMode(context.Background(), &vectorpb.EnableMirrorModeRequest{Enable: true})
	for i := 3; i > 0; i-- {
		if stopImaging {
			return
		}
		time.Sleep(time.Millisecond * 300)
		robot.Conn.SayText(context.Background(), &vectorpb.SayTextRequest{
			Text:           fmt.Sprint(i),
			UseVectorVoice: true,
			DurationScalar: 1.05,
		})
		if stopImaging {
			return
		}
	}
	resp, _ := robot.Conn.CaptureSingleImage(context.Background(), &vectorpb.CaptureSingleImageRequest{EnableHighResolution: true})
	robot.Conn.EnableMirrorMode(context.Background(), &vectorpb.EnableMirrorModeRequest{Enable: false})
	go func() {
		robot.Conn.PlayAnimation(context.Background(), &vectorpb.PlayAnimationRequest{
			Animation: &vectorpb.Animation{Name: "anim_photo_shutter_01"},
			Loops:     1,
		})
	}()
	reqBase64 := base64.StdEncoding.EncodeToString(resp.Data)

	msgs = append(msgs, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleUser,
		MultiContent: []openai.ChatMessagePart{{
			Type: openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{
				URL:    fmt.Sprintf("data:image/jpeg;base64,%s", reqBase64),
				Detail: openai.ImageURLDetailLow,
			},
		}},
	})

	var accumulatedResponseText string
	var completeResponseText string
	var responseChunks []string
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
		MaxTokens:        4095,
		Temperature:      1,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		Messages:         msgs,
		Stream:           true,
	}
	if vars.APIConfig.Knowledge.Provider == "openai" {
		aireq.Model = openai.GPT4o
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
			logger.Println("GPT-4 model cannot be accessed with this API key.")
			aireq.Model = openai.GPT3Dot5Turbo
			logger.Println("Falling back to " + aireq.Model)
			stream, err = c.CreateChatCompletionStream(ctx, aireq)
			if err != nil {
				logger.Println("OpenAI still not returning a response even after falling back.")
				return
			}
		} else {
			logger.Println("LLM error: " + err.Error())
			return
		}
	}

	fmt.Println("LLM stream response: ")
	go func() {
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				isDone = true
				newStr := responseChunks[0]
				for i, str := range responseChunks {
					if i == 0 {
						continue
					}
					newStr = newStr + " " + str
				}
				if strings.TrimSpace(newStr) != strings.TrimSpace(completeResponseText) {
					logger.Println("LLM debug: there is content after the last punctuation mark")
					extraBit := strings.TrimPrefix(accumulatedResponseText, newStr)
					responseChunks = append(responseChunks, extraBit)
				}
				if vars.APIConfig.Knowledge.SaveChat {
					Remember(msgs[len(msgs)-1],
						openai.ChatCompletionMessage{
							Role:    openai.ChatMessageRoleAssistant,
							Content: newStr,
						},
						esn)
				}
				logger.LogUI("LLM response for " + esn + ": " + newStr + "\n\n")
				logger.Println("LLM stream finished")
				return
			}

			if err != nil {
				logger.Println("Stream error: " + err.Error())
				return
			}
			completeResponseText += response.Choices[0].Delta.Content
			accumulatedResponseText += response.Choices[0].Delta.Content
			if strings.Contains(accumulatedResponseText, "...") || strings.Contains(accumulatedResponseText, ".'") || strings.Contains(accumulatedResponseText, ".\"") || strings.Contains(accumulatedResponseText, ".") || strings.Contains(accumulatedResponseText, "?") || strings.Contains(accumulatedResponseText, "!") {
				var sepStr string
				if strings.Contains(accumulatedResponseText, "...") {
					sepStr = "..."
				} else if strings.Contains(accumulatedResponseText, ".'") {
					sepStr = ".'"
				} else if strings.Contains(accumulatedResponseText, ".\"") {
					sepStr = ".\""
				} else if strings.Contains(accumulatedResponseText, ".") {
					sepStr = "."
				} else if strings.Contains(accumulatedResponseText, "?") {
					sepStr = "?"
				} else if strings.Contains(accumulatedResponseText, "!") {
					sepStr = "!"
				}
				splitResp := strings.Split(strings.TrimSpace(accumulatedResponseText), sepStr)
				responseChunks = append(responseChunks, strings.TrimSpace(splitResp[0])+sepStr)
				accumulatedResponseText = splitResp[1]
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
		respSlice := responseChunks
		if len(respSlice)-1 < numInResp {
			if !isDone {
				logger.Println("Waiting for more content from LLM...")
				for range speakReady {
					respSlice = responseChunks
					break
				}
			} else {
				break
			}
		}
		logger.Println(respSlice[numInResp])
		acts := GetActionsFromString(respSlice[numInResp])
		PerformActions(msgs, acts, esn, stopStop) // Directly using esn here
		numInResp = numInResp + 1
		if stopImaging {
			return
		}
	}
}

// Do 'New Request'
func DoNewRequest(robot *vector.Vector) {
	time.Sleep(time.Second / 3)
	robot.Conn.AppIntent(context.Background(), &vectorpb.AppIntentRequest{Intent: "knowledge_question"})
}

// perform the actions
func PerformActions(msgs []openai.ChatCompletionMessage, actions []RobotAction, esn string, stopStop chan bool) bool {
	stopPerforming := false
	var robot *vector.Vector

	// Find robot instance associated with ESN
	for _, r := range vars.BotInfo.Robots {
		if r.Esn == esn {
			guid := r.GUID
			target := r.IPAddress + ":443"
			var err error
			robot, err = vector.New(vector.WithSerialNo(esn), vector.WithToken(guid), vector.WithTarget(target))
			if err != nil {
				logger.Println("Error fetching robot:", err)
				return false
			}
			break
		}
	}

	if robot == nil {
		logger.Println("No robot found with ESN: " + esn)
		return false
	}

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
			if err := DoSayText(action.Parameter, robot); err != nil {
				logger.Println("Error in DoSayText:", err)
				return false
			}
		case action.Action == ActionPlayAnimation:
			if err := DoPlayAnimation(action.Parameter, esn); err != nil {
				logger.Println("Error in DoPlayAnimation:", err)
			}
		case action.Action == ActionPlayAnimationWI:
			if err := DoPlayAnimationWI(action.Parameter, esn); err != nil {
				logger.Println("Error in DoPlayAnimationWI:", err)
			}
		case action.Action == ActionNewRequest:
			go DoNewRequest(robot)
			return true
		case action.Action == ActionGetImage:
			DoGetImage(msgs, action.Parameter, robot, stopStop, esn) // Changing the ESN type
			return true
		case action.Action == ActionPlaySound:
			DoPlaySound(action.Parameter, esn)
		}
	}
	WaitForAnim_Queue(esn)
	return false
}

//
// Animation Queues Section
//
type AnimationQueue struct {
	ESN                  string
	AnimDone             chan struct{}
	AnimCurrentlyPlaying bool
}

var AnimationQueues = make(map[string]*AnimationQueue)

// wait for the queue
func WaitForAnim_Queue(esn string) {
	if q, exists := AnimationQueues[esn]; exists && q.AnimCurrentlyPlaying {
		<-q.AnimDone // Blocks until the animation is done
	}
}

// start the queue
func StartAnim_Queue(esn string) {
	if q, exists := AnimationQueues[esn]; exists {
		if q.AnimCurrentlyPlaying {
			logger.Println("(waiting for animation to be done...)")
			<-q.AnimDone // Blocks until the animation is done
		}
		q.AnimCurrentlyPlaying = true
	} else {
		AnimationQueues[esn] = &AnimationQueue{
			ESN:                  esn,
			AnimDone:             make(chan struct{}),
			AnimCurrentlyPlaying: true,
		}
	}
}

// stop the queue
func StopAnim_Queue(esn string) {
	if q, exists := AnimationQueues[esn]; exists {
		q.AnimCurrentlyPlaying = false
		close(q.AnimDone) // Close the channel to signal that the animation is done
		delete(AnimationQueues, esn) // Optionally remove the queue after stopping
	}
}

//
// kgsim_cmds.go - END
//

