package graph

import (
	"regexp"
	"sort"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"rat/graph/util"
	pathutil "rat/graph/util/path"
)

var (
	// ErrNodeNotFound is returned when a node is not found.
	ErrNodeNotFound = errors.New("node not found")
	// ErrPartialTemplate is returned when root node template is missing or
	// partial.
	ErrPartialTemplate = errors.New("partial or missing template")
)

var allowedPathNameSymbols = regexp.MustCompile(`[a-zA-Z0-9_\-]`)

// Node describes a single node.
type Node struct {
	Path    pathutil.NodePath `json:"path"`
	Header  NodeHeader        `json:"header"`
	Content string            `json:"content"`
}

// NodeHeader describes info stored in nodes header.
type NodeHeader struct {
	ID       uuid.UUID      `yaml:"id"`
	Name     string         `yaml:"name,omitempty"`
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

// Name returns the name of a node. That being either the defined name in node
// header or first element of nodes path.
func (n *Node) Name() string {
	if n.Header.Name != "" {
		return n.Header.Name
	}

	return n.Path.Name()
}

// GetLeafs returns all leafs of node.
func (n *Node) GetLeafs(r Reader) ([]*Node, error) {
	leafs, err := r.GetLeafs(n.Path)
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
				return leafs[i].Name() < leafs[j].Name()
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

	_, err = p.GetByPath(sub.Path)
	if err == nil {
		return nil, errors.Errorf("node %q already exists", sub.Path)
	}

	err = p.Write(sub)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write node")
	}

	return sub, nil
}

// FillID fills nodes id if it is empty.
func (n *Node) FillID() (uuid.UUID, error) {
	if n.Header.ID.IsNil() {
		id, err := uuid.NewV4()
		if err != nil {
			return uuid.Nil, errors.Wrap(err, "failed to generate id")
		}

		n.Header.ID = id
	}

	return n.Header.ID, nil
}

// Walk to every child node recursively starting from n. callback is called
// for every child node. callback is called for n with depth 0.
func (n *Node) Walk(
	r Reader,
	callback func(depth int, node *Node) (shouldWalkLeafs bool, err error),
) error {
	visitChildren, err := callback(0, n)
	if err != nil {
		return errors.Wrap(err, "callback failed")
	}

	if !visitChildren {
		return nil
	}

	return n.walk(r, 1, callback)
}

// Parent returns parent of node.
func (n *Node) Parent(p Provider) (*Node, error) {
	parent, err := p.GetByPath(n.Path.Parent())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get parent")
	}

	return parent, nil
}

// GetTemplate returns the first template encountered when walking up the tree.
func (n *Node) GetTemplate(p Provider) (*NodeTemplate, error) {
	var (
		nt  = &NodeTemplate{}
		err error
	)

	root, err := p.GetByID(RootNodeID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get root node")
	}

	nt.Name, err = getTemplateField(
		p, n, root.Header.Template,
		func(nt *NodeTemplate) string { return nt.Name },
		func(s string) bool { return s == "" },
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get name field")
	}

	nt.Weight, err = getTemplateField(
		p, n, root.Header.Template,
		func(nt *NodeTemplate) string { return nt.Weight },
		func(s string) bool { return s == "" },
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get weight field")
	}

	nt.Content, err = getTemplateField(
		p, n, root.Header.Template,
		func(nt *NodeTemplate) string { return nt.Content },
		func(s string) bool { return s == "" },
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get content field")
	}

	nt.Template, err = getTemplateField(
		p, n, root.Header.Template,
		func(nt *NodeTemplate) *NodeTemplate { return nt.Template },
		func(nt *NodeTemplate) bool { return nt == nil },
	)
	if err != nil {
		if !errors.Is(err, ErrPartialTemplate) {
			return nil, errors.Wrap(err, "failed to get template field")
		}

		nt.Template = nil // allow root recursive template to be nil.
	}

	return nt, nil
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
		func(_ int, node *Node) (bool, error) {
			childNodes = append(childNodes, node)

			return true, nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk graph")
	}

	return childNodes, nil
}

func (n *Node) walk(
	r Reader,
	depth int,
	callback func(int, *Node) (bool, error),
) error {
	leafs, err := n.GetLeafs(r)
	if err != nil {
		return errors.Wrap(err, "failed to get leafs")
	}

	for _, leaf := range leafs {
		walkLeaf, err := callback(depth, leaf)
		if err != nil {
			return errors.Wrap(err, "callback failed")
		}

		if !walkLeaf {
			continue // callback returned false, skip this branch
		}

		err = leaf.walk(r, depth+1, callback)
		if err != nil {
			return errors.Wrap(err, "failed to walk leaf")
		}
	}

	return nil
}

func getTemplateField[T any]( //nolint:ireturn
	p Provider,
	n *Node,
	rootTemplate *NodeTemplate,
	getter func(*NodeTemplate) T,
	empty func(T) bool,
) (T, error) {
	var nilT T

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

	if parent.Header.ID == n.Header.ID {
		if n.Header.Template == nil {
			return getter(rootTemplate), nil
		}

		field := getter(n.Header.Template)

		if empty(field) {
			return getter(rootTemplate), nil
		}

		return field, nil
	}

	return getTemplateField(p, parent, rootTemplate, getter, empty)
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

	pathName := parsePathName(name)
	if pathName == "" {
		return nil, errors.Errorf("empty path name, parsed from %q", name)
	}

	var headerName string

	if pathName != name { // mismatch, need to set.
		headerName = name
	}

	return &Node{
		Path: n.Path.JoinName(pathName),
		Header: NodeHeader{
			ID:       id,
			Name:     headerName,
			Weight:   weight,
			Template: templ.Template,
		},
		Content: content,
	}, nil
}

func parsePathName(name string) string {
	fields := strings.Fields(strings.TrimSpace(strings.ToLower(name)))
	cleanFields := make([]string, 0, len(fields))

	for _, field := range fields {
		var cleanField string

		for _, r := range field {
			if !allowedPathNameSymbols.MatchString(string(r)) {
				continue
			}

			cleanField += string(r)
		}

		if cleanField == "" {
			continue
		}

		cleanFields = append(cleanFields, cleanField)
	}

	return strings.Join(cleanFields, "-")
}
