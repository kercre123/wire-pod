package main

import (
	"fmt"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/chipper/pkg/server"
	"github.com/digital-dream-labs/chipper/pkg/voice_processors/wirepod-vosk"
	//	grpclog "github.com/digital-dream-labs/hugh/grpc/interceptors/logger"
	warnlog "log"
	"os"

	grpcserver "github.com/digital-dream-labs/hugh/grpc/server"
	"github.com/digital-dream-labs/hugh/log"
)

// set false for no warning
const warnIfNoSTT string = "true"

var sttLanguage string = "en-US"

func main() {
	if len(os.Args) > 1 {
		sttLanguage = os.Args[1]
	}

	log.SetJSONFormat("2006-01-02 15:04:05")
	if warnIfNoSTT == "true" {
		if _, err := os.Stat("../vosk"); err == nil {
			warnlog.Println("VOSK directory found!")
			if _, err := os.Stat("../vosk/models"); err == nil {
				warnlog.Println("Models directory found!")
				if _, err := os.Stat("../vosk/models/" + sttLanguage + "/model/am/final.mdl"); err == nil {
					warnlog.Println(sttLanguage + " VOSK model found! Speech-to-text should work like normal.")
				} else {
					warnlog.Println("No " + sttLanguage + " model found. This must be placed at ../vosk/models/" + sttLanguage + "/model. Please read the README. Speech-to-text may not work.")
				}
			} else {
				warnlog.Println("No VOSK models directory found. This must be placed at ../vosk/models. Please read the README. Speech-to-text may not work.")
			}
		} else {
			warnlog.Println("VSOK STT was not found or chipper is being run from outside it's directory. Please read the README. Speech-to-text may not work.")
		}
	}
	startServer()
}

func startServer() {
	srv, err := grpcserver.New(
		grpcserver.WithViper(),
		//grpcserver.WithLogger(logger.Base()),
		grpcserver.WithReflectionService(),

		grpcserver.WithUnaryServerInterceptors(
		//			grpclog.UnaryServerInterceptor(),
		),

		grpcserver.WithStreamServerInterceptors(
		//			grpclog.StreamServerInterceptor(),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	p, err := wirepod.New()
	go wirepod.StartWebServer()
	wirepod.InitHoundify()
	if err != nil {
		log.Fatal(err)
	}

	s, _ := server.New(
		//server.WithLogger(logger.Base()),
		server.WithIntentProcessor(p),
		server.WithKnowledgeGraphProcessor(p),
		server.WithIntentGraphProcessor(p),
	)

	pb.RegisterChipperGrpcServer(srv.Transport(), s)

	srv.Start()
	fmt.Println("\033[33m\033[1mServer started successfully!\033[0m")
	<-srv.Notify(grpcserver.Stopped)
}
