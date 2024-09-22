package provider

import (
	"encoding/json"

	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/services/provider/filesystem"
	"rat/graph/services/provider/pathcache"
	"rat/graph/services/provider/root"
	"rat/logr"
)

// Config contains provider configuration parameters.
type Config struct {
	Dir             string       `yaml:"dir" validate:"nonzero"`
	Root            *root.Config `yaml:"root"`
	EnablePathCache *bool        `yaml:"enablePathCache"`
}

// New creates a new provider.
func New( //nolint:ireturn // i know better.
	c *Config,
	log *logr.LogR,
) (graph.Provider, error) {
	log = log.Prefix("provider")

	fs, err := filesystem.NewProvider(c.Dir, log)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create filesystem provider")
	}

	var p graph.Provider = root.NewProvider(fs, c.Root)

	if c.EnablePathCache == nil || *c.EnablePathCache {
		p = pathcache.NewPathCache(p, log)
	}

	logMetrics(p, log.Prefix("metrics"))

	return p, nil
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

	log.Infof("%s", string(b))
}
