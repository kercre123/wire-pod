package wirepod_ttr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/sashabaranov/go-openai"
)

func GetChat(esn string) vars.RememberedChat {
	for _, chat := range vars.RememberedChats {
		if chat.ESN == esn {
			return chat
		}
	}
	return vars.RememberedChat{
		ESN: esn,
	}
}

func PlaceChat(chat vars.RememberedChat) {
	for i, achat := range vars.RememberedChats {
		if achat.ESN == chat.ESN {
			vars.RememberedChats[i] = chat
			vars.SaveChats()
			return
		}
	}
	vars.RememberedChats = append(vars.RememberedChats, chat)
	vars.SaveChats()
}

// remember last 16 lines of chat
func Remember(user, ai openai.ChatCompletionMessage, esn string) {
	chatAppend := []openai.ChatCompletionMessage{
		user,
		ai,
	}
	currentChat := GetChat(esn)
	if len(currentChat.Chats) == 16 {
		var newChat vars.RememberedChat
		newChat.ESN = currentChat.ESN
		for i, chat := range currentChat.Chats {
			if i < 2 {
				continue
			}
			newChat.Chats = append(newChat.Chats, chat)
		}
		currentChat = newChat
	}
	currentChat.ESN = esn
	currentChat.Chats = append(currentChat.Chats, chatAppend...)
	PlaceChat(currentChat)
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

func removeSpecialCharacters(str string) string {

	// these two lines create a transformation that decomposes characters, removes non-spacing marks (like diacritics), and then recomposes the characters, effectively removing special characters
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, _ := transform.String(t, str)

	// Define the regular expression to match special characters
	re := regexp.MustCompile(`[&^*#@]`)

	// Replace special characters with an empty string
	result = removeEmojis(re.ReplaceAllString(result, ""))

	// Replace special characters with ASCII
    // * COPY/PASTE TO ADD MORE CHARACTERS:
    //   result = strings.ReplaceAll(result, "", "")
	result = strings.ReplaceAll(result, "‘", "'")
	result = strings.ReplaceAll(result, "’", "'")
	result = strings.ReplaceAll(result, "“", "\"")
	result = strings.ReplaceAll(result, "”", "\"")
	result = strings.ReplaceAll(result, "—", "-")
	result = strings.ReplaceAll(result, "–", "-")
	result = strings.ReplaceAll(result, "…", "...")
	result = strings.ReplaceAll(result, "\u00A0", " ")
	result = strings.ReplaceAll(result, "•", "*")
	result = strings.ReplaceAll(result, "¼", "1/4")
	result = strings.ReplaceAll(result, "½", "1/2")
	result = strings.ReplaceAll(result, "¾", "3/4")
	result = strings.ReplaceAll(result, "×", "x")
	result = strings.ReplaceAll(result, "÷", "/")
	result = strings.ReplaceAll(result, "ç", "c")
	result = strings.ReplaceAll(result, "©", "(c)")
	result = strings.ReplaceAll(result, "®", "(r)")
	result = strings.ReplaceAll(result, "™", "(tm)")
	result = strings.ReplaceAll(result, "@", "(a)")
	result = strings.ReplaceAll(result, " AI ", " A. I. ")	
	return result
}


func removeEmojis(input string) string {
	// a mess, but it works!
	re := regexp.MustCompile(`[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]|[\x{1F900}-\x{1F9FF}]|[\x{1F004}]|[\x{1F0CF}]|[\x{1F18E}]|[\x{1F191}-\x{1F251}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]|[\x{1F004}-\x{1F0CF}]|[\x{1F191}-\x{1F251}]|[\x{2B50}]`)
	result := re.ReplaceAllString(input, "")
	return result
}

func CreateAIReq(transcribedText, esn string, gpt3tryagain bool) openai.ChatCompletionRequest {
	defaultPrompt := "You are a helpful, animated robot called Vector. Keep the response concise yet informative."

	var nChat []openai.ChatCompletionMessage

	smsg := openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
	}
	if strings.TrimSpace(vars.APIConfig.Knowledge.OpenAIPrompt) != "" {
		smsg.Content = strings.TrimSpace(vars.APIConfig.Knowledge.OpenAIPrompt)
	} else {
		smsg.Content = defaultPrompt
	}

	var model string

	if gpt3tryagain {
		model = openai.GPT3Dot5Turbo
	} else if vars.APIConfig.Knowledge.Provider == "openai" {
		model = openai.GPT4o
		logger.Println("Using " + model)
	} else {
		logger.Println("Using " + vars.APIConfig.Knowledge.Model)
		model = vars.APIConfig.Knowledge.Model
	}

	smsg.Content = CreatePrompt(smsg.Content, model)

	nChat = append(nChat, smsg)
	if vars.APIConfig.Knowledge.SaveChat {
		rchat := GetChat(esn)
		logger.Println("Using remembered chats, length of " + fmt.Sprint(len(rchat.Chats)) + " messages")
		nChat = append(nChat, rchat.Chats...)
	}
	nChat = append(nChat, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: transcribedText,
	})

	aireq := openai.ChatCompletionRequest{
		Model:            model,
		MaxTokens:        2048,
		Temperature:      1,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		Messages:         nChat,
		Stream:           true,
	}
	return aireq
}

