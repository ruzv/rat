package access

import (
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/services/auth"
	pathutil "rat/graph/util/path"
	"rat/logr"
)

var _ graph.Provider = (*Provider)(nil)

var ErrAccessDenied = errors.New("access denied")

type Provider struct {
	base graph.Provider
	log  *logr.LogR
	role auth.Role
}

// NewProvider creates a new filesystem graph provider.
func NewProvider(
	base graph.Provider,
	log *logr.LogR,
	role auth.Role,
) *Provider {
	return &Provider{
		base: base,
		log:  log.Prefix("access"),
		role: role,
	}
}

// GetByID reads node by id, first checkint if role configured for provides
// allows access to node.
func (p *Provider) GetByID(id uuid.UUID) (*graph.Node, error) {
	n, err := p.base.GetByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get node by ID")
	}

	// access, err := n.Access(p.base)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed to get node access")
	// }

	// if !p.role.Allowed(access) {
	// 	return nil, ErrAccessDenied
	// }

	return n, nil
}

func (p *Provider) GetByPath(path pathutil.NodePath) (*graph.Node, error) {
	return nil, nil
}

func (p *Provider) GetLeafs(path pathutil.NodePath) ([]*graph.Node, error) {
	return nil, nil
}

func (p *Provider) Move(id uuid.UUID, path pathutil.NodePath) error {
	return nil
}

func (p *Provider) Write(n *graph.Node) error {
	return nil
}

func (p *Provider) Delete(n *graph.Node) error {
	return nil
}
