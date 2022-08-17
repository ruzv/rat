package config

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

type Config struct {
	Port  int          `json:"port"`
	Graph *GraphConfig `json:"graph"`
}

type GraphConfig struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	defer f.Close()

	var c Config

	err = json.NewDecoder(f).Decode(&c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode config")
	}

	return &c, nil
}
