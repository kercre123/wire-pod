package wirepod_ttr

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	lcztn "github.com/kercre123/wire-pod/chipper/pkg/wirepod/localization"
)

// stt
func ParamChecker(req interface{}, intent string, speechText string, botSerial string) {
	var intentParam string
	var intentParamValue string
	var newIntent string
	var isParam bool
	var intentParams map[string]string
	var botLocation string = "San Francisco"
	var botUnits string = "F"
	var botPlaySpecific bool = false
	var botIsEarlyOpus bool = false

	// see if jdoc exists
	botJdoc, jdocExists := vars.GetJdoc("vic:"+botSerial, "vic.RobotSettings")
	if jdocExists {
		type robotSettingsJson struct {
			ButtonWakeword int  `json:"button_wakeword"`
			Clock24Hour    bool `json:"clock_24_hour"`
			CustomEyeColor struct {
				Enabled    bool    `json:"enabled"`
				Hue        float64 `json:"hue"`
				Saturation float64 `json:"saturation"`
			} `json:"custom_eye_color"`
			DefaultLocation  string `json:"default_location"`
			DistIsMetric     bool   `json:"dist_is_metric"`
			EyeColor         int    `json:"eye_color"`
			Locale           string `json:"locale"`
			MasterVolume     int    `json:"master_volume"`
			TempIsFahrenheit bool   `json:"temp_is_fahrenheit"`
			TimeZone         string `json:"time_zone"`
		}
		var robotSettings robotSettingsJson
		err := json.Unmarshal([]byte(botJdoc.JsonDoc), &robotSettings)
		if err != nil {
			logger.Println("Error unmarshaling json in paramchecker")
			logger.Println(err)
		} else {
			botLocation = robotSettings.DefaultLocation
			if robotSettings.TempIsFahrenheit {
				botUnits = "F"
			} else {
				botUnits = "C"
			}
		}
	}
	if botPlaySpecific {
		if strings.Contains(intent, "intent_play_blackjack") {
			isParam = true
			newIntent = "intent_play_specific_extend"
			intentParam = "entity_behavior"
			intentParamValue = "blackjack"
			intentParams = map[string]string{intentParam: intentParamValue}
		} else if strings.Contains(intent, "intent_play_fistbump") {
			isParam = true
			newIntent = "intent_play_specific_extend"
			intentParam = "entity_behavior"
			intentParamValue = "fist_bump"
			intentParams = map[string]string{intentParam: intentParamValue}
		} else if strings.Contains(intent, "intent_play_rollcube") {
			isParam = true
			newIntent = "intent_play_specific_extend"
			intentParam = "entity_behavior"
			intentParamValue = "roll_cube"
			intentParams = map[string]string{intentParam: intentParamValue}
		} else if strings.Contains(intent, "intent_play_popawheelie") {
			isParam = true
			newIntent = "intent_play_specific_extend"
			intentParam = "entity_behavior"
			intentParamValue = "pop_a_wheelie"
			intentParams = map[string]string{intentParam: intentParamValue}
		} else if strings.Contains(intent, "intent_play_pickupcube") {
			isParam = true
			newIntent = "intent_play_specific_extend"
			intentParam = "entity_behavior"
			intentParamValue = "pick_up_cube"
			intentParams = map[string]string{intentParam: intentParamValue}
		} else if strings.Contains(intent, "intent_play_keepaway") {
			isParam = true
			newIntent = "intent_play_specific_extend"
			intentParam = "entity_behavior"
			intentParamValue = "keep_away"
			intentParams = map[string]string{intentParam: intentParamValue}
		} else {
			newIntent = intent
			intentParam = ""
			intentParamValue = ""
			isParam = false
			intentParams = map[string]string{intentParam: intentParamValue}
		}
	}
	logger.Println("Checking params for candidate intent " + intent)
	if strings.Contains(intent, "intent_photo_take_extend") {
		isParam = true
		newIntent = intent
		if strings.Contains(speechText, lcztn.GetText(lcztn.STR_ME)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_SELF)) {
			intentParam = "entity_photo_selfie"
			intentParamValue = "photo_selfie"
		} else {
			intentParam = "entity_photo_selfie"
			intentParamValue = ""
		}
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_imperative_eyecolor") {
		isParam = true
		newIntent = "intent_imperative_eyecolor_specific_extend"
		intentParam = "eye_color"
		if strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_PURPLE)) {
			intentParamValue = "COLOR_PURPLE"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_BLUE)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_SAPPHIRE)) {
			intentParamValue = "COLOR_BLUE"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_YELLOW)) {
			intentParamValue = "COLOR_YELLOW"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_TEAL)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_TEAL2)) {
			intentParamValue = "COLOR_TEAL"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_GREEN)) {
			intentParamValue = "COLOR_GREEN"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_ORANGE)) {
			intentParamValue = "COLOR_ORANGE"
		} else {
			newIntent = intent
			intentParamValue = ""
			isParam = false
		}
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_weather_extend") {
		isParam = true
		newIntent = intent
		condition, is_forecast, local_datetime, speakable_location_string, temperature, temperature_unit := weatherParser(speechText, botLocation, botUnits)
		intentParams = map[string]string{"condition": condition, "is_forecast": is_forecast, "local_datetime": local_datetime, "speakable_location_string": speakable_location_string, "temperature": temperature, "temperature_unit": temperature_unit}
	} else if strings.Contains(intent, "intent_imperative_volumelevel_extend") {
		isParam = true
		newIntent = intent
		if strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_MEDIUM_LOW)) {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_2"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_LOW)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_QUIET)) {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_1"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_MEDIUM_HIGH)) {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_4"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_MEDIUM)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_NORMAL)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_REGULAR)) {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_3"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_HIGH)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_LOUD)) {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_5"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_MUTE)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_NOTHING)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_SILENT)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_OFF)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_ZERO)) {
			// there is no VOLUME_0 :(
			intentParam = "volume_level"
			intentParamValue = "VOLUME_1"
		} else {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_1"
		}
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_names_username_extend") {
		if vars.VoskGrammerEnable {
			var guid string
			var target string
			matched := false
			for _, bot := range vars.BotInfo.Robots {
				if botSerial == bot.Esn {
					guid = bot.GUID
					target = bot.IPAddress + ":443"
					matched = true
					break
				}
			}
			if matched {
				vec, err := vector.New(vector.WithSerialNo(botSerial), vector.WithToken(guid), vector.WithTarget(target))
				if err != nil {
					logger.Println("error connecting to vector:", err)
				} else {
					sayText(vec, "You must add a face in the web interface. It cannot be done via voice by default.")
				}
			}
			logger.Println("You must add a face via the web interface (Bot Settings -> Connect -> Faces).")
			logger.LogUI("You must add a face via the web interface (Bot Settings -> Connect -> Faces).")
		}
		var username string
		var nameSplitter string = ""
		isParam = true
		newIntent = intent
		if strings.Contains(speechText, lcztn.GetText(lcztn.STR_NAME_IS)) {
			nameSplitter = lcztn.GetText(lcztn.STR_NAME_IS)
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_NAME_IS2)) {
			nameSplitter = lcztn.GetText(lcztn.STR_NAME_IS2)
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_NAME_IS3)) {
			nameSplitter = lcztn.GetText(lcztn.STR_NAME_IS3)
		}
		if nameSplitter != "" {
			splitPhrase := strings.SplitAfter(speechText, nameSplitter)
			username = strings.TrimSpace(splitPhrase[1])
			if len(splitPhrase) == 3 {
				username = username + " " + strings.TrimSpace(splitPhrase[2])
			} else if len(splitPhrase) == 4 {
				username = username + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
			} else if len(splitPhrase) > 4 {
				username = username + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
			}
			logger.Println("Name parsed from speech: " + "`" + username + "`")
			intentParam = "username"
			intentParamValue = username
			intentParams = map[string]string{intentParam: intentParamValue}
		} else {
			logger.Println("No name parsed from speech")
			intentParam = "username"
			intentParamValue = ""
			intentParams = map[string]string{intentParam: intentParamValue}
		}
	} else if strings.Contains(intent, "intent_clock_settimer_extend") {
		isParam = true
		newIntent = intent
		timerSecs := words2num(speechText)
		logger.Println("Seconds parsed from speech: " + timerSecs)
		intentParam = "timer_duration"
		intentParamValue = timerSecs
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_global_stop_extend") {
		isParam = true
		newIntent = intent
		intentParam = "what_to_stop"
		intentParamValue = "timer"
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_message_playmessage_extend") {
		var given_name string
		isParam = true
		newIntent = intent
		intentParam = "given_name"
		if strings.Contains(speechText, lcztn.GetText(lcztn.STR_FOR)) {
			splitPhrase := strings.SplitAfter(speechText, lcztn.GetText(lcztn.STR_FOR))
			given_name = strings.TrimSpace(splitPhrase[1])
			if len(splitPhrase) == 3 {
				given_name = given_name + " " + strings.TrimSpace(splitPhrase[2])
			} else if len(splitPhrase) == 4 {
				given_name = given_name + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
			} else if len(splitPhrase) > 4 {
				given_name = given_name + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
			}
			intentParamValue = given_name
		} else {
			intentParamValue = ""
		}
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_message_recordmessage_extend") {
		var given_name string
		isParam = true
		newIntent = intent
		intentParam = "given_name"
		if strings.Contains(speechText, lcztn.GetText(lcztn.STR_FOR)) {
			splitPhrase := strings.SplitAfter(speechText, lcztn.GetText(lcztn.STR_FOR))
			given_name = strings.TrimSpace(splitPhrase[1])
			if len(splitPhrase) == 3 {
				given_name = given_name + " " + strings.TrimSpace(splitPhrase[2])
			} else if len(splitPhrase) == 4 {
				given_name = given_name + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
			} else if len(splitPhrase) > 4 {
				given_name = given_name + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
			}
			intentParamValue = given_name
		} else {
			intentParamValue = ""
		}
		intentParams = map[string]string{intentParam: intentParamValue}
	} else {
		if intentParam == "" {
			newIntent = intent
			intentParam = ""
			intentParamValue = ""
			isParam = false
			intentParams = map[string]string{intentParam: intentParamValue}
		}
	}
	if botIsEarlyOpus {
		if strings.Contains(intent, "intent_imperative_praise") {
			isParam = false
			newIntent = "intent_imperative_affirmative"
			intentParam = ""
			intentParamValue = ""
			intentParams = map[string]string{intentParam: intentParamValue}
		} else if strings.Contains(intent, "intent_imperative_abuse") {
			isParam = false
			newIntent = "intent_imperative_negative"
			intentParam = ""
			intentParamValue = ""
			intentParams = map[string]string{intentParam: intentParamValue}
		} else if strings.Contains(intent, "intent_imperative_love") {
			isParam = false
			newIntent = "intent_greeting_hello"
			intentParam = ""
			intentParamValue = ""
			intentParams = map[string]string{intentParam: intentParamValue}
		}
	}
	IntentPass(req, newIntent, speechText, intentParams, isParam)
}

