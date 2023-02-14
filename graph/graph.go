package graph

import (
	"text/template"

	pathutil "private/rat/graph/util/path"

	"github.com/gofrs/uuid"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var log = logging.MustGetLogger("graph")

// Provider describes graph node manipulations.
type Provider interface {
	GetByID(id uuid.UUID) (*Node, error)
	GetByPath(path string) (*Node, error)
	GetLeafs(id uuid.UUID) ([]*Node, error)
	AddLeaf(parent *Node, name string) (*Node, error)
	Root() (*Node, error)
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
}

// GetLeafs returns all leafs of node.
func (n *Node) GetLeafs(p Provider) ([]*Node, error) {
	leafs, err := p.GetLeafs(n.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get leafs")
	}

	return leafs, nil
}

// AddLeaf new node as child with name.
func (n *Node) AddLeaf(p Provider, name string) (*Node, error) {
	node, err := p.AddLeaf(n, name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add node")
	}

	return node, nil
}

// Walk to every child node recursively starting from n. callback is called
// for every child node. callback is not called for n.
func (n *Node) Walk(
	p Provider,
	callback func(depth int, node *Node) bool,
) error {
	return n.walk(p, 0, callback)
}

func (n *Node) walk(
	p Provider,
	depth int,
	callback func(int, *Node) bool,
) error {
	leafs, err := n.GetLeafs(p)
	if err != nil {
		return errors.Wrap(err, "failed to get leafs")
	}

	for _, leaf := range leafs {
		if callback != nil && !callback(depth, leaf) {
			continue // callback returned false, skip this branch
		}

		err = leaf.walk(p, depth+1, callback)
		if err != nil {
			return errors.Wrap(err, "failed to walk leaf")
		}
	}

	return nil
}

// Parent returns parent of node.
func (n *Node) Parent(p Provider) (*Node, error) {
	parent, err := p.GetByPath(pathutil.ParentPath(n.Path))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get parent")
	}

	return parent, nil
}

// GetTemplate returns the first template encountered when walking up the tree.
func (n *Node) GetTemplate(p Provider) (*template.Template, error) {
	if n.Template != "" {
		templ, err := template.New("newNode").Parse(n.Template)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse template")
		}

		return templ, nil
	}

	parent, err := n.Parent(p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get parent")
	}

	return parent.GetTemplate(p)
}

// Metrics groups all nodes metrics.
type Metrics struct {
	Nodes    int
	MaxDepth int
	MaxLeafs int
	AvgLeafs float64
}

// Metrics calculates metrics for node.
func (n *Node) Metrics(p Provider) (*Metrics, error) {
	var (
		m          Metrics
		hasLeafs   int
		totalLeafs int
	)

	err := n.Walk(
		p,
		func(depth int, node *Node) bool {
			m.Nodes++

			if depth > m.MaxDepth {
				m.MaxDepth = depth
			}

			leafs, err := node.GetLeafs(p)
			if err != nil {
				return true
			}

			if len(leafs) == 0 {
				return true
			}

			if len(leafs) > m.MaxLeafs {
				m.MaxLeafs = len(leafs)
			}

			totalLeafs += len(leafs)
			hasLeafs++

			return true
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk graph")
	}

	if hasLeafs > 0 {
		m.AvgLeafs = float64(totalLeafs) / float64(hasLeafs)
	}

	return &m, nil
}
