package main

import (
	"github.com/kercre123/chipper/pkg/initwirepod"
	stt "github.com/kercre123/chipper/pkg/wirepod/stt/leopard"
)

func main() {
	initwirepod.StartServer(stt.Init, stt.STT, stt.Name)
}
