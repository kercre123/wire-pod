package token

import (
	"context"
	"io/ioutil"

	"github.com/digital-dream-labs/vector-cloud/internal/robot"
	"github.com/digital-dream-labs/vector-cloud/internal/token/identity"
	"github.com/digital-dream-labs/vector-cloud/internal/util"

	pb "github.com/digital-dream-labs/api/go/tokenpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type conn struct {
	conn   *grpc.ClientConn
	client pb.TokenClient
}

func newConn(identityProvider identity.Provider, serverURL string, creds credentials.PerRPCCredentials) (*conn, error) {
	dialOpts := append(getDialOptions(identityProvider, creds), util.CommonGRPC()...)
	rpcConn, err := grpc.Dial(serverURL, dialOpts...)
	if err != nil {
		return nil, err
	}

	rpcClient := pb.NewTokenClient(rpcConn)

	ret := &conn{
		conn:   rpcConn,
		client: rpcClient}
	return ret, nil
}

func (c *conn) associatePrimary(session string) (*pb.TokenBundle, error) {
	req := pb.AssociatePrimaryUserRequest{}
	cert, err := ioutil.ReadFile(robot.GatewayCert)
	if err != nil {
		return nil, err
	}
	req.SessionCertificate = cert
	response, err := c.client.AssociatePrimaryUser(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (c *conn) associateSecondary(session, clientName, appID string) (*pb.TokenBundle, error) {
	req := pb.AssociateSecondaryClientRequest{
		UserSession: session,
		ClientName:  clientName,
		AppId:       appID}
	response, err := c.client.AssociateSecondaryClient(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (c *conn) reassociatePrimary(clientName, appID string) (*pb.TokenBundle, error) {
	req := pb.ReassociatePrimaryUserRequest{
		ClientName: clientName,
		AppId:      appID}
	response, err := c.client.ReassociatePrimaryUser(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (c *conn) refreshToken(req pb.RefreshTokenRequest) (*pb.TokenBundle, error) {
	response, err := c.client.RefreshToken(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (c *conn) refreshJwtToken() (*pb.TokenBundle, error) {
	return c.refreshToken(pb.RefreshTokenRequest{RefreshJwtTokens: true})
}

func (c *conn) refreshStsCredentials() (*pb.TokenBundle, error) {
	return c.refreshToken(pb.RefreshTokenRequest{RefreshStsTokens: true})
}

func (c *conn) Close() error {
	return c.conn.Close()
}

func getDialOptions(identityProvider identity.Provider, creds credentials.PerRPCCredentials) []grpc.DialOption {
	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(identityProvider.TransportCredentials()))
	if creds != nil {
		dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(creds))
	}
	return dialOpts
}
