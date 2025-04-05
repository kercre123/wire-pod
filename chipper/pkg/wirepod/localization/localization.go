package localization

import "github.com/kercre123/wire-pod/chipper/pkg/vars"

var ValidVoskModels []string = []string{"en-US", "it-IT", "es-ES", "fr-FR", "de-DE", "pt-BR", "pl-PL", "zh-CN", "tr-TR", "ru-RU", "nt-NL", "uk-UA", "vi-VN"}

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
const STR_ZERO = "str_zero"
const STR_ONE = "str_one"
const STR_TWO = "str_two"
const STR_THREE = "str_three"
const STR_FOUR = "str_four"
const STR_FIVE = "str_five"
const STR_SIX = "str_six"
const STR_SEVEN = "str_seven"
const STR_EIGHT = "str_eight"
const STR_NINE = "str_nine"
const STR_TEN = "str_ten"
const STR_ELEVEN = "str_eleven"
const STR_TWELVE = "str_twelve"
const STR_THIRTEEN = "str_thirteen"
const STR_FOURTEEN = "str_fourteen"
const STR_FIFTEEN = "str_fifteen"
const STR_SIXTEEN = "str_sixteen"
const STR_SEVENTEEN = "str_seventeen"
const STR_EIGHTEEN = "str_eighteen"
const STR_NINETEEN = "str_nineteen"
const STR_TWENTY = "str_twenty"
const STR_THIRTY = "str_thirty"
const STR_FOURTY = "str_fourty"
const STR_FIFTY = "str_fifty"
const STR_SIXTY = "str_sixty"
const STR_SEVENTY = "str_seventy"
const STR_EIGHTY = "str_eighty"
const STR_NINETY = "str_ninety"
const STR_ONE_HUNDRED = "str_one_hundred"
const STR_ONE_HOUR = "str_one_hour"
const STR_ONE_HOUR_ALT = "str_one_hour_alt"
const STR_HOUR = "str_hour"
const STR_MINUTE = "str_minute"
const STR_SECOND = "str_second"

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
	"str_zero",
	"str_one",
	"str_two",
	"str_three",
	"str_four",
	"str_five",
	"str_six",
	"str_seven",
	"str_eight",
	"str_nine",
	"str_ten",
	"str_eleven",
	"str_twelve",
	"str_thirteen",
	"str_fourteen",
	"str_fifteen",
	"str_sixteen",
	"str_seventeen",
	"str_eighteen",
	"str_nineteen",
	"str_twenty",
	"str_thirty",
	"str_fourty",
	"str_fifty",
	"str_sixty",
	"str_seventy",
	"str_eighty",
	"str_ninety",
	"str_one_hundred",
	"str_one_hour",
	"str_one_hour_alt",
	"str_hour",
	"str_minute",
	"str_second",
}

// All text must be lowercase!

