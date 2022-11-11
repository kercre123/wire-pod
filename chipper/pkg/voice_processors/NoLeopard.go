//go:build !leopard
// +build !leopard

package wirepod

func LeopardNew() (*Server, error) {
	return &Server{}, nil
}

func LeopardSttHandler(reqThing interface{}, isKnowledgeGraph bool) (transcribedString string, slots map[string]string, isRhino bool, thisBotNum int, opusUsed bool, err error) {
	return "", nil, false, 0, false, nil
}
