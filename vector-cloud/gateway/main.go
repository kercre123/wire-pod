package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	extint "github.com/digital-dream-labs/vector-cloud/internal/proto/external_interface"

	"github.com/digital-dream-labs/vector-cloud/internal/log"
	"github.com/digital-dream-labs/vector-cloud/internal/robot"

	grpcRuntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
)

// Enables logs about the requests coming and going from the gateway.
// Most useful for debugging the json output being sent to the app.
const (
	logVerbose        = false
	logMessageContent = false
)

var (
	robotHostname          string
	signalHandler          chan os.Signal
	demoKeyPair            *tls.Certificate
	demoCertPool           *x509.CertPool
	cloudCheckLimiter      *MultiLimiter
	debugLogLimiter        *MultiLimiter
	userAuthLimiter        *MultiLimiter
	switchboardManager     SwitchboardIpcManager
	engineProtoManager     EngineProtoIpcManager
	tokenManager           ClientTokenManager
	bleProxy               BLEProxy
	numCommandsSentFromSDK uint32

	// TODO: remove clad socket and map when there are no more clad messages being used
	engineCladManager EngineCladIpcManager
)

func LoggingUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, errOut error) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Recovered from fatal error in %s \"%s\": %s\n", info.FullMethod, err, debug.Stack())
			errOut = grpc.Errorf(codes.Internal, "%s", err)
		}
	}()
	nameList := strings.Split(info.FullMethod, "/")
	name := nameList[len(nameList)-1]
	numCommandsSentFromSDK++
	if logMessageContent {
		log.Printf("Received rpc request %s(%#v)\n", name, req)
	} else {
		log.Printf("Received rpc request %s\n", name)
	}
	resp, err := handler(ctx, req)
	if logMessageContent {
		log.Printf("Sending rpc response %s(%#v)\n", name, resp)
	} else {
		log.Printf("Sending rpc response %s\n", name)
	}
	return resp, err
}

func LoggingStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (errOut error) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Recovered from fatal error in %s \"%s\": %s\n", info.FullMethod, err, debug.Stack())
			errOut = grpc.Errorf(codes.Internal, "%s", err)
		}
	}()
	nameList := strings.Split(info.FullMethod, "/")
	name := nameList[len(nameList)-1]
	numCommandsSentFromSDK++
	if logMessageContent {
		log.Printf("Received stream request %s(%#v)\n", name, srv)
	} else {
		log.Printf("Received stream request %s\n", name)
	}
	return handler(srv, ss)
}

func verboseHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			name, err := checkAuth(w, r) // Note: we only check here because json will forward to grpc, and doesn't need the same auth
			if err != nil {
				http.Error(w, grpc.ErrorDesc(err), http.StatusUnauthorized)
				return
			}
			log.Printf("Authorized connection from '%s'\n", name)
			LogRequest(r, "grpc")
			wrap := WrappedResponseWriter{w, "grpc"}
			grpcServer.ServeHTTP(&wrap, r)
		} else {
			LogRequest(r, "json")
			wrap := WrappedResponseWriter{w, "json"}
			otherHandler.ServeHTTP(&wrap, r)
		}
	})
}

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			_, err := checkAuth(w, r) // Note: we only check here because json will forward to grpc, and doesn't need the same auth
			if err != nil {
				http.Error(w, grpc.ErrorDesc(err), http.StatusUnauthorized)
				return
			}
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func cleanExit() {
	log.Println("Uninstall crash reporter")
	robot.UninstallCrashReporter()

	log.Println("Closed vic-gateway")

	os.Exit(0)
}

