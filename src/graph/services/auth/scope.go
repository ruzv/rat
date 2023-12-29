package auth

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var _ yaml.Unmarshaler = (*Role)(nil)

const (
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

const (
	GraphOwnerNodeRead  Scope = "graph_owner_node_read"
	GraphOwnerNodeWrite Scope = "graph_owner_node_write"

	GraphMemberNodeRead  Scope = "graph_member_node_read"
	GraphMemberNodeWrite Scope = "graph_member_node_write"

	GraphVisitorNodeRead  Scope = "graph_visitor_node_read"
	GraphVisitorNodeWrite Scope = "graph_visitor_node_write"
)

var validScopes = map[Scope]bool{
	GraphOwnerNodeRead:    true,
	GraphOwnerNodeWrite:   true,
	GraphMemberNodeRead:   true,
	GraphMemberNodeWrite:  true,
	GraphVisitorNodeRead:  true,
	GraphVisitorNodeWrite: true,
}

// Role defines a user role - an alias for a list of scopes.
type Role string

// Scope defines a single access scope.
type Scope string

// Scopes defines a list of access scopes. Can be unmarshaled from string
// - a predefined role alias, or a list of strings - list of scopes.
type Scopes struct {
	scopes []Scope
}

func (s Scopes) Slice() []Scope {
	return s.scopes
}

func (r Role) Scopes() []Scope {
	switch r {
	case Owner:
		return []Scope{
			GraphOwnerNodeRead,
			GraphOwnerNodeWrite,
			GraphMemberNodeRead,
			GraphMemberNodeWrite,
			GraphVisitorNodeRead,
			GraphVisitorNodeWrite,
		}
	case Member:
		return []Scope{
			GraphMemberNodeRead,
			GraphMemberNodeWrite,
			GraphVisitorNodeRead,
			GraphVisitorNodeWrite,
		}
	case Viewer:
		return []Scope{
			GraphMemberNodeRead,
			GraphVisitorNodeRead,
		}
	case Visitor:
		return []Scope{
			GraphVisitorNodeRead,
		}

	default:
		return []Scope{}
	}
}

// UnmarshalYAML implements yaml.Unmarshaler for Scopes to unmarshal either
// string or list of strings.
func (s *Scopes) UnmarshalYAML(unmarshal func(any) error) error {
	err := unmarshal(&([]string{}))
	if err != nil {

		var role Role

		err = unmarshal(&role)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal scopes as role")
		}

		s.scopes = role.Scopes()

		return nil
	}

	scopes := []Scope{}

	err = unmarshal(&scopes)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal scopes as list")
	}

	s.scopes = scopes

	return nil
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

// UnmarshalYAML implements yaml.Unmarshaler for Scope type to check for valid
// values.
func (s *Scope) UnmarshalYAML(unmarshal func(any) error) error {
	var raw string

	err := unmarshal(&raw)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal scope")
	}

	scope := Scope(raw)

	if !validScopes[scope] {
		return errors.Errorf("invalid scope: %s", raw)
	}

	*s = scope

	return nil
}
