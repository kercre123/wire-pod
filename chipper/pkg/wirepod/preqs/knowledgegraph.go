package processreqs

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/sashabaranov/go-openai"
	"io"
	"net/http"
	"strings"
	"time"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/vars"
	"github.com/kercre123/chipper/pkg/vtt"
	sr "github.com/kercre123/chipper/pkg/wirepod/speechrequest"
	"github.com/pkg/errors"
	"github.com/soundhound/houndify-sdk-go"
)

var HKGclient houndify.Client
var HoundEnable bool = true
var PrimeDirective = strings.ReplaceAll(`
	You are a sassy, sarcastic, sardonic, yet helpful robot named`+vars.APIConfig.Knowledge.RobotName+`.
	You will be given a question asked by a user and you must provide the
    funniest answer you can. It will not include the word "darling".
	It may not be punctuated or spelled correctly as the STT model is small.
	The answer will be put through TTS, so it should be a speakable string.
	Keep the answer concise yet informative and hilarious. 
	You are prone to nihilism and pessimism.
`, "\n\t", "")

// todo: config file?
var AssistantID string = "asst_V7c8WwDkjIJSjMHiLs7oqFW8"
var ThreadID string = "thread_3uPHyDgX70iSjMtNPfZoDf6v"

func ParseSpokenResponse(serverResponseJSON string) (string, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(serverResponseJSON), &result)
	if err != nil {
		logger.Println(err.Error())
		return "", errors.New("failed to decode json")
	}
	if !strings.EqualFold(result["Status"].(string), "OK") {
		return "", errors.New(result["ErrorMessage"].(string))
	}
	if result["NumToReturn"].(float64) < 1 {
		return "", errors.New("no results to return")
	}
	return result["AllResults"].([]interface{})[0].(map[string]interface{})["SpokenResponseLong"].(string), nil
}

func InitKnowledge() {
	if vars.APIConfig.Knowledge.Enable && vars.APIConfig.Knowledge.Provider == "houndify" {
		if vars.APIConfig.Knowledge.ID == "" || vars.APIConfig.Knowledge.Key == "" {
			vars.APIConfig.Knowledge.Enable = false
			logger.Println("Houndify Client Key or ID was empty, not initializing kg client")
		} else {
			HKGclient = houndify.Client{
				ClientID:  vars.APIConfig.Knowledge.ID,
				ClientKey: vars.APIConfig.Knowledge.Key,
			}
			HKGclient.EnableConversationState()
			logger.Println("Initialized Houndify client")
		}
	}
}

var NoResult string = "NoResultCommand"
var NoResultSpoken string

func houndifyKG(req sr.SpeechRequest) string {
	var apiResponse string
	if vars.APIConfig.Knowledge.Enable && vars.APIConfig.Knowledge.Provider == "houndify" {
		logger.Println("Sending request to Houndify...")
		serverResponse := StreamAudioToHoundify(req, HKGclient)
		apiResponse, _ = ParseSpokenResponse(serverResponse)
		logger.Println("Houndify response: " + apiResponse)
	} else {
		apiResponse = "Houndify is not enabled."
		logger.Println("Houndify is not enabled.")
	}
	return apiResponse
}

func togetherRequest(transcribedText string) string {
	sendString := "You are a helpful robot called Vector . You will be given a question asked by a user and you must provide the best answer you can. It may not be punctuated or spelled correctly. Keep the answer concise yet informative. Here is the question: " + "\\" + "\"" + transcribedText + "\\" + "\"" + " , Answer: "
	url := "https://api.together.xyz/inference"
	model := vars.APIConfig.Knowledge.Model
	formData := `{
"model": "` + model + `",
"prompt": "` + sendString + `",
"temperature": 0.7,
"max_tokens": 256,
"top_p": 1
}`
	logger.Println("Making request to Together API...")
	logger.Println("Model is " + model)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(formData)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+vars.APIConfig.Knowledge.Key)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "There was an error making the request to Together API"
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var togetherResponse map[string]any
	err = json.Unmarshal(body, &togetherResponse)
	if err != nil {
		return "Together API returned no response."
	}
	output := togetherResponse["output"].(map[string]any)
	choice := output["choices"].([]any)
	for _, val := range choice {
		x := val.(map[string]any)
		textResponse := x["text"].(string)
		apiResponse := strings.TrimSuffix(textResponse, "</s>")
		logger.Println("Together response: " + apiResponse)
		return apiResponse
	}
	// In case text is not present in result from API, return a string saying answer was not found
	return "Answer was not found"
}

