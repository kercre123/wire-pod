package wirepod_coqui

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/asticode/go-asticoqui"
	"github.com/kercre123/chipper/pkg/logger"
	sr "github.com/kercre123/chipper/pkg/wirepod/speechrequest"
)

var Name string = "coqui"

// Init should be defined as `func() error`
func Init() error {
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
	cTime := time.Now()
	micData = sr.SplitVAD(pcmBytes)
	for _, sample := range micData {
		coquiStream.FeedAudioContent(sr.BytesToSamples(sample))
	}
	res, err := coquiStream.Finish()
	tTime := time.Now().Sub(cTime)
	if err != nil {
		log.Fatal("Failed testing speech to text: ", err)
	}
	logger.Println("Text:", res)
	if tTime.Seconds() > 3 {
		logger.Println("Coqui test took a while, performance may be degraded. (" + fmt.Sprint(tTime) + ")")
	}
	logger.Println("Coqui test successful! (Took " + fmt.Sprint(tTime) + ")")
	return nil
}

// STT funcs should be defined as func(sr.SpeechRequest) (string, error)

func STT(req sr.SpeechRequest) (string, error) {
	logger.Println("(Bot " + req.Device + ", Coqui) Processing...")
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
		chunk, err = req.GetNextStreamChunk()
		if err != nil {
			return "", err
		}
		coquiStream.FeedAudioContent(sr.BytesToSamples(chunk))
		speechIsDone = req.DetectEndOfSpeech()
		if speechIsDone {
			break
		}
	}
	transcribedText, _ := coquiStream.Finish()
	logger.Println("Bot " + req.Device + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}
