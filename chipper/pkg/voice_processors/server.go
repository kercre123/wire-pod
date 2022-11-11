package wirepod

import (
	"github.com/digital-dream-labs/chipper/pkg/wirepod-vosk"
)

// Server stores the config
type Server struct{}

var matchListList [][]string

var intentsList = []string{}

func sttHandler(reqThing interface{}, isKnowledgeGraph bool) (transcribedString string, slots map[string]string, isRhino bool, thisBotNum int, opusUsed bool, err error) {
	return wirepod_vosk.SttHandlerVosk(reqThing, isKnowledgeGraph)
}

// New returns a new server
func New() (*Server, error) {
	wirepod_vosk.VoskNew()
	matchListList = wirepod_vosk.MatchListList
	intentsList = wirepod_vosk.IntentsList
	return &Server{}, nil
}
