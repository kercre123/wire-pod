// +build vicos

package identity

import (
	"crypto/tls"

	"github.com/digital-dream-labs/vector-cloud/internal/robot"

	"github.com/gwatts/rootcerts"
	"google.golang.org/grpc/credentials"
)

const DefaultTokenPath = "/data/data/com.anki.victor/persistent/token"

func getTLSCert(cloudDir string) (credentials.TransportCredentials, error) {
	cert, err := robot.TLSKeyPair(cloudDir)
	if err != nil {
		return nil, err
	}
	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      rootcerts.ServerCertPool(),
	}), nil
}
