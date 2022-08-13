package wirepod

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/chipper/pkg/vtt"
	"github.com/pkg/errors"
	"github.com/soundhound/houndify-sdk-go"
)

var hKGclient houndify.Client
var houndEnable bool = true

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
		houndEnable = false
	}
	if os.Getenv("HOUNDIFY_CLIENT_KEY") == "" {
		logger("Houndify Client Key not provided.")
		houndEnable = false
	}
	if houndEnable {
		hKGclient = houndify.Client{
			ClientID:  os.Getenv("HOUNDIFY_CLIENT_ID"),
			ClientKey: os.Getenv("HOUNDIFY_CLIENT_KEY"),
		}
		hKGclient.EnableConversationState()
		logger("Houndify for knowledge graph initialized!")
	}
}

var NoResult string = "NoResultCommand"
var NoResultSpoken string

func (s *Server) ProcessKnowledgeGraph(req *vtt.KnowledgeGraphRequest) (*vtt.KnowledgeGraphResponse, error) {
	transcribedText, _, _, justThisBotNum, _, err := sttHandler(req, true)
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
	NoResultSpoken = transcribedText
	kg := pb.KnowledgeGraphResponse{
		Session:     req.Session,
		DeviceId:    req.Device,
		CommandType: NoResult,
		SpokenText:  NoResultSpoken,
	}
	logger("(KG) Bot " + strconv.Itoa(justThisBotNum) + " request served.")
	if err := req.Stream.Send(&kg); err != nil {
		return nil, err
	}
	return &vtt.KnowledgeGraphResponse{
		Intent: &kg,
	}, nil

}
