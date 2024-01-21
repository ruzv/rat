package api

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/services"
	"rat/graph/services/api/router"
	"rat/graph/services/index"
	"rat/graph/services/urlresolve"
	"rat/logr"
)

var _ services.Service = (*API)(nil)

// DefaultTimeouts is the default timeout values for the api server.
//
//nolint:gochecknoglobals
var DefaultTimeouts = &timeouts{
	Read:  15 * time.Second,
	Write: 15 * time.Second,
}

// Config defines configuration parameters for the api server.
type Config struct {
	Port     int       `yaml:"port" validate:"nonzero"`
	Timeouts *timeouts `yaml:"timeouts"`
}

// timeouts defines timeout values for the api server.
type timeouts struct {
	Read  time.Duration `yaml:"read"`
	Write time.Duration `yaml:"write"`
}

// API is the API server service. Implements services.Service.
type API struct {
	log    *logr.LogR
	config *Config
	server *http.Server
}

// New creates a new API server service.
func New(
	config *Config,
	log *logr.LogR,
	provider graph.Provider,
	resolver *urlresolve.Resolver,
	graphIndex *index.GraphIndex,
	webStaticContent fs.FS,
) (*API, error) {
	log = log.Prefix("api")

	r, err := router.New(log, provider, resolver, graphIndex, webStaticContent)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create router")
	}

	timeouts := config.Timeouts.fillDefaults()

	return &API{
		log:    log,
		config: config,

		server: &http.Server{
			Handler:      r,
			Addr:         fmt.Sprintf(":%d", config.Port),
			WriteTimeout: timeouts.Write,
			ReadTimeout:  timeouts.Read,
		},
	}, nil
}

// Run runs the API server.
func (api *API) Run() error {
	api.log.Infof("serving on http://localhost:%d", api.config.Port)

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

func (t *timeouts) fillDefaults() *timeouts {
	if t == nil {
		return DefaultTimeouts
	}

	fill := *t

	if fill.Read == 0 {
		fill.Read = DefaultTimeouts.Read
	}

	if fill.Write == 0 {
		fill.Write = DefaultTimeouts.Write
	}

	return &fill
}
