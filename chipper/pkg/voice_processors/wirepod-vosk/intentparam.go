package wirepod

func paramCheckerSlots(req interface{}, intent string, slots map[string]string, isOpus bool, justThisBotNum int, botSerial string) {
	if sttLanguage=="en-US" {
	    paramCheckerSlotsEnUS(req, intent, slots, isOpus, justThisBotNum, botSerial)
	} else if sttLanguage=="it-IT" {
	    paramCheckerSlotsItIT(req, intent, slots, isOpus, justThisBotNum, botSerial)
	}
}

func paramChecker(req interface{}, intent string, speechText string, justThisBotNum int, botSerial string) {
	if sttLanguage=="en-US" {
	    paramCheckerEnUS(req, intent, speechText, justThisBotNum, botSerial)
	} else if sttLanguage=="it-IT" {
	    paramCheckerItIT(req, intent, speechText, justThisBotNum, botSerial)
	}
}

func prehistoricParamChecker(req interface{}, intent string, speechText string, justThisBotNum int, botSerial string) {
	if sttLanguage=="en-US" {
	    prehistoricParamCheckerEnUS(req, intent, speechText, justThisBotNum, botSerial)
	} else if sttLanguage=="it-IT" {
	    prehistoricParamCheckerItIT(req, intent, speechText, justThisBotNum, botSerial)
	}
}
