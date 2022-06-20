package wirepod

import (
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/chipper/pkg/vtt"
	opus "github.com/digital-dream-labs/opus-go/opus"
)

var NoResult string = "NoResultCommand"
var NoResultSpoken string

var botNumKG int = 0

func (s *Server) ProcessKnowledgeGraph(req *vtt.KnowledgeGraphRequest) (*vtt.KnowledgeGraphResponse, error) {
	var voiceTimer int = 0
	var transcription1 string = ""
	var transcription2 string = ""
	var transcription3 string = ""
	var transcription4 string = ""
	var successMatch bool = false
	var transcribedText string
	var die bool = false
	if os.Getenv("DEBUG_LOGGING") != "true" && os.Getenv("DEBUG_LOGGING") != "false" {
		log.Println("No valid value for DEBUG_LOGGING, setting to true")
		debugLogging = true
	} else {
		if os.Getenv("DEBUG_LOGGING") == "true" {
			debugLogging = true
		} else {
			debugLogging = false
		}
	}
	botNumKG = botNumKG + 1
	var justThisbotNumKG int = botNumKG
	if debugLogging == true {
		log.Println("Stream " + strconv.Itoa(botNumKG) + " opened.")
	}
	data := []byte{}
	data = append(data, req.FirstReq.InputAudio...)
	stream := opus.OggStream{}
	go func() {
		time.Sleep(time.Millisecond * 500)
		for voiceTimer < 7 {
			voiceTimer = voiceTimer + 1
			time.Sleep(time.Second * 1)
		}
	}()
	go func() {
		var processOne bool = false
		var processTwo bool = false
		var processThree bool = false
		var processFour bool = false
		time.Sleep(time.Millisecond * 500)
		for voiceTimer < 7 {
			if processOne == false {
				if _, err := os.Stat("/tmp/" + strconv.Itoa(justThisbotNumKG) + "dumped1"); err == nil {
					processOne = true
					process1 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(justThisbotNumKG)+"voice1.wav")
					process1out, err := process1.Output()
					if err != nil {
						//
					}
					transcription1 = strings.TrimSpace(string(process1out))
					log.Println(strconv.Itoa(justThisbotNumKG) + ", 1: " + transcription1)
				}
			}
			if processTwo == false {
				if _, err := os.Stat("/tmp/" + strconv.Itoa(justThisbotNumKG) + "dumped2"); err == nil {
					processTwo = true
					process2 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(justThisbotNumKG)+"voice2.wav")
					process2out, err := process2.Output()
					if err != nil {
						//
					}
					transcription2 = strings.TrimSpace(string(process2out))
					log.Println(strconv.Itoa(justThisbotNumKG) + ", 2: " + transcription2)
				}
			}
			if processThree == false {
				if _, err := os.Stat("/tmp/" + strconv.Itoa(justThisbotNumKG) + "dumped3"); err == nil {
					processThree = true
					process3 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(justThisbotNumKG)+"voice3.wav")
					process3out, err := process3.Output()
					if err != nil {
						//
					}
					transcription3 = strings.TrimSpace(string(process3out))
					log.Println(strconv.Itoa(justThisbotNumKG) + ", 3: " + transcription3)
				}
			}
			if processFour == false {
				if _, err := os.Stat("/tmp/" + strconv.Itoa(justThisbotNumKG) + "dumped4"); err == nil {
					processFour = true
					process4 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(justThisbotNumKG)+"voice4.wav")
					process4out, err := process4.Output()
					if err != nil {
						//
					}
					transcription4 = strings.TrimSpace(string(process4out))
					log.Println(strconv.Itoa(justThisbotNumKG) + ", 4: " + transcription4)
					successMatch = true
				}
			}
		}
	}()
	for {
		chunk, err := req.Stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		if transcription2 != "" {
			if transcription1 == transcription2 {
				log.Println("Speech stopped, 2: " + transcription1)
				transcribedText = transcription1
				die = true
				break
			} else if transcription2 != "" {
				if transcription2 == transcription3 {
					log.Println("Speech stopped, 3: " + transcription2)
					transcribedText = transcription2
					die = true
					break
				} else if transcription3 != "" {
					if transcription3 == transcription4 {
						log.Println("Speech stopped, 4: " + transcription3)
						transcribedText = transcription3
						die = true
						break
					} else if transcription4 != "" {
						if transcription3 == transcription4 {
							log.Println("Speech stopped, 4: " + transcription4)
							transcribedText = transcription4
							die = true
							break
						} else {
							log.Println("Speech stopped, 4 (nm): " + transcription4)
							transcribedText = transcription4
							die = true
							break
						}
					}
				}
			}
		}
		if transcription2 == "" && transcription3 != "" {
			if transcription4 != "" {
				if transcription3 == transcription4 {
					log.Println("Speech stopped, 4: " + transcription4)
					transcribedText = transcription4
					die = true
					break
				} else {
					log.Println("Speech stopped, 4 (nm): " + transcription4)
					transcribedText = transcription4
					die = true
					break
				}
			}
		}
		if transcription4 == "" && successMatch == true {
			transcribedText = ""
			break
		}
		data = append(data, chunk.InputAudio...)
		go bytesToInt(stream, data, justThisbotNumKG, voiceTimer, die)
	}
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voiceKG.pcm").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voiceKG1.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voiceKG2.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voiceKG3.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voiceKG4.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumpedKG1").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumpedKG2").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumpedKG3").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumpedKG4").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voice.pcm").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voice1.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voice2.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voice3.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"voice4.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumped1").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumped2").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumped3").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNumKG)+"dumped4").Run()
	NoResultSpoken = "This is a placeholder! You said: " + transcribedText
	kg := pb.KnowledgeGraphResponse{
		Session:     req.Session,
		DeviceId:    req.Device,
		CommandType: NoResult,
		SpokenText:  NoResultSpoken,
	}
	botNumKG = botNumKG - 1
	if err := req.Stream.Send(&kg); err != nil {
		return nil, err
	}
	return &vtt.KnowledgeGraphResponse{
		Intent: &kg,
	}, nil

}
