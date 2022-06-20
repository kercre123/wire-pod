package wirepod

import (
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/digital-dream-labs/chipper/pkg/vtt"

	opus "github.com/digital-dream-labs/opus-go/opus"
)

var debugLogging bool
var slowSys bool = false

var botNum int = 0

func check(e error) {
	if e != nil {
		panic(e)
	}
}

/*
	TODO:
	1. Implement jdocs. These are files which are stored on the bot which contain the bot's
	default location, unit settings, etc. Helpful for weather.
		- workaround, ask the user for settings during setup.sh
	3. Overall take shell out of the picture (https://github.com/asticode/go-asticoqui)
	4. Maybe find a way to detect silence in the audio for better end handling.
		- probably unnecessary
*/

func (s *Server) ProcessIntent(req *vtt.IntentRequest) (*vtt.IntentResponse, error) {
	if _, err := os.Stat("./slowsys"); err == nil {
		log.Println("slowsys file found")
		slowSys = true
	}
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
	botNum = botNum + 1
	log.Println("Bot " + strconv.Itoa(botNum) + " " + req.Device)
	log.Println("Bot " + strconv.Itoa(botNum) + " " + req.Session)
	log.Println("Bot " + strconv.Itoa(botNum) + " " + req.LangString)
	var justThisBotNum int = botNum
	if debugLogging == true {
		log.Println("Stream " + strconv.Itoa(botNum) + " opened.")
	}
	data := []byte{}
	data = append(data, req.FirstReq.InputAudio...)
	stream := opus.OggStream{}
	go func() {
		time.Sleep(time.Millisecond * 500)
		for voiceTimer < 7 {
			voiceTimer = voiceTimer + 1
			time.Sleep(time.Millisecond * 800)
		}
	}()
	go func() {
		var processOne bool = false
		var processTwo bool = false
		var processThree bool = false
		var processFour bool = false
		time.Sleep(time.Millisecond * 500)
		if slowSys == false {
			for voiceTimer < 7 {
				if processOne == false {
					if _, err := os.Stat("/tmp/" + strconv.Itoa(justThisBotNum) + "dumped1"); err == nil {
						processOne = true
						process1 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(justThisBotNum)+"voice1.wav")
						process1out, err := process1.Output()
						if err != nil {
							//
						}
						transcription1 = strings.TrimSpace(string(process1out))
						log.Println("Bot " + strconv.Itoa(justThisBotNum) + ", Transcription 1: " + transcription1)
					}
				}
				if processTwo == false {
					if _, err := os.Stat("/tmp/" + strconv.Itoa(justThisBotNum) + "dumped2"); err == nil {
						processTwo = true
						process2 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(justThisBotNum)+"voice2.wav")
						process2out, err := process2.Output()
						if err != nil {
							//
						}
						transcription2 = strings.TrimSpace(string(process2out))
						log.Println("Bot " + strconv.Itoa(justThisBotNum) + ", Transcription 2: " + transcription2)
					}
				}
				if processThree == false {
					if _, err := os.Stat("/tmp/" + strconv.Itoa(justThisBotNum) + "dumped3"); err == nil {
						processThree = true
						process3 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(justThisBotNum)+"voice3.wav")
						process3out, err := process3.Output()
						if err != nil {
							//
						}
						transcription3 = strings.TrimSpace(string(process3out))
						log.Println("Bot " + strconv.Itoa(justThisBotNum) + ", Transcription 3: " + transcription3)
					}
				}
				if processFour == false {
					if _, err := os.Stat("/tmp/" + strconv.Itoa(justThisBotNum) + "dumped4"); err == nil {
						processFour = true
						process4 := exec.Command("../stt/stt", "--model", "../stt/model.tflite", "--scorer", "../stt/large_vocabulary.scorer", "--audio", "/tmp/"+strconv.Itoa(justThisBotNum)+"voice4.wav")
						process4out, err := process4.Output()
						if err != nil {
							//
						}
						transcription4 = strings.TrimSpace(string(process4out))
						log.Println("Bot " + strconv.Itoa(justThisBotNum) + ", Transcription 4: " + transcription4)
						successMatch = true
					}
				}
			}
		} else {
			//idk
		}
	}()
	for {
		chunk, err := req.Stream.Recv()
		if err != nil {
			if err == io.EOF {
				IntentPass(req, "intent_system_noaudio", "EOF error", map[string]string{"error": "EOF"}, true)
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
		go bytesToInt(stream, data, justThisBotNum, voiceTimer, die)
	}
	successMatched := processTextAll(req, transcribedText, matchListList, intentsList)
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"voice.pcm").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"voice1.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"voice2.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"voice3.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"voice4.wav").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"dumped1").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"dumped2").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"dumped3").Run()
	exec.Command("/bin/rm", "/tmp/"+strconv.Itoa(botNum)+"dumped4").Run()
	if successMatched == 0 {
		if debugLogging == true {
			log.Println("No intent was matched.")
		}
		IntentPass(req, "intent_system_noaudio", transcribedText, map[string]string{"": ""}, false)
	}
	botNum = botNum - 1
	return nil, nil
}