// stintent
func ParamCheckerSlotsEnUS(req interface{}, intent string, slots map[string]string, isOpus bool, botSerial string) {
	var intentParam string
	var intentParamValue string
	var newIntent string
	var isParam bool
	var intentParams map[string]string
	var botLocation string = "San Francisco"
	var botUnits string = "F"
	var botPlaySpecific bool = false
	var botIsEarlyOpus bool = false
	// see if jdoc exists
	botJdoc, jdocExists := vars.GetJdoc("vic:"+botSerial, "vic.RobotSettings")
	if jdocExists {
		type robotSettingsJson struct {
			ButtonWakeword int  `json:"button_wakeword"`
			Clock24Hour    bool `json:"clock_24_hour"`
			CustomEyeColor struct {
				Enabled    bool    `json:"enabled"`
				Hue        float64 `json:"hue"`
				Saturation float64 `json:"saturation"`
			} `json:"custom_eye_color"`
			DefaultLocation  string `json:"default_location"`
			DistIsMetric     bool   `json:"dist_is_metric"`
			EyeColor         int    `json:"eye_color"`
			Locale           string `json:"locale"`
			MasterVolume     int    `json:"master_volume"`
			TempIsFahrenheit bool   `json:"temp_is_fahrenheit"`
			TimeZone         string `json:"time_zone"`
		}
		var robotSettings robotSettingsJson
		err := json.Unmarshal([]byte(botJdoc.JsonDoc), &robotSettings)
		if err != nil {
			logger.Println("Error unmarshaling json in paramchecker")
			logger.Println(err)
		} else {
			botLocation = robotSettings.DefaultLocation
			if robotSettings.TempIsFahrenheit {
				botUnits = "F"
			} else {
				botUnits = "C"
			}
		}
	}
	if strings.Contains(intent, "volume") {
		if slots["volume"] != "" {
			newIntent = "intent_imperative_volumelevel_extend"
			isParam = true
			intentParam = "volume_level"
			if strings.Contains(slots["volume"], "medium low") {
				intentParamValue = "VOLUME_2"
			} else if strings.Contains(slots["volume"], "low") {
				intentParamValue = "VOLUME_1"
			} else if strings.Contains(slots["volume"], "medium high") {
				intentParamValue = "VOLUME_4"
			} else if strings.Contains(slots["volume"], "high") {
				intentParamValue = "VOLUME_5"
			} else if strings.Contains(slots["volume"], "medium") {
				intentParamValue = "VOLUME_3"
			} else {
				intentParamValue = "VOLUME_1"
			}
		} else {
			isParam = false
			intentParam = ""
			intentParamValue = ""
		}
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "eyecolor") {
		isParam = true
		newIntent = "intent_imperative_eyecolor_specific_extend"
		intentParam = "eye_color"
		if strings.Contains(slots["eye_color"], "purple") {
			intentParamValue = "COLOR_PURPLE"
		} else if strings.Contains(slots["eye_color"], "blue") || strings.Contains(slots["eye_color"], "sapphire") {
			intentParamValue = "COLOR_BLUE"
		} else if strings.Contains(slots["eye_color"], "yellow") {
			intentParamValue = "COLOR_YELLOW"
		} else if strings.Contains(slots["eye_color"], "teal") || strings.Contains(slots["eye_color"], "tell") {
			intentParamValue = "COLOR_TEAL"
		} else if strings.Contains(slots["eye_color"], "green") {
			intentParamValue = "COLOR_GREEN"
		} else if strings.Contains(slots["eye_color"], "orange") {
			intentParamValue = "COLOR_ORANGE"
		} else {
			newIntent = intent
			intentParamValue = ""
			isParam = false
		}
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "_selfie") {
		newIntent = "intent_photo_take_extend"
		intentParam = "entity_photo_selfie"
		intentParamValue = "photo_selfie"
		isParam = true
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "_noselfie") {
		newIntent = "intent_photo_take_extend"
		intentParam = "entity_photo_selfie"
		intentParamValue = ""
		isParam = true
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "settimer") {
		isParam = true
		newIntent = intent
		slotNum := slots["num"]
		slotUnit := slots["unit"]
		timerSecs, err := strconv.Atoi(slotNum)
		if err != nil {
			logger.Println(err)
		}
		if slotNum != "" && slotUnit != "" {
			if strings.Contains(slotUnit, "minute") {
				timerSecs = timerSecs * 60
			} else if strings.Contains(slotUnit, "hour") {
				timerSecs = timerSecs * 60 * 60
			}
		}
		logger.Println("Seconds parsed from speech: " + strconv.Itoa(timerSecs))
		intentParam = "timer_duration"
		intentParamValue = strconv.Itoa(timerSecs)
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "global_stop_extend") {
		isParam = true
		newIntent = intent
		intentParam = "what_to_stop"
		intentParamValue = "timer"
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_knowledgegraph_prompt") {
		isParam = false
		newIntent = "intent_knowledge_promptquestion"
		intentParam = ""
		intentParamValue = ""
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_weather_extend") {
		isParam = true
		newIntent = intent
		condition, is_forecast, local_datetime, speakable_location_string, temperature, temperature_unit := weatherParser("what's the weather", botLocation, botUnits)
		intentParams = map[string]string{"condition": condition, "is_forecast": is_forecast, "local_datetime": local_datetime, "speakable_location_string": speakable_location_string, "temperature": temperature, "temperature_unit": temperature_unit}
	} else {
		if intentParam == "" {
			newIntent = intent
			intentParam = ""
			intentParamValue = ""
			isParam = false
			intentParams = map[string]string{intentParam: intentParamValue}
		}
	}
	if isOpus || botIsEarlyOpus || botPlaySpecific {
		if strings.Contains(intent, "intent_play_blackjack") {
			isParam = true
			newIntent = "intent_play_specific_extend"
			intentParam = "entity_behavior"
			intentParamValue = "blackjack"
			intentParams = map[string]string{intentParam: intentParamValue}
		} else if strings.Contains(intent, "intent_play_fistbump") {
			isParam = true
			newIntent = "intent_play_specific_extend"
			intentParam = "entity_behavior"
			intentParamValue = "fist_bump"
			intentParams = map[string]string{intentParam: intentParamValue}
		} else if strings.Contains(intent, "intent_play_rollcube") {
			isParam = true
			newIntent = "intent_play_specific_extend"
			intentParam = "entity_behavior"
			intentParamValue = "roll_cube"
			intentParams = map[string]string{intentParam: intentParamValue}
		} else if strings.Contains(intent, "intent_imperative_praise") {
			isParam = false
			newIntent = "intent_imperative_affirmative"
			intentParam = ""
			intentParamValue = ""
			intentParams = map[string]string{intentParam: intentParamValue}
		} else if strings.Contains(intent, "intent_imperative_love") {
			isParam = false
			newIntent = "intent_greeting_hello"
			intentParam = ""
			intentParamValue = ""
			intentParams = map[string]string{intentParam: intentParamValue}
		} else if strings.Contains(intent, "intent_imperative_abuse") {
			isParam = false
			newIntent = "intent_imperative_negative"
			intentParam = ""
			intentParamValue = ""
			intentParams = map[string]string{intentParam: intentParamValue}
		}
	}
	IntentPass(req, newIntent, intent, intentParams, isParam)
}

