package auth

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

// Token defines data stored in JWT token.
type Token struct {
	Username string   `json:"username"`
	Expires  int64    `json:"expires"`
	Scopes   []*Scope `json:"scopes"`
}

// FromMapClaims converts jwt.MapClaims to token claims.
func FromMapClaims(mc jwt.MapClaims) (*Token, error) {
	t := &Token{}

	b, err := json.Marshal(mc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal token claims")
	}

	err = json.Unmarshal(b, t)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal token claims")
	}

	return t, nil
}

// ToMapClaims converts token claims to jwt.MapClaims.
func (t Token) ToMapClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"username": t.Username,
		"expires":  t.Expires,
		"scopes":   t.Scopes,
	}
}

// Expired returns true if the token is expired.
func (t *Token) Expired() bool {
	return time.Unix(t.Expires, 0).Before(time.Now())
}
