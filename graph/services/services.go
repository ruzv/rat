package services

import (
	"encoding/json"

	"rat/config"
	"rat/graph"
	"rat/graph/index"
	"rat/graph/pathcache"
	"rat/graph/singlefile"
	"rat/graph/sync"
	"rat/logr"

	"github.com/pkg/errors"
)

// GraphServices contains service components of a graph.
type GraphServices struct {
	Graph  graph.Provider
	Syncer *sync.Syncer
	Index  *index.GraphIndex
}

// NewGraphServices creates a new graph services.
func NewGraphServices(
	log *logr.LogR, graphConf *config.GraphConfig,
) (*GraphServices, error) {
	p := pathcache.NewPathCache(
		singlefile.NewSingleFile(graphConf.Name, graphConf.Path),
	)

	s, err := sync.NewSyncer(log, graphConf.Path, graphConf.Sync)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create syncer")
	}

	s.Start()

	idx, err := index.NewIndex(log, p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create index")
	}

	logMetrics(log.Prefix("metrics"), p)

	return &GraphServices{
		Graph:  p,
		Syncer: s,
		Index:  idx,
	}, nil
}

// Close closes the graph services.
func (gs *GraphServices) Close() error {
	gs.Syncer.Stop()

	err := gs.Index.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close index")
	}

	return nil
}

func logMetrics(log *logr.LogR, p graph.Provider) {
	r, err := p.Root()
	if err != nil {
		log.Errorf("failed to log metrics: %s", err.Error())

		return
	}

	m, err := r.Metrics(p)
	if err != nil {
		log.Errorf("failed to log metrics: %s", err.Error())

		return
	}

	b, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		log.Errorf("failed to log metrics: %s", err.Error())

		return
	}

	log.Infof(string(b))
}
