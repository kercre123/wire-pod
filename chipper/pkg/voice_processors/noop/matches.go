package noop

// This is where you can add intents and more possible utterances for intents

var meetVictorList = []string{"name is", "native is", "names", "name's"}
var nameAskList = []string{"my name"}
var eyeColorList = []string{"eye color", "colo", "i call her", "i foller", "icolor", "ecce", "erior", "ichor", "agricola",
	"change", "oracular", "oracle"}
var howOldList = []string{"older", "how old", "old are you", "old or yo"}
var exploreStartList = []string{"start", "plor", "owing", "tailoring", "oding", "oring", "pling"}
var chargerList = []string{"charge", "home", "go to your", "church", "find your ch"}
var sleepList = []string{"flee", "sleep", "sheep"}
var morningList = []string{"morning", "mourning", "mooning", "it bore"}
var nightList = []string{"night", "might"}
var byeList = []string{"good bye", "good by", "good buy", "goodbye"}
var newYearList = []string{"fireworks", "new year", "happy new", "happy to", "have been", "i now you", "no year", "enee",
	"i never", "knew her", "hobhouse", "bennie"}
var holidaysList = []string{"he holds", "christmas", "behold", "holiday"}
var signInAlexaList = []string{"in intellect", "fine in electa", "in alex", "ing alex", "in an elect", "to alex",
	"in angelica"}
var signOutAlexaList = []string{"in outlet", "i now of elea", "out alexa", "out of ale"}
var forwardList = []string{"forward", "for ward", "for word"}
var turnAroundList = []string{"around", "one eighty", "one ate he"}
var turnLeftList = []string{"rn left", "go left", "e left", "ed left", "ernest"}
var turnRightList = []string{"rn right", "go right", "e right", "ernie", "credit", "ed right"}
var rollCubeList = []string{"roll cu", "roll your cu", "all your cu", "roll human", "yorke", "old your he"}
var wheelieList = []string{"pop a w", "polwhele", "olwen", "i wieland", "do a wheel", "doorstone", "thibetan", "powell",
	"welst", "a wheel"}
var fistbumpList = []string{"this pomp", "this pump", "bump", "fistb", "fistf", "this book", "pisto", "with pomp",
	"fison", "first", "fifth", "were fifteen", "if bump", "wisdom", "this bu"}
var blackjackList = []string{"black", "cards", "game"}
var affirmativeList = []string{"yes", "correct", "hit"}
var negativeList = []string{"no", "dont", "stand"}
var photoList = []string{"photo", "foto", "selby", "capture", "picture"}
var praiseList = []string{"good", "awesome", "also", "as some", "of them", "battle", "t rob", "the ro", "amazing", "woodcourt"}
var abuseList = []string{"bad", "that ro", "ad ro", "a root", "hate", "horrible"}
var weatherList = []string{"weather", "whether", "the other"}
var apologizeList = []string{"sorry", "apologize", "apologise", "the tory"}
var backupList = []string{"back", "a way", "away"}
var volumeDownList = []string{"all you down", "volume down", "down volume", "down the volume", "quieter"}
var volumeUpList = []string{"all you up", "volume up", "up volume", "up the volume", "louder"}
var lookAtMeList = []string{"stare", "at me"}
var volumeSpecificList = []string{"all you", "volume", "loudness"}
var shutUpList = []string{"shut up"}
var helloList = []string{"hello", "are you", "high", "below", "little", "follow"}
var comeList = []string{"come", "to me"}
var loveList = []string{"love", "dove"}
var questionList = []string{"question"}
var checkTimerList = []string{"check timer", "check the timer", "check the time her", "check time her",
	"check time her", "check time of her", "checked the timer", "checked the time her", "checked the time of her"}
var stopTimerList = []string{"stop timer", "stop clock", "stop be", "stopped t", "stopped be", "stopped at", "stop the"}
var timerList = []string{"timer", "time for", "time of for", "time or"}
var timeList = []string{"time is it", "the time", "what time", "time of", "clock"}
var quietList = []string{"quiet", "stop"}
var danceList = []string{"dance", "dancing", "thence"}
var pickUpList = []string{"pickup", "pick up", "bring to me", "bring me", "the beat", "boogie"}
var fetchCubeList = []string{"fetch your cu", "fetch cu", "fetch the cu"}
var findCubeList = []string{"your cu", "the cu"}
var trickList = []string{"trick", "something cool", "some thing cool"}

// make sure intentsList perfectly matches up with matchListList

var intentsList = []string{"intent_names_username_extend", "intent_names_ask", "intent_imperative_eyecolor",
	"intent_character_age", "intent_explore_start", "intent_system_charger", "intent_system_sleep",
	"intent_greeting_goodmorning", "intent_greeting_goodnight", "intent_greeting_goodbye", "intent_seasonal_happynewyear",
	"intent_seasonal_happyholidays", "intent_amazon_signin", "intent_amazon_signin", "intent_imperative_forward",
	"intent_imperative_turnaround", "intent_imperative_turnleft", "intent_imperative_turnright", "intent_play_rollcube",
	"intent_play_popawheelie", "intent_play_fistbump", "intent_play_blackjack", "intent_imperative_affirmative",
	"intent_imperative_negative", "intent_photo_take_extend", "intent_imperative_praise", "intent_imperative_abuse",
	"intent_weather_extend", "intent_imperative_apologize", "intent_imperative_backup", "intent_imperative_volumedown",
	"intent_imperative_volumeup", "intent_imperative_lookatme", "intent_imperative_volumelevel_extend",
	"intent_imperative_shutup", "intent_greeting_hello", "intent_imperative_come", "intent_imperative_love",
	"intent_knowledge_promptquestion", "intent_clock_checktimer", "intent_global_stop_extend", "intent_clock_settimer_extend",
	"intent_clock_time", "intent_imperative_quiet", "intent_imperative_dance", "intent_play_pickupcube",
	"intent_imperative_fetchcube", "intent_imperative_findcube", "intent_play_anytrick"}

var matchListList = [][]string{meetVictorList, nameAskList, eyeColorList, howOldList, exploreStartList,
	chargerList, sleepList, morningList, nightList, byeList,
	newYearList, holidaysList, signInAlexaList, signOutAlexaList, forwardList, turnAroundList, turnLeftList,
	turnRightList, rollCubeList, wheelieList, fistbumpList, blackjackList, affirmativeList,
	negativeList, photoList, praiseList, abuseList, weatherList, apologizeList,
	backupList, volumeDownList, volumeUpList, lookAtMeList, volumeSpecificList,
	shutUpList, helloList, comeList, loveList, questionList, checkTimerList, stopTimerList,
	timerList, timeList, quietList, danceList, pickUpList, fetchCubeList, findCubeList, trickList}
