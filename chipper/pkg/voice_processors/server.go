package wirepod

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/digital-dream-labs/opus-go/opus"
	"os"
	"strconv"
)

const (
	// FallbackIntent is the failure-mode intent response
	FallbackIntent          = "intent_system_unsupported"
	IntentWeather           = "intent_weather"
	IntentWeatherExtend     = "intent_weather_extend"
	IntentNoLocation        = "intent_weather_unknownlocation"
	IntentNoDefaultLocation = "intent_weather_nodefaultlocation"

	IntentClockSetTimer                    = "intent_clock_settimer"
	IntentClockSetTimerExtend              = "intent_clock_settimer_extend"
	IntentNamesUsername                    = "intent_names_username"
	IntentNamesUsernameExtend              = "intent_names_username_extend"
	IntentPlaySpecific                     = "intent_play_specific"
	IntentPlaySpecificExtend               = "intent_play_specific_extend"
	IntentMessaqePlayMessage               = "intent_message_playmessage"
	IntentMessagePlayMessageExtend         = "intent_message_playmessage_extend"
	IntentMessageRecordMessage             = "intent_message_recordmessage"
	IntentMessageRecordMessageExtend       = "intent_message_recordmessage_extend"
	IntentGlobalStop                       = "intent_global_stop"
	IntentGlobalStopExtend                 = "intent_global_stop_extend"
	IntentGlobalDelete                     = "intent_global_delete"
	IntentGlobalDeleteExtend               = "intent_global_delete_extend"
	IntentPhotoTake                        = "intent_photo_take"
	IntentPhotoTakeExtend                  = "intent_photo_take_extend"
	IntentSystemDiscovery                  = "intent_system_discovery"
	IntentSystemDiscoveryExtend            = "intent_system_discovery_extend"
	IntentImperativeVolumeLevelExtend      = "intent_imperative_volumelevel_extend"
	IntentImperativeEyeColorSpecificExtend = "intent_imperative_eyecolor_specific_extend"
)

// Server stores the config
type Server struct{}

var VoiceProcessor = ""

const (
	VoiceProcessorCoqui   = "coqui"
	VoiceProcessorLeopard = "leopard"
	VoiceProcessorVosk    = "vosk"
)

type JsonIntent struct {
	Name       string   `json:"name"`
	Keyphrases []string `json:"keyphrases"`
}

var sttLanguage string = "en-US"

var matchListList [][]string
var intentsList = []string{}

var botNum int = 0

func split(buf []byte) [][]byte {
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
		byteArray := split(n)
		return byteArray
	} else {
		// pcm
		byteArray := split(data)
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

func sttHandler(reqThing interface{}, isKnowledgeGraph bool) (transcribedString string, slots map[string]string, isRhino bool, thisBotNum int, opusUsed bool, err error) {
	if VoiceProcessorCoqui == VoiceProcessor {
		return CoquiSttHandler(reqThing, isKnowledgeGraph)
	} else if VoiceProcessorLeopard == VoiceProcessor {
		return LeopardSttHandler(reqThing, isKnowledgeGraph)
	}
	return VOSKSttHandler(reqThing, isKnowledgeGraph)
}

func loadIntents(voiceProcessor string, language string) ([][]string, []string, error) {
	jsonFile, err := os.ReadFile("./intent-data/" + voiceProcessor + "/" + language + ".json")

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
	logger("Instantiating " + voiceProcessor + " voice processor with language " + sttLanguage)

	// Instantiate the chosen voice processor and load intents from json
	VoiceProcessor = voiceProcessor
	err := errors.New("Error parsing intents JSON!")

	if VoiceProcessorCoqui == voiceProcessor {
		matchListList, intentsList, err = loadIntents(VoiceProcessorCoqui, sttLanguage)
		if err == nil {
			return CoquiNew()
		}
	} else if VoiceProcessorLeopard == voiceProcessor {
		matchListList, intentsList, err = loadIntents(VoiceProcessorLeopard, sttLanguage)
		if err == nil {
			return LeopardNew()
		}
	} else if VoiceProcessorVosk == voiceProcessor {
		matchListList, intentsList, err = loadIntents(VoiceProcessorVosk, sttLanguage)
		if err == nil {
			return VOSKNew()
		}
	} else {
		err = errors.New("Unknown voice processor")
	}

	return nil, err
}

var debugLogging bool

func logger(a ...any) {
	if debugLogging {
		fmt.Println(a...)
	}
}
