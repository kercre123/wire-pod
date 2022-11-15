//go:build vosk
// +build vosk

package wirepod

import (
	"encoding/json"
	"log"
	"strconv"

	vosk "github.com/alphacep/vosk-api/go"
	"github.com/maxhawkins/go-webrtcvad"
)

var model *vosk.VoskModel

// New returns a new server
func VoskInit() (*Server, error) {
	// Open model
	logger("Opening model")
	aModel, err := vosk.NewModel("../vosk/models/" + sttLanguage + "/model")
	if err != nil {
		log.Fatal(err)
	}
	model = aModel
	logger("Model open!")

	logger("Server OK")

	return &Server{}, nil
}

func VoskSTTHandler(req SpeechRequest) (string, error) {
	logger("NewVoskSTTHandler: Processing...")
	sampleRate := 16000.0
	rec, err := vosk.NewRecognizer(model, sampleRate)
	if err != nil {
		log.Fatal(err)
	}
	rec.SetWords(1)
	rec.AcceptWaveform(req.FirstReq)
	vad, err := webrtcvad.New()
	if err != nil {
		logger(err)
	}
	vad.SetMode(3)
	var activeNum int = 0
	var inactiveNum int = 0
	var inactiveNumMax int = 20
	var die bool = false
	for {
		var chunk []byte
		req, chunk, err = getNextStreamChunk(req)
		if err != nil {
			return "", err
		}
		rec.AcceptWaveform(chunk)
		// has to be split into 320 []byte chunks for VAD
		splitChunk := splitVAD(chunk)
		for _, chunk := range splitChunk {
			active, err := vad.Process(16000, chunk)
			if err != nil {
				logger("err:")
				logger(err)
				return "", err
			}
			if active {
				activeNum = activeNum + 1
				inactiveNum = 0
			} else {
				inactiveNum = inactiveNum + 1
			}
			if inactiveNum >= inactiveNumMax && activeNum > 20 {
				logger("Speech completed.")
				die = true
				break
			}
		}
		if die {
			break
		}
	}
	var jres map[string]interface{}
	json.Unmarshal([]byte(rec.FinalResult()), &jres)
	transcribedText := jres["text"].(string)
	logger("Bot " + strconv.Itoa(req.BotNum) + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}
