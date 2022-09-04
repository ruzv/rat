package main

import (
	"fmt"

	"private/rat/args"
	"private/rat/config"
	"private/rat/errors"
	"private/rat/handler/router"
	"private/rat/logger"
)

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

	err := logger.NewDefault(cmdArgs.LogPath)
	if err != nil {
		return errors.Wrap(err, "failed to create default logger")
	}

	defer logger.Close()

	logger.Infof("rat server starting")

	conf, err := config.Load(cmdArgs.ConfigPath)
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}

	logger.Infof("have config")

	router, err := router.New(conf)
	if err != nil {
		return errors.Wrap(err, "failed to create new router")
	}

	err = router.Run(fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		return errors.Wrap(err, "failed to run router")
	}

	logger.Infof("rat server stopped")

	return nil
}