func main() {
	log.Tag = "vic-gateway"
	log.Println("Launching vic-gateway")

	log.Println("Install crash reporter")
	robot.InstallCrashReporter("vic-gateway")

	signalHandler = make(chan os.Signal, 1)
	signal.Notify(signalHandler, syscall.SIGTERM)
	go func() {
		sig := <-signalHandler
		log.Println("Received signal:", sig)
		cleanExit()
	}()

	if _, err := os.Stat(robot.GatewayCert); os.IsNotExist(err) {
		log.Println("Cannot find cert:", robot.GatewayCert)
		os.Exit(1)
	}
	if _, err := os.Stat(robot.GatewayKey); os.IsNotExist(err) {
		log.Println("Cannot find key: ", robot.GatewayKey)
		os.Exit(1)
	}

	pair, err := tls.LoadX509KeyPair(robot.GatewayCert, robot.GatewayKey)
	if err != nil {
		log.Println("Failed to initialize key pair")
		os.Exit(1)
	}
	demoKeyPair = &pair
	demoCertPool = x509.NewCertPool()
	caCert, err := ioutil.ReadFile(robot.GatewayCert)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	ok := demoCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		log.Println("Error: Bad certificates.")
		panic("Bad certificates.")
	}
	addr := fmt.Sprintf("localhost:%d", Port)

	engineCladManager.Init()
	defer engineCladManager.Close()

	engineProtoManager.Init()
	defer engineProtoManager.Close()

	if IsOnRobot {
		switchboardManager.Init()
		defer switchboardManager.Close()

		tokenManager.Init()
		defer tokenManager.Close()
	}

	log.Println("Sockets successfully created")

	cloudCheckLimiter = NewMultiLimiter(
		rate.NewLimiter(rate.Every(10*time.Second), 1),
		rate.NewLimiter(rate.Every(time.Minute), 3),
	)

	debugLogLimiter = NewMultiLimiter(
		rate.NewLimiter(rate.Every(time.Minute), 1),
		rate.NewLimiter(rate.Every(time.Hour), 3),
	)

	userAuthLimiter = NewMultiLimiter(
		rate.NewLimiter(rate.Every(10*time.Second), 10),
		rate.NewLimiter(rate.Every(10*time.Minute), 25),
	)

	creds, err := credentials.NewServerTLSFromFile(robot.GatewayCert, robot.GatewayKey)
	if err != nil {
		log.Println("Error creating server tls:", err)
		os.Exit(1)
	}
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(LoggingUnaryInterceptor),
		grpc.StreamInterceptor(LoggingStreamInterceptor),
	)
	extint.RegisterExternalInterfaceServer(grpcServer, newServer())
	ctx := context.Background()

	if runtime.GOOS == "darwin" {
		robotHostname = "Vector-Local"
	} else {
		robotHostname, err = os.Hostname()
		if err != nil {
			log.Println("Failed to get Hostname:", err)
			os.Exit(1)
		}
	}

	log.Println("Hostname:", robotHostname)

	tlsConf := &tls.Config{
		ServerName:   robotHostname,
		Certificates: []tls.Certificate{*demoKeyPair},
		RootCAs:      demoCertPool,
	}
	bleProxy = BLEProxy{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConf,
			},
		},
		Address: addr,
	}
	bleProxy.initialize(grpcServer.GetServiceInfo())
	dcreds := credentials.NewTLS(tlsConf)
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}

	gwmux := grpcRuntime.NewServeMux(grpcRuntime.WithMarshalerOption(grpcRuntime.MIMEWildcard, &grpcRuntime.JSONPb{EmitDefaults: true, OrigName: true, EnumsAsInts: true}))
	err = extint.RegisterExternalInterfaceHandlerFromEndpoint(ctx, gwmux, addr, dopts)
	if err != nil {
		log.Println("Error during RegisterExternalInterfaceHandlerFromEndpoint:", err)
		os.Exit(1)
	}

	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", Port))
	if err != nil {
		log.Println("Error during Listen:", err)
		panic(err)
	}

	handlerFunc := grpcHandlerFunc(grpcServer, gwmux)
	if logVerbose {
		handlerFunc = verboseHandlerFunc(grpcServer, gwmux)
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: handlerFunc,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*demoKeyPair},
			NextProtos:   []string{"h2"},
		},
	}

	go engineCladManager.ProcessMessages()
	go engineProtoManager.ProcessMessages()
	if IsOnRobot {
		go switchboardManager.ProcessMessages()
		go tokenManager.StartUpdateListener()
	}

	log.Println("Listening on Port:", Port)
	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))

	if err != http.ErrServerClosed {
		log.Println("Error during Serve:", err)
		os.Exit(1)
	}

	cleanExit()
}
