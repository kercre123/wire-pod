package noop

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/chipper/pkg/vtt"
)

var debugLoggingKG bool

var NoResult string = "NoResultCommand"
var NoResultSpoken string

var botNumKG int = 0
var matchedKG int = 0

// ProcessKnowledgeGraph handles knowledge graph interactions
func (s *Server) ProcessKnowledgeGraph(req *vtt.KnowledgeGraphRequest) (*vtt.KnowledgeGraphResponse, error) {
	if os.Getenv("DEBUG_LOGGING") != "true" && os.Getenv("DEBUG_LOGGING") != "false" {
		log.Println("No valid value for DEBUG_LOGGING, setting to true")
		debugLoggingKG = true
	} else {
		if os.Getenv("DEBUG_LOGGING") == "true" {
			debugLoggingKG = true
		} else {
			debugLoggingKG = false
		}
	}
	var finished1 string
	var finished2 string
	var finished3 string
	var finished4 string
	var transcribedText string
	matchedKG = 0
	botNumKG = botNumKG + 1
	if debugLoggingKG == true {
		log.Println("KG: Stream " + strconv.Itoa(botNumKG) + " opened.")
	}
	f, err := os.Create("/tmp/" + strconv.Itoa(botNumKG) + "voicekg.ogg")
	check(err)
	cmd1 := exec.Command("/bin/bash", "../stt.sh", strconv.Itoa(botNumKG), "kg")
	data := []byte{}
	data = append(data, req.FirstReq.InputAudio...)
	cmd1.Run()
	f.Write(data)
	for {
		chunk, err := req.Stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("EOF error")
				break
			}
		}
		data = append(data, chunk.InputAudio...)
		f.Write(chunk.InputAudio)
		fileBytes1, _ := ioutil.ReadFile("/tmp/" + strconv.Itoa(botNumKG) + "utterance1kg")
		transcribedText1 := strings.TrimSpace(string(fileBytes1))
		if _, err := os.Stat("/tmp/" + strconv.Itoa(botNumKG) + "utterance1kg"); err == nil {
			finished1 = transcribedText1
		}
		if _, err := os.Stat("./slowsys"); err == nil {
			if _, err := os.Stat("/tmp/" + strconv.Itoa(botNumKG) + "sttDonekg"); err == nil {
				transcribedText = finished1
				if debugLoggingKG == true {
					log.Println("KG 1: Speech has stopped, transcribed text is: " + finished1)
				}
				break
			}
		} else {
			fileBytes2, _ := ioutil.ReadFile("/tmp/" + strconv.Itoa(botNumKG) + "utterance2kg")
			transcribedText2 := strings.TrimSpace(string(fileBytes2))
			if _, err := os.Stat("/tmp/" + strconv.Itoa(botNumKG) + "utterance2kg"); err == nil {
				finished2 = transcribedText2
				if finished1 == finished2 {
					transcribedText = finished2
					if debugLoggingKG == true {
						log.Println("KG 2: Speech has stopped, transcribed text is: " + finished2)
					}
					break
				}
			}
			fileBytes3, _ := ioutil.ReadFile("/tmp/" + strconv.Itoa(botNumKG) + "utterance3kg")
			transcribedText3 := strings.TrimSpace(string(fileBytes3))
			if _, err := os.Stat("/tmp/" + strconv.Itoa(botNumKG) + "utterance3kg"); err == nil {
				finished3 = transcribedText3
				if finished2 == finished3 {
					transcribedText = finished3
					if debugLoggingKG == true {
						log.Println("KG 3: Speech has stopped, transcribed text is: " + finished3)
					}
					break
				}
			}
			fileBytes4, _ := ioutil.ReadFile("/tmp/" + strconv.Itoa(botNumKG) + "utterance4kg")
			transcribedText4 := strings.TrimSpace(string(fileBytes4))
			if _, err := os.Stat("/tmp/" + strconv.Itoa(botNumKG) + "utterance4kg"); err == nil {
				finished4 = transcribedText4
				if finished3 == finished4 {
					transcribedText = finished4
					if debugLoggingKG == true {
						log.Println("KG 4: Speech has stopped, transcribed text is: " + finished4)
					}
					break
				} else if _, err := os.Stat("/tmp/" + strconv.Itoa(botNumKG) + "sttDonekg"); err == nil {
					transcribedText = finished4
					if debugLoggingKG == true {
						log.Println("KG 4 (nm): Speech has stopped, transcribed text is: " + finished4)
					}
					break
				}
			}
		}
	}
	NoResultSpoken = "This is a placeholder! You said: " + transcribedText
	kg := pb.KnowledgeGraphResponse{
		Session:     req.Session,
		DeviceId:    req.Device,
		CommandType: NoResult,
		SpokenText:  NoResultSpoken,
	}

	if err := req.Stream.Send(&kg); err != nil {
		return nil, err
	}
	botNumKG = botNumKG - 1
	return &vtt.KnowledgeGraphResponse{
		Intent: &kg,
	}, nil

}
