package main

import (
	"fmt"

	"private/rat/config"
	"private/rat/errors"
	"private/rat/handler/graphhttp"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
)

var (
	configPath = pflag.StringP(
		"config", "c", "./config.json", "path to config file",
	)
	help = pflag.BoolP("help", "h", false, "show help")
)

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}

func run() error {
	pflag.Parse()

	if *help {
		pflag.PrintDefaults()

		return nil
	}

	conf, err := config.Load(*configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}

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

	return nil
}
