package initwirepod

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	chipperpb "github.com/digital-dream-labs/api/go/chipperpb"
	"github.com/digital-dream-labs/api/go/jdocspb"
	"github.com/digital-dream-labs/api/go/tokenpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/mdnshandler"
	chipperserver "github.com/kercre123/wire-pod/chipper/pkg/servers/chipper"
	jdocsserver "github.com/kercre123/wire-pod/chipper/pkg/servers/jdocs"
	tokenserver "github.com/kercre123/wire-pod/chipper/pkg/servers/token"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	wpweb "github.com/kercre123/wire-pod/chipper/pkg/wirepod/config-ws"
	wp "github.com/kercre123/wire-pod/chipper/pkg/wirepod/preqs"
	sdkWeb "github.com/kercre123/wire-pod/chipper/pkg/wirepod/sdkapp"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var voiceProcessor *wp.Server

var (
	chipperMu      sync.Mutex
	chipperClosers []func()
	chipperDone    chan struct{}
)

func serveOk(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func httpServe(l net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok:80", serveOk)
	mux.HandleFunc("/ok", serveOk)
	s := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
	return s.Serve(l)
}

func logPanic(method string, r interface{}) error {
	logger.Println("gRPC: panic in " + method + ": " + fmt.Sprint(r) + "\n" + string(debug.Stack()))
	return status.Error(codes.Internal, "internal error")
}

func recoverUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = logPanic(info.FullMethod, r)
		}
	}()
	return handler(ctx, req)
}

func recoverStream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = logPanic(info.FullMethod, r)
		}
	}()
	return handler(srv, ss)
}

func grpcServe(l net.Listener, p *wp.Server) error {
	return NewGRPCServer(p).Serve(l)
}

func NewGRPCServer(p *wp.Server) *grpc.Server {
	srv := grpc.NewServer(
		grpc.Creds(cmuxTLSCreds{}),
		grpc.ChainUnaryInterceptor(recoverUnary),
		grpc.ChainStreamInterceptor(recoverStream),
	)
	reflection.Register(srv)

	s, _ := chipperserver.New(
		chipperserver.WithIntentProcessor(p),
		chipperserver.WithKnowledgeGraphProcessor(p),
		chipperserver.WithIntentGraphProcessor(p),
	)

	tokenServer := tokenserver.NewTokenServer()
	jdocsServer := jdocsserver.NewJdocsServer()

	chipperpb.RegisterChipperGrpcServer(srv, s)
	jdocspb.RegisterJdocsServer(srv, jdocsServer)
	tokenpb.RegisterTokenServer(srv, tokenServer)
	return srv
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
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		os.Setenv("DEBUG_LOGGING", "true")
		os.Setenv("STT_SERVICE", "vosk")
	}
	err := BeginWirepodSpecific(sttInitFunc, sttHandlerFunc, voiceProcessorName)
	if err != nil {
		logger.Println("\033[33m\033[1mWire-pod is not setup. Use the webserver at port 8080 to set up wire-pod.\033[0m")
	} else if !vars.Config().PastInitialSetup {
		logger.Println("\033[33m\033[1mWire-pod is not setup. Use the webserver at port 8080 to set up wire-pod.\033[0m")
	} else if (vars.Config().STT.Service == "vosk" || vars.Config().STT.Service == "whisper.cpp") && vars.Config().STT.Language == "" {
		logger.Println("\033[33m\033[1mLanguage value is blank, but STT service is " + vars.Config().STT.Service + ". Reinitiating setup process.\033[0m")
		logger.Println("\033[33m\033[1mWire-pod is not setup. Use the webserver at port 8080 to set up wire-pod.\033[0m")
		vars.UpdateConfig(func(c *vars.APIConfig) {
			c.PastInitialSetup = false
		})
	} else {
		go StartChipper()
	}
	// main thread is configuration ws
	wpweb.StartWebServer()
}

func StopServer() {
	mdnshandler.StopmDNS()
	chipperMu.Lock()
	closers := chipperClosers
	done := chipperDone
	chipperClosers = nil
	chipperDone = nil
	chipperMu.Unlock()
	for _, c := range closers {
		c()
	}
	if done == nil {
		return
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		logger.Println("Timed out waiting for chipper server to stop")
	}
}

func RestartServer() {
	StopServer()
	go StartChipper()
}

func loadChipperCert(cfg vars.APIConfig) (tls.Certificate, error) {
	var certPub, certPriv []byte
	var err error
	switch {
	case cfg.Server.EPConfig && (runtime.GOOS == "android" || runtime.GOOS == "ios"):
		certPub, _ = os.ReadFile(vars.AndroidPath + "/static/epod/ep.crt")
		certPriv, err = os.ReadFile(vars.AndroidPath + "/static/epod/ep.key")
	case cfg.Server.EPConfig:
		certPub, _ = os.ReadFile("./epod/ep.crt")
		certPriv, err = os.ReadFile("./epod/ep.key")
	case runtime.GOOS == "android" || runtime.GOOS == "ios":
		certPub, _ = os.ReadFile(vars.AndroidPath + "/wire-pod/certs/cert.crt")
		certPriv, err = os.ReadFile(vars.AndroidPath + "/wire-pod/certs/cert.key")
	default:
		certPub, _ = os.ReadFile("../certs/cert.crt")
		certPriv, err = os.ReadFile("../certs/cert.key")
	}
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.X509KeyPair(certPub, certPriv)
}

func StartChipper() {
	cfg := vars.Config()
	if cfg.Server.EPConfig && runtime.GOOS != "android" {
		go mdnshandler.PostmDNS()
	}
	cert, err := loadChipperCert(cfg)
	if err != nil {
		logger.Println("wire-pod is not setup (couldn't load chipper cert): ", err)
		return
	}

	logger.Println("Initiating TLS listener, cmux, gRPC handler, and REST handler")
	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequestClientCert,
	}
	var ports []string
	if runtime.GOOS == "android" && cfg.Server.Port == "443" {
		logger.Println("not starting chipper at port 443 because android")
	} else {
		ports = append(ports, cfg.Server.Port)
	}
	if cfg.Server.EPConfig && os.Getenv("NO8084") != "true" {
		ports = append(ports, "8084")
	}

	var muxes []cmux.CMux
	var closers []func()
	closeAll := func() {
		for _, c := range closers {
			c()
		}
	}
	for _, port := range ports {
		if port == "8084" {
			logger.Println("Starting chipper server at port 8084 for 2.0.1 compatibility")
		} else {
			logger.Println("Starting chipper server at port " + port)
		}
		l, err := tls.Listen("tcp", ":"+port, tlsConf)
		if err != nil {
			logger.Println("Error starting chipper server at port "+port+": ", err)
			closeAll()
			return
		}
		m := cmux.New(l)
		go grpcServe(m.Match(cmux.HTTP2()), voiceProcessor)
		go httpServe(m.Match(cmux.HTTP1Fast()))
		muxes = append(muxes, m)
		closers = append(closers, m.Close, func() { l.Close() })
	}
	if len(muxes) == 0 {
		logger.Println("No chipper ports to serve")
		return
	}

	done := make(chan struct{})
	chipperMu.Lock()
	chipperClosers = closers
	chipperDone = done
	chipperMu.Unlock()

	fmt.Println("\033[33m\033[1mwire-pod started successfully!\033[0m")

	var wg sync.WaitGroup
	for _, m := range muxes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Serve()
		}()
	}
	wg.Wait()
	logger.Println("Stopping chipper server")
	close(done)
}
