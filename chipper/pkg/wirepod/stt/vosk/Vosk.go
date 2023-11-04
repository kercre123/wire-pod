package wirepod_vosk

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	vosk "github.com/alphacep/vosk-api/go"
	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/vars"
	sr "github.com/kercre123/chipper/pkg/wirepod/speechrequest"
)

var Name string = "vosk"

var model *vosk.VoskModel
var recsmu sync.Mutex
var recs []ARec
var modelLoaded bool

type ARec struct {
	InUse bool
	Rec   *vosk.VoskRecognizer
}

var Grammer string

func Init() error {
	if vars.APIConfig.PastInitialSetup {
		vosk.SetLogLevel(-1)
		if modelLoaded {
			logger.Println("A model was already loaded, freeing")
			model.Free()
		}
		sttLanguage := vars.APIConfig.STT.Language
		if len(sttLanguage) == 0 {
			sttLanguage = "en-US"
		}
		modelPath := "../vosk/models/" + sttLanguage + "/model"
		logger.Println("Opening VOSK model (" + modelPath + ")")
		aModel, err := vosk.NewModel(modelPath)
		if err != nil {
			log.Fatal(err)
			return err
		}
		model = aModel
		// just one rec for now. if more bots request later, those will get newly created and added to list of active recs
		//aRecognizer, err := vosk.NewRecognizerGrm(aModel, 16000.0, Grammer)
		aRecognizer, err := vosk.NewRecognizer(aModel, 16000.0)
		if err != nil {
			log.Fatal(err)
			return err
		}
		var arec ARec
		arec.Rec = aRecognizer
		arec.InUse = false
		recs = append(recs, arec)
		modelLoaded = true
		logger.Println("VOSK initiated successfully")
	}
	return nil
}

func getRec() (*vosk.VoskRecognizer, int) {
	recsmu.Lock()
	defer recsmu.Unlock()
	for ind, rec := range recs {
		if !rec.InUse {
			recs[ind].InUse = true
			fmt.Println("Returning already-created rec")
			return recs[ind].Rec, ind
		}
	}
	var newrec ARec
	newrec.InUse = true
	newRec, err := vosk.NewRecognizer(model, 16000.0)
	if err != nil {
		log.Fatal(err)
	}
	newrec.Rec = newRec
	recs = append(recs, newrec)
	return recs[len(recs)-1].Rec, len(recs) - 1
}

func STT(req sr.SpeechRequest) (string, error) {
	logger.Println("(Bot " + strconv.Itoa(req.BotNum) + ", Vosk) Processing...")
	speechIsDone := false
	bTime := time.Now()
	rec, recind := getRec()
	rec.SetWords(0)
	rec.AcceptWaveform(req.FirstReq)
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
	recs[recind].InUse = false
	fmt.Println("Process took: ", time.Now().Sub(bTime))
	transcribedText := jres["text"].(string)
	logger.Println("Bot " + strconv.Itoa(req.BotNum) + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}

// more performance can be gotten via grammar

func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, val := range slice {
		if _, ok := seen[val]; !ok {
			seen[val] = true
			result = append(result, val)
		}
	}
	return result
}

func GetGrammerList(lang string) string {
	var wordsList []string
	var grammer string
	for _, words := range vars.MatchListList {
		for _, word := range words {
			wors := strings.Split(word, " ")
			for _, wor := range wors {
				found := model.FindWord(wor)
				if found != -1 {
					wordsList = append(wordsList, wor)
				}
			}
		}
	}
	wordsList = removeDuplicates(wordsList)
	for i, word := range wordsList {
		if i == len(wordsList)-1 {
			grammer = grammer + `"` + word + `"`
		} else {
			grammer = grammer + `"` + word + `"` + ", "
		}
	}
	grammer = "[" + grammer + "]"
	return grammer
}
