package config

import (
	"encoding/json"
	"os"

	pathutil "rat/graph/util/path"

	"github.com/pkg/errors"
)

// Config is the configuration for the application.
type Config struct {
	Port  int          `json:"port"`
	Graph *GraphConfig `json:"graph"`
}

// GraphConfig is the configuration for the graph.
type GraphConfig struct {
	Name pathutil.NodePath `json:"name"`
	Path string            `json:"path"`
}

// Load loads the configuration from a file.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	defer f.Close()

	c := &Config{}

	err = json.NewDecoder(f).Decode(c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode config")
	}

	return c, nil
}
