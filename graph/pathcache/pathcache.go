package pathcache

import (
	"sync"

	"private/rat/graph"
	pathutil "private/rat/graph/util/path"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

var _ graph.Provider = (*PathCache)(nil)

// PathCache implements graph.Store reading and writing to the filesystem.
// Node paths are cached in memory.
type PathCache struct {
	p       graph.Provider
	cache   map[uuid.UUID]pathutil.NodePath
	cacheMu sync.Mutex
}

// NewPathCache returns a new PathCache.
func NewPathCache(p graph.Provider) *PathCache {
	return &PathCache{
		p:     p,
		cache: make(map[uuid.UUID]pathutil.NodePath),
	}
}

// GetByID attempts to get node by ID, fist checking cache for path.
func (pc *PathCache) GetByID(id uuid.UUID) (*graph.Node, error) {
	pc.cacheMu.Lock()
	defer pc.cacheMu.Unlock()

	path, ok := pc.cache[id]
	if !ok {
		n, err := pc.p.GetByID(id)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node by id")
		}

		pc.cache[id] = n.Path

		return n, nil
	}

	n, err := pc.p.GetByPath(path)
	if err != nil {
		// this could mean that cache is stale and produces invalid paths
		n, err := pc.p.GetByID(id)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node by id")
		}

		pc.cache[id] = n.Path

		return n, nil
	}

	if n.ID != id { // node changed, but path is the same
		pc.cache[n.ID] = n.Path

		n, err := pc.p.GetByID(id)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get node by id")
		}

		pc.cache[id] = n.Path

		return n, nil
	}

	return n, nil
}

// GetByPath returns a node by path. Caches the read node's path.
func (pc *PathCache) GetByPath(path pathutil.NodePath) (*graph.Node, error) {
	n, err := pc.p.GetByPath(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get node by path")
	}

	pc.cacheMu.Lock()
	pc.cache[n.ID] = n.Path
	pc.cacheMu.Unlock()

	return n, nil
}

// GetLeafs returns all leaf nodes of a node. Caches the read nodes' paths.
func (pc *PathCache) GetLeafs(path pathutil.NodePath) ([]*graph.Node, error) {
	leafs, err := pc.p.GetLeafs(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get leafs")
	}

	pc.cacheMu.Lock()
	for _, n := range leafs {
		pc.cache[n.ID] = n.Path
	}
	pc.cacheMu.Unlock()

	return leafs, nil
}

// AddLeaf adds a leaf to the graph. Caches the newly added nodes path.
func (pc *PathCache) AddLeaf(
	parent *graph.Node,
	name string,
) (*graph.Node, error) {
	n, err := pc.p.AddLeaf(parent, name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add leaf")
	}

	pc.cacheMu.Lock()
	pc.cache[n.ID] = n.Path
	pc.cacheMu.Unlock()

	return n, nil
}

// Root returns the root node.
func (pc *PathCache) Root() (*graph.Node, error) {
	n, err := pc.p.Root()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get root")
	}

	pc.cacheMu.Lock()
	pc.cache[n.ID] = n.Path
	pc.cacheMu.Unlock()

	return n, nil
}

// Move moves a node to a new path. Caches the moved nodes path.
func (pc *PathCache) Move(id uuid.UUID, path pathutil.NodePath) error {
	err := pc.p.Move(id, path)
	if err != nil {
		return errors.Wrap(err, "failed to move node")
	}

	pc.cacheMu.Lock()
	// remove old path
	delete(pc.cache, id)
	// add new path
	pc.cache[id] = path
	pc.cacheMu.Unlock()

	return nil
}
