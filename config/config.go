package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/validator.v2"
	"gopkg.in/yaml.v2"
	pathutil "rat/graph/util/path"
	"rat/logr"
)

// Config is the configuration for the application.
type Config struct {
	Port     int           `yaml:"port" validate:"min=1"`
	Graph    *GraphConfig  `yaml:"graph" validate:"nonnil"`
	LogLevel logr.LogLevel `yaml:"logLevel"`
}

// GraphConfig is the configuration for the graph.
type GraphConfig struct {
	Name        pathutil.NodePath   `yaml:"name" validate:"nonzero"`
	Path        string              `yaml:"path" validate:"nonzero"`
	Sync        *SyncConfig         `yaml:"sync"`
	Fileservers []*FileserverConfig `yaml:"fileservers"`
}

// SyncConfig defines configuration params for periodically syncing graph to a
// git repository.
type SyncConfig struct {
	Interval    time.Duration `yaml:"interval" validate:"nonzero"`
	KeyPath     string        `yaml:"keyPath" validate:"nonzero"`
	KeyPassword string        `yaml:"keyPassword"`
}

// FileserverConfig defines configuration parameters for a fileserver that
// web app can use to retrieve files.
type FileserverConfig struct {
	Authority string `yaml:"authority" validate:"nonzero"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
}

// Load loads the configuration from a file.
func Load(path string) (*Config, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	defer file.Close() //nolint:errcheck // ignore.

	c := &Config{}

	err = yaml.NewDecoder(file).Decode(c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode config")
	}

	err = validator.Validate(c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate config")
	}

	return c, nil
}
