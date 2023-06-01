package stream

import (
	"context"
	"sync"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"
)

type Streamer struct {
	conn        Conn
	byteChan    chan []byte
	audioStream chan []byte
	respOnce    sync.Once
	closed      bool
	opts        options
	receiver    Receiver
	ctx         context.Context
	cancel      func()
}

type Receiver interface {
	OnError(cloud.ErrorType, error)
	OnStreamOpen(string)
	OnIntent(*cloud.IntentResult)
	OnConnectionResult(*cloud.ConnectionResult)
}

type CloudError struct {
	Kind cloud.ErrorType
	Err  error
}
