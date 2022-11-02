package wirepod

// This is where you can add intents and more possible utterances for intents

var meetVictorList []string
var weatherList []string
var nameAskList []string
var eyeColorList []string
var howOldList []string
var exploreStartList []string
var chargerList []string
var sleepList []string
var morningList []string
var nightList []string
var byeList []string
var newYearList []string
var holidaysList []string
var signInAlexaList []string
var signOutAlexaList []string
var forwardList []string
var turnAroundList []string
var turnLeftList []string
var turnRightList []string
var rollCubeList []string
var wheelieList []string
var fistbumpList []string
var blackjackList []string
var affirmativeList []string
var negativeList []string
var photoList []string
var praiseList []string
var abuseList []string
var apologizeList []string
var backupList []string
var volumeDownList []string
var volumeUpList []string
var lookAtMeList []string
var volumeSpecificList []string
var shutUpList []string
var helloList []string
var comeList []string
var loveList []string
var questionList []string
var checkTimerList []string
var stopTimerList []string
var timerList []string
var timeList []string
var quietList []string
var danceList []string
var pickUpList []string
var fetchCubeList []string
var findCubeList []string
var trickList []string
var recordMessageList []string
var playMessageList []string
var blackjackHitList []string
var blackjackStandList []string
var keepawayList []string

var matchListList [][]string

// make sure intentsList perfectly matches up with matchListList
var intentsList = []string{"intent_names_username_extend", "intent_weather_extend", "intent_names_ask", "intent_imperative_eyecolor",
	"intent_character_age", "intent_explore_start", "intent_system_charger", "intent_system_sleep",
	"intent_greeting_goodmorning", "intent_greeting_goodnight", "intent_greeting_goodbye", "intent_seasonal_happynewyear",
	"intent_seasonal_happyholidays", "intent_amazon_signin", "intent_amazon_signin", "intent_imperative_forward",
	"intent_imperative_turnaround", "intent_imperative_turnleft", "intent_imperative_turnright", "intent_play_rollcube",
	"intent_play_popawheelie", "intent_play_fistbump", "intent_play_blackjack", "intent_imperative_affirmative",
	"intent_imperative_negative", "intent_photo_take_extend", "intent_imperative_praise", "intent_imperative_abuse",
	"intent_imperative_apologize", "intent_imperative_backup", "intent_imperative_volumedown",
	"intent_imperative_volumeup", "intent_imperative_lookatme", "intent_imperative_volumelevel_extend",
	"intent_imperative_shutup", "intent_greeting_hello", "intent_imperative_come", "intent_imperative_love",
	"intent_knowledge_promptquestion", "intent_clock_checktimer", "intent_global_stop_extend", "intent_clock_settimer_extend",
	"intent_clock_time", "intent_imperative_quiet", "intent_imperative_dance", "intent_play_pickupcube",
	"intent_imperative_fetchcube", "intent_imperative_findcube", "intent_play_anytrick", "intent_message_recordmessage_extend",
	"intent_message_playmessage_extend", "intent_blackjack_hit", "intent_blackjack_stand", "intent_play_keepaway"}

