package graphhttp

import (
	"path/filepath"
	"strings"

	"private/rat/config"
	"private/rat/errors"
	"private/rat/graph"
	"private/rat/handler"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	graphName string
	graph     *graph.Graph
}

// creates a new Handler
func new(conf *config.Config) (*Handler, error) {
	graph, err := graph.Init(conf.Graph.Name, conf.Graph.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init graph")
	}

	return &Handler{
			graphName: conf.Graph.Name,
			graph:     graph,
		},
		nil
}

func RegisterRoutes(conf *config.Config, router gin.RouterGroup) error {
	h, err := new(conf)
	if err != nil {
		return errors.Wrap(err, "failed create new graph handler")
	}

	subroute := router.Group(conf.Graph.Name)

	subroute.GET("/*path", handler.Wrap(h.read))

	return nil
}

func (h *Handler) read(c *gin.Context) error {
	path := filepath.Join(h.graphName, strings.TrimPrefix(c.Param("path"), "/"))

	n, err := h.graph.Get(path)
	if err != nil {
		return errors.Wrap(err, "failed to get node")
	}

	c.JSON(
		200,
		gin.H{
			"path":    path,
			"content": n.Content(),
		},
	)

	return nil
}
