//go:build !coqui
// +build !coqui

package wirepod

func CoquiInit() (*Server, error) {
	logger("No Coqui... this function shall never be called")
	return &Server{}, nil
}

func CoquiSttHandler(req SpeechRequest) (string, error) {
	logger("No Coqui... this function shall never be called")
	return "", nil
}
