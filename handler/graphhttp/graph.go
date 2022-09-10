package graphhttp

import (
	"net/http"
	"strings"

	"private/rat/config"
	"private/rat/errors"
	"private/rat/graph"
	"private/rat/graph/storefilesystem"
	"private/rat/handler"
	"private/rat/logger"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	graphName string
	// graph     *graph.Graph // remove
	store graph.Store
}

// creates a new Handler.
func newHandler(conf *config.Config) (*Handler, error) {
	store, err := storefilesystem.NewFileSystem(
		conf.Graph.Name,
		conf.Graph.Path,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create net fs")
	}

	logger.Debugf("loaded graph:\n%s", conf.Graph.Name)

	return &Handler{
			graphName: conf.Graph.Name,
			store:     store,
		},
		nil
}

// RegisterRoutes registers graph routes on given router.
func RegisterRoutes(conf *config.Config, router gin.RouterGroup) error {
	h, err := newHandler(conf)
	if err != nil {
		return errors.Wrap(err, "failed create new graph handler")
	}

	graphsRoute := router.Group("/graphs")

	router.POST("/move", handler.Wrap(h.move))

	nodesRoute := graphsRoute.Group("/*path")

	nodesRoute.POST("", handler.Wrap(h.create))
	nodesRoute.GET("", handler.Wrap(h.read))
	nodesRoute.PUT("", handler.Wrap(h.update))
	nodesRoute.DELETE("", handler.Wrap(h.delete))

	return nil
}

func (h *Handler) move(c *gin.Context) error {
	body, err := handler.Body[struct {
		Src  string `json:"src" binding:"required"`
		Dest string `json:"dest" binding:"required"`
	}](c)
	if err != nil {
		return errors.Wrap(err, "failed to get body")
	}

	src, err := h.store.GetByPath(body.Src)
	if err != nil {
		handler.WriteJSON(
			c,
			http.StatusInternalServerError,
			"failed to get src node",
		)

		return errors.Wrap(err, "failed to get src node")
	}

	err = src.Move(body.Dest)
	if err != nil {
		handler.WriteJSON(
			c,
			http.StatusInternalServerError,
			"failed to move to dest",
		)

		return errors.Wrap(err, "failed to move to dest")
	}

	c.Status(http.StatusNoContent)

	return nil
}

// -------------------------------------------------------------------------- //
// CREATE
// -------------------------------------------------------------------------- //

func (h *Handler) create(c *gin.Context) error {
	body, err := handler.Body[struct {
		Name string `json:"name" binding:"required"`
	}](c)
	if err != nil {
		return errors.Wrap(err, "failed to get body")
	}

	parent, err := h.getNode(c)
	if err != nil {
		return errors.Wrap(err, "failed to get node")
	}

	n, err := parent.Add(body.Name)
	if err != nil {
		handler.WriteJSON(
			c,
			http.StatusInternalServerError,
			"failed to create node",
		)

		return errors.Wrap(err, "failed to create node")
	}

	c.JSON(http.StatusOK, gin.H{"path": n.Path})

	return nil
}

// -------------------------------------------------------------------------- //
// READ
// -------------------------------------------------------------------------- //

func (h *Handler) read(c *gin.Context) error {
	n, err := h.getNode(c)
	if err != nil {
		return errors.Wrap(err, "failed to get node")
	}

	c.HTML(
		http.StatusOK,
		"index.html",
		gin.H{
			"html":     n.HTML(),
			"markdown": n.Markdown(),
			"raw":      n.Content,
		},
	)

	return nil
}

// -------------------------------------------------------------------------- //
// UPDATE
// -------------------------------------------------------------------------- //

func (h *Handler) update(c *gin.Context) error {
	body, err := handler.Body[struct {
		Name    string `json:"name"`
		Content string `json:"content"`
		Clear   bool   `json:"clear"`
	}](c)
	if err != nil {
		return errors.Wrap(err, "failed to get body")
	}

	n, err := h.getNode(c)
	if err != nil {
		return errors.Wrap(err, "failed to get node")
	}

	if body.Name != "" {
		err = n.Rename(body.Name)
		if err != nil {
			handler.WriteJSON(
				c,
				http.StatusInternalServerError,
				"failed to update node name",
			)

			return errors.Wrap(err, "failed to update node name")
		}
	}

	if body.Content != "" {
		n.Content = body.Content

		err = n.Update()
		if err != nil {
			handler.WriteJSON(
				c,
				http.StatusInternalServerError,
				"failed to update node content",
			)

			return errors.Wrap(err, "failed to update node content")
		}
	}

	return nil
}

// -------------------------------------------------------------------------- //
// DELETE
// -------------------------------------------------------------------------- //

func (h *Handler) delete(c *gin.Context) error {
	n, err := h.getNode(c)
	if err != nil {
		return errors.Wrap(err, "failed to get node")
	}

	err = n.DeleteSingle()
	if err != nil {
		handler.WriteJSON(
			c,
			http.StatusInternalServerError,
			"failed to delete node",
		)

		return errors.Wrapf(err, "failed to delete node")
	}

	return nil
}

func getPath(c *gin.Context) string {
	return strings.Trim(c.Param("path"), "/")
}

// getNode reads the node specified by path route param. on error writes JSON
// response.
func (h *Handler) getNode(c *gin.Context) (*graph.Node, error) {
	n, err := h.store.GetByPath(getPath(c))
	if err != nil {
		handler.WriteJSON(
			c,
			http.StatusInternalServerError,
			"failed to get node",
		)

		return nil, errors.Wrap(err, "failed to get node")
	}

	return n, nil
}
