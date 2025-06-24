package index

import pathutil "github.com/ruzv/rat/internal/graph/util/path"

// SearchResultLimit is the maximum number of results returned by the index
// search.
const SearchResultLimit = 20

// Indexer is an interface for searching and managing a graph path index.
type Indexer interface {
	Search(query string) ([]string, error)
	Add(path pathutil.NodePath)
	Remove(path pathutil.NodePath)
}