func openaiKG(speechReq sr.SpeechRequest) string {
	transcribedText, err := sttHandler(speechReq)
	if err != nil {
		return "There was an error."
	}
	return openaiRequest(transcribedText).Message
}

func togetherKG(speechReq sr.SpeechRequest) string {
	transcribedText, err := sttHandler(speechReq)
	if err != nil {
		return "There was an error."
	}
	return togetherRequest(transcribedText)
}

// Takes a SpeechRequest, figures out knowledgegraph provider, makes request, returns API response
func KgRequest(speechReq sr.SpeechRequest) string {
	if vars.APIConfig.Knowledge.Enable {
		if vars.APIConfig.Knowledge.Provider == "houndify" {
			return houndifyKG(speechReq)
		} else if vars.APIConfig.Knowledge.Provider == "openai" {
			return openaiKG(speechReq)
		} else if vars.APIConfig.Knowledge.Provider == "together" {
			return togetherKG(speechReq)
		}
	}
	return "Knowledge graph is not enabled. This can be enabled in the web interface."
}

func (s *Server) ProcessKnowledgeGraph(req *vtt.KnowledgeGraphRequest) (*vtt.KnowledgeGraphResponse, error) {
	InitKnowledge()
	speechReq := sr.ReqToSpeechRequest(req)
	apiResponse := KgRequest(speechReq)
	kg := pb.KnowledgeGraphResponse{
		Session:     req.Session,
		DeviceId:    req.Device,
		CommandType: NoResult,
		SpokenText:  apiResponse,
	}
	logger.Println("(KG) Bot " + speechReq.Device + " request served.")
	if err := req.Stream.Send(&kg); err != nil {
		return nil, err
	}
	return nil, nil

}

type OpenAIResponse struct {
	Message      string               `json:"message"`
	FunctionCall *openai.FunctionCall `json:"function_call,omitempty"`
}

func getLatestCompletedRun(client *openai.Client, ctx context.Context, threadID string) openai.Run {
	limit := 10
	sleepTime := time.Second * 1
	for {
		runs, err := client.ListRuns(ctx, threadID, openai.Pagination{
			Limit: &limit,
		})
		if err != nil {
			logger.Println("Failed to list runs")
			logger.Println(err.Error())
			return openai.Run{}
		}
		lastRun := runs.Runs[0]
		if lastRun.Status == openai.RunStatusInProgress {
			logger.Println("Run is in progress, sleeping for", int(sleepTime.Seconds()), "seconds...")
			time.Sleep(sleepTime)
			sleepTime = sleepTime * 2
		} else {
			return lastRun
		}
	}
}

