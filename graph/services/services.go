package services

import (
	"encoding/json"

	"github.com/pkg/errors"
	"rat/config"
	"rat/graph"
	"rat/graph/index"
	"rat/graph/pathcache"
	"rat/graph/singlefile"
	"rat/graph/sync"
	"rat/logr"
)

// GraphServices contains service components of a graph.
type GraphServices struct {
	Graph  graph.Provider
	Syncer *sync.Syncer
	Index  *index.GraphIndex
	log    *logr.LogR
}

// NewGraphServices creates a new graph services.
func NewGraphServices(
	log *logr.LogR, graphConf *config.GraphConfig,
) (*GraphServices, error) {
	gs := &GraphServices{
		log: log.Prefix("services"),
	}

	sf, err := singlefile.NewSingleFile(graphConf.Name, graphConf.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create single file graph")
	}

	gs.Graph = pathcache.NewPathCache(sf)

	if graphConf.Sync != nil {
		log.Infof("starting syncer")

		gs.Syncer, err = sync.NewSyncer(log, graphConf.Path, graphConf.Sync)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create syncer")
		}

		gs.Syncer.Start()
	}

	gs.Index, err = index.NewIndex(log, gs.Graph)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create index")
	}

	logMetrics(log.Prefix("metrics"), gs.Graph)

	return gs, nil
}

// Close closes the graph services.
func (gs *GraphServices) Close() error {
	if gs.Syncer != nil {
		gs.log.Infof("stopping syncer")

		gs.Syncer.Stop()
	}

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
