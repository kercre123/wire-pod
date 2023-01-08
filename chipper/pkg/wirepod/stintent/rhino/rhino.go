package wirepod_rhino

import (
	"fmt"
	"os"
	"strconv"

	rhino "github.com/Picovoice/rhino/binding/go/v2"
	"github.com/kercre123/chipper/pkg/logger"
	sr "github.com/kercre123/chipper/pkg/wirepod/speechrequest"
)

var Name string = "rhino"

var rhinoSTIArray []rhino.Rhino
var picovoiceInstancesOS string = os.Getenv("PICOVOICE_INSTANCES")
var picovoiceInstances int

// New returns a new server
func Init() error {
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
		rhinoSTIArray = append(rhinoSTIArray, rhino.Rhino{AccessKey: picovoiceKey, ContextPath: "./rhino.rhn", Sensitivity: 0.5, EndpointDurationSec: 0.5})
		rhinoSTIArray[i].Init()
	}
	return nil
}

func STT(req sr.SpeechRequest) (intent string, slots map[string]string, err error) {
	logger.Println("(Bot " + strconv.Itoa(req.BotNum) + ", Rhino) Processing...")
	var rhinoSTI rhino.Rhino
	if req.BotNum > picovoiceInstances {
		fmt.Println("Too many bots are connected, sending error to bot " + strconv.Itoa(req.BotNum))
		return "", map[string]string{}, fmt.Errorf("too many bots are connected, max is 3")
	} else {
		rhinoSTI = rhinoSTIArray[req.BotNum-1]
	}
	breakOut := false
	for {
		var chunk []byte
		req, chunk, err = sr.GetNextStreamChunk(req)
		if err != nil {
			return "", map[string]string{}, err
		}
		nint := sr.BytesToSamples(chunk)
		chunks := make([][]int16, len(nint)/512)
		for i := 0; i < len(nint)/512; i++ {
			chunks[i] = nint[i*512 : (i+1)*512]
		}
		for _, bytes := range chunks {
			isFinal, err := rhinoSTI.Process(bytes)
			if err != nil {
				return "", map[string]string{}, err
			}
			if isFinal {
				breakOut = true
				break
			}
		}
		if breakOut {
			break
		}
	}
	inf, err := rhinoSTI.GetInference()
	if err != nil {
		logger.Println(err)
		return "", map[string]string{}, err
	}
	if !inf.IsUnderstood {
		return "", map[string]string{}, fmt.Errorf("inference not understood")
	}
	logger.Println("Bot " + strconv.Itoa(req.BotNum) + " intent: " + inf.Intent)
	return inf.Intent, inf.Slots, nil
}
