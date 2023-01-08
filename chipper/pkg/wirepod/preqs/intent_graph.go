package processreqs

import (
	"os"
	"strconv"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/vtt"
	sr "github.com/kercre123/chipper/pkg/wirepod/speechrequest"
	ttr "github.com/kercre123/chipper/pkg/wirepod/ttr"
)

func (s *Server) ProcessIntentGraph(req *vtt.IntentGraphRequest) (*vtt.IntentGraphResponse, error) {
	sr.BotNum = sr.BotNum + 1
	var successMatched bool
	speechReq := sr.ReqToSpeechRequest(req)
	var transcribedText string

	if !isSti {
		var err error
		transcribedText, err = sttHandler(speechReq)
		if err != nil {
			sr.BotNum = sr.BotNum - 1
			ttr.IntentPass(req, "intent_system_noaudio", "voice processing error", map[string]string{"error": err.Error()}, true, speechReq.BotNum)
			return nil, nil
		}
		successMatched = ttr.ProcessTextAll(req, transcribedText, matchListList, intentsList, speechReq.IsOpus, speechReq.BotNum)
	} else {
		intent, slots, err := stiHandler(speechReq)
		if err != nil {
			if err.Error() == "inference not understood" {
				logger.Println("No intent was matched")
				sr.BotNum = sr.BotNum - 1
				ttr.IntentPass(req, "intent_system_noaudio", "voice processing error", map[string]string{"error": err.Error()}, true, speechReq.BotNum)
				return nil, nil
			}
			logger.Println(err)
			sr.BotNum = sr.BotNum - 1
			ttr.IntentPass(req, "intent_system_noaudio", "voice processing error", map[string]string{"error": err.Error()}, true, speechReq.BotNum)
			return nil, nil
		}
		ttr.ParamCheckerSlotsEnUS(req, intent, slots, speechReq.IsOpus, speechReq.BotNum, speechReq.Device)
		sr.BotNum = sr.BotNum - 1
		return nil, nil
	}

	if !successMatched {
		logger.Println("No intent was matched.")
		if os.Getenv("KNOWLEDGE_INTENT_GRAPH") == "true" && len([]rune(transcribedText)) > 8 {
			apiResponse, err := openaiRequest(transcribedText)
			if err != nil {
				sr.BotNum = sr.BotNum - 1
				ttr.IntentPass(req, "intent_system_noaudio", transcribedText, map[string]string{"": ""}, false, speechReq.BotNum)
				return nil, nil
			}
			sr.BotNum = sr.BotNum - 1
			response := &pb.IntentGraphResponse{
				Session:      req.Session,
				DeviceId:     req.Device,
				ResponseType: pb.IntentGraphMode_KNOWLEDGE_GRAPH,
				SpokenText:   apiResponse,
				QueryText:    transcribedText,
				IsFinal:      true,
			}
			req.Stream.Send(response)
			return nil, nil
		}
		sr.BotNum = sr.BotNum - 1
		ttr.IntentPass(req, "intent_system_noaudio", transcribedText, map[string]string{"": ""}, false, speechReq.BotNum)
		return nil, nil
	}
	sr.BotNum = sr.BotNum - 1
	logger.Println("Bot " + strconv.Itoa(speechReq.BotNum) + " request served.")
	return nil, nil
}
