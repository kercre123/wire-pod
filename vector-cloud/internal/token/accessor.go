package token

import (
	"fmt"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"

	"github.com/digital-dream-labs/vector-cloud/internal/util"

	"github.com/digital-dream-labs/vector-cloud/internal/token/identity"

	ac "github.com/aws/aws-sdk-go/aws/credentials"
	gc "google.golang.org/grpc/credentials"
)

type Accessor interface {
	Credentials() (gc.PerRPCCredentials, error)
	GetStsCredentials() (*ac.Credentials, error)
	IdentityProvider() identity.Provider
	UserID() string
}

type accessor struct {
	identityProvider identity.Provider
	stsCache         stsCredentialsCache
	handler          RequestHandler
}

func (a accessor) Credentials() (gc.PerRPCCredentials, error) {
	req := cloud.NewTokenRequestWithJwt(&cloud.JwtRequest{})
	resp, err := a.handler.handleRequest(req)
	if err != nil {
		return nil, err
	}
	jwt := resp.GetJwt()
	if jwt.Error != cloud.TokenError_NoError {
		return nil, fmt.Errorf("jwt error code %d", jwt.Error)
	}
	return tokenMetadata(resp.GetJwt().JwtToken), nil
}

func (a accessor) GetStsCredentials() (*ac.Credentials, error) {
	return a.stsCache.getStsCredentials(a)
}

func (a accessor) UserID() string {
	token := a.identityProvider.GetToken()
	if token == nil {
		return ""
	}
	return token.UserID()
}

func (a accessor) IdentityProvider() identity.Provider {
	return a.identityProvider
}

func GetAccessor(identityProvider identity.Provider, handler RequestHandler) Accessor {
	return &accessor{identityProvider: identityProvider, handler: handler}
}

func tokenMetadata(jwtToken string) util.MapCredentials {
	ret := util.AppkeyMetadata()
	ret["anki-access-token"] = jwtToken
	return ret
}
