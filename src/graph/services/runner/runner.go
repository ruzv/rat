package runner

import (
	"context"
	stderrors "errors"
	"io/fs"
	stdsync "sync"
	"time"

	"github.com/pkg/errors"
	"rat/graph/services"
	"rat/graph/services/api"
	"rat/graph/services/index"
	"rat/graph/services/provider"
	"rat/graph/services/sync"
	"rat/graph/services/urlresolve"
	"rat/logr"
)

var _ services.Service = (*Runner)(nil)

// Config contains graph services configuration parameters.
type Config struct {
	Provider    *provider.Config   `yaml:"provider"`
	URLResolver *urlresolve.Config `yaml:"urlResolver"`
	Sync        *sync.Config       `yaml:"sync"`
	API         *api.Config        `yaml:"api"`
}

// Runner contains service components of a graph.
type Runner struct {
	services []services.Service
	log      *logr.LogR
}

// NewGraphServices creates a new graph services.
func New(
	c *Config, log *logr.LogR, webStaticContent fs.FS,
) (*Runner, error) {
	log = log.Prefix("services-runner")
	services := []services.Service{}

	resolver := urlresolve.NewResolver(c.URLResolver, log)

	provider, err := provider.New(c.Provider, log)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create graph provider")
	}

	if c.Sync != nil {
		syncer, err := sync.NewSyncer(c.Sync, log)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create syncer")
		}

		services = append(services, syncer)
	}

	graphIndex, err := index.NewIndex(log, provider)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create index")
	}

	graphAPI, err := api.New(
		c.API,
		log,
		provider,
		resolver,
		graphIndex,
		webStaticContent,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create api")
	}

	services = append(services, graphAPI)

	return &Runner{
		services: services,
		log:      log,
	}, nil
}

func (r *Runner) Run() error {
	var (
		runWG   = stdsync.WaitGroup{}
		errsWG  = stdsync.WaitGroup{}
		stopErr error
		runErr  error
		errs    = make(chan error)
		stop    = stdsync.OnceFunc(func() {
			r.log.Debugf("stopping services (once func)")

			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			stopErr = r.Stop(ctx)
		})
	)

	errsWG.Add(1)
	go func() {
		defer errsWG.Done()

		for err := range errs {
			r.log.Debugf("received error from service: %s", err.Error())
			runErr = stderrors.Join(runErr, err)
		}
	}()

	for _, s := range r.services {
		runWG.Add(1)

		go func(s services.Service) {
			defer runWG.Done()

			err := s.Run()
			if err != nil {
				r.log.Errorf("service failed: %s", err.Error())
				stop()
				errs <- err
			}
		}(s)
	}

	runWG.Wait()
	close(errs)
	errsWG.Wait()

	return stderrors.Join(runErr, stopErr)
}

func (r *Runner) Stop(ctx context.Context) error {
	var stopErr error

	for _, s := range r.services {
		err := s.Stop(ctx)
		if err != nil {
			stopErr = stderrors.Join(stopErr, err)
		}
	}

	return stopErr
}