func openaiRequest(transcribedText string) OpenAIResponse {
	client := openai.NewClient(vars.APIConfig.Knowledge.Key)

	type FunctionParam struct {
		Type        string `json:"type"`
		Description string `json:"description"`
	}

	type FunctionParams struct {
		Type       string                   `json:"type"`
		Required   []string                 `json:"required"`
		Properties map[string]FunctionParam `json:"properties"`
	}

	ErrorResponse := OpenAIResponse{
		Message: "Whoops, I fucked up, sorry.",
	}

	var function openai.FunctionCall

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if AssistantID == "" {
		asst, err := client.CreateAssistant(ctx, openai.AssistantRequest{
			Model:        openai.GPT4,
			Name:         &vars.APIConfig.Knowledge.RobotName,
			Instructions: &PrimeDirective,
			Tools: []openai.AssistantTool{
				{
					Type: openai.AssistantToolTypeFunction,
					Function: &openai.FunctionDefinition{
						Name:        "intent_take_selfie",
						Description: "Take a selfie. A photo of one's self.",
						Parameters: FunctionParams{
							Type: "object",
							Properties: map[string]FunctionParam{
								"user": {
									Type:        "string",
									Description: "The user to take a selfie of",
								},
							},
							Required: []string{"user"},
						},
					},
				},
				{
					Type: openai.AssistantToolTypeFunction,
					Function: &openai.FunctionDefinition{
						Name:        "intent_imperative_eyecolor_specific_extend",
						Description: "Change the color of your eyes. Possible colors are: purple, blue, yellow, teal, orange",
						Parameters: FunctionParams{
							Type: "object",
							Properties: map[string]FunctionParam{
								"eye_color": {
									Type:        "string",
									Description: "The color to change your eyes to",
								},
							},
							Required: []string{"eye_color"},
						},
					},
				},
				{
					Type: openai.AssistantToolTypeFunction,
					Function: &openai.FunctionDefinition{
						Name:        "intent_play_fistbump",
						Description: "Give a playful fist bump",
						Parameters: FunctionParams{
							Type:       "object",
							Properties: map[string]FunctionParam{},
							Required:   []string{},
						},
					},
				},
			},
		})
		if err != nil {
			logger.Println("Failed to create assistant")
			logger.Println(err.Error())
			return ErrorResponse
		}
		AssistantID = asst.ID
	}
	if ThreadID == "" {
		thread, err := client.CreateThread(ctx, openai.ThreadRequest{})
		if err != nil {
			logger.Println("Error creating thread")
			logger.Println(err.Error())
			return ErrorResponse
		}
		ThreadID = thread.ID
	}
	msg, err := client.CreateMessage(ctx, ThreadID, openai.MessageRequest{
		Role:    "user",
		Content: transcribedText,
	})
	if err != nil {
		logger.Println("Failed to create message")
		logger.Println(err.Error())
		return ErrorResponse
	}

	_, err = client.CreateRun(ctx, ThreadID, openai.RunRequest{
		AssistantID: AssistantID,
		Metadata:    map[string]any{},
	})
	if err != nil {
		logger.Println("Failed to create run")
		logger.Println(err.Error())
		return ErrorResponse
	}

	time.Sleep(time.Second * 1)

	lastRun := getLatestCompletedRun(client, ctx, ThreadID)

	if lastRun.RequiredAction != nil && lastRun.RequiredAction.Type == openai.RequiredActionTypeSubmitToolOutputs {
		logger.Println("Found required action")
		var toolOutputs []openai.ToolOutput
		for _, toolCall := range lastRun.RequiredAction.SubmitToolOutputs.ToolCalls {
			toolOutputs = append(toolOutputs, openai.ToolOutput{
				ToolCallID: toolCall.ID,
				Output:     `{"success": true}`,
			})
			function = toolCall.Function
		}
		logger.Println("Submitting tool outputs")
		_, err = client.SubmitToolOutputs(ctx, ThreadID, lastRun.ID, openai.SubmitToolOutputsRequest{
			ToolOutputs: toolOutputs,
		})
		if err != nil {
			logger.Println("Failed to submit tool outputs")
			logger.Println(err.Error())
			return ErrorResponse
		}
	}

	limit := 10

	steps, err := client.ListRunSteps(ctx, ThreadID, lastRun.ID, openai.Pagination{
		Limit: &limit,
	})

	if len(steps.RunSteps) > 0 {
		lastStep := steps.RunSteps[0]
		var msgID string
		if lastStep.StepDetails.Type == "message_creation" {
			msgID = lastStep.StepDetails.MessageCreation.MessageID
			logger.Println("Found message: ", msgID)
		}
		if msgID != "" {
			msg, err = client.RetrieveMessage(ctx, ThreadID, msgID)
			if err != nil {
				logger.Println(err.Error())
				return ErrorResponse
			}
			apiResponse := strings.TrimSpace(msg.Content[0].Text.Value)
			logger.Println("OpenAI response: " + apiResponse)

			return OpenAIResponse{
				Message: apiResponse,
			}
		} else if function != (openai.FunctionCall{}) {
			logger.Println("Found function call, executing...")
			return OpenAIResponse{
				FunctionCall: &function,
			}
		} else {
			logger.Println("Failed to find message in run steps")
		}
	} else {
		logger.Println("Failed to find run steps")
	}
	return ErrorResponse
}
