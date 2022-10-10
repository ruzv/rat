package nodeshttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"private/rat/config"

	"private/rat/graph"
	"private/rat/graph/render"
	"private/rat/graph/storefilesystem"
	"private/rat/handler"

	"github.com/op/go-logging"
	"github.com/pkg/errors"

	"github.com/gorilla/mux"
)

var log = logging.MustGetLogger("nodeshttp")

type Handler struct {
	store graph.Store
	rend  *render.NodeRender
}

// creates a new Handler.
func newHandler(conf *config.Config) (*Handler, error) {
	log.Info("loading graph")

	store, err := storefilesystem.NewFileSystem(
		conf.Graph.Name,
		conf.Graph.Path,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create net fs")
	}

	r, err := store.Root()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get root")
	}

	m, err := r.Metrics()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get metrics")
	}

	log.Info("nodes -", m.Nodes)
	log.Info("max depth -", m.MaxDepth)

	log.Notice("loaded graph -", conf.Graph.Name)

	return &Handler{
		store: store,
		rend:  render.NewNodeRender(),
	}, nil
}

// RegisterRoutes registers graph routes on given router.
func RegisterRoutes(conf *config.Config, router *mux.Router) error {
	h, err := newHandler(conf)
	if err != nil {
		return errors.Wrap(err, "failed create new graph handler")
	}

	nodesRouter := router.PathPrefix("/nodes").Subrouter()

	nodesRouter.HandleFunc("/", handler.Wrap(h.read)).Methods(http.MethodGet)
	nodesRouter.HandleFunc("/", handler.Wrap(h.create)).Methods(http.MethodPost)

	pathRe := regexp.MustCompile(
		`[[:alnum:]]+(?:-(?:[[:alnum:]]+))*(?:\/[[:alnum:]]+(?:-(?:[[:alnum:]]+))*)*`, //nolint:lll
	)

	nodeRouter := nodesRouter.
		PathPrefix(fmt.Sprintf("/{path:%s}", pathRe.String())).
		Subrouter()

	nodeRouter.HandleFunc("/", handler.Wrap(h.read)).Methods(http.MethodGet)
	nodeRouter.HandleFunc("/", handler.Wrap(h.create)).Methods(http.MethodPost)

	return nil
}

// func (h *Handler) move(c *gin.Context) error {
// 	body, err := handler.Body[struct {
// 		Src  string `json:"src" binding:"required"`
// 		Dest string `json:"dest" binding:"required"`
// 	}](c)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to get body")
// 	}

// 	src, err := h.store.GetByPath(body.Src)
// 	if err != nil {
// 		handler.WriteJSON(
// 			c,
// 			http.StatusInternalServerError,
// 			"failed to get src node",
// 		)

// 		return errors.Wrap(err, "failed to get src node")
// 	}

// 	err = src.Move(body.Dest)
// 	if err != nil {
// 		handler.WriteJSON(
// 			c,
// 			http.StatusInternalServerError,
// 			"failed to move to dest",
// 		)

// 		return errors.Wrap(err, "failed to move to dest")
// 	}

// 	c.Status(http.StatusNoContent)

// 	return nil
// }

// -------------------------------------------------------------------------- //
// CREATE
// -------------------------------------------------------------------------- //

func (h *Handler) create(w http.ResponseWriter, r *http.Request) error {
	body, err := handler.Body[struct {
		Name string `json:"name" binding:"required"`
	}](w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get body")
	}

	n, err := h.getNode(w, r)
	if err != nil {
		handler.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to get node",
		)

		return errors.Wrap(err, "failed to get node error")
	}

	_, err = n.Add(body.Name)
	if err != nil {
		handler.WriteError(
			w, http.StatusInternalServerError, "failed to create node",
		)

		return errors.Wrap(err, "failed to create node")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

// -------------------------------------------------------------------------- //
// READ
// -------------------------------------------------------------------------- //

func (h *Handler) read(w http.ResponseWriter, r *http.Request) error {
	n, err := h.getNode(w, r)
	if err != nil {
		handler.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to get node",
		)

		return errors.Wrap(err, "failed to get node")
	}

	f, err := h.format(w, r, n)
	if err != nil {
		return errors.Wrap(err, "failed to format")
	}

	if includeLeafs(r) {
		leafs, err := getLeafPaths(w, n)
		if err != nil {
			return errors.Wrap(err, "failed to get leaf paths")
		}

		f["leafs"] = leafs
	}

	handler.WriteResponse(w, http.StatusOK, f)
	if err != nil {
		return errors.Wrap(err, "failed to write response")
	}

	return nil
}

func getLeafPaths(w http.ResponseWriter, n *graph.Node) ([]string, error) {
	leafNodes, err := n.Leafs()
	if err != nil {
		handler.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to get leafs",
		)

		return nil, errors.Wrap(err, "failed to get leafs")
	}

	var leafs []string

	for _, lf := range leafNodes {
		leafs = append(leafs, lf.Path)
	}

	return leafs, nil
}

