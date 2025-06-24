package fuzzy

import (
	"sort"
	"sync"

	"github.com/pkg/errors"
	"github.com/ruzv/rat/internal/graph"
	"github.com/ruzv/rat/internal/graph/services/index"
	"github.com/ruzv/rat/internal/graph/util"
	pathutil "github.com/ruzv/rat/internal/graph/util/path"
	"github.com/ruzv/rat/pkg/logr"
	"github.com/sahilm/fuzzy"
)

var _ index.Indexer = (*Index)(nil)

// Index describes a graphs index.
type Index struct {
	log     *logr.LogR
	pathsMu sync.RWMutex
	paths   []string
}

// NewIndex loads or creates a index in the specified location.
func NewIndex(
	log *logr.LogR, provider graph.Provider,
) (*Index, error) {
	idx := &Index{
		log: log.Prefix("index"),
	}

	err := idx.load(provider)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update index")
	}

	return idx, nil
}

// Search queries the index.
func (idx *Index) Search(query string) ([]string, error) {
	idx.pathsMu.RLock()
	defer idx.pathsMu.RUnlock()

	matches := fuzzy.Find(query, idx.paths)

	if len(matches) > index.SearchResultLimit {
		matches = matches[:index.SearchResultLimit]
	}

	return util.Map(matches, func(m fuzzy.Match) string { return m.Str }), nil
}

// Add adds node path to index.
func (idx *Index) Add(path pathutil.NodePath) {
	idx.pathsMu.Lock()
	defer idx.pathsMu.Unlock()

	pos := sort.SearchStrings(idx.paths, path.String())
	if pos < len(idx.paths) && idx.paths[pos] == path.String() {
		return // already in index
	}

	// paths         [a b c d e f g -]
	// pos           3      ^
	// paths[pos+1:]         [e f g -] dst
	// paths[pos:]         [d e f g -] src
	// copy          [a b c - d e f g]
	//                      ^
	//               [a b c x d e f g]

	idx.paths = append(idx.paths, "")
	// shift by one position
	copy(idx.paths[pos+1:], idx.paths[pos:])
	idx.paths[pos] = path.String()
}

// Remove node path from index.
func (idx *Index) Remove(path pathutil.NodePath) {
	idx.pathsMu.Lock()
	defer idx.pathsMu.Unlock()

	pos := sort.SearchStrings(idx.paths, path.String())
	if pos == len(idx.paths) || idx.paths[pos] != path.String() {
		return // not in index
	}

	copy(idx.paths[pos:], idx.paths[pos+1:])
	idx.paths = idx.paths[:len(idx.paths)-1]
}

func (idx *Index) load(provider graph.Provider) error {
	idx.pathsMu.Lock()
	idx.paths = []string{}
	idx.pathsMu.Unlock()

	r, err := provider.GetByID(graph.RootNodeID)
	if err != nil {
		return errors.Wrap(err, "failed to get root node")
	}

	err = r.Walk(
		provider,
		func(_ int, node *graph.Node) (bool, error) {
			idx.Add(node.Path)

			return true, nil
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to walk root node")
	}

	idx.log.Infof("loaded %d paths", len(idx.paths))

	return nil
}
