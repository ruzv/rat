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
var DefaultTimeouts = &Timeouts{
	Read:  15 * time.Second,
	Write: 15 * time.Second,
}

// Config defines configuration parameters for the api server.
type Config struct {
	Port     int       `yaml:"port" validate:"nonzero"`
	Timeouts *Timeouts `yaml:"timeouts"`
}

// Timeouts defines timeout values for the api server.
type Timeouts struct {
	Read  time.Duration `yaml:"read"`
	Write time.Duration `yaml:"write"`
}

// API is the API server service. Implements services.Service.
type API struct {
	log    *logr.LogR
	auth   bool
	config *Config
	server *http.Server
}

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

	timeouts := config.Timeouts.FillDefaults()

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

func (api *API) Stop(ctx context.Context) error {
	api.log.Infof("shutting down")

	err := api.server.Shutdown(ctx) // trigger exit
	if err != nil {
		return errors.Wrap(err, "failed to shutdown server")
	}

	return nil
}

func (t *Timeouts) FillDefaults() *Timeouts {
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