func StreamingKGSim(req interface{}, esn string, transcribedText string) (string, error) {
	matched := false
	var robot *vector.Vector
	var guid string
	var target string
	for _, bot := range vars.BotInfo.Robots {
		if esn == bot.Esn {
			guid = bot.GUID
			target = bot.IPAddress + ":443"
			matched = true
			break
		}
	}
	if matched {
		var err error
		robot, err = vector.New(vector.WithSerialNo(esn), vector.WithToken(guid), vector.WithTarget(target))
		if err != nil {
			return err.Error(), err
		}
	}
	_, err := robot.Conn.BatteryState(context.Background(), &vectorpb.BatteryStateRequest{})
	if err != nil {
		return "", err
	}
	var fullRespText string
	var fullfullRespText string
	var fullRespSlice []string
	var isDone bool
	var c *openai.Client
	if vars.APIConfig.Knowledge.Provider == "together" {
		if vars.APIConfig.Knowledge.Model == "" {
			vars.APIConfig.Knowledge.Model = "meta-llama/Llama-3-70b-chat-hf"
			vars.WriteConfigToDisk()
		}
		conf := openai.DefaultConfig(vars.APIConfig.Knowledge.Key)
		conf.BaseURL = "https://api.together.xyz/v1"
		c = openai.NewClientWithConfig(conf)
	} else if vars.APIConfig.Knowledge.Provider == "custom" {
		conf := openai.DefaultConfig(vars.APIConfig.Knowledge.Key)
		conf.BaseURL = vars.APIConfig.Knowledge.Endpoint
		c = openai.NewClientWithConfig(conf)
	} else if vars.APIConfig.Knowledge.Provider == "openai" {
		c = openai.NewClient(vars.APIConfig.Knowledge.Key)
	}
	ctx := context.Background()
	speakReady := make(chan string)
	successIntent := make(chan bool)

	aireq := CreateAIReq(transcribedText, esn, false)

	stream, err := c.CreateChatCompletionStream(ctx, aireq)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") && vars.APIConfig.Knowledge.Provider == "openai" {
			logger.Println("GPT-4 model cannot be accessed with this API key. You likely need to add more than $5 dollars of funds to your OpenAI account.")
			logger.LogUI("GPT-4 model cannot be accessed with this API key. You likely need to add more than $5 dollars of funds to your OpenAI account.")
			aireq := CreateAIReq(transcribedText, esn, true)
			logger.Println("Falling back to " + aireq.Model)
			logger.LogUI("Falling back to " + aireq.Model)
			stream, err = c.CreateChatCompletionStream(ctx, aireq)
			if err != nil {
				logger.Println("OpenAI still not returning a response even after falling back. Erroring.")
				return "", err
			}
		} else {
			return "", err
		}
	}
	nChat := aireq.Messages
	nChat = append(nChat, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleAssistant,
	})
	fmt.Println("LLM stream response: ")
	go func() {
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				// if fullRespSlice != fullRespText, add that missing bit to fullRespSlice
				if len(fullRespSlice) == 0 {
					logger.Println("LLM returned no response")
					successIntent <- false
					break
				}
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
					Remember(openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleUser,
						Content: transcribedText,
					},
						openai.ChatCompletionMessage{
							Role:    openai.ChatMessageRoleAssistant,
							Content: newStr,
						},
						esn)
				}
				logger.LogUI("LLM response for " + esn + ": " + newStr)
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
				case successIntent <- true:
				default:
				}
				select {
				case speakReady <- strings.TrimSpace(splitResp[0]) + sepStr:
				default:
				}
			}
		}
	}()
	for is := range successIntent {
		if is {
			IntentPass(req, "intent_greeting_hello", transcribedText, map[string]string{}, false)
			break
		} else {
			return "", errors.New("llm returned no response")
		}
	}
	time.Sleep(time.Millisecond * 200)
	controlRequest := &vectorpb.BehaviorControlRequest{
		RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
			ControlRequest: &vectorpb.ControlRequest{
				Priority: vectorpb.ControlRequest_OVERRIDE_BEHAVIORS,
			},
		},
	}
	start := make(chan bool)
	stop := make(chan bool)

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

	var stopTTSLoop bool
	TTSLoopStopped := make(chan bool)
	for range start {
		time.Sleep(time.Millisecond * 300)
		robot.Conn.PlayAnimation(
			ctx,
			&vectorpb.PlayAnimationRequest{
				Animation: &vectorpb.Animation{
					Name: "anim_getin_tts_01",
				},
				Loops: 1,
			},
		)
		if !vars.APIConfig.Knowledge.CommandsEnable {
			go func() {
				for {
					if stopTTSLoop {
						TTSLoopStopped <- true
						break
					}
					robot.Conn.PlayAnimation(
						ctx,
						&vectorpb.PlayAnimationRequest{
							Animation: &vectorpb.Animation{
								Name: "anim_tts_loop_02",
							},
							Loops: 1,
						},
					)
				}
			}()
		}
		var disconnect bool
		numInResp := 0
		for {
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
			nChat[len(nChat)-1].Content = fullRespText
			disconnect = PerformActions(nChat, acts, robot)
			if disconnect {
				break
			}
			numInResp = numInResp + 1
		}
		if !vars.APIConfig.Knowledge.CommandsEnable {
			stopTTSLoop = true
			for range TTSLoopStopped {
				break
			}
		}
		time.Sleep(time.Millisecond * 100)
		// robot.Conn.PlayAnimation(
		// 	ctx,
		// 	&vectorpb.PlayAnimationRequest{
		// 		Animation: &vectorpb.Animation{
		// 			Name: "anim_knowledgegraph_success_01",
		// 		},
		// 		Loops: 1,
		// 	},
		// )
		//time.Sleep(time.Millisecond * 3300)
		stop <- true
	}
	return "", nil
}

