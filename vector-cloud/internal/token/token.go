package token

import (
	"bytes"
	"context"
	"fmt"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"

	"github.com/digital-dream-labs/vector-cloud/internal/ipc"
	"github.com/digital-dream-labs/vector-cloud/internal/log"
	"github.com/digital-dream-labs/vector-cloud/internal/token/identity"
	"github.com/digital-dream-labs/vector-cloud/internal/util"
)

// Server encapsulates the receiving and queueing of token requests by other robot processes
type Server struct {
	initialized      bool
	queue            tokenQueue
	identityProvider identity.Provider
	backoffHandler   *backoffHandler
}

type RequestHandler interface {
	handleRequest(req *cloud.TokenRequest) (*cloud.TokenResponse, error)
}

func (s *Server) ErrorListener() util.ErrorListener {
	return s.backoffHandler
}

// Init initializes the token service in advance of other services that depend on it
func (s *Server) Init(identityProvider identity.Provider) error {
	s.backoffHandler = NewBackoffHandler(s)

	if identityProvider == nil {
		return fmt.Errorf("Error initializing identity provider")
	}
	s.identityProvider = identityProvider

	if err := identityProvider.Init(); err != nil {
		log.Println("Error initializing jwt store:", err)
		return err
	}
	s.initialized = true
	return nil
}

// Run starts the token service for other code/processes to connect to and
// request tokens
func (s *Server) Run(ctx context.Context, optionValues ...Option) {
	var opts options
	for _, o := range optionValues {
		o(&opts)
	}

	if !s.initialized {
		if err := s.identityProvider.Init(); err != nil {
			return
		}
	}

	if err := s.queue.init(ctx, s.backoffHandler, s.identityProvider); err != nil {
		log.Println("Error initializing request queue:", err)
		return
	}

	initRefresher(ctx, &s.queue, s.identityProvider)

	if opts.server {
		socketName := "token_server"
		if opts.socketNameSuffix != "" {
			socketName = fmt.Sprintf("%s_%s", socketName, opts.socketNameSuffix)
		}

		serv, err := s.initServer(ctx, socketName)
		if err != nil {
			log.Println("Error creating token server:", err)
			return
		}

		for c := range serv.NewConns() {
			go s.handleConn(c)
		}
	}
	// if server isn't requested, our background routines will handle requests
	// and there's no need for this function to block
}

func (s *Server) handleConn(c ipc.Conn) {
	for {
		buf := c.ReadBlock()
		// TODO: will this ever close?
		if buf == nil || len(buf) == 0 {
			return
		}
		var msg cloud.TokenRequest
		if err := msg.Unpack(bytes.NewBuffer(buf)); err != nil {
			log.Println("Could not unpack token request:", err)
			continue
		}

		resp, err := s.handleRequest(&msg)
		if err != nil {
			log.Println("Error handling token request:", err)
		}
		if resp != nil {
			var buf bytes.Buffer
			if err := resp.Pack(&buf); err != nil {
				log.Println("Error packing token response:", err)
			} else if n, err := c.Write(buf.Bytes()); n != buf.Len() || err != nil {
				log.Println("Error sending token response:", fmt.Sprintf("%d/%d,", n, buf.Len()), err)
			}
		}
	}
}

// HandleRequest will process the given request and return a response. It may block,
// either due to waiting for other requests to process or due to waiting for gRPC.
func (s *Server) handleRequest(m *cloud.TokenRequest) (*cloud.TokenResponse, error) {
	req := request{m: m, ch: make(chan *response)}
	defer close(req.ch)
	s.queue.queue <- req
	resp := <-req.ch
	return resp.resp, resp.err
}

func (s *Server) initServer(ctx context.Context, socketName string) (ipc.Server, error) {
	serv, err := ipc.NewUnixgramServer(ipc.GetSocketPath(socketName))
	if err != nil {
		return nil, err
	}

	if done := ctx.Done(); done != nil {
		go func() {
			<-done
			if err := serv.Close(); err != nil {
				log.Println("error closing token server:", err)
			}
		}()
	}
	return serv, nil
}
