package wirepod

import (
	"github.com/digital-dream-labs/chipper/pkg/logger"
	"github.com/digital-dream-labs/chipper/pkg/vtt"
	"strconv"
)

func (s *Server) ProcessIntent(req *vtt.IntentRequest) (*vtt.IntentResponse, error) {
	var successMatched bool
	transcribedText, transcribedSlots, isRhino, justThisBotNum, isOpus, err := sttHandler(req, false)
	if err != nil {
		IntentPass(req, "intent_system_noaudio", "voice processing error", map[string]string{"error": err.Error()}, true, justThisBotNum)
		return nil, nil
	}
	if isRhino {
		successMatched = true
		paramCheckerSlots(req, transcribedText, transcribedSlots, isOpus, justThisBotNum, req.Device)
	} else {
		successMatched = processTextAll(req, transcribedText, matchListList, intentsList, isOpus, justThisBotNum)
	}
	if !successMatched {
		logger.Log("No intent was matched.")
		IntentPass(req, "intent_system_noaudio", transcribedText, map[string]string{"": ""}, false, justThisBotNum)
	}
	logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " request served.")
	return nil, nil
}
