package main

import (
	"fmt"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
	jdocspb "github.com/digital-dream-labs/api/go/jdocspb"
	tokenpb "github.com/digital-dream-labs/api/go/tokenpb"
	"github.com/digital-dream-labs/chipper/pkg/server"
	tokenserver "github.com/digital-dream-labs/chipper/pkg/tokenserver"
	wp "github.com/digital-dream-labs/chipper/pkg/voice_processors"
	grpclog "github.com/digital-dream-labs/hugh/grpc/interceptors/log"

	//	grpclog "github.com/digital-dream-labs/hugh/grpc/interceptors/log"
	warnlog "log"
	"os"

	"context"

	grpcserver "github.com/digital-dream-labs/hugh/grpc/server"
	"github.com/digital-dream-labs/hugh/log"
)

// set false for no warning
const warnIfNoSTT string = "true"

var sttLanguage string = "en-US"

type JdocServer struct {
	jdocspb.UnimplementedJdocsServer
}

func (s *JdocServer) WriteDoc(ctx context.Context, req *jdocspb.WriteDocReq) (*jdocspb.WriteDocResp, error) {
	fmt.Println("test")
	fmt.Println(req.Doc)
	fmt.Println(req.DocName)
	fmt.Println(req.Thing)
	fmt.Println(req.UserId)
	return &jdocspb.WriteDocResp{Status: 1}, nil
}
func (s *JdocServer) ReadDoc(ctx context.Context, req *jdocspb.ReadDocsReq) (*jdocspb.ReadDocsReq, error) {
	fmt.Println("test")
	fmt.Println(req.Items)
	fmt.Println(req.Thing)
	fmt.Println(req.UserId)
	return &jdocspb.ReadDocsReq{}, nil
}

func main() {
	sttLanguage = os.Getenv("STT_LANGUAGE")

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
		grpcserver.WithLogger(log.Base()),
		grpcserver.WithReflectionService(),

		grpcserver.WithUnaryServerInterceptors(
			grpclog.UnaryServerInterceptor(),
		),

		grpcserver.WithStreamServerInterceptors(
			grpclog.StreamServerInterceptor(),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	p, err := wp.New(wp.VoiceProcessorVosk)
	go wp.StartWebServer()
	wp.InitHoundify()
	if err != nil {
		log.Fatal(err)
	}

	s, _ := server.New(
		//server.WithLogger(log.Base()),
		server.WithIntentProcessor(p),
		server.WithKnowledgeGraphProcessor(p),
		server.WithIntentGraphProcessor(p),
	)

	tokenServer := tokenserver.NewTokenServer()

	pb.RegisterChipperGrpcServer(srv.Transport(), s)
	jdocspb.RegisterJdocsServer(srv.Transport(), &JdocServer{})
	tokenpb.RegisterTokenServer(srv.Transport(), tokenServer)

	srv.Start()
	fmt.Println("\033[33m\033[1mServer started successfully!\033[0m")
	<-srv.Notify(grpcserver.Stopped)
}
