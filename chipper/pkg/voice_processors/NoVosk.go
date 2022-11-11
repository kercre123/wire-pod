//go:build !vosk
// +build !vosk

package wirepod

func VOSKNew() (*Server, error) {
	return &Server{}, nil
}

func VOSKSttHandler(reqThing interface{}, isKnowledgeGraph bool) (transcribedString string, slots map[string]string, isRhino bool, thisBotNum int, opusUsed bool, err error) {
	return "", nil, false, 0, false, nil
}
