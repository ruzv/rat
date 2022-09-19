package edithttp

import (
	"net/http"
	"strings"

	"private/rat/handler"

	"github.com/gin-gonic/gin"
)

// type Handler struct {
// 	graphName string
// 	// graph     *graph.Graph // remove
// 	store graph.Store
// }

// // creates a new Handler.
// func newHandler(conf *config.Config) (*Handler, error) {
// 	store, err := storefilesystem.NewFileSystem(
// 		conf.Graph.Name,
// 		conf.Graph.Path,
// 	)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "failed to create net fs")
// 	}

// 	logger.Debugf("loaded graph:\n%s", conf.Graph.Name)

// 	return &Handler{
// 			graphName: conf.Graph.Name,
// 			store:     store,
// 		},
// 		nil
// }

// RegisterRoutes registers graph routes on given router.
func RegisterRoutes(router *gin.RouterGroup) error {
	// h, err := newHandler(conf)
	// if err != nil {
	// 	return errors.Wrap(err, "failed create new graph handler")
	// }

	editRoute := router.Group("/edit")

	nodeRoute := editRoute.Group("/*path")

	nodeRoute.GET("", handler.Wrap(read))

	return nil
}

// -------------------------------------------------------------------------- //
// READ
// -------------------------------------------------------------------------- //

func read(c *gin.Context) error {
	p := strings.Trim(c.Param("path"), "/")

	c.HTML(
		http.StatusOK,
		"index.html",
		gin.H{
			"path": p,
		},
	)

	return nil
}
