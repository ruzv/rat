package root

import (
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"rat/graph"
	pathutil "rat/graph/util/path"
)

var _ graph.Provider = (*Provider)(nil)

var defaultConfig = Config{ //nolint:gochecknoglobals // constant.
	Content: "# 🐀🐀🐀 Welcome to rat! 🐀🐀🐀",
	Template: &graph.NodeTemplate{
		Name:    "{{ .RawName }}",
		Weight:  "0",
		Content: "# {{ .Name }}\n\n<rat graph />\n",
	},
}

// Config contains root node provider configuration parameters.
type Config struct {
	Content  string              `yaml:"content"`
	Template *graph.NodeTemplate `yaml:"template"`
}

// Provider wraps a provider, catching method calls for root node.
type Provider struct {
	graph.RootsProvider
	root Config
}

// NewProvider creates a new root node provider.
func NewProvider(base graph.RootsProvider, root *Config) *Provider {
	return &Provider{
		RootsProvider: base,
		root:          root.fillDefaults(),
	}
}

// GetByID returns node by ID.
func (p *Provider) GetByID(id uuid.UUID) (*graph.Node, error) {
	if id != graph.RootNodeID {
		return p.RootsProvider.GetByID(id) //nolint:wrapcheck // avoid stutter.
	}

	return p.rootNode(), nil
}

// GetByPath returns node by path.
func (p *Provider) GetByPath(path pathutil.NodePath) (*graph.Node, error) {
	if path != graph.RootNodePath {
		return p.RootsProvider.GetByPath( //nolint:wrapcheck // avoid stutter.
			path,
		)
	}

	return p.rootNode(), nil
}

// GetLeafs returns leafs of node by path.
func (p *Provider) GetLeafs(path pathutil.NodePath) ([]*graph.Node, error) {
	if path != graph.RootNodePath {
		return p.RootsProvider.GetLeafs( //nolint:wrapcheck // avoid stutter.
			path,
		)
	}

	leafs, err := p.RootsProvider.Roots()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get root nodes")
	}

	return leafs, nil
}

// Move moves node to a new path.
func (p *Provider) Move(id uuid.UUID, path pathutil.NodePath) error {
	if id != graph.RootNodeID {
		return p.RootsProvider.Move( //nolint:wrapcheck // avoid stutter.
			id,
			path,
		)
	}

	return errors.New("cannot move root node")
}

// Write writes node to a storage.
func (p *Provider) Write(node *graph.Node) error {
	if node.Header.ID != graph.RootNodeID {
		return p.RootsProvider.Write(node) //nolint:wrapcheck // avoid stutter.
	}

	return errors.New("cannot write root node, update config to edit root node")
}

// Delete deletes node from a storage.
func (p *Provider) Delete(node *graph.Node) error {
	if node.Header.ID != graph.RootNodeID {
		return p.RootsProvider.Delete(node) //nolint:wrapcheck // avoid stutter.
	}

	return errors.New("cannot delete root node")
}

func (p *Provider) rootNode() *graph.Node {
	return &graph.Node{
		Header: graph.NodeHeader{
			ID:       uuid.Nil,
			Template: p.root.Template,
		},
		Content: p.root.Content,
	}
}

func (c *Config) fillDefaults() Config {
	if c == nil {
		return defaultConfig
	}

	fill := *c

	if c.Content == "" {
		fill.Content = defaultConfig.Content
	}

	if c.Template == nil {
		fill.Template = defaultConfig.Template
	} else {
		if c.Template.Name == "" {
			fill.Template.Name = defaultConfig.Template.Name
		}
		if c.Template.Weight == "" {
			fill.Template.Weight = defaultConfig.Template.Weight
		}
		if c.Template.Content == "" {
			fill.Template.Content = defaultConfig.Template.Content
		}
	}

	return fill
}
