package util

import (
	"context"

	"github.com/digital-dream-labs/vector-cloud/internal/config"

	"google.golang.org/grpc"
)

type MapCredentials map[string]string

func (r MapCredentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return r, nil
}

func (r MapCredentials) RequireTransportSecurity() bool {
	return true
}

func AppkeyMetadata() MapCredentials {
	ret := MapCredentials{
		"anki-app-key": config.Env.AppKey,
	}
	return ret
}

var platformOpts []grpc.DialOption

// CommonGRPC returns a set of commonly used GRPC dial options for Anki's cloud services,
// if any are defined for the current platform.
func CommonGRPC() []grpc.DialOption {
	return platformOpts
}
