package wirepod_vosk

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	vosk "github.com/alphacep/vosk-api/go"
	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/vars"
	sr "github.com/kercre123/chipper/pkg/wirepod/speechrequest"
)

var Name string = "vosk"

var model *vosk.VoskModel
var rec *vosk.VoskRecognizer
var mainRecInUse bool
var modelLoaded bool
var Grammer string

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
			wordsList = append(wordsList, word)
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

func Init() error {
	if vars.APIConfig.PastInitialSetup {
		Grammer = GetGrammerList(vars.APIConfig.STT.Language)
		vosk.SetLogLevel(-1)
		if modelLoaded {
			logger.Println("A model was already loaded, freeing")
			model.Free()
		}
		sttLanguage := vars.APIConfig.STT.Language
		if len(sttLanguage) == 0 {
			sttLanguage = "en-US"
		}
		// Open model
		modelPath := "../vosk/models/" + sttLanguage + "/model"
		logger.Println("Opening VOSK model (" + modelPath + ")")
		aModel, err := vosk.NewModel(modelPath)
		if err != nil {
			log.Fatal(err)
			return err
		}
		model = aModel
		aRecognizer, err := vosk.NewRecognizerGrm(aModel, 16000.0, Grammer)
		//aRecognizer, err := vosk.NewRecognizer(aModel, 16000.0)
		if err != nil {
			log.Fatal(err)
			return err
		}
		rec = aRecognizer
		modelLoaded = true
		logger.Println("VOSK initiated successfully")
	}
	return nil
}

func setRecFalse() {
	mainRecInUse = false
}

func STT(req sr.SpeechRequest) (string, error) {
	logger.Println("(Bot " + strconv.Itoa(req.BotNum) + ", Vosk) Processing...")
	speechIsDone := false
	var thisRec *vosk.VoskRecognizer
	if mainRecInUse {
		newRec, err := vosk.NewRecognizer(model, 16000.0)
		if err != nil {
			log.Fatal(err)
		}
		thisRec = newRec
	} else {
		mainRecInUse = true
		defer setRecFalse()
		fmt.Println("using main")
		thisRec = rec
	}
	//sampleRate := 16000.0
	// rec, err := vosk.NewRecognizerGrm(model, sampleRate, Grammer)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//thisRec.SetWords(2)
	bTime := time.Now()
	thisRec.AcceptWaveform(req.FirstReq)
	for {
		chunk, err := req.GetNextStreamChunk()
		if err != nil {
			return "", err
		}
		thisRec.AcceptWaveform(chunk)
		// has to be split into 320 []byte chunks for VAD
		speechIsDone = req.DetectEndOfSpeech()
		if speechIsDone {
			break
		}
	}
	var jres map[string]interface{}
	json.Unmarshal([]byte(thisRec.FinalResult()), &jres)
	fmt.Println("Process took: ", time.Now().Sub(bTime))
	transcribedText := jres["text"].(string)
	logger.Println("Bot " + strconv.Itoa(req.BotNum) + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}
