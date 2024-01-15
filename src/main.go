package main

import (
	"context"
	"embed"
	stderrors "errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"rat/args"
	"rat/buildinfo"
	"rat/config"
	"rat/graph/services"
	"rat/logr"
)

//go:embed web/build/*
var embedStaticContent embed.FS

//go:embed logo.txt
var logo string

// Rat describes the rat server.
type Rat struct {
	log *logr.LogR
	gs  *services.GraphServices
	s   *http.Server
}

// NewRat creates a new rat server.
func NewRat(cmdArgs *args.Args) (*Rat, error) {
	conf, err := config.Load(cmdArgs.ConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	log := logr.NewLogR(os.Stdout, "rat", conf.LogLevel)
	log.Infof("%s\nversion: %s", logo, buildinfo.Version())

	gs, err := services.NewGraphServices(conf.Services, log)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create graph services")
	}

	webStaticContent, err := fs.Sub(embedStaticContent, "web/build")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create sub fs from embed")
	}

	r, err := router.NewRouter(log, gs, webStaticContent)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new router")
	}

	return &Rat{
		log: log.Prefix("server"),
		gs:  gs,
		s: &http.Server{
			Handler:      r,
			Addr:         fmt.Sprintf(":%d", conf.Port),
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		},
	}, nil
}

// Serve starts the rat server. Blocks.
func (r *Rat) Serve(exit chan error) {
	r.log.Infof("serving on %s", r.s.Addr)

	start := time.Now()

	err := r.s.ListenAndServe()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
	}

	r.log.Infof("uptime: %s", time.Since(start).String())

	exit <- errors.Wrap(err, "listen and serve error")
}

// RunServer runs the rat server. Listens for SIGINT, to stop server.
func (r *Rat) RunServer() error {
	var (
		exit = make(chan error)
		stop = make(chan os.Signal, 1)
	)

	defer func() {
		close(exit)
		close(stop)
	}()

	signal.Notify(stop, syscall.SIGINT)

	go r.Serve(exit)

	select {
	case err := <-exit:
		if err != nil {
			return errors.Wrap(err, "serve failed")
		}

	case <-stop:
		r.log.Infof("stopping server")

		err := r.s.Shutdown(context.Background()) // trigger exit
		if err != nil {
			r.log.Debugf("not waiting for server exit: %s", err.Error())

			return errors.Wrap(err, "failed to shutdown server")
		}

		err = <-exit // wait for exit
		if err != nil {
			return errors.Wrap(err, "failed to serve")
		}
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}

func run() error {
	cmdArgs, ok := args.Load()
	if !ok {
		return nil
	}

	rat, err := NewRat(cmdArgs)
	if err != nil {
		return errors.Wrap(err, "failed to create rat")
	}

	rErr := rat.RunServer()

	cErr := rat.gs.Close()

	if rErr != nil || cErr != nil {
		return errors.Wrap(
			stderrors.Join(rErr, cErr),
			"failed to run server and/or close services",
		)
	}

	rat.log.Infof("bye")

	return nil
}
