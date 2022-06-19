package jdocs

import (
	"context"
)

// Run starts the jdocs service
func Run(ctx context.Context, optionValues ...Option) {
	var opts options
	for _, o := range optionValues {
		o(&opts)
	}

	if opts.server {
		runServer(ctx, &opts)
	}
}
