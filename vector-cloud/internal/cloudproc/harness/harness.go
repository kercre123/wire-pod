package harness

import (
	"context"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"

	"github.com/digital-dream-labs/vector-cloud/internal/cloudproc"
	"github.com/digital-dream-labs/vector-cloud/internal/voice"
)

type Harness interface {
	voice.MsgIO
	ReadMessage() (*cloud.Message, error)
}

type memHarness struct {
	voice.MsgIO
	intent chan *cloud.Message
}

func (h *memHarness) ReadMessage() (*cloud.Message, error) {
	return <-h.intent, nil
}

func CreateMemProcess(ctx context.Context, options ...cloudproc.Option) (Harness, error) {
	intentResult := make(chan *cloud.Message)

	io, receiver := voice.NewMemPipe()
	process := &voice.Process{}
	process.AddReceiver(receiver)
	process.AddIntentWriter(&voice.ChanMsgSender{Ch: intentResult})

	options = append(options, cloudproc.WithVoice(process))

	go cloudproc.Run(ctx, options...)

	return &memHarness{
		MsgIO:  io,
		intent: intentResult}, nil
}
