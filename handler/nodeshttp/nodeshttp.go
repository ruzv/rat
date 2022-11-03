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
	hUtil "private/rat/handler"

	"github.com/op/go-logging"
	"github.com/pkg/errors"

	"github.com/gorilla/mux"
)

var log = logging.MustGetLogger("nodeshttp")

type handler struct {
	store graph.Store
	rend  *render.NodeRender
}

// creates a new Handler.
func newHandler(conf *config.Config) (*handler, error) {
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

	return &handler{
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

	nodesRouter.HandleFunc("/", hUtil.Wrap(h.read)).Methods(http.MethodGet)
	nodesRouter.HandleFunc("/", hUtil.Wrap(h.create)).Methods(http.MethodPost)

	pathRe := regexp.MustCompile(
		`[[:alnum:]]+(?:-(?:[[:alnum:]]+))*(?:\/[[:alnum:]]+(?:-(?:[[:alnum:]]+))*)*`, //nolint:lll
	)

	nodeRouter := nodesRouter.
		PathPrefix(fmt.Sprintf("/{path:%s}", pathRe.String())).
		Subrouter()

	nodeRouter.HandleFunc("/", hUtil.Wrap(h.read)).Methods(http.MethodGet)
	nodeRouter.HandleFunc("/", hUtil.Wrap(h.create)).Methods(http.MethodPost)

	return nil
}

// -------------------------------------------------------------------------- //
// CREATE
// -------------------------------------------------------------------------- //

func (h *handler) create(w http.ResponseWriter, r *http.Request) error {
	body, err := hUtil.Body[struct {
		Name string `json:"name" binding:"required"`
	}](w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get body")
	}

	n, err := h.getNode(w, r)
	if err != nil {
		hUtil.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to get node",
		)

		return errors.Wrap(err, "failed to get node error")
	}

	_, err = n.Add(body.Name)
	if err != nil {
		hUtil.WriteError(
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

func (h *handler) read(w http.ResponseWriter, r *http.Request) error {
	n, err := h.getNode(w, r)
	if err != nil {
		hUtil.WriteError(
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

	err = hUtil.WriteResponse(w, http.StatusOK, f)
	if err != nil {
		return errors.Wrap(err, "failed to write response")
	}

	return nil
}

func getLeafPaths(w http.ResponseWriter, n *graph.Node) ([]string, error) {
	leafNodes, err := n.Leafs()
	if err != nil {
		hUtil.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to get leafs",
		)

		return nil, errors.Wrap(err, "failed to get leafs")
	}

	leafs := make([]string, 0, len(leafNodes))

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

func (h *handler) getNode(
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
			hUtil.WriteError(
				w,
				http.StatusInternalServerError,
				"failed to get node",
			)

			return nil, errors.Wrap(err, "failed to write error")
		}
	} else {
		n, err = h.store.GetByPath(path)
		if err != nil {
			hUtil.WriteError(
				w,
				http.StatusInternalServerError,
				"failed to get node",
			)

			return nil, errors.Wrap(err, "failed to write error")
		}
	}

	return n, nil
}

func (h *handler) format(
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
		hUtil.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to encode format",
		)

		return nil, errors.Wrap(err, "failed to encode format")
	}

	res := make(map[string]interface{})

	err = json.NewDecoder(buff).Decode(&res)
	if err != nil {
		hUtil.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to decode format",
		)

		return nil, errors.Wrap(err, "failed to decode format")
	}

	return res, nil
}
