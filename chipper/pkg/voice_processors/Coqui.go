//go:build coqui
// +build coqui

package wirepod

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/asticode/go-asticoqui"
	"github.com/maxhawkins/go-webrtcvad"
)

func CoquiInit() error {
	var testTimer float64
	var timerDie bool = false
	logger("Running a Coqui test...")
	coquiInstance, _ := asticoqui.New("../stt/model.tflite")
	if _, err := os.Stat("../stt/large_vocabulary.scorer"); err == nil {
		coquiInstance.EnableExternalScorer("../stt/large_vocabulary.scorer")
	} else if _, err := os.Stat("../stt/model.scorer"); err == nil {
		coquiInstance.EnableExternalScorer("../stt/model.scorer")
	} else {
		logger("No .scorer file found.")
	}
	coquiStream, err := coquiInstance.NewStream()
	if err != nil {
		log.Fatal(err)
	}
	pcmBytes, _ := os.ReadFile("./stttest.pcm")
	var micData [][]byte
	micData = splitVAD(pcmBytes)
	for _, sample := range micData {
		coquiStream.FeedAudioContent(bytesToSamples(sample))
	}
	go func() {
		for testTimer <= 7.00 {
			if timerDie {
				break
			}
			time.Sleep(time.Millisecond * 10)
			testTimer = testTimer + 0.01
			if testTimer > 6.50 {
				logger("The STT test is taking too long, this hardware may not be adequate.")
			}
		}
	}()
	res, err := coquiStream.Finish()
	if err != nil {
		log.Fatal("Failed testing speech to text: ", err)
	}
	logger("Text:", res)
	timerDie = true
	logger("Coqui test successful! (Took " + strconv.FormatFloat(testTimer, 'f', 2, 64) + " seconds)")
	return nil
}

func CoquiSttHandler(req SpeechRequest) (string, error) {
	logger("Processing...")
	activeNum := 0
	inactiveNum := 0
	inactiveNumMax := 20
	die := false
	coquiInstance, _ := asticoqui.New("../stt/model.tflite")
	if _, err := os.Stat("../stt/large_vocabulary.scorer"); err == nil {
		coquiInstance.EnableExternalScorer("../stt/large_vocabulary.scorer")
	} else if _, err := os.Stat("../stt/model.scorer"); err == nil {
		coquiInstance.EnableExternalScorer("../stt/model.scorer")
	} else {
		logger("No .scorer file found.")
	}
	coquiStream, _ := coquiInstance.NewStream()
	coquiStream.FeedAudioContent(bytesToSamples(req.FirstReq))
	vad, err := webrtcvad.New()
	if err != nil {
		logger(err)
	}
	vad.SetMode(3)
	for {
		var chunk []byte
		req, chunk, err = getNextStreamChunk(req)
		if err != nil {
			return "", err
		}
		coquiStream.FeedAudioContent(bytesToSamples(chunk))
		splitChunk := splitVAD(chunk)
		for _, chunk := range splitChunk {
			active, err := vad.Process(16000, chunk)
			if err != nil {
				logger(err)
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
	transcribedText, _ := coquiStream.Finish()
	logger("Bot " + strconv.Itoa(req.BotNum) + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}
