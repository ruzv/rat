package graph

import (
	"time"

	"github.com/gofrs/uuid"
	pathutil "rat/graph/util/path"
)

var (
	// RootNodeID is empty UUDI, filled with zeros, id of the root node.
	RootNodeID = uuid.Nil //nolint:gochecknoglobals // constant.
	// RootNodePath is empty path, root node path.
	RootNodePath = pathutil.NodePath("") //nolint:gochecknoglobals // constant.
)

// Provider describes graph node manipulations.
type Provider interface {
	ReadWriteProvider
	TimeZoner
}

// ReadWriteProvider describes graph node read/write manipulations.
type ReadWriteProvider interface {
	Reader
	Writer
}

// RootsProvider describes a provider implementation with an additional Roots
// method that can retrieve root nodes separately.
type RootsProvider interface {
	Roots() ([]*Node, error)
	Reader
	Writer
}

// Reader describes graph read operations.
type Reader interface {
	GetByID(id uuid.UUID) (*Node, error)
	GetByPath(path pathutil.NodePath) (*Node, error)
	GetLeafs(path pathutil.NodePath) ([]*Node, error)
}

// Writer describes graph write operations.
type Writer interface {
	Move(id uuid.UUID, path pathutil.NodePath) error
	Write(node *Node) error
	Delete(node *Node) error
}

// TimeZoner describes method for retrieving the timezone that the graph is in.
type TimeZoner interface {
	TimeZone() *time.Location
}
