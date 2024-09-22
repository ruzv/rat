package main

import (
	"context"
	"embed"
	"io/fs"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"rat/args"
	"rat/config"
	"rat/graph/services/runner"
)

//go:embed web/build/*
var embedStaticContent embed.FS

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}

//nolint:gocyclo,cyclop
func run() error {
	cmdArgs, err := args.Load()
	if err != nil {
		if errors.Is(err, args.ErrExitZero) {
			return nil
		}

		return errors.Wrap(err, "failed to load args")
	}

	conf, err := config.Load(cmdArgs.ConfigPath)
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}

	webStaticContent, err := fs.Sub(embedStaticContent, "web/build")
	if err != nil {
		return errors.Wrap(err, "failed to create sub fs from embed")
	}

	servicesRunner, log, err := runner.New(conf.Services, webStaticContent)
	if err != nil {
		return errors.Wrap(err, "failed to create services runner")
	}

	var (
		exit = make(chan error)
		stop = make(chan os.Signal, 1)
	)

	defer func() {
		close(exit)
		close(stop)
	}()

	signal.Notify(stop, syscall.SIGINT)

	go func() {
		exit <- servicesRunner.Run()

		log.Debugf("runner.Run() returned")
	}()

	select {
	case err := <-exit:
		log.Debugf("runner exit received")

		if err != nil {
			return errors.Wrap(err, "runner failed")
		}

	case <-stop:
		log.Infof("stopping runner")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()

		err := servicesRunner.Stop(ctx)
		if err != nil {
			log.Debugf("not waiting for runner exit: %s", err.Error())

			return errors.Wrap(err, "failed to shutdown runner")
		}

		err = <-exit // wait for exit
		if err != nil {
			return errors.Wrap(err, "failed to serve")
		}
	}

	log.Infof("bye")

	return nil
}
