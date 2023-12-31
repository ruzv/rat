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
	base   graph.Provider
	log    *logr.LogR
	scopes []*auth.Scope
}

// NewProvider creates a new filesystem graph provider.
func NewProvider(
	base graph.Provider,
	log *logr.LogR,
	scopes []*auth.Scope,
) *Provider {
	return &Provider{
		base:   base,
		log:    log.Prefix("access"),
		scopes: scopes,
	}
}

// GetByID reads node by id, first checkint if role configured for provides
// allows access to node.
func (p *Provider) GetByID(id uuid.UUID) (*graph.Node, error) {
	p.log.Debugf("GetByID %s", id.String())

	n, err := p.base.GetByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get node by ID")
	}

	domain, err := n.Domain(p.base)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get node domain")
	}

	requiredScope := auth.NewScope(auth.GraphNode, domain, auth.Read)

	err = requiredScope.Satisfied(p.scopes)
	if err != nil {
		return nil, errors.Wrap(err, "scope requirement not satisfied")
	}

	return n, nil
}

func (p *Provider) GetByPath(path pathutil.NodePath) (*graph.Node, error) {
	p.log.Debugf("GetByPath %s", path.String())

	n, err := p.base.GetByPath(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get node by path")
	}

	domain, err := n.Domain(p.base)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get node domain")
	}

	requiredScope := auth.NewScope(auth.GraphNode, domain, auth.Read)

	err = requiredScope.Satisfied(p.scopes)
	if err != nil {
		return nil, errors.Wrap(err, "scope requirement not satisfied")
	}

	return n, nil
}

func (p *Provider) GetLeafs(path pathutil.NodePath) ([]*graph.Node, error) {
	p.log.Debugf("GetLeafs %s", path.String())

	// check access to parent node
	_, err := p.GetByPath(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get node by path")
	}

	p.log.Debugf("base.GetLeafs %s", path.String())

	leafs, err := p.base.GetLeafs(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get node by path")
	}

	p.log.Debugf(
		"iterate leafs for node %q, len(%d)", path.String(), len(leafs),
	)

	var allowedLeafs []*graph.Node

	for _, leaf := range leafs {
		domain, err := leaf.Domain(p.base)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node domain")
		}

		p.log.Debugf("got domain %q for leaf %q", domain, leaf.Path)

		requiredScope := auth.NewScope(auth.GraphNode, domain, auth.Read)

		p.log.Debugf("leaf %q requires scope %q", leaf.Path, requiredScope)

		err = requiredScope.Satisfied(p.scopes)
		if err != nil {
			p.log.Debugf("leaf %q not allowed: %s", leaf.Path, err.Error())

			continue
		}

		p.log.Debugf("leaf %q allowed", leaf.Path)

		allowedLeafs = append(allowedLeafs, leaf)
	}

	return allowedLeafs, nil
}

func (p *Provider) Move(id uuid.UUID, path pathutil.NodePath) error {
	p.log.Debugf("Move %s %s", id.String(), path.String())

	n, err := p.base.GetByID(id)
	if err != nil {
		return errors.Wrap(err, "failed to get node by ID")
	}

	domain, err := n.Domain(p.base)
	if err != nil {
		return errors.Wrap(err, "failed to get node domain")
	}

	requiredScope := auth.NewScope(auth.GraphNode, domain, auth.Write)

	err = requiredScope.Satisfied(p.scopes)
	if err != nil {
		return errors.Wrap(err, "scope requirement not satisfied")
	}

	n, err = p.GetByPath(path.Parent())

	domain, err = n.Domain(p.base)
	if err != nil {
		return errors.Wrap(err, "failed to get node domain")
	}

	requiredScope = auth.NewScope(auth.GraphNode, domain, auth.Write)

	err = requiredScope.Satisfied(p.scopes)
	if err != nil {
		return errors.Wrap(err, "scope requirement not satisfied")
	}

	err = p.base.Move(id, path)
	if err != nil {
		return errors.Wrap(err, "failed to move node")
	}

	return nil
}

// TODO: write is accting like a create and update at the same time. this is
// not ideal and the interface should be separated into two methods.
func (p *Provider) Write(n *graph.Node) error {
	p.log.Debugf("Write %s", n.Path.String())

	// curretly access is only checked for a create scenario, for which write
	// has the only use case. this remains sub-optimal and should be fixes.
	parent, err := p.base.GetByPath(n.Path.Parent())
	if err != nil {
		return errors.Wrap(err, "failed to get parent node")
	}

	domain, err := parent.Domain(p.base)
	if err != nil {
		return errors.Wrap(err, "failed to get node domain")
	}

	requiredScope := auth.NewScope(auth.GraphNode, domain, auth.Write)

	err = requiredScope.Satisfied(p.scopes)
	if err != nil {
		return errors.Wrap(err, "scope requirement not satisfied")
	}

	err = p.base.Write(n)
	if err != nil {
		return errors.Wrap(err, "failed to write node")
	}

	return nil
}

func (p *Provider) Delete(n *graph.Node) error {
	p.log.Debugf("Delete %s", n.Path.String())

	existing, err := p.base.GetByPath(n.Path)
	if err != nil {
		return errors.Wrap(err, "failed to get node by path")
	}

	domain, err := existing.Domain(p.base)
	if err != nil {
		return errors.Wrap(err, "failed to get node domain")
	}

	requiredScope := auth.NewScope(auth.GraphNode, domain, auth.Write)

	err = requiredScope.Satisfied(p.scopes)
	if err != nil {
		return errors.Wrap(err, "scope requirement not satisfied")
	}

	err = p.base.Delete(n)
	if err != nil {
		return errors.Wrap(err, "failed to delete node")
	}

	return nil
}
