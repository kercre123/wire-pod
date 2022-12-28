package wirepod_leopard

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	leopard "github.com/Picovoice/leopard/binding/go"
	"github.com/kercre123/chipper/pkg/logger"
	sr "github.com/kercre123/chipper/pkg/speechrequest"
)

var Name string = "leopard"

var leopardSTTArray []leopard.Leopard
var picovoiceInstancesOS string = os.Getenv("PICOVOICE_INSTANCES")
var picovoiceInstances int

// New returns a new server
var Init func() error = func() error {
	var picovoiceKey string
	picovoiceKeyOS := os.Getenv("PICOVOICE_APIKEY")
	leopardKeyOS := os.Getenv("LEOPARD_APIKEY")
	if picovoiceInstancesOS == "" {
		picovoiceInstances = 3
	} else {
		picovoiceInstancesToInt, err := strconv.Atoi(picovoiceInstancesOS)
		picovoiceInstances = picovoiceInstancesToInt
		if err != nil {
			fmt.Println("PICOVOICE_INSTANCES is not a valid integer, using default value of 3")
			picovoiceInstances = 3
		}
	}
	if picovoiceKeyOS == "" {
		if leopardKeyOS == "" {
			fmt.Println("You must set PICOVOICE_APIKEY to a value.")
			os.Exit(1)
		} else {
			fmt.Println("PICOVOICE_APIKEY is not set, using LEOPARD_APIKEY")
			picovoiceKey = leopardKeyOS
		}
	} else {
		picovoiceKey = picovoiceKeyOS
	}
	fmt.Println("Initializing " + strconv.Itoa(picovoiceInstances) + " Picovoice Instances...")
	for i := 0; i < picovoiceInstances; i++ {
		fmt.Println("Initializing Picovoice Instance " + strconv.Itoa(i))
		leopardSTTArray = append(leopardSTTArray, leopard.NewLeopard(picovoiceKey))
		leopardSTTArray[i].Init()
	}
	return nil
}

var STT func(sr.SpeechRequest) (string, error) = func(req sr.SpeechRequest) (transcribedText string, err error) {
	logger.Println("(Bot " + strconv.Itoa(req.BotNum) + ", Leopard) Processing...")
	var leopardSTT leopard.Leopard
	speechIsDone := false
	if req.BotNum > picovoiceInstances {
		fmt.Println("Too many bots are connected, sending error to bot " + strconv.Itoa(req.BotNum))
		return "", fmt.Errorf("too many bots are connected, max is 3")
	} else {
		leopardSTT = leopardSTTArray[req.BotNum-1]
	}
	for {
		req, _, err = sr.GetNextStreamChunk(req)
		if err != nil {
			return "", err
		}
		req, speechIsDone = sr.DetectEndOfSpeech(req)
		if speechIsDone {
			break
		}
	}
	transcribedTextPre, _, err := leopardSTT.Process(sr.BytesToSamples(req.DecodedMicData))
	if err != nil {
		logger.Println(err)
	}
	transcribedText = strings.ToLower(transcribedTextPre)
	logger.Println("Bot " + strconv.Itoa(req.BotNum) + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}
