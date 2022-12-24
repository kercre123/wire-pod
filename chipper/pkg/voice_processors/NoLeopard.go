//go:build !leopard
// +build !leopard

package wirepod

func LeopardInit() (*Server, error) {
	logger("No Leopard... this function shall never be called")
	return &Server{}, nil
}

func LeopardSttHandler(req SpeechRequest) (string, error) {
	logger("No Leopard... this function shall never be called")
	return "", nil
}
