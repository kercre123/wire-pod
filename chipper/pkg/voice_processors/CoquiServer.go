//go:build coqui
// +build coqui

package wirepod

import (
	"github.com/asticode/go-asticoqui"
	"log"
	"os"
	"strconv"
	"time"
)

// New returns a new server
func CoquiNew() (*Server, error) {
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
		logger(err)
	}
	pcmBytes, _ := os.ReadFile("./stttest.pcm")
	var micData [][]byte
	micData = split(pcmBytes)
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
	return &Server{}, nil
}
