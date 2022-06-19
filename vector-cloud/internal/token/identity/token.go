package identity

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	JwtClaimTokenID     = "token_id"
	JwtClaimTokenType   = "token_type"
	JwtClaimRequestorID = "requestor_id"
	JwtClaimUserID      = "user_id"
	JwtClaimIssuedAt    = "iat"
	JwtClaimExpiresAt   = "expires"
	JwtClaimPermissions = "permissions"
)

var (
	errorNilToken = errors.New("missing token")
	errorNoClaims = errors.New("missing claims")
	locUTC, _     = time.LoadLocation("UTC")
)

// Token is a structured representation of an access token.
type TokenInfo struct {
	// Id is the unique ID of the token.
	Id string

	// Type - Only 'user+robot' is supported right now.
	Type string

	// RequestorId is an identifier for the entity which requested the
	// token. Likely to be the common name of the robot cert
	// (i.e. 'vic:<ESN>', or later an Anki Principal URN)
	RequestorId string

	// UserId is the accounts system ID of the user associated with
	// the requesting entity.
	UserId string

	// IssuedAt is the UTC time when the token was issued.
	IssuedAt time.Time

	// ExpiresAt is the UTC time at which the token is no longer
	// valid. Generally equal to IssuedAt + 24 hours.
	ExpiresAt time.Time

	// PurgeAt is the UTC time at which Dynamo will automatically
	// delete the token. Only used within the Token Service.
	PurgeAt time.Time

	// RevokedAt is the UTC time at which the token was revoked, if it
	// has been revoked. Tokens can be revoked due to account system
	// password changes or account deletion. Only used within the
	// Token Service.
	RevokedAt time.Time

	// Revoked is true if this token has been revoked due to account
	// system changes. Only used within the Token Service.
	Revoked bool

	// Raw is the raw string form of the JWT token, if this Token
	// object was parsed from a JWT token. Only used within the Token
	// Service.
	Raw string

	Permissions map[string]interface{} `json:"permissions,omitempty"`
}

// IsExpired is a simple predicate indicating whether the token's
// expiration time has passed or not.
func (t TokenInfo) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

// JwtToken converts this Token object to a jwt.Token object suitable
// for hashing and conversion to a signed string.
func (t TokenInfo) JwtToken(method jwt.SigningMethod) *jwt.Token {
	return jwt.NewWithClaims(method, jwt.MapClaims{
		JwtClaimTokenID:     t.Id,
		JwtClaimTokenType:   t.Type,
		JwtClaimUserID:      t.UserId,
		JwtClaimRequestorID: t.RequestorId,
		JwtClaimIssuedAt:    t.IssuedAt.Format(time.RFC3339Nano),
		JwtClaimExpiresAt:   t.ExpiresAt.Format(time.RFC3339Nano),
		JwtClaimPermissions: t.Permissions,
	})
}

// FromJwtToken converts a generic jwt.Token object, parsed from a
// signed token string, into a Token structure, validating that all
// the required Anki token claims are present.
func FromJwtToken(t *jwt.Token) (*TokenInfo, error) {
	if t == nil {
		return nil, errorNilToken
	}
	if claims, ok := t.Claims.(jwt.MapClaims); ok {
		id, ok := claims[JwtClaimTokenID].(string)
		if !ok {
			return nil, errorMissingClaim(JwtClaimTokenID)
		}

		tp, ok := claims[JwtClaimTokenType].(string)
		if !ok {
			return nil, errorMissingClaim(JwtClaimTokenType)
		}

		userID, ok := claims[JwtClaimUserID].(string)
		if !ok {
			return nil, errorMissingClaim(JwtClaimUserID)
		}

		requestorID, ok := claims[JwtClaimRequestorID].(string)
		if !ok {
			return nil, errorMissingClaim(JwtClaimRequestorID)
		}

		claimsIAT, ok := claims[JwtClaimIssuedAt].(string)
		if !ok {
			return nil, errorMissingClaim(JwtClaimIssuedAt)
		}

		iat, err := time.ParseInLocation(time.RFC3339, claimsIAT, locUTC)
		if err != nil {
			return nil, err
		}

		expiresAt, ok := claims[JwtClaimExpiresAt].(string)
		if !ok {
			return nil, errorMissingClaim(JwtClaimExpiresAt)
		}
		expires, err := time.ParseInLocation(time.RFC3339, expiresAt, locUTC)
		if err != nil {
			return nil, err
		}

		tok := &TokenInfo{
			Raw:         t.Raw,
			Id:          id,
			Type:        tp,
			UserId:      userID,
			RequestorId: requestorID,
			IssuedAt:    iat,
			ExpiresAt:   expires,
		}

		// IDEA:  Its fine if this isn't populated, but i think the whole token parsing strategy
		// needs to be rewritten..  Having to parse the same values in differing contexts (jtw vs actual token)
		// is super ugly..
		permissions, ok := claims[JwtClaimPermissions].(map[string]interface{})
		if ok {
			tok.Permissions = permissions
		}

		return tok, nil
	}
	return nil, errorNoClaims
}

//
// helpers
//

func errorMissingClaim(claim string) error {
	return fmt.Errorf("missing claim %s", claim)
}
