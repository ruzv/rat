package provider

import (
	"encoding/json"

	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/services/provider/filesystem"
	"rat/graph/services/provider/pathcache"
	"rat/graph/services/provider/root"
	"rat/graph/services/provider/timezone"
	"rat/logr"
)

// Config contains provider configuration parameters.
type Config struct {
	Dir             string       `yaml:"dir" validate:"nonzero"`
	EnablePathCache *bool        `yaml:"enablePathCache"`
	Root            *root.Config `yaml:"root"`
	TimeZone        string       `yaml:"timeZone"`
}

// New creates a new provider.
func New(
	c *Config,
	log *logr.LogR,
) (graph.Provider, error) {
	log = log.Prefix("provider")

	fs, err := filesystem.NewProvider(c.Dir, log)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create filesystem provider")
	}

	var rwProvider graph.ReadWriteProvider = root.NewProvider(fs, c.Root)

	if c.EnablePathCache == nil || *c.EnablePathCache {
		rwProvider = pathcache.NewPathCache(rwProvider, log)
	}

	p, err := timezone.New(rwProvider, c.TimeZone)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new time zone provider")
	}

	// log metrics only when construction of provider is fully done
	logMetrics(rwProvider, log.Prefix("metrics"))

	return p, nil
}

func logMetrics(p graph.ReadWriteProvider, log *logr.LogR) {
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
