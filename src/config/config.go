package config

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/validator.v2"
	"gopkg.in/yaml.v2"
	"rat/graph/services/runner"
)

// Config is the configuration for the application.
type Config struct {
	Services *runner.Config `yaml:"services"`
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
