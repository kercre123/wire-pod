package processreqs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/vtt"
	sr "github.com/kercre123/chipper/pkg/wirepod/speechrequest"
	"github.com/pkg/errors"
	"github.com/soundhound/houndify-sdk-go"
)

var HKGclient houndify.Client
var HoundEnable bool = true

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

func InitHoundify() {
	if os.Getenv("KNOWLEDGE_ENABLED") == "" {
		// initing with old source.sh
		if os.Getenv("HOUNDIFY_CLIENT_ID") == "" {
			logger.Println("Houndify Client ID not found.")
			HoundEnable = false
			return
		}
		if os.Getenv("HOUNDIFY_CLIENT_KEY") == "" {
			logger.Println("Houndify Client Key not found.")
			HoundEnable = false
			return
		}
		if HoundEnable {
			HKGclient = houndify.Client{
				ClientID:  os.Getenv("HOUNDIFY_CLIENT_ID"),
				ClientKey: os.Getenv("HOUNDIFY_CLIENT_KEY"),
			}
			HKGclient.EnableConversationState()
			logger.Println("Houndify for knowledge graph initialized!")
		}
	} else {
		if os.Getenv("KNOWLEDGE_PROVIDER") == "houndify" {
			if os.Getenv("KNOWLEDGE_ID") == "" {
				logger.Println("Houndify Client ID not found.")
				HoundEnable = false
				return
			}
			if os.Getenv("KNOWLEDGE_KEY") == "" {
				logger.Println("Houndify Client Key not found.")
				HoundEnable = false
				return
			}
			if HoundEnable {
				HKGclient = houndify.Client{
					ClientID:  os.Getenv("KNOWLEDGE_ID"),
					ClientKey: os.Getenv("KNOWLEDGE_KEY"),
				}
				HKGclient.EnableConversationState()
				logger.Println("Houndify for knowledge graph initialized!")
			}
		} else {
			logger.Println("Knowledge provider: " + os.Getenv("KNOWLEDGE_PROVIDER"))
		}
	}
}

var NoResult string = "NoResultCommand"
var NoResultSpoken string

func houndifyKG(req sr.SpeechRequest) (string, error) {
	var apiResponse string
	if HoundEnable {
		logger.Println("Sending request to Houndify...")
		if os.Getenv("HOUNDIFY_CLIENT_KEY") != "" || os.Getenv("KNOWLEDGE_KEY") != "" {
			serverResponse := StreamAudioToHoundify(req, HKGclient)
			apiResponse, _ = ParseSpokenResponse(serverResponse)
			logger.Println("Houndify response: " + apiResponse)
		} else {
			apiResponse = "Houndify API Key missing."
			logger.Println("Houndify API Key missing.")
		}
	} else {
		apiResponse = "Houndify is not enabled."
		logger.Println("Houndify is not enabled.")
	}
	return apiResponse, nil
}

func openaiRequest(transcribedText string) (string, error) {
	sendString := "You are a helpful robot called the Anki Vector. You will be given a question asked by a user and you must provide the best answer you can. It may not be punctuated or spelled correctly. Keep the answer concise yet informative. Here is the question: " + "\\" + "\"" + transcribedText + "\\" + "\"" + " , Answer: "
	logger.Println("Making request to OpenAI...")
	url := "https://api.openai.com/v1/completions"
	formData := `{
"model": "text-davinci-003",
"prompt": "` + sendString + `",
"temperature": 1,
"max_tokens": 256,
"top_p": 1,
"frequency_penalty": 0.2,
"presence_penalty": 0
}`
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(formData)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("KNOWLEDGE_KEY"))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	type openAIStruct struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int    `json:"created"`
		Model   string `json:"model"`
		Choices []struct {
			Text         string      `json:"text"`
			Index        int         `json:"index"`
			Logprobs     interface{} `json:"logprobs"`
			FinishReason string      `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	var openAIResponse openAIStruct
	err = json.Unmarshal(body, &openAIResponse)
	if err != nil || len(openAIResponse.Choices) == 0 {
		fmt.Println("ERROR: " + string(body))
		fmt.Println("")
		return "", err
	}
	apiResponse := strings.TrimSpace(openAIResponse.Choices[0].Text)
	logger.Println("OpenAI response: " + apiResponse)
	return apiResponse, nil
}

func openaiKG(speechReq sr.SpeechRequest) (string, error) {
	transcribedText, err := sttHandler(speechReq)
	if err != nil {
		return "", nil
	}
	return openaiRequest(transcribedText)
}

// Takes a SpeechRequest, figures out knowledgegraph provider, makes request, returns API response
func KgRequest(speechReq sr.SpeechRequest) (string, error) {
	if os.Getenv("KNOWLEDGE_PROVIDER") == "houndify" {
		return houndifyKG(speechReq)
	} else if os.Getenv("KNOWLEDGE_PROVIDER") == "openai" {
		return openaiKG(speechReq)
	}
	return "", fmt.Errorf("invalid kg provider")
}

func (s *Server) ProcessKnowledgeGraph(req *vtt.KnowledgeGraphRequest) (*vtt.KnowledgeGraphResponse, error) {
	sr.BotNum = sr.BotNum + 1
	speechReq := sr.ReqToSpeechRequest(req)
	apiResponse, err := KgRequest(speechReq)
	if err != nil {
		logger.Println(err)
		NoResultSpoken = err.Error()
		kg := pb.KnowledgeGraphResponse{
			Session:     req.Session,
			DeviceId:    req.Device,
			CommandType: NoResult,
			SpokenText:  NoResultSpoken,
		}
		if err := req.Stream.Send(&kg); err != nil {
			return nil, err
		}
		return &vtt.KnowledgeGraphResponse{
			Intent: &kg,
		}, nil
	}
	NoResultSpoken = apiResponse
	kg := pb.KnowledgeGraphResponse{
		Session:     req.Session,
		DeviceId:    req.Device,
		CommandType: NoResult,
		SpokenText:  NoResultSpoken,
	}
	sr.BotNum = sr.BotNum - 1
	logger.Println("(KG) Bot " + strconv.Itoa(speechReq.BotNum) + " request served.")
	if err := req.Stream.Send(&kg); err != nil {
		return nil, err
	}
	return &vtt.KnowledgeGraphResponse{
		Intent: &kg,
	}, nil

}