func initMatches() {
	if sttLanguage=="en-US" {
		meetVictorList = []string{"name is", "native is", "names", "name's"}
		weatherList = []string{"what's the weather", "weather", "whether", "the other", "the water", "no other"}
		nameAskList = []string{"my name"}
		eyeColorList = []string{"eye color", "colo", "i call her", "i foller", "icolor", "ecce", "erior", "ichor", "agricola",
			"change", "oracular", "oracle"}
		howOldList = []string{"older", "how old", "old are you", "old or yo", "how there you"}
		exploreStartList = []string{"start", "plor", "owing", "tailoring", "oding", "oring", "pling"}
		chargerList = []string{"charge", "home", "go to your", "church", "find your ch"}
		sleepList = []string{"flee", "sleep", "sheep"}
		morningList = []string{"morning", "mourning", "mooning", "it bore", "afternoon", "after noon", "after whom"}
		nightList = []string{"night", "might"}
		byeList = []string{"good bye", "good by", "good buy", "goodbye"}
		newYearList = []string{"fireworks", "new year", "happy new", "happy to", "have been", "i now you", "no year", "enee",
			"i never", "knew her", "hobhouse", "bennie"}
		holidaysList = []string{"he holds", "christmas", "behold", "holiday"}
		signInAlexaList = []string{"in intellect", "fine in electa", "in alex", "ing alex", "in an elect", "to alex",
			"in angelica", "up alexa"}
		signOutAlexaList = []string{"in outlet", "i now of elea", "out alexa", "out of ale"}
		forwardList = []string{"forward", "for ward", "for word"}
		turnAroundList = []string{"around", "one eighty", "one ate he"}
		turnLeftList = []string{"rn left", "go left", "e left", "ed left", "ernest"}
		turnRightList = []string{"rn right", "go right", "e right", "ernie", "credit", "ed right"}
		rollCubeList = []string{"roll cu", "roll your cu", "all your cu", "roll human", "yorke", "old your he"}
		wheelieList = []string{"pop a w", "polwhele", "olwen", "i wieland", "do a wheel", "doorstone", "thibetan", "powell",
			"welst", "a wheel", "willie", "a really", "o' billy"}
		fistbumpList = []string{"this pomp", "this pump", "bump", "fistb", "fistf", "this book", "pisto", "with pomp",
			"fison", "first", "fifth", "were fifteen", "if bump", "wisdom", "this bu", "fist bomb", "fist ball", "this ball", "system"}
		blackjackList = []string{"black", "cards", "game"}
		affirmativeList = []string{"yes", "correct", "sure"}
		negativeList = []string{"no", "dont"}
		photoList = []string{"photo", "foto", "selby", "capture", "picture"}
		praiseList = []string{"good", "awesome", "also", "as some", "of them", "battle", "t rob", "the ro", "amazing", "woodcourt"}
		abuseList = []string{"bad", "that ro", "ad ro", "a root", "hate", "horrible"}
		apologizeList = []string{"sorry", "apologize", "apologise", "the tory", "nevermind", "never mind"}
		backupList = []string{"back"}
		volumeDownList = []string{"all you down", "volume down", "down volume", "down the volume", "quieter"}
		volumeUpList = []string{"all you up", "volume up", "up volume", "up the volume", "louder"}
		lookAtMeList = []string{"stare", "at me"}
		volumeSpecificList = []string{"all you", "volume", "loudness"}
		shutUpList = []string{"shut up"}
		helloList = []string{"hello", "are you", "high", "below", "little", "follow", "for you", "far you"}
		comeList = []string{"come", "to me"}
		loveList = []string{"love", "dove"}
		questionList = []string{"question", "weston"}
		checkTimerList = []string{"check timer", "check the timer", "check the time her", "check time her",
			"check time her", "check time of her", "checked the timer", "checked the time her", "checked the time of her"}
		stopTimerList = []string{"up the timer", "stop timer", "stop clock", "stop be", "stopped t", "stopped be", "stopped at", "stop the"}
		timerList = []string{"timer", "time for", "time of for", "time or", "time of"}
		timeList = []string{"time is it", "the time", "what time", "clock"}
		quietList = []string{"quiet", "stop"}
		danceList = []string{"dance", "dancing", "thence"}
		pickUpList = []string{"pickup", "pick up", "bring to me", "bring me", "the beat", "boogie"}
		fetchCubeList = []string{"fetch your cu", "fetch cu", "fetch the cu"}
		findCubeList = []string{"your cu", "the cu"}
		trickList = []string{"trick", "something cool", "some thing cool"}
		recordMessageList = []string{"record"}
		playMessageList = []string{"play message", "play method", "play a message", "play a method"}
		blackjackHitList = []string{"hit", "it", "hot"}
		blackjackStandList = []string{"stand", "stan"}
		keepawayList = []string{"keepaway", "keep away"}
	} else if sttLanguage=="it-IT" {
		meetVictorList = []string{"il mio nome è", "mi chiamo", "io sono", "qui c'è"}
		weatherList = []string{"che tempo fa", "com'è il tempo", "com'è fuori"}
		nameAskList = []string{"qual è il mio nome", "come mi chiamo", "chi sono"}
		eyeColorList = []string{"colore degli occhi", "colore agli occhi", "cambia colore", "occhi"}
		howOldList = []string{"quanti anni hai", "qual è la tua età", "quanto sei vecchio"}
		exploreStartList = []string{"vai ad esplorare", "esplora", "vai in esplorazione", "fatti un giro"}
		chargerList = []string{"vai a casa", "a casa", "ricaricati", "mettiti in carica", "trova il caricabatterie", "vai in carica"}
		sleepList = []string{"dormi", "vai a dormire", "a nanna", "vai a nanna", "fai la nanna"}
		morningList = []string{"giorno", "mattina", "pomeriggio"}
		nightList = []string{"notte", "sera"}
		byeList = []string{"ciao", "arrivederci", "ci vediamo"}
		newYearList = []string{"fuochi d'artificio", "buon anno", "buon anno nuovo"}
		holidaysList = []string{"natale", "vacanza", "vacanze", "feste"}
		signInAlexaList = []string{"entra in alexa", "registrati su alexa", "attiva alexa", "accendi alexa"}
		signOutAlexaList = []string{"esci da alexa", "disattiva alexa", "spegni alexa"}
		forwardList = []string{"avanti"}
		turnAroundList = []string{"gira"}
		turnLeftList = []string{"gira a sinistra", "vai a sinistra"}
		turnRightList = []string{"gira a destra", "vai a destra"}
		rollCubeList = []string{"gioca col cubo", "fai rotolare il cubo", "sposta il cubo"}
		wheelieList = []string{"fischia", "fischio", "fischietta"}
		fistbumpList = []string{"dammi cinque", "dammi il cinque"}
		blackjackList = []string{"ventuno", "blackjack"}
		affirmativeList = []string{"si", "giusto", "corretto", "sì", "vero"}
		negativeList = []string{"no", "non", "sbagliato", "falso"}
		photoList = []string{"foto", "selfie", "immagine"}
		praiseList = []string{"bravo", "grande", "mitico", "forte", "impressionante"}
		abuseList = []string{"cattivo", "stupido", "così non va"}
		apologizeList = []string{"mi dispiace", "scusa", "scusami", "sono dispiaciuto"}
		backupList = []string{"indietro", "annulla"}
		volumeDownList = []string{"abbassa il volume"}
		volumeUpList = []string{"alza il volume"}
		lookAtMeList = []string{"guardami"}
		volumeSpecificList = []string{"volume"}
		shutUpList = []string{"zitto", "fai silenzio", "stai zitto"}
		helloList = []string{"ciao", "come stai", "buongiorno", "buonasera", "buon pomeriggio", "ehi", "salve"}
		comeList = []string{"vieni", "da me", "qui"}
		loveList = []string{"ti voglio bene", "sei il mio amore", "ti amo"}
		questionList = []string{"domanda"}
		checkTimerList = []string{"controlla il cronometro", "controlla il timer"}
		stopTimerList = []string{"ferma il cronometro", "ferma il timer", "stoppa il timer"}
		timerList = []string{"timer", "cronometro"}
		timeList = []string{"che ora è", "che ore sono", "qual è l'ora", "che ora fai"}
		quietList = []string{"basta", "fermo", "stai buono", "stai calmo"}
		danceList = []string{"balla", "fai un balletto", "danza"}
		pickUpList = []string{"raccogli"}
		fetchCubeList = []string{"prendi il cubo", "prendi il tuo cubo", "raccogli il cubo"}
		findCubeList = []string{"trova il cubo", "trova il tuo cubo"}
		trickList = []string{"fai qualcosa di figo", "stupiscimi", "impressionami", "inventa qualcosa"}
		recordMessageList = []string{"registra"}
		playMessageList = []string{"riproduci il messaggio", "leggi il messaggio"}
		blackjackHitList = []string{"hit"}
		blackjackStandList = []string{"stand"}
		keepawayList = []string{"stai lontano", "via", "allontanati", "indietro"}
	}

	matchListList = [][]string{meetVictorList, weatherList, nameAskList, eyeColorList, howOldList, exploreStartList,
		chargerList, sleepList, morningList, nightList, byeList,
		newYearList, holidaysList, signInAlexaList, signOutAlexaList, forwardList, turnAroundList, turnLeftList,
		turnRightList, rollCubeList, wheelieList, fistbumpList, blackjackList, affirmativeList,
		negativeList, photoList, praiseList, abuseList, apologizeList,
		backupList, volumeDownList, volumeUpList, lookAtMeList, volumeSpecificList,
		shutUpList, helloList, comeList, loveList, questionList, checkTimerList, stopTimerList,
		timerList, timeList, quietList, danceList, pickUpList, fetchCubeList, findCubeList, trickList,
		recordMessageList, playMessageList, blackjackHitList, blackjackStandList, keepawayList}
} 
