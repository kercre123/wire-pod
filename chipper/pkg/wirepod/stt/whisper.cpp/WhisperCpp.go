package wirepod_whispercpp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"math"
	"log"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/vars"
	sr "github.com/kercre123/chipper/pkg/wirepod/speechrequest"
)


var Name string = "whisper.cpp"

var model whisper.Model

func Init() error {
	// sttLanguage := vars.APIConfig.STT.Language
	// if len(sttLanguage) == 0 {
	// 	sttLanguage = "en-US"
	// }

	modelPath := filepath.Join(vars.WhisperModelPath, "ggml-base.bin")
	if _, err := os.Stat(modelPath); err != nil {
		fmt.Println("Path does not exist: " + modelPath)
		return err
	}
	logger.Println("Opening Whisper model (" + modelPath + ")")

	aModel, err := whisper.New(modelPath)
	if err != nil {
		log.Fatal(err)
		return err
	}
	model = aModel

	return nil
}

func STT(req sr.SpeechRequest) (string, error) {
	logger.Println("(Bot " + req.Device + ", Whisper) Processing...")
	speechIsDone := false
	var err error
	for {
		_, err = req.GetNextStreamChunk()
		if err != nil {
			return "", err
		}
		if err != nil {
			return "", err
		}
		// has to be split into 320 []byte chunks for VAD
		speechIsDone = req.DetectEndOfSpeech()
		if speechIsDone {
			break
		}
	}

	transcribedText, err := process(asFloat32Buffer(sr.BytesToSamples(req.DecodedMicData)))
	if err != nil {
		return "", err
	}
	transcribedText = strings.ToLower(transcribedText)
	logger.Println("Bot " + req.Device + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}

func process(data []float32) (string, error) {
	context, err := model.NewContext()
	if err != nil {
		log.Fatal(err)
	}

	var cb whisper.SegmentCallback
	var transcribedText string
	cb = func(segment whisper.Segment) {
		transcribedText = segment.Text
	}

	if err := context.Process(data, cb, nil); err != nil {
		return "", err
	}
	
	return transcribedText, nil
}

func asFloat32Buffer(buf []int16) []float32 {
	newB := make([]float32, len(buf))
	factor := math.Pow(2, float64(16)-1)
	for i := 0; i < len(buf); i++ {
		newB[i] = float32(float64(buf[i]) / factor)
	}
	return newB
}