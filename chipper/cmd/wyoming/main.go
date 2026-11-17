package main

import (
	"os"

	"github.com/kercre123/wire-pod/chipper/pkg/initwirepod"
	stt "github.com/kercre123/wire-pod/chipper/pkg/wirepod/stt/wyoming"
)

func main() {
	// the config keeps STT.Service in sync with this variable
	if os.Getenv("STT_SERVICE") == "" {
		os.Setenv("STT_SERVICE", stt.Name)
	}
	initwirepod.StartFromProgramInit(stt.Init, stt.STT, stt.Name)
}
