package processreqs

import (
	"encoding/json"
	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/vars"
	"github.com/kercre123/chipper/pkg/vtt"
	sr "github.com/kercre123/chipper/pkg/wirepod/speechrequest"
	ttr "github.com/kercre123/chipper/pkg/wirepod/ttr"
)

func (s *Server) ProcessIntentGraph(req *vtt.IntentGraphRequest) (*vtt.IntentGraphResponse, error) {
	var successMatched bool
	speechReq := sr.ReqToSpeechRequest(req)
	var transcribedText string
	if !isSti {
		var err error
		transcribedText, err = sttHandler(speechReq)
		if err != nil {
			ttr.IntentPass(req, "intent_system_noaudio", "voice processing error", map[string]string{"error": err.Error()}, true)
			return nil, nil
		}
		successMatched = ttr.ProcessTextAll(req, transcribedText, vars.MatchListList, vars.IntentsList, speechReq.IsOpus)
	} else {
		intent, slots, err := stiHandler(speechReq)
		if err != nil {
			if err.Error() == "inference not understood" {
				logger.Println("No intent was matched")
				ttr.IntentPass(req, "intent_system_noaudio", "voice processing error", map[string]string{"error": err.Error()}, true)
				return nil, nil
			}
			logger.Println(err)
			ttr.IntentPass(req, "intent_system_noaudio", "voice processing error", map[string]string{"error": err.Error()}, true)
			return nil, nil
		}
		ttr.ParamCheckerSlotsEnUS(req, intent, slots, speechReq.IsOpus, speechReq.Device)
		return nil, nil
	}
	if !successMatched {
		logger.Println("No intent was matched.")
		if vars.APIConfig.Knowledge.Enable && vars.APIConfig.Knowledge.Provider == "openai" && len([]rune(transcribedText)) >= 8 {
			apiResponse := openaiRequest(transcribedText)
			if apiResponse.FunctionCall != nil {
				arguments := map[string]string{}
				err := json.Unmarshal([]byte(apiResponse.FunctionCall.Arguments), &arguments)
				if err != nil {
					logger.Println("Error unmarshalling arguments")
					logger.Println(err.Error())
				}
				logger.Println("Passing intent to ParamCheckerSlotsEnUS: ", apiResponse.FunctionCall.Name, "with arguments: ", arguments)
				ttr.ParamCheckerSlotsEnUS(req, apiResponse.FunctionCall.Name, arguments, speechReq.IsOpus, speechReq.Device)
			} else {
				response := &pb.IntentGraphResponse{
					Session:      req.Session,
					DeviceId:     req.Device,
					ResponseType: pb.IntentGraphMode_KNOWLEDGE_GRAPH,
					SpokenText:   apiResponse.Message,
					QueryText:    transcribedText,
					IsFinal:      true,
				}
				err := req.Stream.Send(response)
				if err != nil {
					logger.Println("Error sending IntentGraphResponse")
					logger.Println(err.Error())
				}
			}
			return nil, nil
		}
		ttr.IntentPass(req, "intent_system_noaudio", transcribedText, map[string]string{"": ""}, false)
		return nil, nil
	}
	logger.Println("Bot " + speechReq.Device + " request served.")
	return nil, nil
}
