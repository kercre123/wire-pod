package wirepod

const STR_WEATHER_IN = "str_weather_in"
const STR_WEATHER_FORECAST = "str_weather_forecast"

var texts = map[string][]string{
	//  key                 en-US   it-IT
	STR_WEATHER_IN:       {" in ", " a "},
	STR_WEATHER_FORECAST: {"forecast", "previsioni"},
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
