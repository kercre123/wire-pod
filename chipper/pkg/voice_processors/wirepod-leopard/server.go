package wirepod

import (
	"fmt"
	"os"
	"strconv"

	leopard "github.com/Picovoice/leopard/binding/go"
)

var debugLogging bool

var leopardSTTArray []leopard.Leopard
var picovoiceInstancesOS string = os.Getenv("PICOVOICE_INSTANCES")
var picovoiceInstances int

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
	var picovoiceKey string
	picovoiceKeyOS := os.Getenv("PICOVOICE_APIKEY")
	leopardKeyOS := os.Getenv("LEOPARD_APIKEY")
	if picovoiceInstancesOS == "" {
		picovoiceInstances = 3
	} else {
		picovoiceInstancesToInt, err := strconv.Atoi(picovoiceInstancesOS)
		picovoiceInstances = picovoiceInstancesToInt
		if err != nil {
			fmt.Println("PICOVOICE_INSTANCES is not a valid integer, using default value of 3")
			picovoiceInstances = 3
		}
	}
	if picovoiceKeyOS == "" {
		if leopardKeyOS == "" {
			fmt.Println("You must set PICOVOICE_APIKEY to a value.")
			os.Exit(1)
		} else {
			fmt.Println("PICOVOICE_APIKEY is not set, using LEOPARD_APIKEY")
			picovoiceKey = leopardKeyOS
		}
	} else {
		picovoiceKey = picovoiceKeyOS
	}
	fmt.Println("Initializing " + strconv.Itoa(picovoiceInstances) + " Picovoice Instances...")
	for i := 0; i < picovoiceInstances; i++ {
		fmt.Println("Initializing Picovoice Instance " + strconv.Itoa(i))
		leopardSTTArray = append(leopardSTTArray, leopard.NewLeopard(picovoiceKey))
		leopardSTTArray[i].Init()
	}
	return &Server{}, nil
}

func logger(a ...any) {
	if debugLogging {
		fmt.Println(a...)
	}
}
