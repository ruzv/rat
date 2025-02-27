package pathcache

import (
	"sync"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"rat/graph"
	pathutil "rat/graph/util/path"
	"rat/logr"
)

var _ graph.Provider = (*Provider)(nil)

// Provider implements graph.Store reading and writing to the filesystem.
// Node paths are cached in memory.
type Provider struct {
	base    graph.Provider
	cache   map[uuid.UUID]pathutil.NodePath
	cacheMu sync.Mutex
}

// NewPathCache returns a new PathCache.
func NewPathCache(base graph.Provider, log *logr.LogR) *Provider {
	log = log.Prefix("pathcache")
	log.Infof("enabled")

	return &Provider{
		base:  base,
		cache: make(map[uuid.UUID]pathutil.NodePath),
	}
}

// GetByID attempts to get node by ID, fist checking cache for path.
func (p *Provider) GetByID(id uuid.UUID) (*graph.Node, error) {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()

	path, ok := p.cache[id]
	if !ok {
		n, err := p.base.GetByID(id)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node by id")
		}

		p.cache[id] = n.Path

		return n, nil
	}

	n, err := p.base.GetByPath(path)
	if err != nil {
		// this could mean that cache is stale and produces invalid paths
		n, err = p.base.GetByID(id) //
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node by id")
		}

		p.cache[id] = n.Path

		return n, nil
	}

	if n.Header.ID != id { // node changed, but path is the same
		p.cache[n.Header.ID] = n.Path

		n, err = p.base.GetByID(id)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node by id")
		}

		p.cache[id] = n.Path

		return n, nil
	}

	return n, nil
}

// GetByPath returns a node by path. Caches the read node's path.
func (p *Provider) GetByPath(path pathutil.NodePath) (*graph.Node, error) {
	n, err := p.base.GetByPath(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get node by path")
	}

	p.cacheMu.Lock()
	p.cache[n.Header.ID] = n.Path
	p.cacheMu.Unlock()

	return n, nil
}

// GetLeafs returns all leaf nodes of a node. Caches the read nodes' paths.
func (p *Provider) GetLeafs(path pathutil.NodePath) ([]*graph.Node, error) {
	leafs, err := p.base.GetLeafs(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get leafs")
	}

	p.cacheMu.Lock()
	for _, n := range leafs {
		p.cache[n.Header.ID] = n.Path
	}
	p.cacheMu.Unlock()

	return leafs, nil
}

// Move moves a node to a new path. Caches the moved nodes path.
func (p *Provider) Move(id uuid.UUID, path pathutil.NodePath) error {
	err := p.base.Move(id, path)
	if err != nil {
		return errors.Wrap(err, "failed to move node")
	}

	p.cacheMu.Lock()
	// remove old path
	delete(p.cache, id)
	// add new path
	p.cache[id] = path
	p.cacheMu.Unlock()

	return nil
}

func (p *Provider) Write(n *graph.Node) error {
	err := p.base.Write(n)
	if err != nil {
		return errors.Wrap(err, "failed to write node")
	}

	p.cacheMu.Lock()
	p.cache[n.Header.ID] = n.Path
	p.cacheMu.Unlock()

	return nil
}

// Delete wraps underlying implementation delete also deleting node from cache.
func (p *Provider) Delete(n *graph.Node) error {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()

	err := p.base.Delete(n)
	if err != nil {
		return errors.Wrap(err, "failed to delete node")
	}

	delete(p.cache, n.Header.ID)

	return nil
}
