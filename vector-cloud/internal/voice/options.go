package voice

import (
	"github.com/digital-dream-labs/vector-cloud/internal/token"
	"github.com/digital-dream-labs/vector-cloud/internal/util"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
)

// Handler aliases the IntentService field from our protobuf that determines
// which backend service (Google, MS, etc) should handle this request
type Handler = pb.IntentService

const (
	// HandlerDefault represents the default backend service chosen by chipper
	HandlerDefault Handler = pb.IntentService_DEFAULT
	// HandlerGoogle will have Google's DialogFlow service handle requests
	HandlerGoogle Handler = pb.IntentService_DIALOGFLOW
	// HandlerMicrosoft will have Microsoft's Bing/LUIS speech service handle requests
	HandlerMicrosoft Handler = pb.IntentService_BING_LUIS
	// HandlerAmazon will have Amazon's Lex service handle requests
	HandlerAmazon Handler = pb.IntentService_LEX
)

// Option defines an option that can be set on the cloud process
type Option func(o *options)

type options struct {
	compress     bool
	chunkMs      uint
	handler      Handler
	saveAudio    bool
	tokener      token.Accessor
	requireToken bool
	errListener  util.ErrorListener
}

// WithCompression sets whether compression will be performed on audio
// before uploading (and returns the same Options struct to allow method
// chaining)
func WithCompression(value bool) Option {
	return func(o *options) {
		o.compress = value
	}
}

// WithChunkMs determines how often the cloud process will stream data to the cloud
func WithChunkMs(value uint) Option {
	return func(o *options) {
		o.chunkMs = value
	}
}

// WithHandler sets the intent service (MS, Google, etc) that should handle this
// request on the server, if one is desired
func WithHandler(value Handler) Option {
	return func(o *options) {
		o.handler = value
	}
}

// WithSaveAudio sets whether the chipper server should save the audio we send to it
func WithSaveAudio(value bool) Option {
	return func(o *options) {
		o.saveAudio = value
	}
}

// WithRequireToken specifies that a failure to obtain an access token from the TMS
// should be treated as a fatal error for requests (temporary - eventually TMS tokens
// will be required)
func WithRequireToken() Option {
	return func(o *options) {
		o.requireToken = true
	}
}

// WithTokener sets the accessor that should be used to obtain tokens
func WithTokener(value token.Accessor) Option {
	return func(o *options) {
		o.tokener = value
	}
}

// WithErrorListener specifies that the given ErrorListener should be passed errors
// that result from voice requests
func WithErrorListener(value util.ErrorListener) Option {
	return func(o *options) {
		o.errListener = value
	}
}
