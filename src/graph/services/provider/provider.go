package provider

import (
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
		p = pathcache.NewProvider(p, log)
	}

	return p, nil
}
