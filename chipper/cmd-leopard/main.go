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

	grpcserver "github.com/digital-dream-labs/hugh/grpc/server"
	"github.com/digital-dream-labs/hugh/log"
)

func main() {
	log.SetJSONFormat("2006-01-02 15:04:05")
	startServer()
}

func startServer() {
	srv, err := grpcserver.New(
		grpcserver.WithViper(),
		//grpcserver.WithLogger(log.Base()),
		grpcserver.WithReflectionService(),

		grpcserver.WithUnaryServerInterceptors(
		//	grpclog.UnaryServerInterceptor(),
		),

		grpcserver.WithStreamServerInterceptors(
		//	grpclog.StreamServerInterceptor(),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	p, err := wp.New(wp.VoiceProcessorLeopard)
	go wp.StartWebServer()
	go sdkWeb.BeginServer()
	wp.InitHoundify()
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
	//jdocsserver.IniToJson()

	pb.RegisterChipperGrpcServer(srv.Transport(), s)
	jdocspb.RegisterJdocsServer(srv.Transport(), jdocsServer)
	tokenpb.RegisterTokenServer(srv.Transport(), tokenServer)

	srv.Start()
	fmt.Println("\033[33m\033[1mServer started successfully!\033[0m")
	<-srv.Notify(grpcserver.Stopped)
}
