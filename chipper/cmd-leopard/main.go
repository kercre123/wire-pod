package main

import (
	"fmt"
	wp "github.com/digital-dream-labs/chipper/pkg/voice_processors"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/chipper/pkg/server"
	//	grpclog "github.com/digital-dream-labs/hugh/grpc/interceptors/logger"

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

	p, err := wp.New()
	go wp.StartWebServer()
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

	pb.RegisterChipperGrpcServer(srv.Transport(), s)

	srv.Start()
	fmt.Println("\033[33m\033[1mServer started successfully!\033[0m")
	<-srv.Notify(grpcserver.Stopped)
}
