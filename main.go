package main

import (
	"fmt"

	"private/rat/args"
	"private/rat/config"
	"private/rat/errors"
	"private/rat/handler/graphhttp"
	"private/rat/logger"

	"github.com/gin-gonic/gin"
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

	router := gin.Default()
	router.LoadHTMLFiles("./templates/index.html")

	err = graphhttp.RegisterRoutes(conf, router.RouterGroup)
	if err != nil {
		return errors.Wrap(err, "failed to register routes")
	}

	err = router.Run(fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		return errors.Wrap(err, "failed to run router")
	}

	logger.Infof("rat server stopped")

	return nil
}
