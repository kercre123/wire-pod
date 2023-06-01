package token

import (
	"context"
	"time"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"

	"github.com/digital-dream-labs/vector-cloud/internal/log"

	"github.com/cenkalti/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type backoffHandler struct {
	handler RequestHandler
	cancel  context.CancelFunc
	backoff backoff.BackOff
	denied  bool
}

func (b *backoffHandler) reset() {
	// reset can be called from multiple routines, so store cancel in a temp var before
	// executing it so something else doesn't set it to nil first
	f := b.cancel
	b.cancel = nil
	if f != nil {
		f()
	}
	if b.backoff != nil {
		b.backoff = nil
	}
}

func (b *backoffHandler) onSuccess() {
	// if backoff retries were previously disabled due to repeated PermissionDenied errors,
	// a successful call to the server should reset that state
	b.denied = false
}

func (b *backoffHandler) OnError(err error) {
	// we only care about grpc PermissionDenied errors
	if status.Code(err) != codes.PermissionDenied {
		return
	}

	// if we already have a backoff running, ignore
	if b.backoff != nil {
		return
	}

	// if we already retried and got PermissionDenied, no point in trying again
	if b.denied {
		return
	}

	// create backoff to manage retry attempts
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 10 * time.Second
	bo.Multiplier = 1.5
	bo.MaxInterval = 2 * time.Minute
	bo.MaxElapsedTime = 60 * time.Minute
	ctx, cancel := context.WithCancel(context.Background())
	b.backoff = backoff.WithContext(bo, ctx)
	b.cancel = cancel

	log.Println("Got PermissionDenied, creating backoff for token refresh...")

	go func() {
		// this blocks until the time limit expires or the request succeeds, at which point we'll lift the
		// backoff and allow further retries:
		backoff.Retry(b.retry, b.backoff)
		b.reset()
	}()
}

func (b *backoffHandler) retry() error {
	// make request
	_, err := b.handler.handleRequest(cloud.NewTokenRequestWithJwt(&cloud.JwtRequest{ForceRefresh: true}))
	if status.Code(err) == codes.PermissionDenied {
		b.denied = true
		log.Println("Token retry got PermissionDenied, stopping")
		return &backoff.PermanentError{Err: err}
	} else if err != nil {
		log.Println("Token retry failed with err, still trying:", err)
	} else {
		log.Println("Token force refresh succeeded")
	}
	return err
}

// NewBackoffHandler creates a backoff handler that will respond to PermissionDenied errors
// in other services by attempting to refresh the current JWT token
func NewBackoffHandler(handler RequestHandler) *backoffHandler {
	return &backoffHandler{handler: handler}
}
