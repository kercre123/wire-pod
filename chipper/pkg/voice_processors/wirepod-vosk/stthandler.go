package wirepod

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/digital-dream-labs/chipper/pkg/voice_processors/logger"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	vosk "github.com/alphacep/vosk-api/go"
	"github.com/digital-dream-labs/chipper/pkg/vtt"
	opus "github.com/digital-dream-labs/opus-go/opus"
)

var botNum int = 0

func split(buf []byte) [][]byte {
	var chunk [][]byte
	for len(buf) >= 320 {
		chunk = append(chunk, buf[:320])
		buf = buf[320:]
	}
	return chunk
}

func bytesToIntVAD(stream opus.OggStream, data []byte, die bool, isOpus bool) [][]byte {
	// detect if data is pcm or opus
	if die {
		return nil
	}
	if isOpus {
		// opus
		n, err := stream.Decode(data)
		if err != nil {
			logger.Log(err)
		}
		byteArray := split(n)
		return byteArray
	} else {
		// pcm
		byteArray := split(data)
		return byteArray
	}
}

func bytesToSamples(buf []byte) []int16 {
	samples := make([]int16, len(buf)/2)
	for i := 0; i < len(buf)/2; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(buf[i*2:]))
	}
	return samples
}

func sttHandler(reqThing interface{}, isKnowledgeGraph bool) (transcribedString string, slots map[string]string, isRhino bool, thisBotNum int, opusUsed bool, err error) {
	var req2 *vtt.IntentRequest
	var req1 *vtt.KnowledgeGraphRequest
	var req3 *vtt.IntentGraphRequest
	var isIntentGraph bool
	if str, ok := reqThing.(*vtt.IntentRequest); ok {
		req2 = str
	} else if str, ok := reqThing.(*vtt.KnowledgeGraphRequest); ok {
		req1 = str
	} else if str, ok := reqThing.(*vtt.IntentGraphRequest); ok {
		req3 = str
		isIntentGraph = true
	}
	var voiceTimer int = 0
	var transcribedText string = ""
	var isOpus bool
	var micData [][]byte
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

	botNum = botNum + 1
	justThisBotNum := botNum
	if isKnowledgeGraph {
		logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " ESN: " + req1.Device)
		logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " Session: " + req1.Session)
		logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " Language: " + req1.LangString)
		logger.Log("KG Stream " + strconv.Itoa(justThisBotNum) + " opened.")
		deviceESN = req1.Device
		deviceSession = req1.Session
	} else if isIntentGraph {
		logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " ESN: " + req3.Device)
		logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " Session: " + req3.Session)
		logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " Language: " + req3.LangString)
		deviceESN = req3.Device
		deviceSession = req3.Session
		logger.Log("Stream " + strconv.Itoa(justThisBotNum) + " opened.")
	} else {
		logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " ESN: " + req2.Device)
		logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " Session: " + req2.Session)
		logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " Language: " + req2.LangString)
		deviceESN = req2.Device
		deviceSession = req2.Session
		logger.Log("Stream " + strconv.Itoa(justThisBotNum) + " opened.")
	}
	data := []byte{}
	if isKnowledgeGraph {
		data = append(data, req1.FirstReq.InputAudio...)
	} else if isIntentGraph {
		data = append(data, req3.FirstReq.InputAudio...)
	} else {
		data = append(data, req2.FirstReq.InputAudio...)
	}
	if len(data) > 0 {
		if data[0] == 0x4f {
			isOpus = true
			logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " Stream Type: Opus")
		} else {
			isOpus = false
			logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " Stream Type: PCM")
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
	logger.Log("Processing...")

	sampleRate := 16000.0
	rec, err := vosk.NewRecognizer(model, sampleRate)
	if err != nil {
		log.Fatal(err)
	}
	rec.SetWords(1)

	inactiveNumMax := 20
	if !isKnowledgeGraph {
		micData = bytesToIntVAD(stream, data, die, isOpus)
		for _, sample := range micData {
			//buf := bytesToSamples(sample)
			if rec.AcceptWaveform(sample) != 0 {
				fmt.Println(rec.Result())
			}
		}
	}
	vad, err := webrtcvad.New()
	if err != nil {
		logger.Log(err)
	}
	vad.SetMode(3)
	for {
		if isKnowledgeGraph {
			chunk, chunkErr := req1.Stream.Recv()
			if chunkErr != nil {
				if chunkErr == io.EOF {
					botNum = botNum - 1
					return "", transcribedSlots, false, justThisBotNum, isOpus, fmt.Errorf("EOF error")
				} else {
					logger("Bot " + strconv.Itoa(justThisBotNum) + " Error: " + chunkErr.Error())
					botNum = botNum - 1
					return "", transcribedSlots, false, justThisBotNum, isOpus, fmt.Errorf("unknown error")
				}
			}
			data = append(data, chunk.InputAudio...)
		} else if isIntentGraph {
			chunk, chunkErr := req3.Stream.Recv()
			if chunkErr != nil {
				if chunkErr == io.EOF {
					botNum = botNum - 1
					return "", transcribedSlots, false, justThisBotNum, isOpus, fmt.Errorf("EOF error")
				} else {
					logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " Error: " + chunkErr.Error())
					botNum = botNum - 1
					return "", transcribedSlots, false, justThisBotNum, isOpus, fmt.Errorf("unknown error")
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
					logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " Error: " + chunkErr.Error())
					botNum = botNum - 1
					return "", transcribedSlots, false, justThisBotNum, isOpus, fmt.Errorf("unknown error")
				}
			}
			data = append(data, chunk.InputAudio...)
		}
		if die {
			break
		}
		micData = bytesToIntVAD(stream, data, die, isOpus)
		numInRange = 0
		for _, sample := range micData {
			if !speechDone {
				if numInRange >= oldDataLength {
					if !isKnowledgeGraph {
						//buf := bytesToSamples(sample)
						if rec.AcceptWaveform(sample) != 0 {
							fmt.Println(rec.Result())
						}
					}
					active, err := vad.Process(16000, sample)
					if err != nil {
						logger.Log(err)
					}
					if active {
						activeNum = activeNum + 1
						inactiveNum = 0
					} else {
						inactiveNum = inactiveNum + 1
					}
					if inactiveNum >= inactiveNumMax && activeNum > 20 {
						logger.Log("Speech completed in " + strconv.FormatFloat(requestTimer, 'f', 2, 64) + " seconds.")
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
		if voiceTimer >= 5 {
			logger.Log("Voice timeout threshold reached.")
			speechDone = true
		}
		if speechDone {
			if isKnowledgeGraph {
				if houndEnable {
					logger.Log("Sending requst to Houndify...")
					if os.Getenv("HOUNDIFY_CLIENT_KEY") != "" {
						req := houndify.VoiceRequest{
							AudioStream:       bytes.NewReader(data),
							UserID:            deviceESN,
							RequestID:         deviceSession,
							RequestInfoFields: make(map[string]interface{}),
						}
						partialTranscripts := make(chan houndify.PartialTranscript)
						serverResponse, err := hKGclient.VoiceSearch(req, partialTranscripts)
						if err != nil {
							logger.Log.Log(err)
						}
						transcribedText, _ = ParseSpokenResponse(serverResponse)
						logger.Log("Transcribed text: " + transcribedText)
						die = true
					}
				} else {
					transcribedText = "Houndify is not enabled."
					logger.Log("Houndify is not enabled.")
					die = true
				}
			} else {
				var jres map[string]interface{}
				json.Unmarshal([]byte(rec.FinalResult()), &jres)
				transcribedText = jres["text"].(string)
				logger.Log("transcribed text: " + transcribedText)
				logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " Transcribed text: " + transcribedText)
				die = true
			}
		}
		if die {
			break
		}
	}
	botNum = botNum - 1
	logger.Log("Bot " + strconv.Itoa(justThisBotNum) + " request served.")
	var rhinoUsed bool
	if rhinoSucceeded {
		rhinoUsed = true
	} else {
		rhinoUsed = false
	}
	return transcribedText, transcribedSlots, rhinoUsed, justThisBotNum, isOpus, nil
}
