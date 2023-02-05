package wirepod_ttr

import "github.com/kercre123/chipper/pkg/vars"

const STR_WEATHER_IN = "str_weather_in"
const STR_WEATHER_FORECAST = "str_weather_forecast"
const STR_WEATHER_TOMORROW = "str_weather_tomorrow"
const STR_WEATHER_THE_DAY_AFTER_TOMORROW = "str_weather_the_day_after_tomorrow"
const STR_WEATHER_TONIGHT = "str_weather_tonight"
const STR_WEATHER_THIS_AFTERNOON = "str_weather_this_afternoon"
const STR_EYE_COLOR_PURPLE = "str_eye_color_purple"
const STR_EYE_COLOR_BLUE = "str_eye_color_blue"
const STR_EYE_COLOR_SAPPHIRE = "str_eye_color_sapphire"
const STR_EYE_COLOR_YELLOW = "str_eye_color_yellow"
const STR_EYE_COLOR_TEAL = "str_eye_color_teal"
const STR_EYE_COLOR_TEAL2 = "str_eye_color_teal2"
const STR_EYE_COLOR_GREEN = "str_eye_color_green"
const STR_EYE_COLOR_ORANGE = "str_eye_color_orange"
const STR_ME = "str_me"
const STR_SELF = "str_self"
const STR_VOLUME_LOW = "str_volume_low"
const STR_VOLUME_QUIET = "str_volume_quiet"
const STR_VOLUME_MEDIUM_LOW = "str_volume_medium_low"
const STR_VOLUME_MEDIUM = "str_volume_medium"
const STR_VOLUME_NORMAL = "str_volume_normal"
const STR_VOLUME_REGULAR = "str_volume_regular"
const STR_VOLUME_MEDIUM_HIGH = "str_volume_medium_high"
const STR_VOLUME_HIGH = "str_volume_high"
const STR_VOLUME_LOUD = "str_volume_loud"
const STR_VOLUME_MUTE = "str_volume_mute"
const STR_VOLUME_NOTHING = "str_volume_nothing"
const STR_VOLUME_SILENT = "str_volume_silent"
const STR_VOLUME_OFF = "str_volume_off"
const STR_VOLUME_ZERO = "str_volume_zero"
const STR_NAME_IS = "str_name_is"
const STR_NAME_IS2 = "str_name_is1"
const STR_NAME_IS3 = "str_name_is2"
const STR_FOR = "str_for"

// All text must be lowercase!

var texts = map[string][]string{
	//  key                 			en-US   it-IT   es-ES    fr-FR    de-DE
	STR_WEATHER_IN:                     {" in ", " a ", " en ", " en ", " in "},
	STR_WEATHER_FORECAST:               {"forecast", "previsioni", "pronóstico", "prévisions", "wettervorhersage"},
	STR_WEATHER_TOMORROW:               {"tomorrow", "domani", "mañana", "demain", "morgen"},
	STR_WEATHER_THE_DAY_AFTER_TOMORROW: {"day after tomorrow", "dopodomani", "el día después de mañana", "lendemain de demain", "am tag nach morgen"},
	STR_WEATHER_TONIGHT:                {"tonight", "stasera", "esta noche", "ce soir", "heute abend"},
	STR_WEATHER_THIS_AFTERNOON:         {"afternoon", "pomeriggio", "esta tarde", "après-midi", "heute nachmittag"},
	STR_EYE_COLOR_PURPLE:               {"purple", "lilla", "violeta", "violet", "violett"},
	STR_EYE_COLOR_BLUE:                 {"blue", "blu", "azul", "bleu", "blau"},
	STR_EYE_COLOR_SAPPHIRE:             {"sapphire", "zaffiro", "zafiro", "saphir", "saphir"},
	STR_EYE_COLOR_YELLOW:               {"yellow", "giallo", "amarillo", "jaune", "gelb"},
	STR_EYE_COLOR_TEAL:                 {"teal", "verde acqua", "verde azulado", "sarcelle", "blaugrün"},
	STR_EYE_COLOR_TEAL2:                {"tell", "acquamarina", "aguamarina", "acquamarina", "acquamarina"},
	STR_EYE_COLOR_GREEN:                {"green", "verde", "verde", "vert", "grün"},
	STR_EYE_COLOR_ORANGE:               {"orange", "arancio", "naranja", "orange", "orange"},
	STR_ME:                             {"me", "me", "me", "moi", "mir"},
	STR_SELF:                           {"self", "mi", "mía", "moi", "mein"},
	STR_VOLUME_LOW:                     {"low", "basso", "bajo", "bas", "niedrig"},
	STR_VOLUME_QUIET:                   {"quiet", "poco rumoroso", "tranquilo", "silencieux", "ruhig"},
	STR_VOLUME_MEDIUM_LOW:              {"medium low", "medio basso", "medio-bajo", "moyen-doux", "mittelschwer"},
	STR_VOLUME_MEDIUM:                  {"medium", "medio", "medio", "moyen", "mittel"},
	STR_VOLUME_NORMAL:                  {"normal", "normale", "normal", "normal", "normal"},
	STR_VOLUME_REGULAR:                 {"regular", "regolare", "regular", "régulier", "regulär"},
	STR_VOLUME_MEDIUM_HIGH:             {"medium high", "medio alto", "medio-alto", "moyen-élevé", "mittelhoch"},
	STR_VOLUME_HIGH:                    {"high", "alto", "alto", "élevé", "hoch"},
	STR_VOLUME_LOUD:                    {"loud", "rumoroso", "fuerte", "fort", "laut"},
	STR_VOLUME_MUTE:                    {"mute", "muto", "mudo", "", "stumm"},
	STR_VOLUME_NOTHING:                 {"nothing", "nessuno", "nada", "rien", "nichts"},
	STR_VOLUME_SILENT:                  {"silent", "silenzioso", "silencio", "silencieux", "still"},
	STR_VOLUME_OFF:                     {"off", "spento", "apagado", "éteindre", "aus"},
	STR_VOLUME_ZERO:                    {"zero", "zero", "cero", "zéro", "null"},
	STR_NAME_IS:                        {" is ", " è ", " es ", " est ", " ist "},
	STR_NAME_IS2:                       {"'s", "sono ", "soy ", "suis ", "bin "},
	STR_NAME_IS3:                       {"names", " chiamo ", " llamo ", "appelle ", "werde"},
	STR_FOR:                            {" for ", " per ", " para ", " pour ", " für "},
}

func getText(key string) string {
	var data = texts[key]
	if data != nil {
		if vars.APIConfig.STT.Language == "it-IT" {
			return data[1]
		} else if vars.APIConfig.STT.Language == "es-ES" {
			return data[2]
		} else if vars.APIConfig.STT.Language == "fr-FR" {
			return data[3]
		} else if vars.APIConfig.STT.Language == "de-DE" {
			return data[4]
		}
	}
	return data[0]
}
