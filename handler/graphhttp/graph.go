package graphhttp

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"private/rat/config"
	"private/rat/errors"
	"private/rat/graph"
	"private/rat/handler"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
)

type Handler struct {
	graphName string
	graph     *graph.Graph
}

// creates a new Handler.
func newHandler(conf *config.Config) (*Handler, error) {
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

// RegisterRoutes registers graph routes on given router.
func RegisterRoutes(conf *config.Config, router gin.RouterGroup) error {
	h, err := newHandler(conf)
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

	out := markdown.ToHTML([]byte(n.Content()), nil, nil)

	//nolint:gosec
	c.HTML(http.StatusOK, "index.html", gin.H{"content": template.HTML(out)})

	return nil
}
