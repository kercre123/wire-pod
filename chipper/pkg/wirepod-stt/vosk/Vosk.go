package wirepod_vosk

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	vosk "github.com/alphacep/vosk-api/go"
	"github.com/kercre123/chipper/pkg/logger"
	sr "github.com/kercre123/chipper/pkg/speechrequest"
)

var Name string = "vosk"

var model *vosk.VoskModel

var Init func() error = func() error {
	sttLanguage := os.Getenv("STT_LANGUAGE")
	if len(sttLanguage) == 0 {
		sttLanguage = "en-US"
	}
	// Open model
	logger.Println("Opening model")
	aModel, err := vosk.NewModel("../vosk/models/" + sttLanguage + "/model")
	if err != nil {
		log.Fatal(err)
		return err
	}
	model = aModel
	logger.Println("Model open!")

	logger.Println("Server OK")

	return nil
}

var STT func(sr.SpeechRequest) (string, error) = func(req sr.SpeechRequest) (string, error) {
	logger.Println("(Bot " + strconv.Itoa(req.BotNum) + ", Vosk) Processing...")
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
		req, chunk, err = sr.GetNextStreamChunk(req)
		if err != nil {
			return "", err
		}
		rec.AcceptWaveform(chunk)
		// has to be split into 320 []byte chunks for VAD
		req, speechIsDone = sr.DetectEndOfSpeech(req)
		if speechIsDone {
			break
		}
	}
	var jres map[string]interface{}
	json.Unmarshal([]byte(rec.FinalResult()), &jres)
	transcribedText := jres["text"].(string)
	logger.Println("Bot " + strconv.Itoa(req.BotNum) + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}
