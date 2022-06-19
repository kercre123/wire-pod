package token

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

const (
	tokenSize = 16
	saltSize  = 16
	hashSize  = sha256.Size

	errMismatchedTokenAndHash = "Hash mismatch"
	errHashTooLong            = "Hash too long"
	errHashTooShort           = "Hash too short"
	errTokenTooLong           = "Token too long"
	errTokenTooShort          = "Token too short"
)

type hashed struct {
	// hash the hash of the token on its own (without the appended
	// salt)
	hash []byte
	// salt is the salt of the token
	salt []byte
}

// CompareHashAndPassword compares a hashed client token with its
// possible plain text equivalent. Returns nil on success, or an error
// on failure. Modeled after the crypto/bcrypt interface.
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
