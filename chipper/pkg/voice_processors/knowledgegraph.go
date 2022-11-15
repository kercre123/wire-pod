package wirepod

import (
	"bytes"
	"encoding/json"
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
	if os.Getenv("HOUNDIFY_CLIENT_ID") == "" {
		logger("Houndify Client ID not provided.")
		HoundEnable = false
	}
	if os.Getenv("HOUNDIFY_CLIENT_KEY") == "" {
		logger("Houndify Client Key not provided.")
		HoundEnable = false
	}
	if HoundEnable {
		HKGclient = houndify.Client{
			ClientID:  os.Getenv("HOUNDIFY_CLIENT_ID"),
			ClientKey: os.Getenv("HOUNDIFY_CLIENT_KEY"),
		}
		HKGclient.EnableConversationState()
		logger("Houndify for knowledge graph initialized!")
	}
}

var NoResult string = "NoResultCommand"
var NoResultSpoken string

func kgRequestHandler(req SpeechRequest) (string, error) {
	var transcribedText string
	if HoundEnable {
		logger("Sending requst to Houndify...")
		if os.Getenv("HOUNDIFY_CLIENT_KEY") != "" {
			req := houndify.VoiceRequest{
				AudioStream:       bytes.NewReader(req.DecodedMicData),
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

func (s *Server) ProcessKnowledgeGraph(req *vtt.KnowledgeGraphRequest) (*vtt.KnowledgeGraphResponse, error) {
	botNum = botNum + 1
	speechReq := reqToSpeechRequest(req)
	speechReq.BotNum = botNum
	var err error
	speechReq, err = kgVADHandler(speechReq)
	apiResponse, _ := kgRequestHandler(speechReq)
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
