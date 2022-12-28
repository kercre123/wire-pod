package processreqs

import (
	"strconv"

	"github.com/digital-dream-labs/chipper/pkg/logger"
	sr "github.com/digital-dream-labs/chipper/pkg/speechrequest"
	"github.com/digital-dream-labs/chipper/pkg/vtt"
	ttr "github.com/digital-dream-labs/chipper/pkg/wirepod-ttr"
)

func (s *Server) ProcessIntent(req *vtt.IntentRequest) (*vtt.IntentResponse, error) {
	sr.BotNum = sr.BotNum + 1
	var successMatched bool
	speechReq := sr.ReqToSpeechRequest(req)
	transcribedText, err := sttHandler(speechReq)
	if err != nil {
		sr.BotNum = sr.BotNum - 1
		ttr.IntentPass(req, "intent_system_noaudio", "voice processing error", map[string]string{"error": err.Error()}, true, speechReq.BotNum)
		return nil, nil
	}
	successMatched = ttr.ProcessTextAll(req, transcribedText, matchListList, intentsList, speechReq.IsOpus, speechReq.BotNum)
	if !successMatched {
		logger.Println("No intent was matched.")
		sr.BotNum = sr.BotNum - 1
		ttr.IntentPass(req, "intent_system_noaudio", transcribedText, map[string]string{"": ""}, false, speechReq.BotNum)
		return nil, nil
	}
	sr.BotNum = sr.BotNum - 1
	logger.Println("Bot " + strconv.Itoa(speechReq.BotNum) + " request served.")
	return nil, nil
}
