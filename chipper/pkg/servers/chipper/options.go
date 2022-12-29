package server

import "github.com/digital-dream-labs/hugh/log"

type options struct {
	log         log.Logger
	intent      intentProcessor
	kg          kgProcessor
	intentGraph intentGraphProcessor
}

// Option is the list of options
type Option func(*options)

// WithLogger sets the logger
func WithLogger(l log.Logger) Option {
	return func(o *options) {
		o.log = l
	}
}

// WithIntentProcessor sets the intent processor
func WithIntentProcessor(s intentProcessor) Option {
	return func(o *options) {
		o.intent = s
	}
}

// WithKnowledgeGraphProcessor sets the knowledge graph processor
func WithKnowledgeGraphProcessor(s kgProcessor) Option {
	return func(o *options) {
		o.kg = s
	}
}

// WithKnowledgeGraphProcessor sets the knowledge graph processor
func WithIntentGraphProcessor(s intentGraphProcessor) Option {
	return func(o *options) {
		o.intentGraph = s
	}
}
