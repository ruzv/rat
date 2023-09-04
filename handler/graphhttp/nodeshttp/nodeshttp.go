package nodeshttp

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"rat/graph"
	"rat/graph/render"
	"rat/graph/render/jsonast"
	pathutil "rat/graph/util/path"
	"rat/handler/shared"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var log = logging.MustGetLogger("nodeshttp")

type handler struct {
	ss *shared.Services
	r  jsonast.Renderer
}

// RegisterRoutes registers graph routes on given router.
func RegisterRoutes(
	router *mux.Router, ss *shared.Services,
) error {
	h := &handler{
		ss: ss,
		r:  render.NewJSONRenderer(ss.Graph),
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

	nodeRouter.HandleFunc("/", shared.Wrap(h.deconstruct)).
		Methods(http.MethodGet)
	nodeRouter.HandleFunc("/", shared.Wrap(h.create)).Methods(http.MethodPost)

	return nil
}

type response struct {
	ID         uuid.UUID         `json:"id"`
	Name       string            `json:"name"`
	Path       pathutil.NodePath `json:"path"`
	Length     int               `json:"length"`
	AST        *jsonast.AstPart  `json:"ast"`
	ChildNodes []*response       `json:"childNodes,omitempty"`
}

func (h *handler) deconstruct(w http.ResponseWriter, r *http.Request) error {
	n, err := h.getNode(w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get node error")
	}

	root := jsonast.NewRootAstPart("document")

	err = h.r.Render(root, n)
	if err != nil {
		shared.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to render node content to JSON",
		)

		return errors.Wrap(err, "failed to render node content to JSON")
	}

	childNodes, err := h.getChildNodes(w, n.Path)
	if err != nil {
		return errors.Wrap(err, "failed to get child node paths")
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	err = shared.WriteResponse(
		w,
		http.StatusOK,
		response{
			ID:         n.ID,
			Name:       n.Name,
			Path:       n.Path,
			Length:     len(strings.Split(n.Content, "\n")),
			ChildNodes: childNodes,
			AST:        root,
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to write response")
	}

	return nil
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) error {
	body, err := shared.Body[struct {
		Name string `json:"name" binding:"required"`
	}](w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get body")
	}

	n, err := h.getNode(w, r)
	if err != nil {
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

	// if includeLeafs(r) {
	// 	leafs, err := h.getChildNodes(w, n.Path)
	// 	if err != nil {
	// 		return errors.Wrap(err, "failed to get leaf paths")
	// 	}
	//
	// 	resp.Leafs = leafs
	// }

	err = shared.WriteResponse(w, http.StatusOK, resp)
	if err != nil {
		return errors.Wrap(err, "failed to write response")
	}

	return nil
}

func (h *handler) getChildNodes(
	w http.ResponseWriter,
	path pathutil.NodePath,
) ([]*response, error) {
	children, err := h.ss.Graph.GetLeafs(path)
	if err != nil {
		shared.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to get child nodes",
		)

		return nil, errors.Wrap(err, "failed to get leafs")
	}

	childNodes := make([]*response, 0, len(children))

	for _, child := range children {
		root := jsonast.NewRootAstPart("document")
		err := h.r.Render(root, child)
		if err != nil {
			shared.WriteError(
				w,
				http.StatusInternalServerError,
				"failed to render child node",
			)

			return nil, errors.Wrap(err, "failed to render child node")
		}

		childNodes = append(
			childNodes,
			&response{
				ID:     child.ID,
				Name:   child.Name,
				Path:   child.Path,
				Length: len(strings.Split(child.Content, "\n")),
				AST:    root,
			},
		)
	}

	return childNodes, nil
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
