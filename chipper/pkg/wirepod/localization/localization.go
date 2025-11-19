package localization

import "github.com/kercre123/wire-pod/chipper/pkg/vars"

var ValidVoskModels []string = []string{"en-US", "it-IT", "es-ES", "fr-FR", "de-DE", "pt-BR", "pl-PL", "zh-CN", "tr-TR", "ru-RU", "nt-NL", "uk-UA", "vi-VN", "ko-KR"}

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
    // key                  en-US           it-IT           es-ES           fr-FR           de-DE           pl-PL           zh-CN    tr-TR           ru-RU           nt-NL           uk-UA           vi-VN           ko-KR
    STR_WEATHER_IN:                     {" in ", " a ", " en ", " en ", " in ", " w ", " 在 ", " içinde ", " в ", " in ", " в ", " ở ", "의 "},
    STR_WEATHER_FORECAST:               {"forecast", "previsioni", "pronóstico", "prévisions", "wettervorhersage", "prognoza", "预报", "tahmin", "прогноз", "voorspelling", "прогноз", "dự báo", "일기 예보"},
    STR_WEATHER_TOMORROW:               {"tomorrow", "domani", "mañana", "demain", "morgen", "jutro", "明天", "yarın", "завтра", "morgen", "завтра", "ngày mai", "내일"},
    STR_WEATHER_THE_DAY_AFTER_TOMORROW: {"day after tomorrow", "dopodomani", "el día después de mañana", "lendemain de demain", "am tag nach morgen", "pojutrze", "后天", "yarından sonra", "послезавтра", "overmorgen", "післязавтра", "ngày mốt", "모레"},
    STR_WEATHER_TONIGHT:                {"tonight", "stasera", "esta noche", "ce soir", "heute abend", "dziś wieczorem", "今晚", "bu gece", "сегодня вечером", "vanavond", "сьогодні ввечері", "tối nay", "오늘 밤"},
    STR_WEATHER_THIS_AFTERNOON:         {"afternoon", "pomeriggio", "esta tarde", "après-midi", "heute nachmittag", "popołudniu", "下午", "bu öğleden sonra", "после полудня", "middag", "після полудня", "chiều nay", "오후"},
    STR_EYE_COLOR_PURPLE:               {"purple", "lilla", "violeta", "violet", "violett", "fioletowy", "紫色", "mor", "фиолетовый", "paars", "фіолетовий", "màu tím", "보라색"},
    STR_EYE_COLOR_BLUE:                 {"blue", "blu", "azul", "bleu", "blau", "niebieski", "蓝色", "mavi", "голубой", "blauw", "голубий", "màu xanh", "파란색"},
    STR_EYE_COLOR_SAPPHIRE:             {"sapphire", "zaffiro", "zafiro", "saphir", "saphir", "szafir", "天蓝", "safir", "синий", "saffier", "синій", "màu ngọc bích", "사파이어색"},
    STR_EYE_COLOR_YELLOW:               {"yellow", "giallo", "amarillo", "jaune", "gelb", "żółty", "黄色", "sarı", "жёлтый", "geel", "жовтий", "màu vàng", "노란색"},
    STR_EYE_COLOR_TEAL:                 {"teal", "verde acqua", "verde azulado", "sarcelle", "blaugrün", "morski", "浅绿", "teal", "бирюзовый", "wintertaling", "бірюзовий", "xanh lá cây", "청록색"},
    STR_EYE_COLOR_TEAL2:                {"tell", "acquamarina", "aguamarina", "acquamarine", "acquamarina", "akwamaryn", "蓝绿", "turkuaz", "аквамарин", "vertellen", "аквамариновий", "màu xanh ngọc", "아쿠아마린색"},
    STR_EYE_COLOR_GREEN:                {"green", "verde", "verde", "vert", "grün", "zielony", "绿色", "yeşil", "зелёный", "groente", "зелений", "màu xanh lá", "초록색"},
    STR_EYE_COLOR_ORANGE:               {"orange", "arancio", "naranja", "orange", "orange", "pomarańczowy", "橙色", "turuncu", "оранжевый", "oranje", "оранжевий", "màu cam", "주황색"},
    STR_ME:                             {"me", "me", "me", "moi", "mir", "mnie", "我", "ben", "меня", "mij", "мене", "tôi", "나"},
    STR_SELF:                           {"self", "mi", "mía", "moi", "mein", "ja", "自己", "kendim", "себя", "zelf", "себе", "bản thân", "자신"},
    STR_VOLUME_LOW:                     {"low", "minimo", "bajo", "bas", "niedrig", "niski", "低", "düşük", "низкий", "laag", "на мінімум", "thấp", "아주 작게"},
    STR_VOLUME_QUIET:                   {"quiet", "basso", "tranquilo", "silencieux", "ruhig", "cichy", "安静", "sessiz", "тихо", "rustig", "тихо", "yên tĩnh", "작게"},
    STR_VOLUME_MEDIUM_LOW:              {"medium low", "medio basso", "medio-bajo", "moyen bas", "mittelschwer", "średnio niski", "中低", "orta düşük", "ниже среднего", "middel laag", "нижче середнього", "vừa thấp", "조금 작게"},
    STR_VOLUME_MEDIUM:                  {"medium", "medio", "medio", "moyen", "mittel", "średni", "中档", "orta", "средний", "medium", "середню", "vừa", "중간"},
    STR_VOLUME_NORMAL:                  {"normal", "normale", "normal", "normal", "normal", "normalny", "正常", "normal", "нормальный", "normaal", "нормальна", "bình thường", "보통"},
    STR_VOLUME_REGULAR:                 {"regular", "regolare", "regular", "standard", "regulär", "zwyczajny", "标准", "düzenli", "обычный", "normaal", "звичайна", "thông thường", "보통"},
    STR_VOLUME_MEDIUM_HIGH:             {"medium high", "medio alto", "medio-alto", "moyen-élevé", "mittelhoch", "średno wysoki", "中高", "orta yüksek", "выше среднего", "gemiddeld hoog", "вище середнього", "vừa cao", "조금 크게"},
    STR_VOLUME_HIGH:                    {"high", "alto", "alto", "élevé", "hoch", "wysoki", "高档", "yüksek", "высокий", "hoog", "висока", "cao", "크게"},
    STR_VOLUME_LOUD:                    {"loud", "massimo", "fuerte", "fort", "laut", "głośny", "高", "gürültülü", "громкий", "luidruchtig", "гучний", "to", "아주 크게"},
    STR_VOLUME_MUTE:                    {"mute", "muto", "mudo", "muet", "stumm", "wyciszony", "静音", "sessiz", "без звука", "stom", "німий", "im lặng", "음소거"},
    STR_VOLUME_NOTHING:                 {"nothing", "nessuno", "nada", "rien", "nichts", "nic", "无声", "hiçbir şey", "ничего", "niets", "нічого", "không có gì", "음소거"},
    STR_VOLUME_SILENT:                  {"silent", "silenzioso", "silencio", "silencieux", "still", "cichy", "悄声", "sessiz", "тихий", "stil", "тихий", "yên lặng", "조용"},
    STR_VOLUME_OFF:                     {"off", "spento", "apagado", "éteindre", "aus", "wyłączony", "关闭", "kapalı", "выключить", "uit", "вимкнути", "tắt", "꺼"},
    STR_VOLUME_ZERO:                    {"zero", "zero", "cero", "zéro", "null", "zero", "零", "sıfır", "ноль", "nul", "нуль", "không", "영"},
    STR_NAME_IS:                        {" is ", " è ", " es ", " est ", " ist ", " to ", "是", " olan ", " это ", " is ", " це ", " là ", "은 "},
    STR_NAME_IS2:                       {"'s", "sono ", "soy ", "suis ", "bin ", " się ", "的", "'nin", " зовут ", "'s", " звати ", "của", "의 "},
    STR_NAME_IS3:                       {"names", " chiamo ", " llamo ", "appelle ", "werde", "imię", "名字", "adlar", "имя", "namen", "імена", "tên", "이름은"},
    STR_FOR:                            {" for ", " per ", " para ", " pour ", " für ", " dla ", "给", " için ", "для", " voor ", " для ", " cho ", " 위해 "},
    STR_ZERO:                           {"zero", "zero", "cero", "zéro", "null", "zero", "零", "sıfır", "ноль", "nul", "нуль", "không", "영"},
    STR_ONE:                            {"one", "uno", "uno", "un", "eins", "jeden", "一", "bir", "один", "één", "один", "một", "일"},
    STR_TWO:                            {"two", "due", "dos", "deux", "zwei", "dwa", "二", "iki", "два", "twee", "два", "hai", "이"},
    STR_THREE:                          {"three", "tre", "tres", "trois", "drei", "trzy", "三", "üç", "три", "drie", "три", "ba", "삼"},
    STR_FOUR:                           {"four", "quattro", "cuatro", "quatre", "vier", "cztery", "四", "dört", "четыре", "vier", "чотири", "bốn", "사"},
    STR_FIVE:                           {"five", "cinque", "cinco", "cinq", "fünf", "pięć", "五", "beş", "пять", "vijf", "п'ять", "năm", "오"},
    STR_SIX:                            {"six", "sei", "seis", "six", "sechs", "sześć", "六", "altı", "шесть", "zes", "шість", "sáu", "육"},
    STR_SEVEN:                          {"seven", "sette", "siete", "sept", "sieben", "siedem", "七", "yedi", "семь", "zeven", "сім", "bảy", "칠"},
    STR_EIGHT:                          {"eight", "otto", "ocho", "huit", "acht", "osiem", "八", "sekiz", "восемь", "acht", "вісім", "tám", "팔"},
    STR_NINE:                           {"nine", "nove", "nueve", "neuf", "neun", "dziewięć", "九", "dokuz", "девять", "negen", "дев'ять", "chín", "구"},
    STR_TEN:                            {"ten", "dieci", "diez", "dix", "zehn", "dziesięć", "十", "on", "десять", "tien", "десять", "mười", "십"},
    STR_ELEVEN:                         {"eleven", "undici", "once", "onze", "elf", "jedenaście", "十一", "on bir", "одиннадцать", "elf", "одинадцять", "mười một", "십일"},
    STR_TWELVE:                         {"twelve", "dodici", "doce", "douze", "zwölf", "dwanaście", "十二", "on iki", "двенадцать", "twaalf", "дванадцять", "mười hai", "십이"},
    STR_THIRTEEN:                       {"thirteen", "tredici", "trece", "treize", "dreizehn", "trzynaście", "十三", "on üç", "тринадцать", "dertien", "тринадцять", "mười ba", "십삼"},
    STR_FOURTEEN:                       {"fourteen", "quattordici", "catorce", "quatorze", "vierzehn", "czternaście", "十四", "on dört", "четырнадцать", "veertien", "чотирнадцять", "mười bốn", "십사"},
    STR_FIFTEEN:                        {"fifteen", "quindici", "quince", "quinze", "fünfzehn", "piętnaście", "十五", "on beş", "пятнадцать", "vijftien", "п'ятнадцять", "mười lăm", "십오"},
    STR_SIXTEEN:                        {"sixteen", "sedici", "dieciséis", "seize", "sechzehn", "szesnaście", "十六", "on altı", "шестнадцать", "zestien", "шістнадцять", "mười sáu", "십육"},
    STR_SEVENTEEN:                      {"seventeen", "diciassette", "diecisiete", "dix-sept", "siebzehn", "siedemnaście", "十七", "on yedi", "семнадцать", "zeventien", "сімнадцять", "mười bảy", "십칠"},
    STR_EIGHTEEN:                       {"eighteen", "diciotto", "dieciocho", "dix-huit", "achtzehn", "osiemnaście", "十八", "on sekiz", "восемнадцать", "achttien", "вісімнадцять", "mười tám", "십팔"},
    STR_NINETEEN:                       {"nineteen", "diciannove", "diecinueve", "dix-neuf", "neunzehn", "dziewiętnaście", "十九", "on dokuz", "девятнадцать", "negentien", "дев'ятнадцять", "mười chín", "십구"},
    STR_TWENTY:                         {"twenty", "venti", "veinte", "vingt", "zwanzig", "dwadzieścia", "二十", "yirmi", "двадцать", "twintig", "двадцять", "hai mươi", "이십"},
    STR_THIRTY:                         {"thirty", "trenta", "treinta", "trente", "dreißig", "trzydzieści", "三十", "otuz", "тридцать", "dertig", "тридцять", "ba mươi", "삼십"},
    STR_FOURTY:                         {"fourty", "quaranta", "cuarenta", "quarante", "vierzig", "czterdzieści", "四十", "kırk", "сорок", "veertig", "сорок", "bốn mươi", "사십"},
    STR_FIFTY:                          {"fifty", "cinquanta", "cincuenta", "cinquante", "fünfzig", "pięćdziesiąt", "五十", "elli", "пятьдесят", "vijftig", "п'ятдесят", "năm mươi", "오십"},
    STR_SIXTY:                          {"sixty", "sessanta", "sesenta", "soixante", "sechzig", "sześćdziesiąt", "六十", "altmış", "шестьдесят", "zestig", "шістдесят", "sáu mươi", "육십"},
    STR_SEVENTY:                        {"seventy", "settanta", "setenta", "soixante-dix", "siebzig", "siedemdziesiąt", "七十", "yetmiş", "семьдесят", "zeventig", "сімдесят", "bảy mươi", "칠십"},
    STR_EIGHTY:                         {"eighty", "ottanta", "ochenta", "quatre-vingt", "achtzig", "osiemdziesiąt", "八十", "seksen", "восемьдесят", "tachtig", "вісімдесят", "tám mươi", "팔십"},
    STR_NINETY:                         {"ninety", "novanta", "noventa", "quatre-vingt-dix", "neunzig", "dziewięćdziesiąt", "九十", "doksan", "девяносто", "negentig", "дев'яносто", "chín mươi", "구십"},
    STR_ONE_HUNDRED:                    {"one hundred", "cento", "cien", "cent", "einhundert", "sto", "一百", "yüz", "сто", "honderd", "сто", "một trăm", "백"},
    STR_ONE_HOUR:                       {"one hour", "un'ora", "una hora", "une heure", "eine Stunde", "godzina", "一小时", "bir saat", "один час", "een uur", "година", "một giờ", "한 시간"},
    STR_ONE_HOUR_ALT:                   {"an hour", "un'ora", "una hora", "une heure", "eine Stunde", "godzina", "一小时", "bir saat", "час", "een uur", "година", "một giờ", "한 시간"},
    STR_HOUR:                           {"hour", "ora", "hora", "heure", "Stunde", "godzina", "小时", "saat", "час", "uur", "година", "giờ", "시간"},
    STR_MINUTE:                         {"minute", "minuto", "minuto", "minute", "Minute", "minuta", "分钟", "dakika", "минута", "minuut", "хвилина", "phút", "분"},
    STR_SECOND:                         {"second", "secondo", "segundo", "seconde", "Sekunde", "sekunda", "秒", "saniye", "секунда", "seconde", "секунда", "giây", "초"},
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
		} else if vars.APIConfig.STT.Language == "ko-KR" {
			return data[12]
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
