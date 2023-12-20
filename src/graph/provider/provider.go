package provider

import (
	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/provider/filesystem"
	"rat/graph/provider/pathcache"
	"rat/graph/provider/root"
)

// Config contains provider configuration parameters.
type Config struct {
	Dir             string       `yaml:"dir" validate:"nonzero"`
	Root            *root.Config `yaml:"root"`
	EnablePathCache bool         `yaml:"enablePathCache"`
}

// New creates a new provider.
func New(c *Config) (graph.Provider, error) { //nolint:ireturn // i know better.
	fs, err := filesystem.NewProvider(c.Dir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create filesystem provider")
	}

	var p graph.Provider = root.NewProvider(fs, c.Root)

	if c.EnablePathCache {
		p = pathcache.NewPathCache(p)
	}

	return p, nil
}
