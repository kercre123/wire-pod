package wirepod_ttr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/vars"
	"github.com/kercre123/chipper/pkg/vtt"
	sr "github.com/kercre123/chipper/pkg/wirepod/speechrequest"
)

type systemIntentResponseStruct struct {
	Status       string `json:"status"`
	ReturnIntent string `json:"returnIntent"`
}

func IntentPass(req interface{}, intentThing string, speechText string, intentParams map[string]string, isParam bool, justThisBotNum int) (interface{}, error) {
	var esn string
	var req1 *vtt.IntentRequest
	var req2 *vtt.IntentGraphRequest
	var isIntentGraph bool
	if str, ok := req.(*vtt.IntentRequest); ok {
		req1 = str
		esn = req1.Device
		isIntentGraph = false
	} else if str, ok := req.(*vtt.IntentGraphRequest); ok {
		req2 = str
		esn = req2.Device
		isIntentGraph = true
	}
	var intentResult pb.IntentResult
	if isParam {
		intentResult = pb.IntentResult{
			QueryText:  speechText,
			Action:     intentThing,
			Parameters: intentParams,
		}
	} else {
		intentResult = pb.IntentResult{
			QueryText: speechText,
			Action:    intentThing,
		}
	}
	logger.LogUI("Intent matched: " + intentThing + ", transcribed text: '" + speechText + "', device: " + esn)
	if isParam {
		logger.LogUI("Parameters sent: " + fmt.Sprint(intentParams))
	}
	intent := pb.IntentResponse{
		IsFinal:      true,
		IntentResult: &intentResult,
	}
	intentGraphSend := pb.IntentGraphResponse{
		ResponseType: pb.IntentGraphMode_INTENT,
		IsFinal:      true,
		IntentResult: &intentResult,
		CommandType:  pb.RobotMode_VOICE_COMMAND.String(),
	}
	if !isIntentGraph {
		if err := req1.Stream.Send(&intent); err != nil {
			return nil, err
		}
		r := &vtt.IntentResponse{
			Intent: &intent,
		}
		logger.Println("Bot " + strconv.Itoa(justThisBotNum) + " Intent Sent: " + intentThing)
		if isParam {
			logger.Println("Bot "+strconv.Itoa(justThisBotNum)+" Parameters Sent:", intentParams)
		} else {
			logger.Println("No Parameters Sent")
		}
		return r, nil
	} else {
		if err := req2.Stream.Send(&intentGraphSend); err != nil {
			return nil, err
		}
		r := &vtt.IntentGraphResponse{
			Intent: &intentGraphSend,
		}
		logger.Println("Bot " + strconv.Itoa(justThisBotNum) + " Intent Sent: " + intentThing)
		if isParam {
			logger.Println("Bot "+strconv.Itoa(justThisBotNum)+" Parameters Sent:", intentParams)
		} else {
			logger.Println("No Parameters Sent")
		}
		return r, nil
	}
}

