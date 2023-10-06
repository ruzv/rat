package graph

import (
	"sort"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"rat/graph/util"
	pathutil "rat/graph/util/path"
)

// ErrNodeNotFound is returned when a node is not found.
var ErrNodeNotFound = errors.New("node not found")

// Provider describes graph node manipulations.
type Provider interface {
	GetByID(id uuid.UUID) (*Node, error)
	GetByPath(path pathutil.NodePath) (*Node, error)
	GetLeafs(path pathutil.NodePath) ([]*Node, error)
	Move(id uuid.UUID, path pathutil.NodePath) error
	Write(node *Node) error
	Root() (*Node, error)
}

// Node describes a single node.
type Node struct {
	Name    string            `json:"name"`
	Path    pathutil.NodePath `json:"path"`
	Header  NodeHeader        `json:"header"`
	Content string            `json:"content"`
}

// NodeHeader describes info stored in nodes header.
type NodeHeader struct {
	ID       uuid.UUID      `yaml:"id"`
	Weight   int            `yaml:"weight,omitempty"`
	Template *NodeTemplate  `yaml:"template,omitempty"`
	Any      map[string]any `yaml:",inline"`
}

// Metrics groups all nodes metrics.
type Metrics struct {
	Nodes      int    `json:"nodes"`
	FinalNodes int    `json:"finalNodes"`
	Depth      metric `json:"depth"`
	Leafs      metric `json:"leafs"`
}

type metric struct {
	Max int     `json:"max"`
	Avg float64 `json:"avg"`
}

// GetLeafs returns all leafs of node.
func (n *Node) GetLeafs(p Provider) ([]*Node, error) {
	leafs, err := p.GetLeafs(n.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get leafs")
	}

	sort.SliceStable(
		leafs,
		func(i, j int) bool {
			// 0 no weight sort by name
			// 1 - n sort by weight ascending
			if leafs[i].Header.Weight != 0 && leafs[j].Header.Weight != 0 {
				return leafs[i].Header.Weight < leafs[j].Header.Weight
			}

			if leafs[i].Header.Weight == 0 && leafs[j].Header.Weight == 0 {
				return leafs[i].Name < leafs[j].Name
			}

			if leafs[i].Header.Weight == 0 {
				return false
			}

			return true
		},
	)

	return leafs, nil
}

// AddSub new node as child with name.
func (n *Node) AddSub(p Provider, name string) (*Node, error) {
	sub, err := n.sub(p, name)
	if err != nil {
		return nil, errors.Wrapf(
			err, "failed to create new sub node %q for %q", name, n.Path,
		)
	}

	err = p.Write(sub)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write node")
	}

	return sub, nil
}

// Walk to every child node recursively starting from n. callback is called
// for every child node. callback is not called for n.
func (n *Node) Walk(
	p Provider,
	callback func(depth int, node *Node) (shouldWalkLeafs bool, err error),
) error {
	return n.walk(p, 0, callback)
}

func (n *Node) walk(
	p Provider,
	depth int,
	callback func(int, *Node) (bool, error),
) error {
	leafs, err := n.GetLeafs(p)
	if err != nil {
		return errors.Wrap(err, "failed to get leafs")
	}

	for _, leaf := range leafs {
		if callback != nil {
			walkLeaf, err := callback(depth, leaf)
			if err != nil {
				return errors.Wrap(err, "callback failed")
			}

			if !walkLeaf {
				continue // callback returned false, skip this branch
			}
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
func (n *Node) GetTemplate(p Provider) (*NodeTemplate, error) {
	root, err := p.Root()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get root")
	}

	nt := &NodeTemplate{}

	nt.Name, err = getTemplateField(
		p, n, root.Header.ID,
		func(nt *NodeTemplate) string { return nt.Name },
		func(s string) bool { return s == "" },
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get name field")
	}

	nt.Weight, err = getTemplateField(
		p, n, root.Header.ID,
		func(nt *NodeTemplate) string { return nt.Weight },
		func(s string) bool { return s == "" },
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get weight field")
	}

	nt.Content, err = getTemplateField(
		p, n, root.Header.ID,
		func(nt *NodeTemplate) string { return nt.Content },
		func(s string) bool { return s == "" },
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get content field")
	}

	return nt, nil
}

func getTemplateField[T any](
	p Provider, n *Node, rootID uuid.UUID,
	getter func(*NodeTemplate) T,
	empty func(T) bool,
) (T, error) {
	var nilT T

	if rootID == n.Header.ID {
		if n.Header.Template == nil {
			return nilT, errors.New("root node must have a template")
		}

		field := getter(n.Header.Template)

		if empty(field) {
			return nilT, errors.New("root node template fields cannot be empty")
		}

		return field, nil
	}

	if n.Header.Template != nil {
		field := getter(n.Header.Template)

		if !empty(field) {
			return field, nil
		}
	}

	parent, err := n.Parent(p)
	if err != nil {
		return nilT, errors.Wrap(err, "failed to get parent")
	}

	return getTemplateField(p, parent, rootID, getter, empty)
}

// Metrics calculates metrics for node.
func (n *Node) Metrics(p Provider) (*Metrics, error) {
	var (
		m          Metrics
		hasLeafs   int
		totalLeafs int
		totalDepth int
	)

	errs := []error{}

	err := n.Walk(
		p,
		func(depth int, node *Node) (bool, error) {
			m.Nodes++

			if depth > m.Depth.Max {
				m.Depth.Max = depth
			}

			leafs, err := node.GetLeafs(p)
			if err != nil {
				errs = append(errs, errors.Wrap(err, "failed to get leafs"))

				return false, nil
			}

			if len(leafs) == 0 {
				totalDepth += depth
				m.FinalNodes++

				return true, nil
			}

			if len(leafs) > m.Leafs.Max {
				m.Leafs.Max = len(leafs)
			}

			totalLeafs += len(leafs)
			hasLeafs++

			return true, nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk")
	}

	if len(errs) > 0 {
		return nil, errors.Errorf(
			"failed to walk graph:\n%s",
			strings.Join(
				util.Map(errs, func(err error) string { return err.Error() }),
				"\n",
			),
		)
	}

	if hasLeafs > 0 {
		m.Leafs.Avg = float64(totalLeafs) / float64(hasLeafs)
	}

	if m.FinalNodes > 0 {
		m.Depth.Avg = float64(totalDepth) / float64(m.FinalNodes)
	}

	return &m, nil
}

// ChildNodes returns all child nodes of node.
func (n *Node) ChildNodes(p Provider) ([]*Node, error) {
	var childNodes []*Node

	err := n.Walk(
		p,
		func(d int, node *Node) (bool, error) {
			childNodes = append(childNodes, node)

			return true, nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk graph")
	}

	return childNodes, nil
}

func (n *Node) sub(p Provider, name string) (*Node, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate id")
	}

	templ, err := n.GetTemplate(p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get template")
	}

	td := NewTemplateData(name)

	name, err = templ.FillName(&td.RawTemplateData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fill name")
	}

	td.Name = name

	content, err := templ.FillContent(td)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fill content")
	}

	weight, err := templ.FillWeight(td)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fill weight")
	}

	return &Node{
		Name: name,
		Path: n.Path.JoinName(name),
		Header: NodeHeader{
			ID:     id,
			Weight: weight,
		},
		Content: content,
	}, nil
}
