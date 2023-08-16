package nodeshttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"rat/graph"
	"rat/handler/shared"

	pathutil "rat/graph/util/path"

	"github.com/op/go-logging"
	"github.com/pkg/errors"

	"github.com/gorilla/mux"
)

var log = logging.MustGetLogger("nodeshttp")

type handler struct {
	ss *shared.Services
}

// creates a new Handler.
func newHandler(ss *shared.Services) (*handler, error) {
	h := &handler{
		ss: ss,
	}

	log.Info("loading graph")

	r, err := h.ss.Graph.Root()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get root")
	}

	log.Info("reading metrics")

	m, err := r.Metrics(h.ss.Graph)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get metrics")
	}

	b, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal metrics")
	}

	log.Infof("metrics:\n%s", string(b))
	log.Notice("loaded graph")

	return h, nil
}

// RegisterRoutes registers graph routes on given router.
func RegisterRoutes(
	router *mux.Router, ss *shared.Services,
) error {
	h, err := newHandler(ss)
	if err != nil {
		return errors.Wrap(err, "failed create new graph handler")
	}

	nodesRouter := router.PathPrefix("/nodes").Subrouter()

	nodesRouter.HandleFunc("/", shared.Wrap(h.read)).Methods(http.MethodGet)
	nodesRouter.HandleFunc("/", shared.Wrap(h.create)).Methods(http.MethodPost)

	pathRe := regexp.MustCompile(
		`[[:alnum:]]+(?:-(?:[[:alnum:]]+))*(?:\/[[:alnum:]]+(?:-(?:[[:alnum:]]+))*)*`, //nolint:lll
	)

	nodeRouter := nodesRouter.
		PathPrefix(fmt.Sprintf("/{path:%s}", pathRe.String())).
		Subrouter()

	nodeRouter.HandleFunc("/", shared.Wrap(h.read)).Methods(http.MethodGet)
	nodeRouter.HandleFunc("/", shared.Wrap(h.create)).Methods(http.MethodPost)

	return nil
}

// -------------------------------------------------------------------------- //
// CREATE
// -------------------------------------------------------------------------- //

func (h *handler) create(w http.ResponseWriter, r *http.Request) error {
	body, err := shared.Body[struct {
		Name string `json:"name" binding:"required"`
	}](w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get body")
	}

	n, err := h.getNode(w, r)
	if err != nil {
		shared.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to get node",
		)

		return errors.Wrap(err, "failed to get node error")
	}

	_, err = n.AddLeaf(h.ss.Graph, body.Name)
	if err != nil {
		shared.WriteError(
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
	resp := struct {
		graph.Node
		Leafs []pathutil.NodePath `json:"leafs,omitempty"`
	}{}

	n, err := h.getNode(w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get node")
	}

	resp.Node = *n
	resp.Node.Content = h.ss.Renderer.Render(n)

	if includeLeafs(r) {
		leafs, err := h.getLeafPaths(w, n.Path)
		if err != nil {
			return errors.Wrap(err, "failed to get leaf paths")
		}

		resp.Leafs = leafs
	}

	err = shared.WriteResponse(w, http.StatusOK, resp)
	if err != nil {
		return errors.Wrap(err, "failed to write response")
	}

	return nil
}

func (h *handler) getLeafPaths(
	w http.ResponseWriter,
	path pathutil.NodePath,
) ([]pathutil.NodePath, error) {
	leafNodes, err := h.ss.Graph.GetLeafs(path)
	if err != nil {
		shared.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to get leafs",
		)

		return nil, errors.Wrap(err, "failed to get leafs")
	}

	leafs := make([]pathutil.NodePath, 0, len(leafNodes))

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
		n, err = h.ss.Graph.Root()
		if err != nil {
			shared.WriteError(
				w,
				http.StatusInternalServerError,
				"failed to get node",
			)

			return nil, errors.Wrap(err, "failed to write error")
		}
	} else {
		n, err = h.ss.Graph.GetByPath(pathutil.NodePath(path))
		if err != nil {
			shared.WriteError(
				w,
				http.StatusInternalServerError,
				"failed to get node",
			)

			return nil, errors.Wrap(err, "failed to write error")
		}
	}

	return n, nil
}
