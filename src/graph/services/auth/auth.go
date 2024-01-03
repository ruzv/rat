package auth

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"rat/logr"
)

type ctxKey string

const AuthTokenCtxKey ctxKey = "auth-token"

// Config defines configuration params for authentication.
type Config struct {
	Users []*User           `yaml:"users"`
	Roles map[Role][]*Scope `yaml:"roles"`
	Token *TokenConfig      `yaml:"token" validate:"nonzero"`
}

// TokenConfig defines configuration params for JWT token generation.
type TokenConfig struct {
	Secret     string        `yaml:"secret" validate:"nonzero"`
	Expiration time.Duration `yaml:"expiration" validate:"nonzero"`
}

// User defines a user with credentials and role.
type User struct {
	Credentials `yaml:",inline"`
	Scopes      *Scopes `yaml:"scopes" validate:"nonzero"`
}

// Credentials defines users username and password.
type Credentials struct {
	Username string `yaml:"username" validate:"nonzero"`
	Password string `yaml:"password" validate:"nonzero"`
}

type TokenControl struct {
	users       map[string]*User
	roles       map[Role][]*Scope
	tokenConfig *TokenConfig
	log         *logr.LogR
}

func NewTokenControl(config *Config, log *logr.LogR) (*TokenControl, error) {
	log = log.Prefix("token-control")

	lg := log.Group(logr.LogLevelInfo)
	defer lg.Close()

	lg.Log("users:")

	users := make(map[string]*User)
	for _, user := range config.Users {
		users[user.Username] = user

		b, err := json.MarshalIndent(
			struct {
				Username string   `json:"username"`
				Scopes   []*Scope `json:"scopes"`
			}{
				Username: user.Username,
				Scopes:   user.Scopes.Get(config.Roles),
			},
			"",
			"  ",
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal user for log")
		}

		lg.Log("%s", string(b))

	}

	return &TokenControl{
		users:       users,
		roles:       config.Roles,
		tokenConfig: config.Token,
		log:         log,
	}, nil
}

func (tc *TokenControl) Generate(
	username, password string,
) (string, *Token, error) {
	user, ok := tc.users[username]
	if !ok {
		return "", nil, errors.Errorf("user %q not found", username)
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to authenticate user")
	}

	token := &Token{
		Username: user.Username,
		Expires:  time.Now().Add(tc.tokenConfig.Expiration).Unix(),
		Scopes:   user.Scopes.Get(tc.roles),
	}

	signed, err := jwt.NewWithClaims(
		jwt.SigningMethodHS512,
		token.ToMapClaims(),
	).
		SignedString([]byte(tc.tokenConfig.Secret))
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to sign token")
	}

	return signed, token, nil
}

func (tc *TokenControl) Validate(signed string) (*Token, error) {
	token, err := jwt.Parse(
		signed,
		func(token *jwt.Token) (any, error) {
			return []byte(tc.tokenConfig.Secret), nil
		},
		jwt.WithValidMethods([]string{"HS512"}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse token")
	}

	if !token.Valid {
		return nil, errors.New("token not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to parse claims")
	}

	t, err := FromMapClaims(claims)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse token from claims")
	}

	if t.Expired() {
		return nil, errors.New("token expired")
	}

	_, ok = tc.users[t.Username]
	if !ok {
		return nil, errors.New("user not found")
	}

	return t, nil
}
