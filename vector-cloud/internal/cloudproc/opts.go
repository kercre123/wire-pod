package cloudproc

import (
	"github.com/digital-dream-labs/vector-cloud/internal/jdocs"
	"github.com/digital-dream-labs/vector-cloud/internal/logcollector"
	"github.com/digital-dream-labs/vector-cloud/internal/token"
	"github.com/digital-dream-labs/vector-cloud/internal/token/identity"
	"github.com/digital-dream-labs/vector-cloud/internal/voice"
)

type Option func(o *options)

type options struct {
	voice            *voice.Process
	identityProvider identity.Provider
	voiceOpts        []voice.Option
	tokenOpts        []token.Option
	jdocOpts         []jdocs.Option
	logcollectorOpts []logcollector.Option
}

func WithVoice(process *voice.Process) Option {
	return func(o *options) {
		o.voice = process
	}
}

func WithIdentityProvider(identityProvider identity.Provider) Option {
	return func(o *options) {
		o.identityProvider = identityProvider
	}
}

func WithVoiceOptions(voiceOptions ...voice.Option) Option {
	return func(o *options) {
		o.voiceOpts = append(o.voiceOpts, voiceOptions...)
	}
}

func WithTokenOptions(tokenOptions ...token.Option) Option {
	return func(o *options) {
		o.tokenOpts = append(o.tokenOpts, tokenOptions...)
	}
}

func WithJdocs(jdocOptions ...jdocs.Option) Option {
	return func(o *options) {
		o.jdocOpts = append(o.jdocOpts, jdocOptions...)
		if o.jdocOpts == nil {
			// even if no options specified, code is saying "run jdocs plz" by calling this
			o.jdocOpts = []jdocs.Option{}
		}
	}
}

func WithLogCollectorOptions(logcollectorOptions ...logcollector.Option) Option {
	return func(o *options) {
		o.logcollectorOpts = append(o.logcollectorOpts, logcollectorOptions...)
		if o.logcollectorOpts == nil {
			// even if no options specified, code is saying "run log collector plz" by calling this
			o.logcollectorOpts = []logcollector.Option{}
		}
	}
}
