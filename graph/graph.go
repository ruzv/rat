package graph

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// var log = logging.MustGetLogger("graph")

// Store describes a graph store.
type Store interface {
	GetByID(id uuid.UUID) (*Node, error)
	GetByPath(path string) (*Node, error)
	Leafs(path string) ([]*Node, error)
	Add(parent *Node, name, template string) (*Node, error)
	Root() (*Node, error)
	Update(node *Node) error
	Move(node *Node, path string) error
	Delete(node *Node) error
}

// NodePath filesystem like path that describes where a node is located in the
// graph.
type NodePath string

// GraphProvider describes read and write opetations on a graph.
type GraphProvider interface {
	// GetNodeByID returns a node by id.
	GetNodeByID(id uuid.UUID) (*Node, error)
	// GetNodeByPath returns a node by path.
	GetNodeByPath(path NodePath) (*Node, error)
	// GetLeafNodes returns all leaf nodes of a node specified by id.
	GetLeafNodes(id uuid.UUID) ([]*Node, error)
	// GetRoot returns the root node of graph.
	GetRoot() (*Node, error)
	// AddLeafNode adds a new node to the graph.
	AddLeafNode(parentID uuid.UUID, leaf *Node) (*Node, error)
}

// Node describes a single node.
type Node struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Content  string    `json:"content"`
	Template string    `json:"template"`
	Store    Store     `json:"-"`
}

type NodeUpdate struct {
	Todos []string
}

// -------------------------------------------------------------------------- //
// LEAFS
// -------------------------------------------------------------------------- //

// Leafs returns all leafs of node.
func (n *Node) Leafs() ([]*Node, error) {
	leafs, err := n.Store.Leafs(n.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get leafs")
	}

	return leafs, nil
}

// GenID enerates an sets a new ID for node.
func (n *Node) GenID() error {
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "failed to generate uuid")
	}

	n.ID = id

	return nil
}

// Add new node as child with name.
func (n *Node) Add(name string) (*Node, error) {
	template, err := n.getTemplate()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get template")
	}

	node, err := n.Store.Add(n, name, template)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add node")
	}

	return node, nil
}

func (n *Node) getTemplate() (string, error) {
	root, err := n.Store.Root()
	if err != nil {
		return "", errors.Wrap(err, "failed to get root")
	}

	return func() (string, error) {
		if n.Template != "" {
			return n.Template, nil
		}

		p, err := n.Parent()
		if err != nil {
			return "", errors.Wrap(err, "failed to get parent")
		}

		if p.ID == root.ID {
			return n.Template, nil
		}

		return p.getTemplate()
	}()
}

// Walk to every child node recursively starting from n. callback is called
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

// Update updates node.
func (n *Node) Update(nu *NodeUpdate) error {
	for _, todo := range nu.Todos {
		n.Content = strings.Replace(
			n.Content,
			fmt.Sprintf("- %s", todo),
			fmt.Sprintf("x %s", todo),
			1,
		)
	}

	for _, todo := range nu.Todos {
		n.Content = strings.Replace(
			n.Content,
			fmt.Sprintf("- %s", todo),
			fmt.Sprintf("x %s", todo),
			1,
		)
	}

	err := n.Store.Update(n)
	if err != nil {
		return errors.Wrap(err, "failed to update")
	}

	return nil
}

// Rename renames a node.
func (n *Node) Rename(name string) error {
	err := n.Store.Move(n, filepath.Join(ParentPath(n.Path), name))
	if err != nil {
		return errors.Wrap(err, "failed to rename")
	}

	return nil
}

// Move node to new path.
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

// DeleteAll deletes a node and all its children.
func (n *Node) DeleteAll() error {
	err := n.Store.Delete(n)
	if err != nil {
		return errors.Wrap(err, "failed to delete")
	}

	return nil
}

// DeleteSingle deletes only a single node, moving all children to parent.
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

// Metrics groups all nodes metrics.
type Metrics struct {
	Nodes    int
	MaxDepth int
	MaxLeafs int
	AvgLeafs float64
}

// Metrics calculates metrics for node.
func (n *Node) Metrics() (*Metrics, error) {
	var (
		m        Metrics
		hasLeafs int
	)

	err := n.Walk(
		func(depth int, node *Node) bool {
			m.Nodes++

			if depth > m.MaxDepth {
				m.MaxDepth = depth
			}

			leafs, err := node.Leafs()
			if err != nil {
				return true
			}

			if len(leafs) == 0 {
				return true
			}

			if len(leafs) > m.MaxLeafs {
				m.MaxLeafs = len(leafs)
			}

			m.AvgLeafs += float64(len(leafs))
			hasLeafs++

			return true
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk graph")
	}

	m.AvgLeafs /= float64(hasLeafs)

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

// PathDepth returns depth.
func PathDepth(path string) int {
	return len(PathParts(path))
}

// PathParts returns path parts.
func PathParts(path string) []string {
	split := strings.Split(path, "/")
	parts := make([]string, 0, len(split))

	for _, part := range split {
		if part == "" {
			continue
		}

		parts = append(parts, part)
	}

	return parts
}

// ParentPath returns parent path of node. Returns root path for root path.
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

// Parent .
func (n *Node) Parent() (*Node, error) {
	parentPath := ParentPath(n.Path)

	p, err := n.Store.GetByPath(parentPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get parent")
	}

	return p, nil
}
