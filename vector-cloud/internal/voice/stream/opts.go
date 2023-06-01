package stream

import (
	"context"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"

	"github.com/digital-dream-labs/vector-cloud/internal/token"

	"github.com/digital-dream-labs/api-clients/chipper"
)

var platformOpts []chipper.ConnOpt

type ConnectFunc func(context.Context) (Conn, *CloudError)

type options struct {
	tokener         token.Accessor
	requireToken    bool
	mode            cloud.StreamType
	connOpts        []chipper.ConnOpt
	intentOpts      *chipper.IntentOpts
	kgOpts          *chipper.KGOpts
	intentGraphOpts *chipper.IntentGraphOpts
	checkOpts       *chipper.ConnectOpts
	streamOpts      *chipper.StreamOpts
	url             string
	connectFn       ConnectFunc
}

type Option func(o *options)

func WithTokener(t token.Accessor, require bool) Option {
	return func(o *options) {
		o.tokener = t
		o.requireToken = require
	}
}

func WithIntentOptions(opts chipper.IntentOpts, mode cloud.StreamType) Option {
	return func(o *options) {
		o.mode = mode
		o.intentOpts = &opts
		o.streamOpts = &opts.StreamOpts
	}
}

func WithKnowledgeGraphOptions(opts chipper.KGOpts) Option {
	return func(o *options) {
		o.mode = cloud.StreamType_KnowledgeGraph
		o.kgOpts = &opts
		o.streamOpts = &opts.StreamOpts
	}
}

func WithIntentGraphOptions(opts chipper.IntentGraphOpts, mode cloud.StreamType) Option {
	return func(o *options) {
		o.mode = mode
		o.intentGraphOpts = &opts
		o.streamOpts = &opts.StreamOpts
	}
}

func WithConnectionCheckOptions(opts chipper.ConnectOpts) Option {
	return func(o *options) {
		o.checkOpts = &opts
		o.streamOpts = &opts.StreamOpts
	}
}

func WithChipperURL(url string) Option {
	return func(o *options) {
		o.url = url
	}
}

// WithConnectFunc allows tests to provide a separate connection interface for the streamer, to
// mock connections instead of using real ones
func WithConnectFunc(connectFn ConnectFunc) Option {
	return func(o *options) {
		o.connectFn = connectFn
	}
}
