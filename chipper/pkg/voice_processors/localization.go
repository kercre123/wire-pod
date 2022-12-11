package wirepod

const STR_WEATHER_IN = "str_weather_in"
const STR_WEATHER_FORECAST = "str_weather_forecast"
const STR_WEATHER_TOMORROW = "str_weather_tomorrow"
const STR_WEATHER_THE_DAY_AFTER_TOMORROW = "str_weather_the_day_after_tomorrow"
const STR_WEATHER_TONIGHT = "str_weather_tonight"
const STR_WEATHER_THIS_AFTERNOON = "str_weather_this_afternoon"

var texts = map[string][]string{
	//  key                 			en-US   it-IT
	STR_WEATHER_IN:                     {" in ", " a "},
	STR_WEATHER_FORECAST:               {"forecast", "previsioni"},
	STR_WEATHER_TOMORROW:               {"tomorrow", "domani", "to morrow"},
	STR_WEATHER_THE_DAY_AFTER_TOMORROW: {"day after tomorrow", "dopodomani"},
	STR_WEATHER_TONIGHT:                {"tonight", "stasera"},
	STR_WEATHER_THIS_AFTERNOON:         {"afternoon", "pomeriggio"},
}

func getText(key string) string {
	var data = texts[key]
	if data != nil {
		if sttLanguage == "it-IT" {
			return data[1]
		}
	}
	return data[0]
}
