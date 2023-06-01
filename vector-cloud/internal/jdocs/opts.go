package jdocs

import (
	"github.com/digital-dream-labs/vector-cloud/internal/token"
	"github.com/digital-dream-labs/vector-cloud/internal/util"
)

type options struct {
	server           bool
	socketNameSuffix string
	tokener          token.Accessor
	errListener      util.ErrorListener
}

// Option defines an option that can be set on the token server
type Option func(o *options)

// WithServer specifies that an IPC server should be started so other processes
// can request jdocs from this process
func WithServer() Option {
	return func(o *options) {
		o.server = true
	}
}

// WithSocketNameSuffix specifies the (optional) suffix of the socket name
func WithSocketNameSuffix(socketNameSuffix string) Option {
	return func(o *options) {
		o.socketNameSuffix = socketNameSuffix
	}
}

// WithTokener specifies that the given token.Accessor should be used to obtain
// authorization credentials
func WithTokener(value token.Accessor) Option {
	return func(o *options) {
		o.tokener = value
	}
}

// WithErrorListener specifies that the given ErrorListener should be passed errors
// that result from jdoc requests
func WithErrorListener(value util.ErrorListener) Option {
	return func(o *options) {
		o.errListener = value
	}
}
