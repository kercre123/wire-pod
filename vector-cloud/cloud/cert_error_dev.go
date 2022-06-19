// +build !shipping,vicos

package main

import (
	"github.com/digital-dream-labs/vector-cloud/internal/log"
	"github.com/digital-dream-labs/vector-cloud/internal/robot"
)

func init() {
	certErrorFunc = onCertError
}

func onCertError() bool {
	if err := robot.WriteFaceErrorCode(850); err != nil {
		log.Println("Error writing error code (isn't it ironic?):", err)
	}
	return true
}
