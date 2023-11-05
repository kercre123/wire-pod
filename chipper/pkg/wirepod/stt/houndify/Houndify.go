package wirepod_vosk

import (
	"fmt"
	"io"
	"os"

	"github.com/kercre123/chipper/pkg/logger"
	preqs "github.com/kercre123/chipper/pkg/wirepod/preqs"
	sr "github.com/kercre123/chipper/pkg/wirepod/speechrequest"
	"github.com/soundhound/houndify-sdk-go"
)

// to use, you must create a Houndify client with the only domain enabled being "Speech to text only"
// set HOUNDIFY_STT_ID and HOUNDIFY_STT_KEY to the respective strings you will find on the dashboard
// also set STT_SERVICE to "houndify"

var Name string = "houndify"

var houndSTTClient houndify.Client

func Init() error {
	if os.Getenv("HOUNDIFY_STT_ID") == "" {
		logger.Println("Houndify STT Client ID not found.")
		return fmt.Errorf("houndify stt client id not found")
	}
	if os.Getenv("HOUNDIFY_STT_KEY") == "" {
		logger.Println("Houndify STT Client Key not found.")
		return fmt.Errorf("houndify stt client key not found")
	}
	houndSTTClient = houndify.Client{
		ClientID:  os.Getenv("HOUNDIFY_STT_ID"),
		ClientKey: os.Getenv("HOUNDIFY_STT_KEY"),
	}
	houndSTTClient.EnableConversationState()
	logger.Println("Houndify client for speech-to-text initialized!")
	return nil
}

func STT(sreq sr.SpeechRequest) (string, error) {
	logger.Println("Incoming request")
	var err error
	rp, wp := io.Pipe()
	req := houndify.VoiceRequest{
		AudioStream: rp,
		UserID:      sreq.Device,
		RequestID:   sreq.Session,
	}
	done := make(chan bool)
	speechDone := false
	go func(wp *io.PipeWriter) {
		defer wp.Close()

		for {
			select {
			case <-done:
				return
			default:
				var chunk []byte
				chunk, err = sreq.GetNextStreamChunkOpus()
				speechDone = sreq.DetectEndOfSpeech()
				if err != nil {
					fmt.Println("End of stream")
					return
				}
				wp.Write(chunk)
				if speechDone {
					return
				}
			}
		}
	}(wp)

	partialTranscripts := make(chan houndify.PartialTranscript)
	go func() {
		for partial := range partialTranscripts {
			if *partial.SafeToStopAudio {
				fmt.Println("SafeToStopAudio recieved")
				done <- true
				return
			}
		}
	}()

	serverResponse, err := houndSTTClient.VoiceSearch(req, partialTranscripts)
	if err != nil {
		fmt.Println(err)
		fmt.Println(serverResponse)
	}
	resp, _ := preqs.ParseSpokenResponse(serverResponse)
	logger.Println("Houndify response: " + resp)
	return resp, nil
}