func prehistoricParamChecker(req interface{}, intent string, speechText string, botSerial string) {
	// intent.go detects if the stream uses opus or PCM.
	// If the stream is PCM, it is likely a bot with 0.10.
	// This accounts for the newer 0.10.1### builds.
	var intentParam string
	var intentParamValue string
	var newIntent string
	var isParam bool
	var intentParams map[string]string
	var botLocation string = "San Francisco"
	var botUnits string = "F"
	if strings.Contains(intent, "intent_photo_take_extend") {
		isParam = true
		newIntent = intent
		if strings.Contains(speechText, lcztn.GetText(lcztn.STR_ME)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_SELF)) {
			intentParam = "entity_photo_selfie"
			intentParamValue = "photo_selfie"
		} else {
			intentParam = "entity_photo_selfie"
			intentParamValue = ""
		}
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_imperative_eyecolor") {
		// leaving stuff like this in case someone wants to add features like this to older software
		isParam = true
		newIntent = "intent_imperative_eyecolor_specific_extend"
		intentParam = "eye_color"
		if strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_PURPLE)) {
			intentParamValue = "COLOR_PURPLE"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_BLUE)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_SAPPHIRE)) {
			intentParamValue = "COLOR_BLUE"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_YELLOW)) {
			intentParamValue = "COLOR_YELLOW"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_TEAL)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_TEAL2)) {
			intentParamValue = "COLOR_TEAL"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_GREEN)) {
			intentParamValue = "COLOR_GREEN"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_EYE_COLOR_ORANGE)) {
			intentParamValue = "COLOR_ORANGE"
		} else {
			newIntent = intent
			intentParamValue = ""
			isParam = false
		}
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_weather_extend") {
		isParam = true
		newIntent = intent
		condition, is_forecast, local_datetime, speakable_location_string, temperature, temperature_unit := weatherParser(speechText, botLocation, botUnits)
		intentParams = map[string]string{"condition": condition, "is_forecast": is_forecast, "local_datetime": local_datetime, "speakable_location_string": speakable_location_string, "temperature": temperature, "temperature_unit": temperature_unit}
	} else if strings.Contains(intent, "intent_imperative_volumelevel_extend") {
		isParam = true
		newIntent = intent
		if strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_MEDIUM_LOW)) {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_2"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_LOW)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_QUIET)) {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_1"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_MEDIUM_HIGH)) {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_4"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_MEDIUM)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_NORMAL)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_REGULAR)) {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_3"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_HIGH)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_LOUD)) {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_5"
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_MUTE)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_NOTHING)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_SILENT)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_OFF)) || strings.Contains(speechText, lcztn.GetText(lcztn.STR_VOLUME_ZERO)) {
			// there is no VOLUME_0 :(
			intentParam = "volume_level"
			intentParamValue = "VOLUME_1"
		} else {
			intentParam = "volume_level"
			intentParamValue = "VOLUME_1"
		}
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_names_username_extend") {
		var username string
		var nameSplitter string = ""
		isParam = true
		newIntent = "intent_names_username"
		if strings.Contains(speechText, lcztn.GetText(lcztn.STR_NAME_IS)) {
			nameSplitter = lcztn.GetText(lcztn.STR_NAME_IS)
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_NAME_IS2)) {
			nameSplitter = lcztn.GetText(lcztn.STR_NAME_IS2)
		} else if strings.Contains(speechText, lcztn.GetText(lcztn.STR_NAME_IS3)) {
			nameSplitter = lcztn.GetText(lcztn.STR_NAME_IS3)
		}
		if nameSplitter != "" {
			splitPhrase := strings.SplitAfter(speechText, nameSplitter)
			username = strings.TrimSpace(splitPhrase[1])
			if len(splitPhrase) == 3 {
				username = username + " " + strings.TrimSpace(splitPhrase[2])
			} else if len(splitPhrase) == 4 {
				username = username + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
			} else if len(splitPhrase) > 4 {
				username = username + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
			}
			logger.Println("Name parsed from speech: " + "`" + username + "`")
			intentParam = "username"
			intentParamValue = username
			intentParams = map[string]string{intentParam: intentParamValue}
		} else {
			logger.Println("No name parsed from speech")
			intentParam = "username"
			intentParamValue = ""
			intentParams = map[string]string{intentParam: intentParamValue}
		}
	} else if strings.Contains(intent, "intent_clock_settimer_extend") {
		isParam = true
		newIntent = "intent_clock_settimer"
		timerSecs := words2num(speechText)
		logger.Println("Seconds parsed from speech: " + timerSecs)
		intentParam = "timer_duration"
		intentParamValue = timerSecs
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_global_stop_extend") {
		isParam = true
		newIntent = "intent_global_stop"
		intentParam = "what_to_stop"
		intentParamValue = "timer"
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_message_playmessage_extend") {
		var given_name string
		isParam = true
		newIntent = "intent_message_playmessage"
		intentParam = "given_name"
		if strings.Contains(speechText, lcztn.GetText(lcztn.STR_FOR)) {
			splitPhrase := strings.SplitAfter(speechText, lcztn.GetText(lcztn.STR_FOR))
			given_name = strings.TrimSpace(splitPhrase[1])
			if len(splitPhrase) == 3 {
				given_name = given_name + " " + strings.TrimSpace(splitPhrase[2])
			} else if len(splitPhrase) == 4 {
				given_name = given_name + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
			} else if len(splitPhrase) > 4 {
				given_name = given_name + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
			}
			intentParamValue = given_name
		} else {
			intentParamValue = ""
		}
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_message_recordmessage_extend") {
		var given_name string
		isParam = true
		newIntent = "intent_message_recordmessage"
		intentParam = "given_name"
		if strings.Contains(speechText, lcztn.GetText(lcztn.STR_FOR)) {
			splitPhrase := strings.SplitAfter(speechText, lcztn.GetText(lcztn.STR_FOR))
			given_name = strings.TrimSpace(splitPhrase[1])
			if len(splitPhrase) == 3 {
				given_name = given_name + " " + strings.TrimSpace(splitPhrase[2])
			} else if len(splitPhrase) == 4 {
				given_name = given_name + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
			} else if len(splitPhrase) > 4 {
				given_name = given_name + " " + strings.TrimSpace(splitPhrase[2]) + " " + strings.TrimSpace(splitPhrase[3])
			}
			intentParamValue = given_name
		} else {
			intentParamValue = ""
		}
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_play_blackjack") {
		isParam = true
		newIntent = "intent_play_specific_extend"
		intentParam = "entity_behavior"
		intentParamValue = "blackjack"
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_play_fistbump") {
		isParam = true
		newIntent = "intent_play_specific_extend"
		intentParam = "entity_behavior"
		intentParamValue = "fist_bump"
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_play_rollcube") {
		isParam = true
		newIntent = "intent_play_specific_extend"
		intentParam = "entity_behavior"
		intentParamValue = "roll_cube"
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_imperative_praise") {
		isParam = false
		newIntent = "intent_imperative_affirmative"
		intentParam = ""
		intentParamValue = ""
		intentParams = map[string]string{intentParam: intentParamValue}
	} else if strings.Contains(intent, "intent_imperative_abuse") {
		isParam = false
		newIntent = "intent_imperative_negative"
		intentParam = ""
		intentParamValue = ""
		intentParams = map[string]string{intentParam: intentParamValue}
	} else {
		newIntent = intent
		intentParam = ""
		intentParamValue = ""
		isParam = false
		intentParams = map[string]string{intentParam: intentParamValue}
	}
	IntentPass(req, newIntent, speechText, intentParams, isParam)
}
