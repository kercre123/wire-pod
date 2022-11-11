package wirepod

import (
	"errors"
	"fmt"
)

// Server stores the config
type Server struct{}

const (
	VoiceProcessorCoqui   = "coqui"
	VoiceProcessorLeopard = "leopard"
	VoiceProcessorVosk    = "vosk"
)

var matchListList [][]string

var intentsList = []string{}

func sttHandler(reqThing interface{}, isKnowledgeGraph bool) (transcribedString string, slots map[string]string, isRhino bool, thisBotNum int, opusUsed bool, err error) {
	return VOSKSttHandler(reqThing, isKnowledgeGraph)
}

// New returns a new server
func New(voiceProcessor string) (*Server, error) {
	if VoiceProcessorVosk == voiceProcessor {
		VOSKNew()
		matchListList = VOSKMatchListList
		intentsList = VOSKIntentsList
		return &Server{}, nil
	}

	return nil, errors.New("Unknown voice processor")
}

var debugLogging bool

func logger(a ...any) {
	if debugLogging {
		fmt.Println(a...)
	}
}
