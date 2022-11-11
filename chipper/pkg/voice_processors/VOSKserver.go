package wirepod

import (
	"log"
	//"bufio"
	//"io"
	//"encoding/json"
	vosk "github.com/alphacep/vosk-api/go"
)

var model *vosk.VoskModel

// New returns a new server
func VOSKNew() (*Server, error) {
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
