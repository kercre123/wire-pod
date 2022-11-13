//go:build !leopard
// +build !leopard

package wirepod

func LeopardNew() (*Server, error) {
	logger("No Leopard... this function shall never be called")
	return &Server{}, nil
}

func LeopardSttHandler(reqThing interface{}, isKnowledgeGraph bool) (transcribedString string, slots map[string]string, isRhino bool, thisBotNum int, opusUsed bool, err error) {
	logger("No Leopard... this function shall never be called")
	return "", nil, false, 0, false, nil
}
