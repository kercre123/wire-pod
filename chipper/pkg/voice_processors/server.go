package wirepod

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/digital-dream-labs/opus-go/opus"
)

// where a new stt service would be added

const (
	VoiceProcessorCoqui   = "coqui"
	VoiceProcessorLeopard = "leopard"
	VoiceProcessorVosk    = "vosk"
)

func initSTT(voiceProcessor string) {
	switch {
	case voiceProcessor == VoiceProcessorCoqui:
		CoquiInit()
	case voiceProcessor == VoiceProcessorLeopard:
		LeopardInit()
	case voiceProcessor == VoiceProcessorVosk:
		VoskInit()
	}
}

func sttHandler(req SpeechRequest) (transcribedText string, err error) {
	switch {
	case VoiceProcessor == VoiceProcessorCoqui:
		return CoquiSttHandler(req)
	case VoiceProcessor == VoiceProcessorLeopard:
		return LeopardSttHandler(req)
	case VoiceProcessor == VoiceProcessorVosk:
		return VoskSTTHandler(req)
	}
	return "", errors.New("invalid stt service")
}

// Server stores the config
type Server struct{}

var VoiceProcessor = ""

type JsonIntent struct {
	Name       string   `json:"name"`
	Keyphrases []string `json:"keyphrases"`
}

var sttLanguage string = "en-US"

var matchListList [][]string
var intentsList = []string{}

var botNum int = 0

func splitVAD(buf []byte) [][]byte {
	var chunk [][]byte
	for len(buf) >= 320 {
		chunk = append(chunk, buf[:320])
		buf = buf[320:]
	}
	return chunk
}

func bytesToIntVAD(stream opus.OggStream, data []byte, die bool, isOpus bool) [][]byte {
	// detect if data is pcm or opus
	if die {
		return nil
	}
	if isOpus {
		// opus
		n, err := stream.Decode(data)
		if err != nil {
			logger(err)
		}
		byteArray := splitVAD(n)
		return byteArray
	} else {
		// pcm
		byteArray := splitVAD(data)
		return byteArray
	}
}

func bytesToSamples(buf []byte) []int16 {
	samples := make([]int16, len(buf)/2)
	for i := 0; i < len(buf)/2; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(buf[i*2:]))
	}
	return samples
}

func loadIntents(language string) ([][]string, []string, error) {
	jsonFile, err := os.ReadFile("./intent-data/" + language + ".json")

	var matches [][]string
	var intents []string

	if err == nil {
		var jsonIntents []JsonIntent
		json.Unmarshal(jsonFile, &jsonIntents)

		for index, element := range jsonIntents {
			logger("Loading intent " + strconv.Itoa(index) + " --> " + element.Name + "( " + strconv.Itoa(len(element.Keyphrases)) + " keyphrases )")
			intents = append(intents, element.Name)
			matches = append(matches, element.Keyphrases)
		}
	}
	return matches, intents, err
}

// New returns a new server
func New(voiceProcessor string) (*Server, error) {
	// Setup logging
	if os.Getenv("DEBUG_LOGGING") != "true" && os.Getenv("DEBUG_LOGGING") != "false" {
		logger("No valid value for DEBUG_LOGGING, setting to true")
		debugLogging = true
	} else {
		if os.Getenv("DEBUG_LOGGING") == "true" {
			debugLogging = true
		} else {
			debugLogging = false
		}
	}

	// Decide the TTS language
	sttLanguage = os.Getenv("STT_LANGUAGE")
	if len(sttLanguage) == 0 {
		sttLanguage = "en-US"
	}
	logger("Initiating " + voiceProcessor + " voice processor with language " + sttLanguage)
	initSTT(voiceProcessor)

	// Initiating the chosen voice processor and load intents from json
	VoiceProcessor = voiceProcessor
	var err error
	matchListList, intentsList, err = loadIntents(sttLanguage)

	return &Server{}, err
}

var debugLogging bool

func logger(a ...any) {
	if debugLogging {
		fmt.Println(a...)
	}
}
