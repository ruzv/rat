package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// Token defines data stored in JWT token.
type Token struct {
	Username string `mapstructure:"username"`
	Expires  int64  `mapstructure:"expires"`
	Role     Role   `mapstructure:"role"`
}

// FromMapClaims converts jwt.MapClaims to token claims.
func FromMapClaims(mc jwt.MapClaims) (*Token, error) {
	t := &Token{}

	err := mapstructure.Decode(mc, t)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode token claims")
	}

	return t, nil
}

// ToMapClaims converts token claims to jwt.MapClaims.
func (t Token) ToMapClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"username": t.Username,
		"expires":  t.Expires,
		"role":     t.Role,
	}
}

// Expired returns true if the token is expired.
func (t *Token) Expired() bool {
	return time.Unix(t.Expires, 0).Before(time.Now())
}