var texts = map[string][]string{
	//  key                 			en-US   it-IT   es-ES    fr-FR    de-DE    pl-PL   tr-TR	ru-RU    nt-NL     uk-UA  vi-VN
	STR_WEATHER_IN:                     {" in ", " a ", " en ", " en ", " in ", " w ", " 的 ", " içinde ", " в ", " in ", " в ", " ở "},
	STR_WEATHER_FORECAST:               {"forecast", "previsioni", "pronóstico", "prévisions", "wettervorhersage", "prognoza", "预报", "tahmin", "прогноз", "voorspelling", "прогноз", "dự báo"},
	STR_WEATHER_TOMORROW:               {"tomorrow", "domani", "mañana", "demain", "morgen", "jutro", "明天", "yarın", "завтра", "morgen", "завтра", "ngày mai"},
	STR_WEATHER_THE_DAY_AFTER_TOMORROW: {"day after tomorrow", "dopodomani", "el día después de mañana", "lendemain de demain", "am tag nach morgen", "pojutrze", "后天", "yarından sonra", "послезавтра", "overmorgen", "післязавтра", "ngày mốt"},
	STR_WEATHER_TONIGHT:                {"tonight", "stasera", "esta noche", "ce soir", "heute abend", "dziś wieczorem", "今晚", "bu gece", "сегодня вечером", "vanavond", "сьогодні ввечері", "tối nay"},
	STR_WEATHER_THIS_AFTERNOON:         {"afternoon", "pomeriggio", "esta tarde", "après-midi", "heute nachmittag", "popołudniu", "下午", "bu öğleden sonra", "после полудня", "middag", "після полудня", "chiều nay"},
	STR_EYE_COLOR_PURPLE:               {"purple", "lilla", "violeta", "violet", "violett", "fioletowy", "紫色", "mor", "фиолетовый", "paars", "фіолетовий", "màu tím"},
	STR_EYE_COLOR_BLUE:                 {"blue", "blu", "azul", "bleu", "blau", "niebieski", "蓝色", "mavi", "голубой", "blauw", "голубий", "màu xanh"},
	STR_EYE_COLOR_SAPPHIRE:             {"sapphire", "zaffiro", "zafiro", "saphir", "saphir", "szafir", "天蓝", "safir", "синий", "saffier", "синій", "màu ngọc bích"},
	STR_EYE_COLOR_YELLOW:               {"yellow", "giallo", "amarillo", "jaune", "gelb", "żółty", "黄色", "sarı", "жёлтый", "geel", "жовтий", "màu vàng"},
	STR_EYE_COLOR_TEAL:                 {"teal", "verde acqua", "verde azulado", "sarcelle", "blaugrün", "morski", "浅绿", "teal", "бирюзовый", "wintertaling", "бірюзовий", "xanh lá cây"},
	STR_EYE_COLOR_TEAL2:                {"tell", "acquamarina", "aguamarina", "acquamarine", "acquamarina", "akwamaryn", "蓝绿", "turkuaz", "аквамарин", "vertellen", "аквамариновий", "màu xanh ngọc"},
	STR_EYE_COLOR_GREEN:                {"green", "verde", "verde", "vert", "grün", "zielony", "绿色", "yeşil", "зелёный", "groente", "зелений", "màu xanh lá"},
	STR_EYE_COLOR_ORANGE:               {"orange", "arancio", "naranja", "orange", "orange", "pomarańczowy", "橙色", "turuncu", "оранжевый", "oranje", "оранжевий", "màu cam"},
	STR_ME:                             {"me", "me", "me", "moi", "mir", "mnie", "我", "ben", "меня", "mij", "мене", "tôi"},
	STR_SELF:                           {"self", "mi", "mía", "moi", "mein", "ja", "自己", "kendim", "себя", "zelf", "себе", "bản thân"},
	STR_VOLUME_LOW:                     {"low", "basso", "bajo", "bas", "niedrig", "niski", "低", "düşük", "низкий", "laag", "на мінімум", "thấp"},
	STR_VOLUME_QUIET:                   {"quiet", "poco rumoroso", "tranquilo", "silencieux", "ruhig", "cichy", "安静", "sessiz", "тихо", "rustig", "тихо", "yên tĩnh"},
	STR_VOLUME_MEDIUM_LOW:              {"medium low", "medio basso", "medio-bajo", "moyen bas", "mittelschwer", "średnio niski", "中低", "orta düşük", "ниже среднего", "middel laag", "нижче середнього", "vừa thấp"},
	STR_VOLUME_MEDIUM:                  {"medium", "medio", "medio", "moyen", "mittel", "średni", "中档", "orta", "средний", "medium", "середню", "vừa"},
	STR_VOLUME_NORMAL:                  {"normal", "normale", "normal", "normal", "normal", "normalny", "正常", "normal", "нормальный", "normaal", "нормальна", "bình thường"},
	STR_VOLUME_REGULAR:                 {"regular", "regolare", "regular", "standard", "regulär", "zwyczajny", "标准", "düzenli", "обычный", "normaal", "звичайна", "thông thường"},
	STR_VOLUME_MEDIUM_HIGH:             {"medium high", "medio alto", "medio-alto", "moyen-élevé", "mittelhoch", "średno wysoki", "中高", "orta yüksek", "выше среднего", "gemiddeld hoog", "вище середнього", "vừa cao"},
	STR_VOLUME_HIGH:                    {"high", "alto", "alto", "élevé", "hoch", "wysoki", "高档", "yüksek", "высокий", "hoog", "висока", "cao"},
	STR_VOLUME_LOUD:                    {"loud", "rumoroso", "fuerte", "fort", "laut", "głośny", "高", "gürültülü", "громкий", "luidruchtig", "гучний", "to"},
	STR_VOLUME_MUTE:                    {"mute", "muto", "mudo", "muet", "stumm", "wyciszony", "静音", "sessiz", "немой", "stom", "німий", "im lặng"},
	STR_VOLUME_NOTHING:                 {"nothing", "nessuno", "nada", "rien", "nichts", "nic", "无声", "hiçbir şey", "", "Niets", "нічого", "không có gì"},
	STR_VOLUME_SILENT:                  {"silent", "silenzioso", "silencio", "silencieux", "still", "cichy", "悄声", "sessiz", "тихий", "stil", "тихий", "yên lặng"},
	STR_VOLUME_OFF:                     {"off", "spento", "apagado", "éteindre", "aus", "wyłączony", "关闭", "kapalı", "выключить", "uit", "вимкнути", "tắt"},
	STR_VOLUME_ZERO:                    {"zero", "zero", "cero", "zéro", "null", "zero", "零", "sıfır", "ноль", "nul", "нуль", "không"},
	STR_NAME_IS:                        {" is ", " è ", " es ", " est ", " ist ", " to ", "到", " olan ", "", " is ", "", " là "},
	STR_NAME_IS2:                       {"'s", "sono ", "soy ", "suis ", "bin ", " się ", "的", "'nin", "", "", "", "của"},
	STR_NAME_IS3:                       {"names", " chiamo ", " llamo ", "appelle ", "werde", "imię", "名字", "adlar", "имена", "namen", "імена", "tên"},
	STR_FOR:                            {" for ", " per ", " para ", " pour ", " für ", " dla ", "给", " için ", "для", " voor ", " для ", " cho "},
	STR_ZERO:							{"zero","zero","zero","zéro","zero","zero","zero","zero","zero","zero","zero","zero",},
	STR_ONE:							{"one","one","one","un","one","one","one","one","one","one","one","one"},
	STR_TWO:							{"two","two","two","deux","two","two","two","two","two","two","two","two"},
	STR_THREE:							{"three","three","three","trois","three","three","three","three","three","three","three","three"},
	STR_FOUR:							{"four","four","four","quatre","four","four","four","four","four","four","four","four"},
	STR_FIVE:							{"five","five","five","cinq","five","five","five","five","five","five","five","five"},
	STR_SIX:							{"six","six","six","six","six","six","six","six","six","six","six","six"},
	STR_SEVEN:							{"seven","seven","seven","sept","seven","seven","seven","seven","seven","seven","seven","seven"},
	STR_EIGHT:							{"eight","eight","eight","huit","eight","eight","eight","eight","eight","eight","eight","eight"},
	STR_NINE:							{"nine","nine","nine","neuf","nine","nine","nine","nine","nine","nine","nine","nine"},
	STR_TEN:							{"ten","ten","ten","dix","ten","ten","ten","ten","ten","ten","ten","ten"},
	STR_ELEVEN:							{"eleven","eleven","eleven","onze","eleven","eleven","eleven","eleven","eleven","eleven","eleven","eleven"},
	STR_TWELVE:							{"twelve","twelve","twelve","douze","twelve","twelve","twelve","twelve","twelve","twelve","twelve","twelve"},
	STR_THIRTEEN:						{"thirteen","thirteen","thirteen","treize","thirteen","thirteen","thirteen","thirteen","thirteen","thirteen","thirteen","thirteen"},
	STR_FOURTEEN:						{"fourteen","fourteen","fourteen","quatorze","fourteen","fourteen","fourteen","fourteen","fourteen","fourteen","fourteen","fourteen"},
	STR_FIFTEEN:						{"fifteen","fifteen","fifteen","quinze","fifteen","fifteen","fifteen","fifteen","fifteen","fifteen","fifteen","fifteen"},
	STR_SIXTEEN:						{"sixteen","sixteen","sixteen","seize","sixteen","sixteen","sixteen","sixteen","sixteen","sixteen","sixteen","sixteen"},
	STR_SEVENTEEN:						{"seventeen","seventeen","seventeen","dix-sept","seventeen","seventeen","seventeen","seventeen","seventeen","seventeen","seventeen","seventeen"},
	STR_EIGHTEEN:						{"eighteen","eighteen","eighteen","dix-huit","eighteen","eighteen","eighteen","eighteen","eighteen","eighteen","eighteen","eighteen"},
	STR_NINETEEN:						{"nineteen","nineteen","nineteen","dix-neuf","nineteen","nineteen","nineteen","nineteen","nineteen","nineteen","nineteen","nineteen"},
	STR_TWENTY:							{"twenty","twenty","twenty","vingt","twenty","twenty","twenty","twenty","twenty","twenty","twenty","twenty"},
	STR_THIRTY:							{"thirty","thirty","thirty","trente","thirty","thirty","thirty","thirty","thirty","thirty","thirty","thirty"},
	STR_FOURTY:							{"fourty","fourty","fourty","quarante","fourty","fourty","fourty","fourty","fourty","fourty","fourty","fourty"},
	STR_FIFTY:							{"fifty","fifty","fifty","cinquante","fifty","fifty","fifty","fifty","fifty","fifty","fifty","fifty"},
	STR_SIXTY:							{"sixty","sixty","sixty","soixante","sixty","sixty","sixty","sixty","sixty","sixty","sixty","sixty"},
	STR_SEVENTY:						{"seventy","seventy","seventy","soixante-dix","seventy","seventy","seventy","seventy","seventy","seventy","seventy","seventy"},
	STR_EIGHTY:							{"eighty","eighty","eighty","quatre-vingt","eighty","eighty","eighty","eighty","eighty","eighty","eighty","eighty"},
	STR_NINETY:							{"ninety","ninety","ninety","quatre vingt dix","ninety","ninety","ninety","ninety","ninety","ninety","ninety","ninety"},
	STR_ONE_HUNDRED:					{"one hundred","one hundred","one hundred","cent","one hundred","one hundred","one hundred","one hundred","one hundred","one hundred","one hundred","one hundred"},
	STR_ONE_HOUR:						{"one hour","one hour","one hour","une heure","one hour","one hour","one hour","one hour","one hour","one hour","one hour","one hour"},
	STR_ONE_HOUR_ALT:					{"an hour","an hour","an hour","une heure","an hour","an hour","an hour","an hour","an hour","an hour","an hour","an hour"},
	STR_HOUR:							{"hour","hour","hour","heure","hour","hour","hour","hour","hour","hour","hour","hour"},
	STR_MINUTE:							{"minute","minute","minute","minute","minute","minute","minute","minute","minute","minute","minute","minute"},
	STR_SECOND:							{"second","second","second","seconde","second","second","second","second","second","second","second","second"},
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
		} else if vars.APIConfig.STT.Language == "tr-TR" {
			return data[7]
		} else if vars.APIConfig.STT.Language == "ru-RU" {
			return data[8]
		} else if vars.APIConfig.STT.Language == "nt-NL" {
			return data[9]
		} else if vars.APIConfig.STT.Language == "uk-UA" {
			return data[10]
		} else if vars.APIConfig.STT.Language == "vi-VN" {
			return data[11]
		}
	}
	return data[0]
}

func ReloadVosk() {
	if vars.APIConfig.STT.Service == "vosk" || vars.APIConfig.STT.Service == "whisper.cpp" {
		vars.IntentList, _ = vars.LoadIntents()
		vars.SttInitFunc()
	}
}
