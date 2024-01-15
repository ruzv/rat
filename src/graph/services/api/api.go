package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"rat/graph/services"
	"rat/graph/services/api/router"
	"rat/logr"
)

var _ services.Service = (*API)(nil)

var DefaultTimeouts = &Timeouts{
	Read:  15 * time.Second,
	Write: 15 * time.Second,
}

type Config struct {
	Port     int       `yaml:"port" validate:"nonzero"`
	Timeouts *Timeouts `yaml:"timeouts"`
}

type Timeouts struct {
	Read  time.Duration `yaml:"read"`
	Write time.Duration `yaml:"write"`
}

type API struct {
	log    *logr.LogR
	auth   bool
	config *Config
	server *http.Server
}

func NewAPI(config *Config, log *logr.LogR) (*API, error) {
	log = log.Prefix("api")

	r, err := router.New(log, nil, nil)
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

func (api *API) Start() error {
	return nil
}

func (api *API) Stop() error {
	return nil
}

// Serve starts the rat server. Blocks.
func (api *API) Serve(exit chan error) {
	api.log.Infof("serving on http://localhost:%d", api.config.Port)

	start := time.Now()

	err := api.server.ListenAndServe()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
	}

	api.log.Infof("uptime: %s", time.Since(start).String())

	exit <- errors.Wrap(err, "listen and serve error")
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
