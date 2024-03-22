package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/services"
	"rat/graph/services/api/httputil"
	"rat/graph/services/api/router"
	"rat/graph/services/index"
	"rat/graph/services/urlresolve"
	"rat/logr"
)

var _ services.Service = (*API)(nil)

// Config defines configuration parameters for the api server.
type Config struct {
	Port           int                `yaml:"port" validate:"nonzero"`
	Authority      string             `yaml:"authority" validate:"nonzero"`
	AllowedOrigins []string           `yaml:"allowedOrigins" validate:"nonzero"`
	URLResolver    *urlresolve.Config `yaml:"urlResolver"`

	Timeouts *httputil.ServerTimeouts `yaml:"timeouts"`
}

// API is the API server service. Implements services.Service.
type API struct {
	log    *logr.LogR
	config *Config
	server *http.Server
}

// New creates a new API server service.
//
//revive:disable:argument-limit
func New(
	config *Config,
	log *logr.LogR,
	provider graph.Provider,
	// resolver *urlresolve.Resolver,
	graphIndex *index.Index,
) (*API, error) {
	log = log.Prefix("api")

	resolver, err := urlresolve.NewResolver(
		config.URLResolver, log, config.Authority,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create url resolver")
	}

	r, err := router.New(
		log,
		provider,
		resolver,
		graphIndex,
		config.AllowedOrigins,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create router")
	}

	timeouts := config.Timeouts.FillDefaults()

	return &API{
		log:    log,
		config: config,
		server: &http.Server{
			Handler:      r,
			Addr:         fmt.Sprintf(":%d", config.Port),
			WriteTimeout: timeouts.Write,
			ReadTimeout:  timeouts.Read,
			IdleTimeout:  timeouts.Idle,
		},
	}, nil
}

// Run runs the API server.
func (api *API) Run() error {
	api.log.Infof("API available on http://localhost:%d", api.config.Port)

	start := time.Now()

	err := api.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "listen and serve error")
	}

	api.log.Infof("uptime: %s", time.Since(start).String())

	return nil
}

// Stop stops the API server.
func (api *API) Stop(ctx context.Context) error {
	api.log.Infof("shutting down")

	err := api.server.Shutdown(ctx) // trigger exit
	if err != nil {
		return errors.Wrap(err, "failed to shutdown server")
	}

	return nil
}
