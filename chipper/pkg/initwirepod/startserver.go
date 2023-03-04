package initwirepod

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"

	chipperpb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/api/go/jdocspb"
	"github.com/digital-dream-labs/api/go/tokenpb"
	"github.com/digital-dream-labs/hugh/log"
	"github.com/kercre123/chipper/pkg/logger"
	chipperserver "github.com/kercre123/chipper/pkg/servers/chipper"
	jdocsserver "github.com/kercre123/chipper/pkg/servers/jdocs"
	tokenserver "github.com/kercre123/chipper/pkg/servers/token"
	"github.com/kercre123/chipper/pkg/vars"
	wpweb "github.com/kercre123/chipper/pkg/wirepod/config-ws"
	wp "github.com/kercre123/chipper/pkg/wirepod/preqs"
	sdkWeb "github.com/kercre123/chipper/pkg/wirepod/sdkapp"
	"github.com/soheilhy/cmux"

	//	grpclog "github.com/digital-dream-labs/hugh/grpc/interceptors/logger"

	grpcserver "github.com/digital-dream-labs/hugh/grpc/server"
)

var serverOne cmux.CMux
var serverTwo cmux.CMux
var listenerOne net.Listener
var listenerTwo net.Listener
var voiceProcessor *wp.Server

// grpcServer *grpc.Servervar
var chipperServing bool = false

func serveOk(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func httpServe(l net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok:80", serveOk)
	mux.HandleFunc("/ok", serveOk)
	s := &http.Server{
		Handler: mux,
	}
	return s.Serve(l)
}

func grpcServe(l net.Listener, p *wp.Server) error {
	srv, err := grpcserver.New(
		grpcserver.WithViper(),
		grpcserver.WithReflectionService(),
		grpcserver.WithInsecureSkipVerify(),
	)
	if err != nil {
		log.Fatal(err)
	}

	s, _ := chipperserver.New(
		chipperserver.WithIntentProcessor(p),
		chipperserver.WithKnowledgeGraphProcessor(p),
		chipperserver.WithIntentGraphProcessor(p),
	)

	tokenServer := tokenserver.NewTokenServer()
	jdocsServer := jdocsserver.NewJdocsServer()
	//jdocsserver.IniToJson()

	chipperpb.RegisterChipperGrpcServer(srv.Transport(), s)
	jdocspb.RegisterJdocsServer(srv.Transport(), jdocsServer)
	tokenpb.RegisterTokenServer(srv.Transport(), tokenServer)

	return srv.Transport().Serve(l)
}

func BeginWirepodSpecific(sttInitFunc func() error, sttHandlerFunc interface{}, voiceProcessorName string) error {
	logger.Init()

	// begin wirepod stuff
	vars.Init()
	var err error
	voiceProcessor, err = wp.New(sttInitFunc, sttHandlerFunc, voiceProcessorName)
	wpweb.SttInitFunc = sttInitFunc
	go sdkWeb.BeginServer()
	http.HandleFunc("/api-chipper/", ChipperHTTPApi)
	if err != nil {
		return err
	}
	return nil
}

func StartFromProgramInit(sttInitFunc func() error, sttHandlerFunc interface{}, voiceProcessorName string) {
	err := BeginWirepodSpecific(sttInitFunc, sttHandlerFunc, voiceProcessorName)
	if err != nil {
		logger.Println("Wire-pod is not setup. Use the webserver at port 8080 to set up wire-pod.")
	} else if !vars.APIConfig.PastInitialSetup {
		logger.Println("Wire-pod is not setup. Use the webserver at port 8080 to set up wire-pod.")
	} else {
		go StartChipper()
	}
	// main thread is configuration ws
	wpweb.StartWebServer()
}

func CheckHostname() {
	hostname, _ := os.Hostname()
	if hostname != "escapepod" && vars.APIConfig.Server.EPConfig {
		logger.Println("\033[31m\033[1mWARNING: You have chosen the Escape Pod config, but the system hostname is not 'escapepod'. This means your robot will not be able to communicate with wire-pod unless you have a custom network configuration.")
		logger.Println("Actual reported hostname: " + hostname + "\033[0m")
	}
}

func RestartServer() {
	if chipperServing {
		serverOne.Close()
		serverTwo.Close()
		listenerOne.Close()
		listenerTwo.Close()
	}
	go StartChipper()
}

func StartChipper() {
	CheckHostname()
	// load certs
	var certPub []byte
	var certPriv []byte
	if vars.APIConfig.Server.EPConfig {
		certPub, _ = os.ReadFile("./epod/ep.crt")
		certPriv, _ = os.ReadFile("./epod/ep.key")
	} else {
		var err error
		certPub, _ = os.ReadFile("../certs/cert.crt")
		certPriv, err = os.ReadFile("../certs/cert.key")
		if err != nil {
			logger.Println("wire-pod is not setup.")
			return
		}
	}

	logger.Println("Initiating TLS listener, cmux, gRPC handler, and REST handler")
	cert, err := tls.X509KeyPair(certPub, certPriv)
	if err != nil {
		logger.Println(err)
		os.Exit(1)
	}
	logger.Println("Starting chipper server at port " + vars.APIConfig.Server.Port)
	listenerOne, err = tls.Listen("tcp", ":"+vars.APIConfig.Server.Port, &tls.Config{
		Certificates: []tls.Certificate{cert},
		CipherSuites: nil,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	serverOne = cmux.New(listenerOne)
	grpcListenerOne := serverOne.Match(cmux.HTTP2())
	httpListenerOne := serverOne.Match(cmux.HTTP1Fast())
	go grpcServe(grpcListenerOne, voiceProcessor)
	go httpServe(httpListenerOne)

	if vars.APIConfig.Server.EPConfig && os.Getenv("NO8084") != "true" {
		logger.Println("Starting chipper server at port 8084 for 2.0.1 compatibility")
		listenerTwo, err = tls.Listen("tcp", ":8084", &tls.Config{
			Certificates: []tls.Certificate{cert},
			CipherSuites: nil,
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		serverTwo = cmux.New(listenerTwo)
		grpcListenerTwo := serverTwo.Match(cmux.HTTP2())
		httpListenerTwo := serverTwo.Match(cmux.HTTP1Fast())
		go grpcServe(grpcListenerTwo, voiceProcessor)
		go httpServe(httpListenerTwo)
	}

	fmt.Println("\033[33m\033[1mwire-pod started successfully!\033[0m")

	chipperServing = true
	if vars.APIConfig.Server.EPConfig && os.Getenv("NO8084") != "true" {
		go serverOne.Serve()
		serverTwo.Serve()
		logger.Println("Stopping chipper server")
		chipperServing = false
	} else {
		serverOne.Serve()
		logger.Println("Stopping chipper server")
		chipperServing = false
	}
}
