package router

import (
	"private/rat/config"
	"private/rat/errors"
	"private/rat/handler/edithttp"
	"private/rat/handler/graphhttp"

	"github.com/gin-gonic/gin"
)

// New creates a new router, loads templates and registers handlers for routes.
func New(conf *config.Config) (*gin.Engine, error) {
	router := gin.Default()
	// router.LoadHTMLFiles("./templates/index.html", "./templates/edit.html")

	router.LoadHTMLFiles("./statics/index.html")
	router.Static("/statics", "./statics")

	err := graphhttp.RegisterRoutes(conf, &router.RouterGroup)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register routes")
	}

	err = edithttp.RegisterRoutes(&router.RouterGroup)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register routes")
	}

	return router, nil
}
