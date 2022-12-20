package wirepod

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
	"github.com/digital-dream-labs/chipper/pkg/vtt"
	"github.com/maxhawkins/go-webrtcvad"
	"github.com/pkg/errors"
	"github.com/soundhound/houndify-sdk-go"
)

var HKGclient houndify.Client
var HoundEnable bool = true

func ParseSpokenResponse(serverResponseJSON string) (string, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(serverResponseJSON), &result)
	if err != nil {
		logger(err.Error())
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
			logger("Houndify Client ID not found.")
			HoundEnable = false
			return
		}
		if os.Getenv("HOUNDIFY_CLIENT_KEY") == "" {
			logger("Houndify Client Key not found.")
			HoundEnable = false
			return
		}
		if HoundEnable {
			HKGclient = houndify.Client{
				ClientID:  os.Getenv("HOUNDIFY_CLIENT_ID"),
				ClientKey: os.Getenv("HOUNDIFY_CLIENT_KEY"),
			}
			HKGclient.EnableConversationState()
			logger("Houndify for knowledge graph initialized!")
		}
	} else {
		if os.Getenv("KNOWLEDGE_PROVIDER") == "houndify" {
			if os.Getenv("KNOWLEDGE_ID") == "" {
				logger("Houndify Client ID not found.")
				HoundEnable = false
				return
			}
			if os.Getenv("KNOWLEDGE_KEY") == "" {
				logger("Houndify Client Key not found.")
				HoundEnable = false
				return
			}
			if HoundEnable {
				HKGclient = houndify.Client{
					ClientID:  os.Getenv("KNOWLEDGE_ID"),
					ClientKey: os.Getenv("KNOWLEDGE_KEY"),
				}
				HKGclient.EnableConversationState()
				logger("Houndify for knowledge graph initialized!")
			}
		} else {
			logger("Knowledge provider: " + os.Getenv("KNOWLEDGE_PROVIDER"))
		}
	}
}

var NoResult string = "NoResultCommand"
var NoResultSpoken string

func kgHoundifyRequestHandler(req SpeechRequest) (string, error) {
	var transcribedText string
	if HoundEnable {
		logger("Sending requst to Houndify...")
		if os.Getenv("HOUNDIFY_CLIENT_KEY") != "" {
			req := houndify.VoiceRequest{
				AudioStream:       bytes.NewReader(req.MicData),
				UserID:            req.Device,
				RequestID:         req.Session,
				RequestInfoFields: make(map[string]interface{}),
			}
			partialTranscripts := make(chan houndify.PartialTranscript)
			serverResponse, err := HKGclient.VoiceSearch(req, partialTranscripts)
			if err != nil {
				logger(err)
			}
			transcribedText, _ = ParseSpokenResponse(serverResponse)
			logger("Transcribed text: " + transcribedText)
		}
	} else {
		transcribedText = "Houndify is not enabled."
		logger("Houndify is not enabled.")
	}
	return transcribedText, nil
}

func kgVADHandler(req SpeechRequest) (SpeechRequest, error) {
	logger("Processing...")
	activeNum := 0
	inactiveNum := 0
	inactiveNumMax := 20
	die := false
	vad, err := webrtcvad.New()
	if err != nil {
		logger(err)
	}
	vad.SetMode(3)
	for {
		var chunk []byte
		req, chunk, err = getNextStreamChunk(req)
		if err != nil {
			return req, err
		}
		splitChunk := splitVAD(chunk)
		for _, chunk := range splitChunk {
			active, err := vad.Process(16000, chunk)
			if err != nil {
				logger(err)
			}
			if active {
				activeNum = activeNum + 1
				inactiveNum = 0
			} else {
				inactiveNum = inactiveNum + 1
			}
			if inactiveNum >= inactiveNumMax && activeNum > 20 {
				logger("Speech completed.")
				die = true
				break
			}
		}
		if die {
			break
		}
	}
	return req, nil
}

func houndifyKG(req SpeechRequest) (string, error) {
	req, _ = kgVADHandler(req)
	apiResponse, err := kgHoundifyRequestHandler(req)
	return apiResponse, err
}

func openaiRequest(transcribedText string) (string, error) {
	sendString := "You are a helpful robot called the Anki Vector. You will be given a question asked by a user and you must provide the best answer you can. It may not be punctuated or spelled correctly. Keep the answer concise yet informative. Here is the question: " + "\\" + "\"" + transcribedText + "\\" + "\""
	logger("Making request to OpenAI...")
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
	logger("OpenAI response: " + apiResponse)
	return apiResponse, nil
}

func openaiKG(speechReq SpeechRequest) (string, error) {
	transcribedText, err := sttHandler(speechReq)
	if err != nil {
		return "", nil
	}
	return openaiRequest(transcribedText)
}

func (s *Server) ProcessKnowledgeGraph(req *vtt.KnowledgeGraphRequest) (*vtt.KnowledgeGraphResponse, error) {
	botNum = botNum + 1
	var apiResponse string
	var err error
	speechReq := reqToSpeechRequest(req)
	if os.Getenv("KNOWLEDGE_PROVIDER") == "houndify" || os.Getenv("HOUNDIFY_CLIENT_ID") != "" {
		apiResponse, err = houndifyKG(speechReq)
	} else if os.Getenv("KNOWLEDGE_PROVIDER") == "openai" {
		apiResponse, err = openaiKG(speechReq)
	}
	if err != nil {
		logger(err)
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
	botNum = botNum - 1
	logger("(KG) Bot " + strconv.Itoa(speechReq.BotNum) + " request served.")
	if err := req.Stream.Send(&kg); err != nil {
		return nil, err
	}
	return &vtt.KnowledgeGraphResponse{
		Intent: &kg,
	}, nil

}
