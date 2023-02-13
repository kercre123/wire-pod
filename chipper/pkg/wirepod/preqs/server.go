package processreqs

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/vars"
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

var sttInitFunc func() error

// speech-to-text
var sttHandler func(sr.SpeechRequest) (string, error)

// speech-to-intent (rhino)
var stiHandler func(sr.SpeechRequest) (string, map[string]string, error)

var isSti bool = false

var matchListList [][]string
var intentsList = []string{}

func ReloadVosk() {
	if vars.APIConfig.STT.Service == "vosk" {
		sttInitFunc()
		matchListList, intentsList, _ = loadIntents()
	}
}

func loadIntents() ([][]string, []string, error) {
	sttLanguage = vars.APIConfig.STT.Language
	jsonFile, err := os.ReadFile("./intent-data/" + vars.APIConfig.STT.Language + ".json")

	var matches [][]string
	var intents []string

	if err == nil {
		var jsonIntents []JsonIntent
		err = json.Unmarshal(jsonFile, &jsonIntents)
		if err != nil {
			logger.Println("Failed to load intents: " + err.Error())
		}

		for _, element := range jsonIntents {
			//logger.Println("Loading intent " + strconv.Itoa(index) + " --> " + element.Name + "( " + strconv.Itoa(len(element.Keyphrases)) + " keyphrases )")
			intents = append(intents, element.Name)
			matches = append(matches, element.Keyphrases)
		}
		logger.Println("Loaded " + strconv.Itoa(len(jsonIntents)) + " intents and " + strconv.Itoa(len(matches)) + " matches (language: " + vars.APIConfig.STT.Language + ")")
	}
	return matches, intents, err
}

// New returns a new server
func New(InitFunc func() error, SttHandler interface{}, voiceProcessor string) (*Server, error) {

	// Decide the TTS language
	sttLanguage = vars.APIConfig.STT.Language
	logger.Println("Initiating " + voiceProcessor + " voice processor with language " + sttLanguage)
	err := InitFunc()
	if err != nil {
		return nil, err
	}
	sttInitFunc = InitFunc

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
	matchListList, intentsList, err = loadIntents()

	// Load plugins
	ttr.LoadPlugins()

	return &Server{}, err
}