func KGSim(esn string, textToSay string) error {
	ctx := context.Background()
	matched := false
	var robot *vector.Vector
	var guid string
	var target string
	for _, bot := range vars.BotInfo.Robots {
		if esn == bot.Esn {
			guid = bot.GUID
			target = bot.IPAddress + ":443"
			matched = true
			break
		}
	}
	if matched {
		var err error
		robot, err = vector.New(vector.WithSerialNo(esn), vector.WithToken(guid), vector.WithTarget(target))
		if err != nil {
			return err
		}
	}
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

		var stopTTSLoop bool
		var TTSLoopStopped bool
		for range start {
			time.Sleep(time.Millisecond * 300)
			robot.Conn.PlayAnimation(
				ctx,
				&vectorpb.PlayAnimationRequest{
					Animation: &vectorpb.Animation{
						Name: "anim_getin_tts_01",
					},
					Loops: 1,
				},
			)
			go func() {
				for {
					if stopTTSLoop {
						TTSLoopStopped = true
						break
					}
					robot.Conn.PlayAnimation(
						ctx,
						&vectorpb.PlayAnimationRequest{
							Animation: &vectorpb.Animation{
								Name: "anim_tts_loop_02",
							},
							Loops: 1,
						},
					)
				}
			}()
			textToSaySplit := strings.Split(textToSay, ". ")
			for _, str := range textToSaySplit {
				_, err := robot.Conn.SayText(
					ctx,
					&vectorpb.SayTextRequest{
						Text:           str + ".",
						UseVectorVoice: true,
						DurationScalar: 1.0,
					},
				)
				if err != nil {
					logger.Println("KG SayText error: " + err.Error())
					stop <- true
					break
				}
			}
			stopTTSLoop = true
			for {
				if TTSLoopStopped {
					break
				} else {
					time.Sleep(time.Millisecond * 10)
				}
			}
			time.Sleep(time.Millisecond * 100)
			robot.Conn.PlayAnimation(
				ctx,
				&vectorpb.PlayAnimationRequest{
					Animation: &vectorpb.Animation{
						Name: "anim_knowledgegraph_success_01",
					},
					Loops: 1,
				},
			)
			//time.Sleep(time.Millisecond * 3300)
			stop <- true
		}
	}()
	return nil
}
