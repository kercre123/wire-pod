package wirepod_coqui

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/asticode/go-asticoqui"
	"github.com/kercre123/chipper/pkg/logger"
	sr "github.com/kercre123/chipper/pkg/wirepod/speechrequest"
)

var Name string = "coqui"

// Init should be defined as `func() error`
func Init() error {
	var testTimer float64
	var timerDie bool = false
	logger.Println("Running a Coqui test...")
	coquiInstance, _ := asticoqui.New("../stt/model.tflite")
	if _, err := os.Stat("../stt/large_vocabulary.scorer"); err == nil {
		coquiInstance.EnableExternalScorer("../stt/large_vocabulary.scorer")
	} else if _, err := os.Stat("../stt/model.scorer"); err == nil {
		coquiInstance.EnableExternalScorer("../stt/model.scorer")
	} else {
		logger.Println("No .scorer file found.")
	}
	coquiStream, err := coquiInstance.NewStream()
	if err != nil {
		log.Fatal(err)
	}
	pcmBytes, _ := os.ReadFile("./stttest.pcm")
	var micData [][]byte
	micData = sr.SplitVAD(pcmBytes)
	for _, sample := range micData {
		coquiStream.FeedAudioContent(sr.BytesToSamples(sample))
	}
	go func() {
		for testTimer <= 7.00 {
			if timerDie {
				break
			}
			time.Sleep(time.Millisecond * 10)
			testTimer = testTimer + 0.01
			if testTimer > 6.50 {
				logger.Println("The STT test is taking too long, this hardware may not be adequate.")
			}
		}
	}()
	res, err := coquiStream.Finish()
	if err != nil {
		log.Fatal("Failed testing speech to text: ", err)
	}
	logger.Println("Text:", res)
	timerDie = true
	logger.Println("Coqui test successful! (Took " + strconv.FormatFloat(testTimer, 'f', 2, 64) + " seconds)")
	return nil
}

// STT funcs should be defined as func(sr.SpeechRequest) (string, error)

func STT(req sr.SpeechRequest) (string, error) {
	logger.Println("(Bot " + strconv.Itoa(req.BotNum) + ", Coqui) Processing...")
	speechIsDone := false
	coquiInstance, _ := asticoqui.New("../stt/model.tflite")
	if _, err := os.Stat("../stt/large_vocabulary.scorer"); err == nil {
		coquiInstance.EnableExternalScorer("../stt/large_vocabulary.scorer")
	} else if _, err := os.Stat("../stt/model.scorer"); err == nil {
		coquiInstance.EnableExternalScorer("../stt/model.scorer")
	} else {
		logger.Println("No .scorer file found.")
	}
	coquiStream, _ := coquiInstance.NewStream()
	for {
		var chunk []byte
		var err error
		req, chunk, err = sr.GetNextStreamChunk(req)
		if err != nil {
			return "", err
		}
		coquiStream.FeedAudioContent(sr.BytesToSamples(chunk))
		req, speechIsDone = sr.DetectEndOfSpeech(req)
		if speechIsDone {
			break
		}
	}
	transcribedText, _ := coquiStream.Finish()
	logger.Println("Bot " + strconv.Itoa(req.BotNum) + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}
