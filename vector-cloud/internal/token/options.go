package token

type options struct {
	server           bool
	socketNameSuffix string
}

// Option defines an option that can be set on the token server
type Option func(o *options)

// WithServer specifies that an IPC server should be started so other processes
// can request tokens from this process
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
