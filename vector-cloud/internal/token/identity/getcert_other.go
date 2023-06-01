// +build !vicos

package identity

import (
	"crypto/tls"

	"github.com/digital-dream-labs/vector-cloud/internal/robot"

	"github.com/gwatts/rootcerts"
	"google.golang.org/grpc/credentials"
)

// DefaultTokenPath specifies default directory for persistent JWT storage
const DefaultTokenPath = "/tmp/victoken"

// UseClientCert can be set to true to force the use of client certs
var UseClientCert = false

func getTLSCert(cloudDir string) (credentials.TransportCredentials, error) {
	if !UseClientCert {
		return credentials.NewClientTLSFromCert(rootcerts.ServerCertPool(), ""), nil
	}

	cert, err := robot.TLSKeyPair(cloudDir)
	if err != nil {
		return nil, err
	}
	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      rootcerts.ServerCertPool(),
	}), nil
}
