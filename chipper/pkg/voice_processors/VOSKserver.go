package wirepod

import (
	"log"
	"os"
	//"bufio"
	//"io"
	//"encoding/json"
	vosk "github.com/alphacep/vosk-api/go"
)

var model *vosk.VoskModel

// New returns a new server
func VOSKNew() (*Server, error) {
	initMatches()
	if os.Getenv("DEBUG_LOGGING") != "true" && os.Getenv("DEBUG_LOGGING") != "false" {
		logger("No valid value for DEBUG_LOGGING, setting to true")
		debugLogging = true
	} else {
		if os.Getenv("DEBUG_LOGGING") == "true" {
			debugLogging = true
		} else {
			debugLogging = false
		}
	}
	logger("Server START")

	// Open model
	logger("Opening model")
	aModel, err := vosk.NewModel("../vosk/models/" + sttLanguage + "/model")
	if err != nil {
		log.Fatal(err)
	}
	model = aModel
	logger("Model open!")

	logger("Server OK")

	return &Server{}, nil
}
