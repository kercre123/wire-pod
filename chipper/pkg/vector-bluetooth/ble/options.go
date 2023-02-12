package ble

type options struct {
	outputDir  string
	statuschan chan StatusChannel
}

// Option is the list of options
type Option func(*options)

// WithLogDirectory sets the logger output directory
func WithLogDirectory(l string) Option {
	return func(o *options) {
		o.outputDir = l
	}
}

// WithStatusChan sets a status response channel
func WithStatusChan(c chan StatusChannel) Option {
	return func(o *options) {
		o.statuschan = c
	}
}
