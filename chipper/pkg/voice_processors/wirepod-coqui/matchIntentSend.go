package wirepod

import (
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/chipper/pkg/vtt"
)

type intentsStruct []struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Utterances  []string `json:"utterances"`
	Intent      string   `json:"intent"`
	Params      struct {
		ParamName  string `json:"paramname"`
		ParamValue string `json:"paramvalue"`
	} `json:"params"`
	Exec     string   `json:"exec"`
	ExecArgs []string `json:"execargs"`
}

func IntentPass(req interface{}, intentThing string, speechText string, intentParams map[string]string, isParam bool, justThisBotNum int) (interface{}, error) {
	var req1 *vtt.IntentRequest
	var req2 *vtt.IntentGraphRequest
	var isIntentGraph bool
	if str, ok := req.(*vtt.IntentRequest); ok {
		req1 = str
		isIntentGraph = false
	} else if str, ok := req.(*vtt.IntentGraphRequest); ok {
		req2 = str
		isIntentGraph = true
	}
	intent := pb.IntentResponse{
		IsFinal: true,
		IntentResult: &pb.IntentResult{
			QueryText:  speechText,
			Action:     intentThing,
			Parameters: intentParams,
		},
	}
	intentGraphSend := pb.IntentGraphResponse{
		ResponseType: pb.IntentGraphMode_INTENT,
		IsFinal:      true,
		IntentResult: &pb.IntentResult{
			QueryText:  speechText,
			Action:     intentThing,
			Parameters: intentParams,
		},
		CommandType: pb.RobotMode_VOICE_COMMAND.String(),
	}
	if !isIntentGraph {
		if err := req1.Stream.Send(&intent); err != nil {
			return nil, err
		}
		r := &vtt.IntentResponse{
			Intent: &intent,
		}
		logger("Bot " + strconv.Itoa(justThisBotNum) + " Intent Sent: " + intentThing)
		if isParam {
			logger("Bot "+strconv.Itoa(justThisBotNum)+" Parameters Sent:", intentParams)
		} else {
			logger("No Parameters Sent")
		}
		return r, nil
	} else {
		if err := req2.Stream.Send(&intentGraphSend); err != nil {
			return nil, err
		}
		r := &vtt.IntentGraphResponse{
			Intent: &intentGraphSend,
		}
		logger("Bot " + strconv.Itoa(justThisBotNum) + " Intent Sent: " + intentThing)
		if isParam {
			logger("Bot "+strconv.Itoa(justThisBotNum)+" Parameters Sent:", intentParams)
		} else {
			logger("No Parameters Sent")
		}
		return r, nil
	}
}

func customIntentHandler(req interface{}, voiceText string, intentList []string, isOpus bool, justThisBotNum int, botSerial string) bool {
	var successMatched bool = false
	if _, err := os.Stat("./customIntents.json"); err == nil {
		var customIntentJSON intentsStruct
		customIntentJSONFile, err := os.ReadFile("./customIntents.json")
		json.Unmarshal(customIntentJSONFile, &customIntentJSON)
		for _, c := range customIntentJSON {
			for _, v := range c.Utterances {
				if strings.Contains(voiceText, strings.ToLower(strings.TrimSpace(v))) {
					logger("Custom Intent Matched: " + c.Name + " - " + c.Description + " - " + c.Intent)
					var intentParams map[string]string
					var isParam bool = false
					if c.Params.ParamValue != "" {
						logger("Custom Intent Parameter: " + c.Params.ParamName + " - " + c.Params.ParamValue)
						intentParams = map[string]string{c.Params.ParamName: c.Params.ParamValue}
						isParam = true
					}
					var args []string
					for _, arg := range c.ExecArgs {
						if arg == "!botSerial" {
							arg = botSerial
						}
						args = append(args, arg)
					}
					var customIntentExec *exec.Cmd
					if len(args) == 0 {
						logger("Executing: " + c.Exec)
						customIntentExec = exec.Command(c.Exec)
					} else {
						logger("Executing: " + c.Exec + " " + strings.Join(args, " "))
						customIntentExec = exec.Command(c.Exec, args...)
					}
					customOut, err := customIntentExec.Output()
					if err != nil {
						logger(err)
					}
					logger("Custom Intent Exec Output: " + strings.TrimSpace(string(customOut)))
					IntentPass(req, c.Intent, voiceText, intentParams, isParam, justThisBotNum)
					successMatched = true
					break
				}
				if successMatched {
					break
				}
			}
			if successMatched {
				break
			}
		}
		if err != nil {
			logger(err)
		}

	}
	return successMatched
}

func processTextAll(req interface{}, voiceText string, listOfLists [][]string, intentList []string, isOpus bool, justThisBotNum int) bool {
	var botSerial string
	var req2 *vtt.IntentRequest
	var req1 *vtt.KnowledgeGraphRequest
	var req3 *vtt.IntentGraphRequest
	if str, ok := req.(*vtt.IntentRequest); ok {
		req2 = str
		botSerial = req2.Device
	} else if str, ok := req.(*vtt.KnowledgeGraphRequest); ok {
		req1 = str
		botSerial = req1.Device
	} else if str, ok := req.(*vtt.IntentGraphRequest); ok {
		req3 = str
		botSerial = req3.Device
	}
	var matched int = 0
	var intentNum int = 0
	var successMatched bool = false
	customIntentMatched := customIntentHandler(req, voiceText, intentList, isOpus, justThisBotNum, botSerial)
	if !customIntentMatched {
		for _, b := range listOfLists {
			for _, c := range b {
				if strings.Contains(voiceText, c) {
					if isOpus {
						paramChecker(req, intentList[intentNum], voiceText, justThisBotNum, botSerial)
					} else {
						prehistoricParamChecker(req, intentList[intentNum], voiceText, justThisBotNum, botSerial)
					}
					successMatched = true
					matched = 1
					break
				}
			}
			if matched == 1 {
				matched = 0
				break
			}
			intentNum = intentNum + 1
		}
	} else {
		successMatched = true
	}
	return successMatched
}
