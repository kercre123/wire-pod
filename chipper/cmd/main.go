package main

import (
	"fmt"

	"github.com/digital-dream-labs/chipper/pkg/jdocsserver"
	sdkWeb "github.com/digital-dream-labs/chipper/pkg/sdkapp"
	"github.com/digital-dream-labs/chipper/pkg/tokenserver"
	wp "github.com/digital-dream-labs/chipper/pkg/voice_processors"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/api/go/jdocspb"
	"github.com/digital-dream-labs/api/go/tokenpb"
	"github.com/digital-dream-labs/chipper/pkg/server"

	//	grpclog "github.com/digital-dream-labs/hugh/grpc/interceptors/logger"
	warnlog "log"
	"os"

	grpcserver "github.com/digital-dream-labs/hugh/grpc/server"
	"github.com/digital-dream-labs/hugh/log"
)

// set false for no warning
const warnIfNoSTT string = "false"

func main() {
	log.SetJSONFormat("2006-01-02 15:04:05")
	if warnIfNoSTT == "true" {
		if _, err := os.Stat("/root/.coqui/stt"); err == nil {
			warnlog.Println("STT binary found!")
			if _, err := os.Stat("../stt/large_vocabulary.scorer"); err == nil {
				warnlog.Println("STT scorer found!")
				if _, err := os.Stat("../stt/model.tflite"); err == nil {
					warnlog.Println("STT model found! Speech-to-text should work like normal.")
				} else {
					warnlog.Println("No STT model found. This must be placed at ../stt/model.tflite. Please read the README. Speech-to-text may not work.")
				}
			} else {
				warnlog.Println("No scorer file found. This must be placed at ../stt/large_vocabulary.scorer. Please read the README. Speech-to-text may not work.")
			}
		} else {
			warnlog.Println("Coqui STT was not found or chipper is being run from outside it's directory. Please read the README. Speech-to-text may not work.")
		}
	}
	startServer()
}

func startServer() {
	srv, err := grpcserver.New(
		grpcserver.WithViper(),
		grpcserver.WithLogger(log.Base()),
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

	p, err := wp.New(wp.VoiceProcessorCoqui)
	go wp.StartWebServer()
	wp.InitHoundify()
	go sdkWeb.BeginServer()
	if err != nil {
		log.Fatal(err)
	}

	s, _ := server.New(
		//server.WithLogger(logger.Base()),
		server.WithIntentProcessor(p),
		server.WithKnowledgeGraphProcessor(p),
		server.WithIntentGraphProcessor(p),
	)

	tokenServer := tokenserver.NewTokenServer()
	jdocsServer := jdocsserver.NewJdocsServer()
	jdocsserver.IniToJson()

	pb.RegisterChipperGrpcServer(srv.Transport(), s)
	jdocspb.RegisterJdocsServer(srv.Transport(), jdocsServer)
	tokenpb.RegisterTokenServer(srv.Transport(), tokenServer)

	srv.Start()
	fmt.Println("\033[33m\033[1mServer started successfully!\033[0m")
	<-srv.Notify(grpcserver.Stopped)
}
