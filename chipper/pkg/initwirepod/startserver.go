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

func StartServer(sttInitFunc func() error, sttHandlerFunc interface{}, voiceProcessorName string) {
	logger.Init()

	// begin wirepod stuff
	vars.Init()
	p, err := wp.New(sttInitFunc, sttHandlerFunc, voiceProcessorName)
	go wpweb.StartWebServer()
	wpweb.SttInitFunc = sttInitFunc
	go sdkWeb.BeginServer()
	if err != nil {
		logger.Println(err)
		os.Exit(1)
	}

	logger.Println("Initiating TLS listener, cmux, gRPC handler, and REST handler")
	cert, err := tls.X509KeyPair([]byte(os.Getenv("DDL_RPC_TLS_CERTIFICATE")), []byte(os.Getenv("DDL_RPC_TLS_KEY")))
	if err != nil {
		logger.Println(err)
		os.Exit(1)
	}

	listener, err := tls.Listen("tcp", ":"+os.Getenv("DDL_RPC_PORT"), &tls.Config{
		Certificates: []tls.Certificate{cert},
		CipherSuites: nil,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	m := cmux.New(listener)
	grpcListener := m.Match(cmux.HTTP2())
	httpListener := m.Match(cmux.HTTP1Fast())
	go grpcServe(grpcListener, p)
	go httpServe(httpListener)
	var m2 cmux.CMux

	if os.Getenv("DDL_RPC_PORT") == "443" && os.Getenv("NO8084") != "true" {
		logger.Println("Starting server at ports 443 and 8084 for 2.0.1 compatibility")
		listener2, err := tls.Listen("tcp", ":8084", &tls.Config{
			Certificates: []tls.Certificate{cert},
			CipherSuites: nil,
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		m2 = cmux.New(listener2)
		grpcListener := m2.Match(cmux.HTTP2())
		httpListener := m2.Match(cmux.HTTP1Fast())
		go grpcServe(grpcListener, p)
		go httpServe(httpListener)
	}

	fmt.Println("\033[33m\033[1mwire-pod started successfully!\033[0m")

	if os.Getenv("DDL_RPC_PORT") == "443" && os.Getenv("NO8084") != "true" {
		go m.Serve()
		m2.Serve()
	} else {
		m.Serve()
	}
}
