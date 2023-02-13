package ble

import "errors"

func handlerUnsupportedVersionError() (data []byte, cont bool, err error) {
	return nil, false, errors.New("unsupported rts protocol version")
}

func handlerUnsupportedTypeError() (data []byte, cont bool, err error) {
	return nil, false, errors.New("unsupported message type")
}
