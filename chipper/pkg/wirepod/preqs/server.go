package processreqs

import (
	"errors"
	"fmt"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	sr "github.com/kercre123/wire-pod/chipper/pkg/wirepod/speechrequest"
	wyomingstt "github.com/kercre123/wire-pod/chipper/pkg/wirepod/stt/wyoming"
	ttr "github.com/kercre123/wire-pod/chipper/pkg/wirepod/ttr"
)

// Server stores the config
type Server struct{}

var VoiceProcessor = ""

type JsonIntent struct {
	Name              string   `json:"name"`
	Keyphrases        []string `json:"keyphrases"`
	RequireExactMatch bool     `json:"requiresexact"`
}

var sttLanguage string = "en-US"

// speech-to-text
var sttHandler func(sr.SpeechRequest) (string, error)

// speech-to-intent (rhino)
var stiHandler func(sr.SpeechRequest) (string, map[string]string, error)

var isSti bool = false

func ReloadVosk() {
	if vars.APIConfig.STT.Service == "vosk" || vars.APIConfig.STT.Service == "whisper.cpp" {
		vars.SttInitFunc()
		vars.IntentList, _ = vars.LoadIntents()
	}
}

// New returns a new server
func New(InitFunc func() error, SttHandler interface{}, voiceProcessor string) (*Server, error) {

	// Decide the TTS language
	if voiceProcessor != "vosk" && voiceProcessor != "whisper.cpp" && voiceProcessor != "whisper" {
		vars.APIConfig.STT.Language = "en-US"
	}
	sttLanguage = vars.APIConfig.STT.Language
	vars.IntentList, _ = vars.LoadIntents()
	logger.Println("Initiating " + voiceProcessor + " voice processor with language " + sttLanguage)
	vars.SttInitFunc = InitFunc
	err := InitFunc()
	if err != nil {
		return nil, err
	}

	// SttHandler can either be `func(sr.SpeechRequest) (string, error)` or `func (sr.SpeechRequest) (string, map[string]string, error)`
	// second one exists to accomodate Rhino

	// check function type
	if str, is := SttHandler.(func(sr.SpeechRequest) (string, error)); is {
		sttHandler = withWyomingSTT(str)
	} else if str, is := SttHandler.(func(sr.SpeechRequest) (string, map[string]string, error)); is {
		stiHandler = str
		isSti = true
	} else {
		return nil, fmt.Errorf("stthandler not of correct type")
	}

	// Initiating the chosen voice processor and load intents from json
	VoiceProcessor = voiceProcessor

	// Load plugins
	ttr.LoadPlugins()

	return &Server{}, err
}

// withWyomingSTT routes speech to the external Wyoming STT server when it is
// enabled in the web UI (checked on every request, no restart needed). If the
// server cannot be reached, the compiled-in engine handles the request.
func withWyomingSTT(builtin func(sr.SpeechRequest) (string, error)) func(sr.SpeechRequest) (string, error) {
	return func(req sr.SpeechRequest) (string, error) {
		if !wyomingstt.Enabled() {
			return builtin(req)
		}
		text, err := wyomingstt.STT(req)
		if errors.Is(err, wyomingstt.ErrUnavailable) {
			logger.Println("Wyoming STT unavailable, falling back to " + VoiceProcessor)
			return builtin(req)
		}
		return text, err
	}
}
