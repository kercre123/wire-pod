package wirepod

import (
	vosk "github.com/digital-dream-labs/chipper/pkg/voice_processors/wirepod-vosk"
)

// Server stores the config
type Server struct{}

var matchListList [][]string

var intentsList = []string{}

func sttHandler(reqThing interface{}, isKnowledgeGraph bool) (transcribedString string, slots map[string]string, isRhino bool, thisBotNum int, opusUsed bool, err error) {
	return vosk.SttHandlerVosk(reqThing, isKnowledgeGraph)
}

// New returns a new server
func New() (*Server, error) {
	vosk.VoskNew()
	matchListList = vosk.MatchListList
	intentsList = vosk.IntentsList
	return &Server{}, nil
}
