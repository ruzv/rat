package auth

import (
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

//nolint:gochecknoglobals
var (
	// Owner role gives a user unlimited access to the graph.
	Owner Role = "owner"
	// Member role gives a user full access to graph nodes marked with
	// access: member and its sub-nodes.
	Member Role = "member"
	// Viewer role gives a user read access to graph nodes marked with
	// access: viewer and its sub-nodes.
	Viewer Role = "viewer"
	// Visitor role gives read access unauthenticated users to graph nodes
	// marked with access: visitor and its sub-nodes.
	Visitor Role = "visitor"
)

var _ yaml.Unmarshaler = (*Role)(nil)

// Role defines a user role.
type Role string

// Credentials defines users username and password.
type Credentials struct {
	Username string `yaml:"username" validate:"nonzero"`
	Password string `yaml:"password" validate:"nonzero"`
}

// User defines a user with credentials and role.
type User struct {
	Credentials `yaml:",inline"`
	Role        Role `yaml:"role" validate:"nonzero"`
}

// TokenConfig defines configuration params for JWT token generation.
type TokenConfig struct {
	Secret     string        `yaml:"secret" validate:"nonzero"`
	Expiration time.Duration `yaml:"expiration" validate:"nonzero"`
}

// Config defines configuration params for authentication.
type Config struct {
	Owner *Credentials `yaml:"owner" validate:"nonzero"`
	Users []*User      `yaml:"users"`
	Token *TokenConfig `yaml:"token" validate:"nonzero"`
}

// AllUsers returns all users, with roles and credentials, including the owner
// user.
func (c *Config) AllUsers() []*User {
	return append(
		[]*User{
			{
				Credentials: *c.Owner,
				Role:        Owner,
			},
		},
		c.Users...,
	)
}

// UnmarshalYAML implements yaml.Unmarshaler for Role type to check for valid
// values.
func (r *Role) UnmarshalYAML(unmarshal func(any) error) error {
	var raw string

	err := unmarshal(&raw)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal role")
	}

	role := Role(raw)

	switch role {
	case Owner, Member, Viewer:
		*r = role

		return nil
	default:
		return errors.Errorf("invalid role: %s", raw)
	}
}
