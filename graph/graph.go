package graph

import (
	"path/filepath"
	"strings"

	"github.com/op/go-logging"
	"github.com/pkg/errors"

	"github.com/gofrs/uuid"
)

var log = logging.MustGetLogger("graph")

type Store interface {
	GetByID(id uuid.UUID) (*Node, error)
	GetByPath(path string) (*Node, error)
	Leafs(path string) ([]*Node, error)
	Add(parent *Node, name string) (*Node, error)
	Root() (*Node, error)
	Update(node *Node) error
	Move(node *Node, path string) error
	Delete(node *Node) error
}

type Node struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Content string    `json:"content"`
	Store   Store     `json:"-"`
}

// -------------------------------------------------------------------------- //
// LEAFS
// -------------------------------------------------------------------------- //

func (n *Node) Leafs() ([]*Node, error) {
	leafs, err := n.Store.Leafs(n.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get leafs")
	}

	return leafs, nil
}

func (n *Node) GenID() error {
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "failed to generate uuid")
	}

	n.ID = id

	return nil
}

func (n *Node) Add(name string) (*Node, error) {
	node, err := n.Store.Add(n, name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add node")
	}

	return node, nil
}

// walks to every child node recursively starting from n. callback is called
// for every child node. callback is not called for n.
func (n *Node) Walk(callback func(int, *Node) bool) error {
	return n.walk(0, callback)
}

func (n *Node) walk(depth int, callback func(int, *Node) bool) error {
	leafs, err := n.Leafs()
	if err != nil {
		return errors.Wrap(err, "failed to get leafs")
	}

	for _, leaf := range leafs {
		if callback != nil && !callback(depth, leaf) {
			continue // callback returned false, skip this branch
		}

		err = leaf.walk(depth+1, callback)
		if err != nil {
			return errors.Wrap(err, "failed to walk leaf")
		}
	}

	return nil
}

// -------------------------------------------------------------------------- //
// UPDATE
// -------------------------------------------------------------------------- //

func (n *Node) Update() error {
	err := n.Store.Update(n)
	if err != nil {
		return errors.Wrap(err, "failed to update")
	}

	return nil
}

func (n *Node) Rename(name string) error {
	err := n.Store.Move(n, filepath.Join(ParentPath(n.Path), name))
	if err != nil {
		return errors.Wrap(err, "failed to rename")
	}

	return nil
}

func (n *Node) Move(path string) error {
	err := n.Store.Move(n, path)
	if err != nil {
		return errors.Wrap(err, "failed to move")
	}

	return nil
}

// -------------------------------------------------------------------------- //
// DELETE
// -------------------------------------------------------------------------- //

func (n *Node) DeleteAll() error {
	err := n.Store.Delete(n)
	if err != nil {
		return errors.Wrap(err, "failed to delete")
	}

	return nil
}

func (n *Node) DeleteSingle() error {
	leafs, err := n.Leafs()
	if err != nil {
		return errors.Wrap(err, "failed to get leafs")
	}

	parent := ParentPath(n.Path)

	for _, leaf := range leafs {
		err = leaf.Move(filepath.Join(parent, leaf.Name))
		if err != nil {
			return errors.Wrap(err, "failed to move leaf node")
		}
	}

	err = n.DeleteAll()
	if err != nil {
		return errors.Wrap(err, "failed to delete all")
	}

	return nil
}

// -------------------------------------------------------------------------- //
// METRICS
// -------------------------------------------------------------------------- //

type Metrics struct {
	Nodes    int
	MaxDepth int
}

func (n *Node) Metrics() (*Metrics, error) {
	var m Metrics

	err := n.Walk(
		func(depth int, node *Node) bool {
			m.Nodes++

			if depth > m.MaxDepth {
				m.MaxDepth = depth
			}

			return true
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk graph")
	}

	return &m, nil
}

// -------------------------------------------------------------------------- //
// UTILS
// -------------------------------------------------------------------------- //

// NameFromPath returns name of node from its path.
func NameFromPath(path string) string {
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		return ""
	}

	if len(parts) == 1 {
		return parts[0]
	}

	return parts[len(parts)-1]
}

func PathDepth(path string) int {
	return len(PathParts(path))
}

func PathParts(path string) []string {
	split := strings.Split(path, "/")

	var parts []string
	for _, part := range split {
		if part == "" {
			continue
		}

		parts = append(parts, part)
	}

	return parts
}

func ParentPath(path string) string {
	parts := PathParts(path)

	if len(parts) < 2 {
		return path
	}

	return strings.Join(parts[:len(parts)-1], "/")
}

// -------------------------------------------------------------------------- //
// PARENT
// -------------------------------------------------------------------------- //

func (n *Node) Parent() (*Node, error) {
	parentPath := ParentPath(n.Path)

	p, err := n.Store.GetByPath(parentPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get parent")
	}

	return p, nil
}
