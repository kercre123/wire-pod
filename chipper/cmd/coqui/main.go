package main

import (
	"github.com/kercre123/wire-pod/chipper/pkg/initwirepod"
	stt "github.com/kercre123/wire-pod/chipper/pkg/wirepod/stt/coqui"
)

func main() {
	initwirepod.StartFromProgramInit(stt.Init, stt.STT, stt.Name)
}
