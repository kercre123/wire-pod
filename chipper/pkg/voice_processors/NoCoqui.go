//go:build !coqui
// +build !coqui

package wirepod

func CoquiNew() (*Server, error) {
	return &Server{}, nil
}

func CoquiSttHandler(reqThing interface{}, isKnowledgeGraph bool) (transcribedString string, slots map[string]string, isRhino bool, thisBotNum int, opusUsed bool, err error) {
	return "", nil, false, 0, false, nil
}
