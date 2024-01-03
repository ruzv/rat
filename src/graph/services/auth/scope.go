package auth

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"rat/graph/util"
)

// var _ yaml.Unmarshaler = (*Role)(nil)
var (
	_ yaml.Unmarshaler = (*Scopes)(nil)
	_ yaml.Unmarshaler = (*Scope)(nil)
	_ json.Marshaler   = (*Scope)(nil)
)

// const (
// 	// Owner role gives a user unlimited access to the graph.
// 	Owner Role = "owner"
// 	// Member role gives a user full access to graph nodes marked with
// 	// access: member and its sub-nodes.
// 	Member Role = "member"
// 	// Viewer role gives a user read access to graph nodes marked with
// 	// access: viewer and its sub-nodes.
// 	Viewer Role = "viewer"
// 	// Visitor role gives read access unauthenticated users to graph nodes
// 	// marked with access: visitor and its sub-nodes.
// 	Visitor Role = "visitor"
// )
//
// const (
// 	GraphOwnerNodeRead  Scope = "graph_owner_node_read"
// 	GraphOwnerNodeWrite Scope = "graph_owner_node_write"
//
// 	GraphMemberNodeRead  Scope = "graph_member_node_read"
// 	GraphMemberNodeWrite Scope = "graph_member_node_write"
//
// 	GraphVisitorNodeRead  Scope = "graph_visitor_node_read"
// 	GraphVisitorNodeWrite Scope = "graph_visitor_node_write"
// )
//
// var validScopes = map[Scope]bool{
// 	GraphOwnerNodeRead:    true,
// 	GraphOwnerNodeWrite:   true,
// 	GraphMemberNodeRead:   true,
// 	GraphMemberNodeWrite:  true,
// 	GraphVisitorNodeRead:  true,
// 	GraphVisitorNodeWrite: true,
// }

const (
	GraphNode Resource = "graph_node"
)

const (
	Read  Operation = "read"
	Write Operation = "write"
)

// Resource defines a resource, that belongs to a domain and on which a
// user can perform operations.
type Resource string

// Domain defines a access domain a resource has.
type Domain string

// Operation defines a operation a user can perform on a resource.
type Operation string

// Role defines a user role - an alias for a list of scopes.
type Role string

// Scope defines a single access scope. A access scope defines a permission
// for a user (that has the particular scope) to perform an operation, on a
// resource in a domain.
//
// Example:
//   - `graph_node:owner:read` - defines a scope for the `graph_node` resource
//     to perform `read` operation, on nodes that are in `owner` domain.
type Scope struct {
	resource  Resource
	domain    Domain
	operation Operation
}

// NewScope creates a new scope.
func NewScope(resource Resource, domain Domain, operation Operation) *Scope {
	return &Scope{
		resource:  resource,
		domain:    domain,
		operation: operation,
	}
}

func (s *Scope) Satisfied(scopes []*Scope) error {
	scopes = util.Filter(
		scopes,
		func(fs *Scope) bool {
			return fs.resource == s.resource || fs.resource == "*"
		},
	)

	scopes = util.Filter(
		scopes,
		func(fs *Scope) bool {
			return fs.domain == s.domain || fs.domain == "*"
		},
	)

	scopes = util.Filter(
		scopes,
		func(fs *Scope) bool {
			return fs.operation == s.operation || fs.operation == "*"
		},
	)

	if len(scopes) == 0 {
		return errors.Errorf("no scopes found for %s", s)
	}

	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler for Scope type to check for valid
// values.
func (s *Scope) UnmarshalYAML(unmarshal func(any) error) error {
	var raw string

	err := unmarshal(&raw)
	if err != nil {
		return errors.Wrap(err, "failed to YAML unmarshal scope")
	}

	parts := strings.Split(raw, ":")
	if len(parts) != 3 {
		return errors.Errorf("invalid scope: %s", raw)
	}

	s.resource = Resource(parts[0])
	s.domain = Domain(parts[1])
	s.operation = Operation(parts[2])

	return nil
}

func (s Scope) String() string {
	return strings.Join(
		[]string{
			string(s.resource),
			string(s.domain),
			string(s.operation),
		},
		":",
	)
}

func (s Scope) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Scope) UnmarshalJSON(b []byte) error {
	var raw string

	err := json.Unmarshal(b, &raw)
	if err != nil {
		return errors.Wrap(err, "failed to JSON unmarshal scope")
	}

	parts := strings.Split(raw, ":")
	if len(parts) != 3 {
		return errors.Errorf("invalid scope: %s", raw)
	}

	s.resource = Resource(parts[0])
	s.domain = Domain(parts[1])
	s.operation = Operation(parts[2])

	return nil
}

// Scopes defines a list of access scopes or a role.
// Can be unmarshaled from string - a role, defined in the auth configs roles
// map.
// Can be unmarshaled from list of strings - list of scopes.
type Scopes struct {
	role   Role
	scopes []*Scope
}

// Get returns a list of scopes, either from the role or slice of scopes.
func (s Scopes) Get(roles map[Role][]*Scope) []*Scope {
	if len(s.scopes) > 0 {
		return s.scopes
	}

	scopes, ok := roles[s.role]
	if !ok {
		return []*Scope{}
	}

	return scopes
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

		s.role = role
		s.scopes = []*Scope{}

		return nil
	}

	scopes := []*Scope{}

	err = unmarshal(&scopes)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal scopes as list")
	}

	s.scopes = scopes

	return nil
}

// // UnmarshalYAML implements yaml.Unmarshaler for Role type to check for valid
// // values.
// func (r *Role) UnmarshalYAML(unmarshal func(any) error) error {
// 	var raw string
//
// 	err := unmarshal(&raw)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to unmarshal role")
// 	}
//
// 	role := Role(raw)
//
// 	switch role {
// 	case Owner, Member, Viewer:
// 		*r = role
//
// 		return nil
// 	default:
// 		return errors.Errorf("invalid role: %s", raw)
// 	}
// }
