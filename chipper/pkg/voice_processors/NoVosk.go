//go:build !vosk
// +build !vosk

package wirepod

func VoskInit() (*Server, error) {
	logger("No Vosk... this function shall never be called")
	return &Server{}, nil
}

func VoskSTTHandler(req SpeechRequest) (transcribedString string, err error) {
	logger("No Vosk... this function shall never be called")
	return "", nil
}
