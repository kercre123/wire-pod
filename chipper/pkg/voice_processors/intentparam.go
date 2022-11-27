package wirepod

func paramChecker(req interface{}, intent string, speechText string, justThisBotNum int, botSerial string) {
	if sttLanguage == "en-US" {
		paramCheckerEnUS(req, intent, speechText, justThisBotNum, botSerial)
	} else if sttLanguage == "it-IT" {
		paramCheckerItIT(req, intent, speechText, justThisBotNum, botSerial)
	}
}

func prehistoricParamChecker(req interface{}, intent string, speechText string, justThisBotNum int, botSerial string) {
	if sttLanguage == "en-US" {
		prehistoricParamCheckerEnUS(req, intent, speechText, justThisBotNum, botSerial)
	} else if sttLanguage == "it-IT" {
		prehistoricParamCheckerItIT(req, intent, speechText, justThisBotNum, botSerial)
	}
}
