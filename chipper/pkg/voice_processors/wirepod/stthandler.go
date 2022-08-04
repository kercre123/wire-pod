package wirepod

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/asticode/go-asticoqui"
	"github.com/digital-dream-labs/chipper/pkg/vtt"
	opus "github.com/digital-dream-labs/opus-go/opus"
	"github.com/maxhawkins/go-webrtcvad"
	"github.com/soundhound/houndify-sdk-go"
)

var debugLogging bool

var botNum int = 0

func sttHandler(reqThing interface{}, isKnowledgeGraph bool) (transcribedString string, slots map[string]string, isRhino bool, thisBotNum int, opusUsed bool, err error) {
	var req2 *vtt.IntentRequest
	var req1 *vtt.KnowledgeGraphRequest
	if str, ok := reqThing.(*vtt.IntentRequest); ok {
		req2 = str
	} else if str, ok := reqThing.(*vtt.KnowledgeGraphRequest); ok {
		req1 = str
	}
	var voiceTimer int = 0
	var transcribedText string = ""
	var isOpus bool
	var micData [][]byte
	var micDataHound []byte
	var die bool = false
	var numInRange int = 0
	var oldDataLength int = 0
	var speechDone bool = false
	var rhinoSucceeded bool = false
	var transcribedSlots map[string]string
	var activeNum int = 0
	var inactiveNum int = 0
	var deviceESN string
	var deviceSession string
	if os.Getenv("DEBUG_LOGGING") != "true" && os.Getenv("DEBUG_LOGGING") != "false" {
		fmt.Println("No valid value for DEBUG_LOGGING, setting to true")
		debugLogging = true
	} else {
		if os.Getenv("DEBUG_LOGGING") == "true" {
			debugLogging = true
		} else {
			debugLogging = false
		}
	}
	botNum = botNum + 1
	justThisBotNum := botNum
	if botNum > 1 {
		if debugLogging {
			fmt.Println("Multiple bots are streaming, live transcription disabled")
		}
	}
	if os.Getenv("DISABLE_LIVE_TRANSCRIPTION") == "true" {
		if debugLogging {
			fmt.Println("DISABLE_LIVE_TRANSCRIPTION is true, live transcription disabled")
		}
	}
	if debugLogging {
		if isKnowledgeGraph {
			fmt.Println("Bot " + strconv.Itoa(justThisBotNum) + " ESN: " + req1.Device)
			fmt.Println("Bot " + strconv.Itoa(justThisBotNum) + " Session: " + req1.Session)
			fmt.Println("Bot " + strconv.Itoa(justThisBotNum) + " Language: " + req1.LangString)
			fmt.Println("KG Stream " + strconv.Itoa(justThisBotNum) + " opened.")
			deviceESN = req1.Device
			deviceSession = req1.Session
		} else {
			fmt.Println("Bot " + strconv.Itoa(justThisBotNum) + " ESN: " + req2.Device)
			fmt.Println("Bot " + strconv.Itoa(justThisBotNum) + " Session: " + req2.Session)
			fmt.Println("Bot " + strconv.Itoa(justThisBotNum) + " Language: " + req2.LangString)
			deviceESN = req2.Device
			deviceSession = req2.Session
			fmt.Println("Stream " + strconv.Itoa(justThisBotNum) + " opened.")
		}
	}
	data := []byte{}
	if isKnowledgeGraph {
		data = append(data, req1.FirstReq.InputAudio...)
	} else {
		data = append(data, req2.FirstReq.InputAudio...)
	}
	if len(data) > 0 {
		if data[0] == 0x4f {
			isOpus = true
			if debugLogging {
				fmt.Println("Bot " + strconv.Itoa(justThisBotNum) + " Stream Type: Opus")
			}
		} else {
			isOpus = false
			if debugLogging {
				fmt.Println("Bot " + strconv.Itoa(justThisBotNum) + " Stream Type: PCM")
			}
		}
	}
	stream := opus.OggStream{}
	var requestTimer float64 = 0.00
	go func() {
		time.Sleep(time.Millisecond * 500)
		for voiceTimer < 7 {
			voiceTimer = voiceTimer + 1
			time.Sleep(time.Millisecond * 750)
		}
	}()
	go func() {
		// accurate timer
		for requestTimer < 7.00 {
			requestTimer = requestTimer + 0.01
			time.Sleep(time.Millisecond * 10)
		}
	}()
	fmt.Printf("Processing...")
	inactiveNumMax := 35
	coquiInstance, _ := asticoqui.New("../stt/model.tflite")
	coquiInstance.EnableExternalScorer("../stt/large_vocabulary.scorer")
	coquiStream, _ := coquiInstance.NewStream()
	coquiStream.FeedAudioContent(bytesToSamples(data))
	for {
		if isKnowledgeGraph {
			chunk, chunkErr := req1.Stream.Recv()
			if chunkErr != nil {
				if chunkErr == io.EOF {
					botNum = botNum - 1
					return "", transcribedSlots, false, justThisBotNum, isOpus, fmt.Errorf("EOF error")
				}
			}
			data = append(data, chunk.InputAudio...)
		} else {
			chunk, chunkErr := req2.Stream.Recv()
			if chunkErr != nil {
				if chunkErr == io.EOF {
					botNum = botNum - 1
					return "", transcribedSlots, false, justThisBotNum, isOpus, fmt.Errorf("EOF error")
				} else {
					if debugLogging {
						fmt.Println("Bot " + strconv.Itoa(justThisBotNum) + " Error: " + chunkErr.Error())
					}
					botNum = botNum - 1
					return "", transcribedSlots, false, justThisBotNum, isOpus, fmt.Errorf("unknown error")
				}
			}
			data = append(data, chunk.InputAudio...)
		}
		if die {
			break
		}
		// returns []int16, framesize unknown
		// returns [][]int16, 512 framesize
		micData = bytesToIntVAD(stream, data, die, isOpus)
		micDataHound = bytesToIntHound(stream, data, die, isOpus)
		numInRange = 0
		for _, sample := range micData {
			if !speechDone {
				if numInRange >= oldDataLength {
					coquiStream.FeedAudioContent(bytesToSamples(sample))
					vad, err := webrtcvad.New()
					if err != nil {
						log.Fatal(err)
					}

					if err := vad.SetMode(2); err != nil {
						log.Fatal(err)
					}

					rate := 16000 // kHz
					active, err := vad.Process(rate, sample)
					if err != nil {
						log.Fatal(err)
					}
					if active {
						activeNum = activeNum + 1
						inactiveNum = 0
					} else {
						inactiveNum = inactiveNum + 1
					}
					if inactiveNum >= inactiveNumMax && activeNum > 20 {
						fmt.Printf("\n")
						fmt.Printf("Speech completed in %f seconds.\n", requestTimer)
						speechDone = true
						break
					}
					numInRange = numInRange + 1
				} else {
					numInRange = numInRange + 1
				}
			}
		}
		oldDataLength = len(micData)
		if speechDone {
			if isKnowledgeGraph {
				if os.Getenv("HOUNDIFY_CLIENT_KEY") != "" {
					req := houndify.VoiceRequest{
						AudioStream:       bytes.NewReader(micDataHound),
						UserID:            deviceESN,
						RequestID:         deviceSession,
						RequestInfoFields: make(map[string]interface{}),
					}
					partialTranscripts := make(chan houndify.PartialTranscript)
					serverResponse, err := hKGclient.VoiceSearch(req, partialTranscripts)
					if err != nil {
						fmt.Println(err)
					}
					transcribedText, _ = ParseSpokenResponse(serverResponse)
					fmt.Println("Transcribed text: " + transcribedText)
					die = true
				}
			} else {
				transcribedText, _ = coquiStream.Finish()
				fmt.Println("Bot " + strconv.Itoa(justThisBotNum) + " Transcribed text: " + transcribedText)
				die = true
			}
		}
		if die {
			break
		}
	}
	botNum = botNum - 1
	if debugLogging {
		fmt.Println("Bot " + strconv.Itoa(justThisBotNum) + " request served.")
	}
	var rhinoUsed bool
	if rhinoSucceeded {
		rhinoUsed = true
	} else {
		rhinoUsed = false
	}
	return transcribedText, transcribedSlots, rhinoUsed, justThisBotNum, isOpus, nil
}
