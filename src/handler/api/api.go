package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"rat/handler/api/router"
	"rat/logr"
)

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

	r, err := router.NewRouter(log, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create router")
	}

	r.Path("/graph")

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

	r.log.Infof("uptime: %s", time.Since(start).String())

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
