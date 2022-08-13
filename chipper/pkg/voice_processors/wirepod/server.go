package wirepod

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/asticode/go-asticoqui"
)

var debugLogging bool

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

// New returns a new server
func New() (*Server, error) {
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
	var testTimer float64
	var timerDie bool = false
	logger("Running a Coqui test...")
	coquiInstance, _ := asticoqui.New("../stt/model.tflite")
	if _, err := os.Stat("../stt/large_vocabulary.scorer"); err == nil {
		coquiInstance.EnableExternalScorer("../stt/large_vocabulary.scorer")
	} else if _, err := os.Stat("../stt/model.scorer"); err == nil {
		coquiInstance.EnableExternalScorer("../stt/model.scorer")
	} else {
		logger("No .scorer file found.")
	}
	coquiStream, err := coquiInstance.NewStream()
	if err != nil {
		logger(err)
	}
	pcmBytes, _ := os.ReadFile("./stttest.pcm")
	var micData [][]byte
	micData = split(pcmBytes)
	for _, sample := range micData {
		coquiStream.FeedAudioContent(bytesToSamples(sample))
	}
	go func() {
		for testTimer <= 7.00 {
			if timerDie {
				break
			}
			time.Sleep(time.Millisecond * 10)
			testTimer = testTimer + 0.01
			if testTimer > 6.50 {
				logger("The STT test is taking too long, this hardware may not be adequate.")
			}
		}
	}()
	res, err := coquiStream.Finish()
	if err != nil {
		log.Fatal("Failed testing speech to text: ", err)
	}
	logger("Text:", res)
	timerDie = true
	logger("Coqui test successful! (Took " + strconv.FormatFloat(testTimer, 'f', 2, 64) + " seconds)")
	return &Server{}, nil
}

func logger(a ...any) {
	if debugLogging {
		fmt.Println(a...)
	}
}
