package graph

import (
	"text/template"

	pathutil "private/rat/graph/util/path"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// Provider describes graph node manipulations.
type Provider interface {
	GetByID(id uuid.UUID) (*Node, error)
	GetByPath(path pathutil.NodePath) (*Node, error)
	GetLeafs(path pathutil.NodePath) ([]*Node, error)
	AddLeaf(parent *Node, name string) (*Node, error)
	Root() (*Node, error)
}

// Node describes a single node.
type Node struct {
	ID       uuid.UUID         `json:"id"`
	Name     string            `json:"name"`
	Path     pathutil.NodePath `json:"path"`
	Content  string            `json:"content"`
	Template string            `json:"template"`
}

// GetLeafs returns all leafs of node.
func (n *Node) GetLeafs(p Provider) ([]*Node, error) {
	leafs, err := p.GetLeafs(n.Path)
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
	root, err := p.Root()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get root")
	}

	var getTemplate func(n *Node) (*template.Template, error)

	getTemplate = func(n *Node) (*template.Template, error) {
		if root.ID == n.ID || n.Template != "" {
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

		return getTemplate(parent)
	}

	return getTemplate(n)
}

// Metrics groups all nodes metrics.
type Metrics struct {
	Nodes      int    `json:"nodes"`
	FinalNodes int    `json:"final_nodes"`
	Depth      metric `json:"depth"`
	Leafs      metric `json:"leafs"`
}

type metric struct {
	Max int     `json:"max"`
	Avg float64 `json:"avg"`
}

// Metrics calculates metrics for node.
func (n *Node) Metrics(p Provider) (*Metrics, error) {
	var (
		m          Metrics
		hasLeafs   int
		totalLeafs int
		totalDepth int
	)

	err := n.Walk(
		p,
		func(depth int, node *Node) bool {
			m.Nodes++

			if depth > m.Depth.Max {
				m.Depth.Max = depth
			}

			leafs, err := node.GetLeafs(p)
			if err != nil {
				return true
			}

			if len(leafs) == 0 {
				totalDepth += depth
				m.FinalNodes++

				return true
			}

			if len(leafs) > m.Leafs.Max {
				m.Leafs.Max = len(leafs)
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
		m.Leafs.Avg = float64(totalLeafs) / float64(hasLeafs)
	}

	if m.FinalNodes > 0 {
		m.Depth.Avg = float64(totalDepth) / float64(m.FinalNodes)
	}

	return &m, nil
}
