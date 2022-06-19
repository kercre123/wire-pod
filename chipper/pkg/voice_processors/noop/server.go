package noop

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
	return &Server{}, nil
}
