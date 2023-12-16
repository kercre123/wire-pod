package wirepod_vosk

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	vosk "github.com/alphacep/vosk-api/go"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	sr "github.com/kercre123/wire-pod/chipper/pkg/wirepod/speechrequest"
)

var GrammerEnable bool = false

var Name string = "vosk"

var model *vosk.VoskModel
var recsmu sync.Mutex

var grmRecs []ARec
var gpRecs []ARec

var modelLoaded bool

type ARec struct {
	InUse bool
	Rec   *vosk.VoskRecognizer
}

var Grammer string

func Init() error {
	if os.Getenv("VOSK_WITH_GRAMMER") == "true" {
		fmt.Println("Initializing vosk with grammer optimizations")
		GrammerEnable = true
	}
	if vars.APIConfig.PastInitialSetup {
		vosk.SetLogLevel(-1)
		if modelLoaded {
			logger.Println("A model was already loaded, freeing all recognizers and model")
			for ind, _ := range grmRecs {
				grmRecs[ind].Rec.Free()
			}
			for ind, _ := range gpRecs {
				gpRecs[ind].Rec.Free()
			}
			gpRecs = []ARec{}
			grmRecs = []ARec{}
			model.Free()
		}
		sttLanguage := vars.APIConfig.STT.Language
		if len(sttLanguage) == 0 {
			sttLanguage = "en-US"
		}
		modelPath := filepath.Join(vars.VoskModelPath, sttLanguage, "model")
		if _, err := os.Stat(modelPath); err != nil {
			fmt.Println("Path does not exist: " + modelPath)
			return err
		}
		logger.Println("Opening VOSK model (" + modelPath + ")")
		aModel, err := vosk.NewModel(modelPath)
		if err != nil {
			log.Fatal(err)
			return err
		}
		model = aModel
		if GrammerEnable {
			logger.Println("Initializing grammer list")
			Grammer = GetGrammerList(vars.APIConfig.STT.Language)
		}

		logger.Println("Initializing VOSK recognizers")
		if GrammerEnable {
			grmRecognizer, err := vosk.NewRecognizerGrm(aModel, 16000.0, Grammer)
			if err != nil {
				log.Fatal(err)
			}
			var grmrec ARec
			grmrec.Rec = grmRecognizer
			grmrec.InUse = false
			grmRecs = append(grmRecs, grmrec)
		}
		gpRecognizer, err := vosk.NewRecognizer(aModel, 16000.0)
		var gprec ARec
		gprec.Rec = gpRecognizer
		gprec.InUse = false
		gpRecs = append(gpRecs, gprec)
		if err != nil {
			log.Fatal(err)
		}
		modelLoaded = true
		logger.Println("VOSK initiated successfully")
	}
	return nil
}

func getRec(withGrm bool) (*vosk.VoskRecognizer, int) {
	recsmu.Lock()
	defer recsmu.Unlock()
	if withGrm && GrammerEnable {
		for ind, rec := range grmRecs {
			if !rec.InUse {
				grmRecs[ind].InUse = true
				return grmRecs[ind].Rec, ind
			}
		}
	} else {
		for ind, rec := range gpRecs {
			if !rec.InUse {
				gpRecs[ind].InUse = true
				return gpRecs[ind].Rec, ind
			}
		}
	}
	recsmu.Unlock()
	var newrec ARec
	var newRec *vosk.VoskRecognizer
	var err error
	newrec.InUse = true
	if withGrm {
		newRec, err = vosk.NewRecognizerGrm(model, 16000.0, Grammer)
	} else {
		newRec, err = vosk.NewRecognizer(model, 16000.0)
	}
	if err != nil {
		log.Fatal(err)
	}
	newrec.Rec = newRec
	recsmu.Lock()
	if withGrm {
		grmRecs = append(grmRecs, newrec)
		return grmRecs[len(grmRecs)-1].Rec, len(grmRecs) - 1
	} else {
		gpRecs = append(gpRecs, newrec)
		return gpRecs[len(gpRecs)-1].Rec, len(gpRecs) - 1
	}
}

func STT(req sr.SpeechRequest) (string, error) {
	logger.Println("(Bot " + req.Device + ", Vosk) Processing...")
	speechIsDone := false
	var withGrm bool
	if (vars.APIConfig.Knowledge.IntentGraph || req.IsKG) || !GrammerEnable {
		logger.Println("Using general recognizer")
		withGrm = false
	} else {
		logger.Println("Using grammer-optimized recognizer")
		withGrm = true
	}
	rec, recind := getRec(withGrm)
	rec.SetWords(1)
	rec.AcceptWaveform(req.FirstReq)
	req.DetectEndOfSpeech()
	for {
		chunk, err := req.GetNextStreamChunk()
		if err != nil {
			return "", err
		}
		rec.AcceptWaveform(chunk)
		// has to be split into 320 []byte chunks for VAD
		speechIsDone = req.DetectEndOfSpeech()
		if speechIsDone {
			break
		}
	}
	var jres map[string]interface{}
	json.Unmarshal([]byte(rec.FinalResult()), &jres)
	if withGrm {
		grmRecs[recind].InUse = false
	} else {
		gpRecs[recind].InUse = false
	}
	transcribedText := jres["text"].(string)
	logger.Println("Bot " + req.Device + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}
