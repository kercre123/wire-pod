package main

import (
	"github.com/kercre123/chipper/pkg/initwirepod"
	stt "github.com/kercre123/chipper/pkg/wirepod/stt/houndify"
)

func main() {
	initwirepod.StartFromProgramInit(stt.Init, stt.STT, stt.Name)
}
