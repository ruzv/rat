package index

import (
	"github.com/pkg/errors"
	"github.com/sahilm/fuzzy"
	"rat/graph"
	pathutil "rat/graph/util/path"
	"rat/logr"
)

// GraphIndex describes a graphs index.
type GraphIndex struct {
	log *logr.LogR
	p   graph.Provider

	paths []string
}

// NewIndex loads or creates a index in the specified location.
func NewIndex(
	log *logr.LogR, p graph.Provider,
) (*GraphIndex, error) {
	gi := &GraphIndex{
		log: log.Prefix("index"),
		p:   p,
	}

	err := gi.Update()
	if err != nil {
		return nil, errors.Wrap(err, "failed to update index")
	}

	return gi, nil
}

// Close closes the index.
func (*GraphIndex) Close() error {
	return nil
}

// Search queries the index.
func (gi *GraphIndex) Search(query string) ([]*graph.Node, error) {
	const matchLimit = 20

	matches := fuzzy.Find(query, gi.paths)

	if len(matches) > matchLimit {
		matches = matches[:matchLimit]
	}

	nodes := make([]*graph.Node, 0, len(matches))

	for _, match := range matches {
		node, err := gi.p.GetByPath(pathutil.NodePath(match.Str))
		if err != nil {
			if errors.Is(err, graph.ErrNodeNotFound) {
				err := gi.Update()
				if err != nil {
					return nil, errors.Wrap(err, "failed to update index")
				}

				return gi.Search(query)
			}

			return nil, errors.Wrap(err, "failed to get node")
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// Update adds graph nodes to index.
func (gi *GraphIndex) Update() error {
	gi.log.Infof("updating index")

	gi.paths = nil

	r, err := gi.p.GetByID(graph.RootNodeID)
	if err != nil {
		return errors.Wrap(err, "failed to get root node")
	}

	gi.paths = append(gi.paths, r.Path.String())

	err = r.Walk(
		gi.p,
		func(_ int, node *graph.Node) (bool, error) {
			gi.paths = append(gi.paths, node.Path.String())

			return true, nil
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to walk root node")
	}

	return nil
}
