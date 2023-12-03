package wirepod_whispercpp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"math"
	"encoding/binary"
	"runtime"
	
	"github.com/ggerganov/whisper.cpp/bindings/go"
	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/vars"
	sr "github.com/kercre123/chipper/pkg/wirepod/speechrequest"
)


var Name string = "whisper.cpp"

var context *whisper.Context
var params whisper.Params

func Init() error {
	var sttLanguage string
	if len(vars.APIConfig.STT.Language) == 0 {
		sttLanguage = "en"
	} else  {
		sttLanguage = strings.Split(vars.APIConfig.STT.Language, "-")[0]
	}

	modelPath := filepath.Join(vars.WhisperModelPath, "ggml-tiny.bin")
	if _, err := os.Stat(modelPath); err != nil {
		fmt.Println("Path does not exist: " + modelPath)
		return err
	}
	logger.Println("Opening Whisper model (" + modelPath + ")")
	logger.Println(whisper.Whisper_print_system_info())
	context = whisper.Whisper_init(modelPath)
	params = context.Whisper_full_default_params(whisper.SAMPLING_GREEDY)
	params.SetTranslate(false)
	params.SetPrintSpecial(false)
	params.SetPrintProgress(false)
	params.SetPrintRealtime(false)
	params.SetPrintTimestamps(false)
	params.SetThreads(runtime.NumCPU())
	params.SetNoContext(true)
	params.SetSingleSegment(true)
	params.SetLanguage(context.Whisper_lang_id(sttLanguage))
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
		// has to be split into 320 []byte chunks for VAD
		speechIsDone = req.DetectEndOfSpeech()
		if speechIsDone {
			break
		}
	}

	transcribedText, err := process(BytesToFloat32Buffer(req.DecodedMicData))
	if err != nil {
		return "", err
	}
	transcribedText = strings.ToLower(transcribedText)
	logger.Println("Bot " + req.Device + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}

func process(data []float32) (string, error) {
	var transcribedText string
	context.Whisper_full(params, data, nil, func(_ int) {
		transcribedText = strings.TrimSpace(context.Whisper_full_get_segment_text(0))
	}, nil)
	return transcribedText, nil
}

func BytesToFloat32Buffer(buf []byte) []float32 {
	newB := make([]float32, len(buf)/2)
	factor := math.Pow(2, float64(16)-1)
	for i := 0; i < len(buf)/2; i++ {
		newB[i] = float32(float64(int16(binary.LittleEndian.Uint16(buf[i*2:]))) / factor)
	}
	return newB
}