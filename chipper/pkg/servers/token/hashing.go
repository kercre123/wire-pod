package tokenserver

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// mostly copied from vector-cloud

const (
	tokenSize = 16
	saltSize  = 16
	hashSize  = sha256.Size

	errMismatchedTokenAndHash = "hash mismatch"
	errHashTooLong            = "hash too long"
	errHashTooShort           = "hash too short"
	errTokenTooLong           = "token too long"
	errTokenTooShort          = "token too short"
)

type hashed struct {
	// hash the hash of the token on its own (without the appended
	// salt)
	hash []byte
	// salt is the salt of the token
	salt []byte
}

type ClientToken struct {
	Hash       string `json:"hash"`
	ClientName string `json:"client_name"`
	AppId      string `json:"app_id"`
	IssuedAt   string `json:"issued_at"`
}

type ClientTokenManager struct {
	ClientTokens []ClientToken `json:"client_tokens"`
}

func CreateTokenAndHashedToken() (GUID string, GUIDHash string, isError error) {
	// generate random bytes
	tokenBytes := make([]byte, tokenSize)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", "", err
	}
	token := base64.StdEncoding.EncodeToString(tokenBytes)

	// generate a random salt
	saltBytes := make([]byte, saltSize)
	_, err = rand.Read(saltBytes)
	if err != nil {
		return "", "", err
	}

	// hash token
	hashed := hash(tokenBytes, saltBytes)
	hashed = append(hashed, saltBytes...)

	// encode
	hashedToken := base64.StdEncoding.EncodeToString(hashed)

	return token, hashedToken, nil
}

func DecodeAndCompare(tokenHashes string, token string) {
	// debug
	var ctm ClientTokenManager
	json.Unmarshal([]byte(tokenHashes), &ctm)
	for _, tokenHash := range ctm.ClientTokens {
		err := CompareHashAndToken(tokenHash.Hash, token)
		if err == nil {
			logger.Println(tokenHash.Hash + " matched " + token)
		} else {
			logger.Println(err)
		}
	}

}

// copied from vector-cloud
func CompareHashAndToken(hashedToken, token string) error {
	// Decode the hash and the token to their raw bytes
	hashedBytes, err := base64.StdEncoding.DecodeString(hashedToken)
	if err != nil {
		return err
	}
	tokenBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return err
	}

	// extract the hash and salt from the pre-hashed value
	hashed, err := newFromHash(hashedBytes)
	if err != nil {
		return err
	}
	// hash the token using the salt extracted from the hashed input
	// and compare
	newHash := hash(tokenBytes, hashed.salt)

	if subtle.ConstantTimeCompare(hashed.hash, newHash) == 1 {
		return nil
	}

	return fmt.Errorf(errMismatchedTokenAndHash)
}

func newFromHash(hashedToken []byte) (*hashed, error) {
	if len(hashedToken) > (hashSize + saltSize) {
		return nil, fmt.Errorf(errHashTooLong)
	} else if len(hashedToken) < (hashSize + saltSize) {
		return nil, fmt.Errorf(errHashTooShort)
	}

	hash, salt := hashedToken[:hashSize], hashedToken[hashSize:]
	return &hashed{
		hash: hash,
		salt: salt,
	}, nil
}

func hash(token, salt []byte) []byte {
	salted := make([]byte, 0, (tokenSize + saltSize))
	salted = append(salted, token...)
	salted = append(salted, salt...)
	hash := sha256.Sum256(salted)
	return hash[:]
}
