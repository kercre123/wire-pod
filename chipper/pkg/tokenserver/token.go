package tokenserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/digital-dream-labs/api/go/tokenpb"
	"github.com/golang-jwt/jwt"
)

type TokenServer struct {
	tokenpb.UnimplementedTokenServer
}

type RobotSDKInfoStore struct {
	GlobalGUID string `json:"global_guid"`
	Robots     []struct {
		Esn       string `json:"esn"`
		IPAddress string `json:"ip_address"`
	} `json:"robots"`
}

func (s *TokenServer) AssociatePrimaryUser(ctx context.Context, req *tokenpb.AssociatePrimaryUserRequest) (*tokenpb.AssociatePrimaryUserResponse, error) {
	fmt.Println("Token: Incoming Associate Primary User request")
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"expires":      "2029-11-26T16:27:51.997352463Z",
		"iat":          time.Now(),
		"permissions":  nil,
		"requestor_id": "vic:00601b50",
		"token_id":     "11ec68ca-1d4c-4e45-b1a2-715fd5e0abf9",
		"token_type":   "user+robot",
		"user_id":      "2gsE4HbQ8UCBpYqurDgsafX",
	})
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	tokenString, _ := token.SignedString(rsaKey)
	// constant GUID
	clientToken := "tni1TRsTRTaNSapjo0Y+Sw=="
	return &tokenpb.AssociatePrimaryUserResponse{
		Data: &tokenpb.TokenBundle{
			Token:       tokenString,
			ClientToken: clientToken,
		},
	}, nil
}

func (s *TokenServer) AssociateSecondaryClient(ctx context.Context, req *tokenpb.AssociateSecondaryClientRequest) (*tokenpb.AssociateSecondaryClientResponse, error) {
	fmt.Println("Token: Incoming Associate Secondary Client request")
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"expires":      "2029-11-26T16:27:51.997352463Z",
		"iat":          time.Now(),
		"permissions":  nil,
		"requestor_id": "vic:00601b50",
		"token_id":     "11ec68ca-1d4c-4e45-b1a2-715fd5e0abf9",
		"token_type":   "user+robot",
		"user_id":      "2gsE4HbQ8UCBpYqurDgsafX",
	})
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	tokenString, _ := token.SignedString(rsaKey)
	// constant GUID
	clientToken := "tni1TRsTRTaNSapjo0Y+Sw=="
	return &tokenpb.AssociateSecondaryClientResponse{
		Data: &tokenpb.TokenBundle{
			Token:       tokenString,
			ClientToken: clientToken,
		},
	}, nil
}

func (s *TokenServer) RefreshToken(ctx context.Context, req *tokenpb.RefreshTokenRequest) (*tokenpb.RefreshTokenResponse, error) {
	fmt.Println("Token: Incoming Refresh Token Request")
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"expires":      "2029-11-26T16:27:51.997352463Z",
		"iat":          time.Now(),
		"permissions":  nil,
		"requestor_id": "vic:00601b50",
		"token_id":     "11ec68ca-1d4c-4e45-b1a2-715fd5e0abf9",
		"token_type":   "user+robot",
		"user_id":      "2gsE4HbQ8UCBpYqurDgsafX",
	})
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	tokenString, _ := token.SignedString(rsaKey)
	// constant GUID
	clientToken := "tni1TRsTRTaNSapjo0Y+Sw=="
	return &tokenpb.RefreshTokenResponse{
		Data: &tokenpb.TokenBundle{
			Token:       tokenString,
			ClientToken: clientToken,
		},
	}, nil
}

func NewTokenServer() *TokenServer {
	return &TokenServer{}
}
