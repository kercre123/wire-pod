package initwirepod

import (
	"context"
	"crypto/tls"
	"errors"
	"net"

	"github.com/soheilhy/cmux"
	"google.golang.org/grpc/credentials"
)

type cmuxTLSCreds struct{}

func (cmuxTLSCreds) ClientHandshake(ctx context.Context, authority string, conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return nil, nil, errors.New("cmuxTLSCreds is server-only")
}

func (cmuxTLSCreds) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	inner := conn
	if mc, ok := inner.(*cmux.MuxConn); ok {
		inner = mc.Conn
	}
	tc, ok := inner.(*tls.Conn)
	if !ok {
		return conn, nil, nil
	}
	return conn, credentials.TLSInfo{
		State:          tc.ConnectionState(),
		CommonAuthInfo: credentials.CommonAuthInfo{SecurityLevel: credentials.PrivacyAndIntegrity},
	}, nil
}

func (cmuxTLSCreds) Info() credentials.ProtocolInfo {
	return credentials.ProtocolInfo{SecurityProtocol: "tls"}
}

func (c cmuxTLSCreds) Clone() credentials.TransportCredentials {
	return c
}

func (cmuxTLSCreds) OverrideServerName(string) error {
	return nil
}
