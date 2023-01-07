package processreqs

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/kercre123/chipper/pkg/logger"
	sr "github.com/kercre123/chipper/pkg/wirepod/speechrequest"
	ttr "github.com/kercre123/chipper/pkg/wirepod/ttr"
)

// Server stores the config
type Server struct{}

var VoiceProcessor = ""

type JsonIntent struct {
	Name       string   `json:"name"`
	Keyphrases []string `json:"keyphrases"`
}

var sttLanguage string = "en-US"

// speech-to-text
var sttHandler func(sr.SpeechRequest) (string, error)

// speech-to-intent (rhino)
var stiHandler func(sr.SpeechRequest) (string, map[string]string, error)

var isSti bool = false

var matchListList [][]string
var intentsList = []string{}

func loadIntents(language string) ([][]string, []string, error) {
	jsonFile, err := os.ReadFile("./intent-data/" + language + ".json")

	var matches [][]string
	var intents []string

	if err == nil {
		var jsonIntents []JsonIntent
		json.Unmarshal(jsonFile, &jsonIntents)

		for index, element := range jsonIntents {
			logger.Println("Loading intent " + strconv.Itoa(index) + " --> " + element.Name + "( " + strconv.Itoa(len(element.Keyphrases)) + " keyphrases )")
			intents = append(intents, element.Name)
			matches = append(matches, element.Keyphrases)
		}
	}
	return matches, intents, err
}

// New returns a new server
func New(InitFunc func() error, SttHandler interface{}, voiceProcessor string) (*Server, error) {

	// Decide the TTS language
	sr.InitLanguage()
	sttLanguage = sr.SttLanguage
	logger.Println("Initiating " + voiceProcessor + " voice processor with language " + sttLanguage)
	err := InitFunc()
	if err != nil {
		return nil, err
	}

	// SttHandler can either be `func(sr.SpeechRequest) (string, error)` or `func (sr.SpeechRequest) (string, map[string]string, error)`
	// second one exists to accomodate Rhino

	// check function type
	if str, is := SttHandler.(func(sr.SpeechRequest) (string, error)); is {
		sttHandler = str
	} else if str, is := SttHandler.(func(sr.SpeechRequest) (string, map[string]string, error)); is {
		stiHandler = str
		isSti = true
	} else {
		return nil, fmt.Errorf("stthandler not of correct type")
	}

	// Initiating the chosen voice processor and load intents from json
	VoiceProcessor = voiceProcessor
	matchListList, intentsList, err = loadIntents(sttLanguage)

	// Load plugins
	ttr.LoadPlugins()

	return &Server{}, err
}
