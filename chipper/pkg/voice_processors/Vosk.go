//go:build vosk
// +build vosk

package wirepod

import (
	"encoding/json"
	"log"
	"strconv"

	vosk "github.com/alphacep/vosk-api/go"
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
	logger("(Bot " + strconv.Itoa(req.BotNum) + ", Vosk) Processing...")
	speechIsDone := false
	sampleRate := 16000.0
	rec, err := vosk.NewRecognizer(model, sampleRate)
	if err != nil {
		log.Fatal(err)
	}
	rec.SetWords(1)
	rec.AcceptWaveform(req.FirstReq)
	for {
		var chunk []byte
		req, chunk, err = getNextStreamChunk(req)
		if err != nil {
			return "", err
		}
		rec.AcceptWaveform(chunk)
		// has to be split into 320 []byte chunks for VAD
		req, speechIsDone = detectEndOfSpeech(req)
		if speechIsDone {
			break
		}
	}
	var jres map[string]interface{}
	json.Unmarshal([]byte(rec.FinalResult()), &jres)
	transcribedText := jres["text"].(string)
	logger("Bot " + strconv.Itoa(req.BotNum) + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}
