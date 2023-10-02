package config

import (
	"os"
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
	Name pathutil.NodePath `yaml:"name" validate:"nonzero"`
	Path string            `yaml:"path" validate:"nonzero"`
	Sync *SyncConfig       `yaml:"sync" validate:"nonnil"`
	Auth *AuthConfig       `yaml:"auth" validate:"nonnil"`
}

// SyncConfig defines configuration params for periodically syncing graph to a
// git repository.
type SyncConfig struct {
	Interval    time.Duration `yaml:"interval" validate:"nonzero"`
	KeyPath     string        `yaml:"keyPath" validate:"nonzero"`
	KeyPassword string        `yaml:"keyPassword"`
}

// AuthConfig defines configuration params for authentication.
type AuthConfig struct {
	Username        string        `yaml:"username" validate:"nonzero"`
	PasswordHash    string        `yaml:"passwordHash" validate:"nonzero"`
	TokenExpiration time.Duration `yaml:"tokenExpiration" validate:"nonzero"`
	Secret          string        `yaml:"secret" validate:"nonzero"`
}

// Load loads the configuration from a file.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	defer f.Close()

	c := &Config{}

	err = yaml.NewDecoder(f).Decode(c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode config")
	}

	err = validator.Validate(c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate config")
	}

	return c, nil
}