func customIntentHandler(req interface{}, voiceText string, intentList []string, isOpus bool, justThisBotNum int, botSerial string) bool {
	var successMatched bool = false
	if vars.CustomIntentsExist {
		for _, c := range vars.CustomIntents {
			for _, v := range c.Utterances {
				//if strings.Contains(voiceText, strings.ToLower(strings.TrimSpace(v))) {
				// Check whether the custom sentence is either at the end of the spoken text or space-separated...
				var seekText = strings.ToLower(strings.TrimSpace(v))
				// System intents can also match any utterances (*)
				if (c.IsSystemIntent && strings.HasPrefix(seekText, "*")) ||
					strings.HasSuffix(voiceText, seekText) || strings.Contains(voiceText, seekText+" ") {
					logger.Println("Custom Intent Matched: " + c.Name + " - " + c.Description + " - " + c.Intent)
					var intentParams map[string]string
					var isParam bool = false
					if c.Params.ParamValue != "" {
						logger.Println("Custom Intent Parameter: " + c.Params.ParamName + " - " + c.Params.ParamValue)
						intentParams = map[string]string{c.Params.ParamName: c.Params.ParamValue}
						isParam = true
					}
					var args []string
					for _, arg := range c.ExecArgs {
						if arg == "!botSerial" {
							arg = botSerial
						} else if arg == "!speechText" {
							arg = "\"" + voiceText + "\""
						} else if arg == "!intentName" {
							arg = c.Name
						} else if arg == "!locale" {
							arg = sr.SttLanguage
						}
						args = append(args, arg)
					}
					var customIntentExec *exec.Cmd
					if len(args) == 0 {
						logger.Println("Executing: " + c.Exec)
						customIntentExec = exec.Command(c.Exec)
					} else {
						logger.Println("Executing: " + c.Exec + " " + strings.Join(args, " "))
						customIntentExec = exec.Command(c.Exec, args...)
					}
					var out bytes.Buffer
					var stderr bytes.Buffer
					customIntentExec.Stdout = &out
					customIntentExec.Stderr = &stderr
					err := customIntentExec.Run()
					if err != nil {
						fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
					}
					logger.Println("Custom Intent Exec Output: " + strings.TrimSpace(string(out.String())))

					if c.IsSystemIntent {
						// A system intent returns its output in json format
						var resp systemIntentResponseStruct
						err := json.Unmarshal(out.Bytes(), &resp)
						if err == nil && resp.Status == "ok" {
							logger.Println("System intent parsed and executed successfully")
							IntentPass(req, resp.ReturnIntent, voiceText, intentParams, isParam, justThisBotNum)
							successMatched = true
						}
					} else {
						IntentPass(req, c.Intent, voiceText, intentParams, isParam, justThisBotNum)
						successMatched = true
					}
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
	}
	return successMatched
}

func pluginFunctionHandler(req interface{}, voiceText string, justThisBotNum int, botSerial string) bool {
	matched := false
	var intent string
	for num, array := range PluginUtterances {
		array := array
		for _, str := range *array {
			if strings.Contains(voiceText, str) {
				logger.Println("Text matched plugin " + PluginNames[num] + ", executing function")
				intent = PluginFunctions[num](voiceText, botSerial)
				if intent == "" {
					intent = "intent_imperative_praise"
				}
				IntentPass(req, intent, voiceText, make(map[string]string), false, justThisBotNum)
				matched = true
				break
			}
		}
		if matched {
			break
		}
	}
	return matched
}

func ProcessTextAll(req interface{}, voiceText string, listOfLists [][]string, intentList []string, isOpus bool, justThisBotNum int) bool {
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
	voiceText = strings.ToLower(voiceText)
	pluginMatched := pluginFunctionHandler(req, voiceText, justThisBotNum, botSerial)
	customIntentMatched := customIntentHandler(req, voiceText, intentList, isOpus, justThisBotNum, botSerial)
	if !customIntentMatched && !pluginMatched {
		logger.Println("Not a custom intent")
		// Look for a perfect match first
		for _, b := range listOfLists {
			for _, c := range b {
				if voiceText == strings.ToLower(c) {
					logger.Println("Perfect match for intent " + intentList[intentNum] + " (" + strings.ToLower(c) + ")")
					if isOpus {
						ParamChecker(req, intentList[intentNum], voiceText, justThisBotNum, botSerial)
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
		// Not found? Then let's be happy with a bare substring search
		if !successMatched {
			intentNum = 0
			matched = 0
			for _, b := range listOfLists {
				for _, c := range b {
					if strings.Contains(voiceText, strings.ToLower(c)) {
						logger.Println("Partial match for intent " + intentList[intentNum] + " (" + strings.ToLower(c) + ")")
						if isOpus {
							ParamChecker(req, intentList[intentNum], voiceText, justThisBotNum, botSerial)
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
		}
	} else {
		logger.Println("This is a custom intent or plugin!")
		successMatched = true
	}
	return successMatched
}
