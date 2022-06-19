package cloudproc

import (
	"context"
	"sync"

	"github.com/digital-dream-labs/vector-cloud/internal/log"
	"github.com/digital-dream-labs/vector-cloud/internal/logcollector"
	"github.com/digital-dream-labs/vector-cloud/internal/offboard_vision"
	"github.com/digital-dream-labs/vector-cloud/internal/token"
	"github.com/digital-dream-labs/vector-cloud/internal/token/identity"
	"github.com/digital-dream-labs/vector-cloud/internal/voice"

	"github.com/digital-dream-labs/vector-cloud/internal/jdocs"
)

var devServer func() error

func Run(ctx context.Context, procOptions ...Option) {
	var opts options
	for _, o := range procOptions {
		o(&opts)
	}

	var wg sync.WaitGroup
	if devServer != nil {
		launchProcess(&wg, func() {
			if err := devServer(); err != nil {
				log.Println("dev HTTP server reported error:", err)
			}
		})
	}
	// start token service synchronously since everything else depends on it
	identityProvider := opts.identityProvider
	if identityProvider == nil {
		// If the identity provider is not provided as an option, we create a default
		// file backed identity provider with platform specific storage paths for JWT
		// and certs (see getcert_*.go files)
		var err error
		if identityProvider, err = identity.NewFileProvider("", ""); err != nil {
			log.Println("Fatal error initializing default identity provider:", err)
			return
		}
	}

	var tokenServer = new(token.Server)
	if err := tokenServer.Init(identityProvider); err != nil {
		log.Println("Fatal error initializing token service:", err)
		return
	}
	addHandlers(token.GetDevHandlers, tokenServer)
	launchProcess(&wg, func() {
		tokenServer.Run(ctx, opts.tokenOpts...)
	})
	tokener := token.GetAccessor(identityProvider, tokenServer)
	if opts.voice != nil {
		launchProcess(&wg, func() {
			// provide default token accessor
			voiceOpts := append([]voice.Option{voice.WithTokener(tokener),
				voice.WithErrorListener(tokenServer.ErrorListener())},
				opts.voiceOpts...)
			opts.voice.Run(ctx, voiceOpts...)
		})
	}
	if opts.jdocOpts != nil {
		launchProcess(&wg, func() {
			// provide default token accessor
			jdocOpts := append([]jdocs.Option{jdocs.WithTokener(tokener),
				jdocs.WithErrorListener(tokenServer.ErrorListener())},
				opts.jdocOpts...)
			jdocs.Run(ctx, jdocOpts...)
		})
	}
	if opts.logcollectorOpts != nil {
		launchProcess(&wg, func() {
			logcollectorOpts := append([]logcollector.Option{logcollector.WithTokener(tokener),
				logcollector.WithErrorListener(tokenServer.ErrorListener())},
				opts.logcollectorOpts...)
			logcollector.Run(ctx, logcollectorOpts...)
		})
	}
	launchProcess(&wg, func() {
		offboard_vision.Run(ctx)
	})
	addHandlers(offboard_vision.GetDevHandlers, tokenServer)
	wg.Wait()
}

func launchProcess(wg *sync.WaitGroup, launcher func()) {
	wg.Add(1)
	go func() {
		launcher()
		wg.Done()
	}()
}
