package api

import (
	"github.com/gorilla/mux"
	"rat/logr"
)

type Config struct{}

type API struct {
	log  *logr.LogR
	auth bool
}

func NewAPI(config *Config, log *logr.LogR) (*API, error) {
	log = log.Prefix("api")

	router := mux.NewRouter()

	return &API{
		log: log,
	}, nil
}

func (api *API) GraphHandler() {
}
