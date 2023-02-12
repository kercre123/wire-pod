package rts

import (
	"bytes"
)

const (
	errUnsupportedVersion = "unsupported message for this RTS type"

	rtsv2 = 2
	rtsv3 = 3
	rtsv4 = 4
	rtsv5 = 5
)

func buildMessage2(message *RtsConnection_2) ([]byte, error) {
	ec := NewExternalCommsWithRtsConnection(
		NewRtsConnectionWithRtsConnection2(
			message,
		),
	)
	var br bytes.Buffer
	if err := ec.Pack(&br); err != nil {
		return nil, err
	}
	return br.Bytes(), nil
}

func buildMessage3(message *RtsConnection_3) ([]byte, error) {
	ec := NewExternalCommsWithRtsConnection(
		NewRtsConnectionWithRtsConnection3(
			message,
		),
	)
	var br bytes.Buffer
	if err := ec.Pack(&br); err != nil {
		return nil, err
	}
	return br.Bytes(), nil
}

func buildMessage4(message *RtsConnection_4) ([]byte, error) {
	ec := NewExternalCommsWithRtsConnection(
		NewRtsConnectionWithRtsConnection4(
			message,
		),
	)
	var br bytes.Buffer
	if err := ec.Pack(&br); err != nil {
		return nil, err
	}
	return br.Bytes(), nil
}

func buildMessage5(message *RtsConnection_5) ([]byte, error) {
	ec := NewExternalCommsWithRtsConnection(
		NewRtsConnectionWithRtsConnection5(
			message,
		),
	)
	var br bytes.Buffer
	if err := ec.Pack(&br); err != nil {
		return nil, err
	}
	return br.Bytes(), nil
}
