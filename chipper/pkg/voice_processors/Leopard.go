//go:build leopard
// +build leopard

package wirepod

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	leopard "github.com/Picovoice/leopard/binding/go"
	"github.com/digital-dream-labs/opus-go/opus"
	"github.com/maxhawkins/go-webrtcvad"
)

func bytesToIntLeopard(stream opus.OggStream, data []byte, die bool, isOpus bool) []int16 {
	if die {
		return nil
	}
	if isOpus {
		// opus
		n, err := stream.Decode(data)
		if err != nil {
			logger(err)
		}
		return bytesToSamples(n)
	} else {
		// pcm
		return bytesToSamples(data)
	}
}

var leopardSTTArray []leopard.Leopard
var picovoiceInstancesOS string = os.Getenv("PICOVOICE_INSTANCES")
var picovoiceInstances int

// New returns a new server
func LeopardInit() error {
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

func LeopardSttHandler(req SpeechRequest) (transcribedText string, err error) {
	logger("Processing...")
	var leopardSTT leopard.Leopard
	activeNum := 0
	inactiveNum := 0
	inactiveNumMax := 20
	die := false
	vad, err := webrtcvad.New()
	if err != nil {
		logger(err)
	}
	vad.SetMode(3)
	if botNum > picovoiceInstances {
		fmt.Println("Too many bots are connected, sending error to bot " + strconv.Itoa(req.BotNum))
		return "", fmt.Errorf("too many bots are connected, max is 3")
	} else {
		leopardSTT = leopardSTTArray[botNum-1]
	}
	for {
		var chunk []byte
		req, chunk, err = getNextStreamChunk(req)
		if err != nil {
			return "", err
		}
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
	transcribedTextPre, _, err := leopardSTT.Process(bytesToSamples(req.DecodedMicData))
	if err != nil {
		logger(err)
	}
	transcribedText = strings.ToLower(transcribedTextPre)
	logger("Bot " + strconv.Itoa(req.BotNum) + " Transcribed text: " + transcribedText)
	return transcribedText, nil
}
