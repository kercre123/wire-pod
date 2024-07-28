package processreqs

import (
	"encoding/json"
	"strings"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/kercre123/wire-pod/chipper/pkg/vtt"
	sr "github.com/kercre123/wire-pod/chipper/pkg/wirepod/speechrequest"
	ttr "github.com/kercre123/wire-pod/chipper/pkg/wirepod/ttr"
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

func streamingKG(req *vtt.KnowledgeGraphRequest, speechReq sr.SpeechRequest) string {
	// have him start "thinking" right after the text is transcribed
	transcribedText, err := sttHandler(speechReq)
	if err != nil {
		return "There was an error."
	}
	kg := pb.KnowledgeGraphResponse{
		Session:     req.Session,
		DeviceId:    req.Device,
		CommandType: NoResult,
		SpokenText:  "bla bla bla bla bla bla bla bla bla bla",
	}
	req.Stream.Send(&kg)
	_, err = ttr.StreamingKGSim(req, req.Device, transcribedText, true)
	if err != nil {
		logger.Println("LLM error: " + err.Error())
	}
	logger.Println("(KG) Bot " + speechReq.Device + " request served.")
	return ""
}

// Takes a SpeechRequest, figures out knowledgegraph provider, makes request, returns API response
func KgRequest(req *vtt.KnowledgeGraphRequest, speechReq sr.SpeechRequest) string {
	if vars.APIConfig.Knowledge.Enable {
		if vars.APIConfig.Knowledge.Provider == "houndify" {
			return houndifyKG(speechReq)
		}
	}
	return "Knowledge graph is not enabled. This can be enabled in the web interface."
}

func (s *Server) ProcessKnowledgeGraph(req *vtt.KnowledgeGraphRequest) (*vtt.KnowledgeGraphResponse, error) {
	InitKnowledge()
	speechReq := sr.ReqToSpeechRequest(req)
	if vars.APIConfig.Knowledge.Enable && vars.APIConfig.Knowledge.Provider != "houndify" {
		streamingKG(req, speechReq)
	} else {
		apiResponse := KgRequest(req, speechReq)
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
	}
	return nil, nil

}
