package tokenserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	mathrand "math/rand"
	"os"
	"time"

	"github.com/digital-dream-labs/api/go/tokenpb"
	"github.com/golang-jwt/jwt"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789="

var seededRand *mathrand.Rand = mathrand.New(
	mathrand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandomString(length int) string {
	return StringWithCharset(length, charset)
}

type TokenServer struct {
	tokenpb.UnimplementedTokenServer
}

func (s *TokenServer) AssociatePrimaryUser(ctx context.Context, req *tokenpb.AssociatePrimaryUserRequest) (*tokenpb.AssociatePrimaryUserResponse, error) {
	fmt.Println("Token request incoming")
	fmt.Println("string: " + req.String())
	os.WriteFile("./sessionCert.pem", req.SessionCertificate, 0644)
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"expires":      "2029-11-26T16:27:51.997352463Z",
		"iat":          time.Now(),
		"permissions":  nil,
		"requestor_id": "vic:00601b50",
		"token_id":     "11ec68ca-1d4c-4e45-b1a2-715fd5e0abf9",
		"token_type":   "user+robot",
		"user_id":      "2gsE4HbQ8UCBpYqurDgsafX",
	})

	// Sign and get the complete encoded token as a string using the secret
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	fmt.Println(rsaKey)
	tokenString, _ := token.SignedString(rsaKey)
	fmt.Println("")
	fmt.Println(tokenString)
	clientToken := RandomString(24)
	fmt.Println("")
	fmt.Println("GUID: " + clientToken)
	os.WriteFile("./robotGUID", []byte(clientToken), 0644)
	fmt.Println("")
	return &tokenpb.AssociatePrimaryUserResponse{
		Data: &tokenpb.TokenBundle{
			Token:       tokenString,
			ClientToken: clientToken,
		},
	}, nil
}

func NewTokenServer() *TokenServer {
	return &TokenServer{}
}
