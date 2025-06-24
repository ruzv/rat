package fzf

import (
	"sort"
	"sync"

	"github.com/pkg/errors"
	"github.com/reinhrst/fzf-lib"
	"github.com/ruzv/rat/internal/graph"
	"github.com/ruzv/rat/internal/graph/services/index"
	"github.com/ruzv/rat/internal/graph/util"
	pathutil "github.com/ruzv/rat/internal/graph/util/path"
	"github.com/ruzv/rat/pkg/logr"
)

var _ index.Indexer = (*Index)(nil)

// Index describes a graphs index.
type Index struct {
	log        *logr.LogR
	pathsMu    sync.RWMutex
	paths      []string
	fzf        *fzf.Fzf
	fzfOptions fzf.Options
}

// NewIndex loads or creates a index in the specified location.
func NewIndex(
	log *logr.LogR, provider graph.Provider,
) (*Index, error) {
	log = log.Prefix("index-fzf")

	var (
		idx = &Index{
			log:        log,
			fzfOptions: fzf.DefaultOptions(),
		}
		err error
	)

	idx.paths, err = getAllPaths(provider)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update index")
	}

	idx.fzf = fzf.New(idx.paths, idx.fzfOptions)

	log.Infof("index loaded with %d paths", len(idx.paths))

	return idx, nil
}

// Search queries the index.
func (idx *Index) Search(query string) ([]string, error) {
	idx.pathsMu.RLock()
	defer idx.pathsMu.RUnlock()

	idx.fzf.Search(query)
	result := <-idx.fzf.GetResultChannel()

	matches := result.Matches

	if len(matches) > index.SearchResultLimit {
		matches = matches[:index.SearchResultLimit]
	}

	return util.Map(
		matches,
		func(m fzf.MatchResult) string { return m.Key },
	), nil
}

// Add adds node path to index.
func (idx *Index) Add(path pathutil.NodePath) {
	idx.pathsMu.Lock()
	defer idx.pathsMu.Unlock()

	idx.paths = append(idx.paths, path.String())

	idx.fzf.End()

	idx.fzf = fzf.New(idx.paths, idx.fzfOptions)
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

	idx.fzf.End()

	idx.fzf = fzf.New(idx.paths, idx.fzfOptions)
}

func getAllPaths(provider graph.Provider) ([]string, error) {
	var paths []string

	r, err := provider.GetByID(graph.RootNodeID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get root node")
	}

	err = r.Walk(
		provider,
		func(_ int, node *graph.Node) (bool, error) {
			paths = append(paths, node.Path.String())

			return true, nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk root node")
	}

	return paths, nil
}
