package processreqs

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/digital-dream-labs/chipper/pkg/logger"
	sr "github.com/digital-dream-labs/chipper/pkg/speechrequest"
	ttr "github.com/digital-dream-labs/chipper/pkg/wirepod-ttr"
)

// Server stores the config
type Server struct{}

var VoiceProcessor = ""

type JsonIntent struct {
	Name       string   `json:"name"`
	Keyphrases []string `json:"keyphrases"`
}

var sttLanguage string = "en-US"
var sttHandler func(sr.SpeechRequest) (string, error)

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
func New(InitFunc func() error, SttHandler func(sr.SpeechRequest) (string, error), voiceProcessor string) (*Server, error) {
	// Setup logging
	if os.Getenv("DEBUG_LOGGING") != "true" && os.Getenv("DEBUG_LOGGING") != "false" {
		logger.Println("No valid value for DEBUG_LOGGING, setting to true")
		debugLogging = true
	} else {
		if os.Getenv("DEBUG_LOGGING") == "true" {
			debugLogging = true
		} else {
			debugLogging = false
		}
	}

	// Decide the TTS language
	sr.InitLanguage()
	sttLanguage = sr.SttLanguage
	logger.Println("Initiating " + voiceProcessor + " voice processor with language " + sttLanguage)
	err := InitFunc()
	if err != nil {
		return nil, err
	}
	sttHandler = SttHandler

	// Initiating the chosen voice processor and load intents from json
	VoiceProcessor = voiceProcessor
	matchListList, intentsList, err = loadIntents(sttLanguage)

	// Load plugins
	ttr.LoadPlugins()

	return &Server{}, err
}

var debugLogging bool
