package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	pathutil "rat/graph/util/path"

	"github.com/pkg/errors"
	"gopkg.in/validator.v2"
	"gopkg.in/yaml.v2"
)

// Config is the configuration for the application.
type Config struct {
	Port  int          `json:"port" yaml:"port" validate:"min=1"`
	Graph *GraphConfig `json:"graph" yaml:"graph" validate:"nonnil"`
}

// GraphConfig is the configuration for the graph.
type GraphConfig struct {
	Name pathutil.NodePath `json:"name" yaml:"name" validate:"nonzero"`
	Path string            `json:"path" yaml:"path" validate:"nonzero"`
	Sync *SyncConfig       `json:"sync" yaml:"sync" validate:"nonnil"`
}

// SyncConfig defines configuration params for periodically syncing graph to a
// git repository.
type SyncConfig struct {
	Interval    time.Duration `json:"interval" yaml:"interval" validate:"nonzero"` //nolint:lll
	KeyPath     string        `json:"keyPath" yaml:"keyPath" validate:"nonzero"`
	KeyPassword string        `json:"keyPassword" yaml:"keyPassword"`
}

// Load loads the configuration from a file.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	defer f.Close()

	c := &Config{}

	switch filepath.Ext(path) {
	case ".json":
		err = json.NewDecoder(f).Decode(c)
	case ".yaml", ".yml":
		err = yaml.NewDecoder(f).Decode(c)
	default:
		return nil, errors.Errorf("unknown config file format %s", path)
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to decode config")
	}

	err = validator.Validate(c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate config")
	}

	return c, nil
}