func includeLeafs(r *http.Request) bool {
	leafsParam := r.URL.Query().Get("leafs")
	if leafsParam == "" {
		return false
	}

	l, err := strconv.ParseBool(leafsParam)
	if err != nil {
		log.Debug("failed to parse leafs param", err)
		return false
	}

	return l
}

func (h *Handler) getNode(
	w http.ResponseWriter,
	r *http.Request,
) (*graph.Node, error) {
	path := mux.Vars(r)["path"]

	var (
		n   *graph.Node
		err error
	)

	if path == "" {
		n, err = h.store.Root()
		if err != nil {
			handler.WriteError(
				w,
				http.StatusInternalServerError,
				"failed to get node",
			)

			return nil, errors.Wrap(err, "failed to write error")
		}
	} else {
		n, err = h.store.GetByPath(path)
		if err != nil {
			handler.WriteError(
				w,
				http.StatusInternalServerError,
				"failed to get node",
			)

			return nil, errors.Wrap(err, "failed to write error")
		}
	}

	return n, nil
}

func (h *Handler) format(
	w http.ResponseWriter, r *http.Request, n *graph.Node,
) (map[string]interface{}, error) {
	format := r.URL.Query().Get("format")

	cn := *n

	switch format {
	case "html":
		cn.Content = h.rend.HTML(n)
	case "md":
		cn.Content = h.rend.Markdown(n)
	default:

	}

	buff := &bytes.Buffer{}

	err := json.NewEncoder(buff).Encode(cn)
	if err != nil {
		handler.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to encode format",
		)

		return nil, errors.Wrap(err, "failed to encode format")
	}

	var res map[string]interface{} = make(map[string]interface{})

	err = json.NewDecoder(buff).Decode(&res)
	if err != nil {
		handler.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to decode format",
		)

		return nil, errors.Wrap(err, "failed to decode format")
	}

	return res, nil
}

// func writeFormat(w http.ResponseWriter, r *http.Request, n *graph.Node) error
// {
// 	var f interface{}

// 	format := r.URL.Query().Get("format")

// 	switch format {
// 	case "html":
// 		f = n.HTML()
// 	case "md":
// 		f = n.Markdown()
// 	default:
// 		f = n
// 	}

// 	err = handler.WriteResponse(w, http.StatusOK, f)
// 	if err != nil {
// 		handler.WriteError(
// 			w,
// 			http.StatusInternalServerError,
// 			"failed to write response",
// 		)

// 		return errors.Wrap(err, "failed to write response")
// 	}

// 	return nil
// }

// -------------------------------------------------------------------------- //
// UPDATE
// -------------------------------------------------------------------------- //

// func (h *Handler) update(c *gin.Context) error {
// 	body, err := handler.Body[struct {
// 		Name    string `json:"name"`
// 		Content string `json:"content"`
// 		Clear   bool   `json:"clear"`
// 	}](c)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to get body")
// 	}

// 	n, err := h.getNode(c)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to get node")
// 	}

// 	if body.Name != "" {
// 		err = n.Rename(body.Name)
// 		if err != nil {
// 			handler.WriteJSON(
// 				c,
// 				http.StatusInternalServerError,
// 				"failed to update node name",
// 			)

// 			return errors.Wrap(err, "failed to update node name")
// 		}
// 	}

// 	if body.Content != "" {
// 		n.Content = body.Content

// 		err = n.Update()
// 		if err != nil {
// 			handler.WriteJSON(
// 				c,
// 				http.StatusInternalServerError,
// 				"failed to update node content",
// 			)

// 			return errors.Wrap(err, "failed to update node content")
// 		}
// 	}

// 	writeNode(c, n)

// 	return nil
// }

// -------------------------------------------------------------------------- //
// DELETE
// -------------------------------------------------------------------------- //

// func (h *Handler) delete(c *gin.Context) error {
// 	n, err := h.getNode(c)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to get node")
// 	}

// 	err = n.DeleteSingle()
// 	if err != nil {
// 		handler.WriteJSON(
// 			c,
// 			http.StatusInternalServerError,
// 			"failed to delete node",
// 		)

// 		return errors.Wrapf(err, "failed to delete node")
// 	}

// 	return nil
// }

// // getNode reads the node specified by path route param. on error writes JSON
// // response.
// func (h *Handler) getNode(c *gin.Context) (*graph.Node, error) {
// 	n, err := h.store.GetByPath(getPath(c))
// 	if err != nil {
// 		handler.WriteJSON(
// 			c,
// 			http.StatusInternalServerError,
// 			"failed to get node",
// 		)

// 		return nil, errors.Wrap(err, "failed to get node")
// 	}

// 	return n, nil
// }

// func writeNode(c *gin.Context, n *graph.Node) {
// 	// ID      uuid.UUID
// 	// Name    string
// 	// Path    string
// 	// Content string

// 	c.JSON(
// 		http.StatusOK,
// 		gin.H{
// 			"id":       n.ID.String(),
// 			"name":     n.Name,
// 			"path":     n.Path,
// 			"raw":      n.Content,
// 			"markdown": n.Markdown(),
// 			"html":     n.HTML(),
// 		},
// 	)
// }
