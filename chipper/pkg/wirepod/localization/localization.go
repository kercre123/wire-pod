package localization

import "github.com/kercre123/wire-pod/chipper/pkg/vars"

var ValidVoskModels []string = []string{"en-US", "it-IT", "es-ES", "fr-FR", "de-DE", "pt-BR", "pl-PL", "zh-CN"}

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

// for grammer
var ALL_STR []string = []string{
	"str_weather_in",
	"str_weather_forecast",
	"str_weather_tomorrow",
	"str_weather_the_day_after_tomorrow",
	"str_weather_tonight",
	"str_weather_this_afternoon",
	"str_eye_color_purple",
	"str_eye_color_blue",
	"str_eye_color_sapphire",
	"str_eye_color_yellow",
	"str_eye_color_teal",
	"str_eye_color_teal2",
	"str_eye_color_green",
	"str_eye_color_orange",
	"str_me",
	"str_self",
	"str_volume_low",
	"str_volume_quiet",
	"str_volume_medium_low",
	"str_volume_medium",
	"str_volume_normal",
	"str_volume_regular",
	"str_volume_medium_high",
	"str_volume_high",
	"str_volume_loud",
	"str_volume_mute",
	"str_volume_nothing",
	"str_volume_silent",
	"str_volume_off",
	"str_volume_zero",
	"str_name_is",
	"str_name_is1",
	"str_name_is2",
	"str_for",
}

// All text must be lowercase!

var texts = map[string][]string{
	//  key                 			en-US   it-IT   es-ES    fr-FR    de-DE    pl-PL
	STR_WEATHER_IN:                     {" in ", " a ", " en ", " en ", " in ", " w ", " 的 "},
	STR_WEATHER_FORECAST:               {"forecast", "previsioni", "pronóstico", "prévisions", "wettervorhersage", "prognoza", "预报"},
	STR_WEATHER_TOMORROW:               {"tomorrow", "domani", "mañana", "demain", "morgen", "jutro", "明天"},
	STR_WEATHER_THE_DAY_AFTER_TOMORROW: {"day after tomorrow", "dopodomani", "el día después de mañana", "lendemain de demain", "am tag nach morgen", "pojutrze", "后天"},
	STR_WEATHER_TONIGHT:                {"tonight", "stasera", "esta noche", "ce soir", "heute abend", "dziś wieczorem", "今晚"},
	STR_WEATHER_THIS_AFTERNOON:         {"afternoon", "pomeriggio", "esta tarde", "après-midi", "heute nachmittag", "popołudniu", "下午"},
	STR_EYE_COLOR_PURPLE:               {"purple", "lilla", "violeta", "violet", "violett", "fioletowy", "紫色"},
	STR_EYE_COLOR_BLUE:                 {"blue", "blu", "azul", "bleu", "blau", "niebieski", "蓝色"},
	STR_EYE_COLOR_SAPPHIRE:             {"sapphire", "zaffiro", "zafiro", "saphir", "saphir", "szafir", "天蓝"},
	STR_EYE_COLOR_YELLOW:               {"yellow", "giallo", "amarillo", "jaune", "gelb", "żółty", "黄色"},
	STR_EYE_COLOR_TEAL:                 {"teal", "verde acqua", "verde azulado", "sarcelle", "blaugrün", "morski", "浅绿"},
	STR_EYE_COLOR_TEAL2:                {"tell", "acquamarina", "aguamarina", "acquamarina", "acquamarina", "akwamaryn", "蓝绿"},
	STR_EYE_COLOR_GREEN:                {"green", "verde", "verde", "vert", "grün", "zielony", "绿色"},
	STR_EYE_COLOR_ORANGE:               {"orange", "arancio", "naranja", "orange", "orange", "pomarańczowy", "橙色"},
	STR_ME:                             {"me", "me", "me", "moi", "mir", "mnie", "我"},
	STR_SELF:                           {"self", "mi", "mía", "moi", "mein", "ja", "自己"},
	STR_VOLUME_LOW:                     {"low", "basso", "bajo", "bas", "niedrig", "niski", "低"},
	STR_VOLUME_QUIET:                   {"quiet", "poco rumoroso", "tranquilo", "silencieux", "ruhig", "cichy", "安静"},
	STR_VOLUME_MEDIUM_LOW:              {"medium low", "medio basso", "medio-bajo", "moyen-doux", "mittelschwer", "średnio niski", "中低"},
	STR_VOLUME_MEDIUM:                  {"medium", "medio", "medio", "moyen", "mittel", "średni", "中档"},
	STR_VOLUME_NORMAL:                  {"normal", "normale", "normal", "normal", "normal", "normalny", "正常"},
	STR_VOLUME_REGULAR:                 {"regular", "regolare", "regular", "régulier", "regulär", "zwyczajny", "标准"},
	STR_VOLUME_MEDIUM_HIGH:             {"medium high", "medio alto", "medio-alto", "moyen-élevé", "mittelhoch", "średno wysoki", "中高"},
	STR_VOLUME_HIGH:                    {"high", "alto", "alto", "élevé", "hoch", "wysoki", "高档"},
	STR_VOLUME_LOUD:                    {"loud", "rumoroso", "fuerte", "fort", "laut", "głośny", "高"},
	STR_VOLUME_MUTE:                    {"mute", "muto", "mudo", "", "stumm", "wyciszony", "静音"},
	STR_VOLUME_NOTHING:                 {"nothing", "nessuno", "nada", "rien", "nichts", "nic", "无声"},
	STR_VOLUME_SILENT:                  {"silent", "silenzioso", "silencio", "silencieux", "still", "cichy", "悄声"},
	STR_VOLUME_OFF:                     {"off", "spento", "apagado", "éteindre", "aus", "wyłączony", "关闭"},
	STR_VOLUME_ZERO:                    {"zero", "zero", "cero", "zéro", "null", "zero", "零"},
	STR_NAME_IS:                        {" is ", " è ", " es ", " est ", " ist ", " to ", "到"},
	STR_NAME_IS2:                       {"'s", "sono ", "soy ", "suis ", "bin ", " się ", "的"},
	STR_NAME_IS3:                       {"names", " chiamo ", " llamo ", "appelle ", "werde", "imię", "名字"},
	STR_FOR:                            {" for ", " per ", " para ", " pour ", " für ", " dla ", "给"},
}

func GetText(key string) string {
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
		} else if vars.APIConfig.STT.Language == "pl-PL" {
			return data[5]
		} else if vars.APIConfig.STT.Language == "zh-CN" {
			return data[6]
		}
	}
	return data[0]
}

func ReloadVosk() {
	if vars.APIConfig.STT.Service == "vosk" || vars.APIConfig.STT.Service == "whisper.cpp" {
		vars.MatchListList, vars.IntentsList, _ = vars.LoadIntents()
		vars.SttInitFunc()
	}
}
