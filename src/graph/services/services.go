package services

import (
	"encoding/json"

	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/services/index"
	"rat/graph/services/provider"
	"rat/graph/services/sync"
	"rat/graph/services/urlresolve"
	"rat/logr"
)

// Service describes the common methods all rat services must implement.
type Service interface {
	Start() error
	Stop() error
}

// Config contains graph services configuration parameters.
type Config struct {
	Provider    *provider.Config   `yaml:"provider"`
	URLResolver *urlresolve.Config `yaml:"urlResolver"`
	Sync        *sync.Config       `yaml:"sync"`
}

// GraphServices contains service components of a graph.
type GraphServices struct {
	Provider    graph.Provider
	Syncer      *sync.Syncer
	Index       *index.GraphIndex
	URLResolver *urlresolve.Resolver
	log         *logr.LogR
}

// NewGraphServices creates a new graph services.
func NewGraphServices(c *Config, log *logr.LogR) (*GraphServices, error) {
	log = log.Prefix("services")

	var (
		err error
		gs  = &GraphServices{
			URLResolver: urlresolve.NewResolver(c.URLResolver, log),
			log:         log,
		}
	)

	gs.Provider, err = provider.New(c.Provider, log)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create graph provider")
	}

	if c.Sync != nil {
		log.Infof("starting syncer")

		gs.Syncer, err = sync.NewSyncer(c.Sync, log)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create syncer")
		}

		gs.Syncer.Start()
	}

	gs.Index, err = index.NewIndex(log, gs.Provider)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create index")
	}

	logMetrics(gs.Provider, log.Prefix("metrics"))

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

func logMetrics(p graph.Provider, log *logr.LogR) {
	r, err := p.GetByID(graph.RootNodeID)
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
