package nodeshttp

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/render"
	"rat/graph/render/jsonast"
	"rat/graph/services"
	pathutil "rat/graph/util/path"
	"rat/handler/httputil"
	"rat/logr"
)

type handler struct {
	log *logr.LogR
	gs  *services.GraphServices
	r   jsonast.Renderer
}

type response struct {
	ID         uuid.UUID         `json:"id"`
	Name       string            `json:"name"`
	Path       pathutil.NodePath `json:"path"`
	Length     int               `json:"length"`
	AST        *jsonast.AstPart  `json:"ast"`
	ChildNodes []*response       `json:"childNodes,omitempty"`
}

// RegisterRoutes registers graph routes on given router.
func RegisterRoutes(
	router *mux.Router, log *logr.LogR, gs *services.GraphServices,
) error {
	h := &handler{
		log: log,
		gs:  gs,
		r:   render.NewJSONRenderer(log, gs.Graph),
	}

	nodesRouter := router.PathPrefix("/nodes").Subrouter()

	pathRe := regexp.MustCompile(
		`[[:alnum:]]+(?:-(?:[[:alnum:]]+))*(?:\/[[:alnum:]]+(?:-(?:[[:alnum:]]+))*)*`, //nolint:lll // can't split.
	)

	nodeRouter := nodesRouter.
		PathPrefix(fmt.Sprintf("/{path:%s}", pathRe.String())).
		Subrouter()

	nodeRouter.HandleFunc("/", httputil.Wrap(h.log, h.read)).
		Methods(http.MethodGet)

	nodeRouter.HandleFunc("/", httputil.Wrap(h.log, h.create)).
		Methods(http.MethodPost)

	nodeRouter.HandleFunc("", httputil.Wrap(h.log, h.delete)).
		Methods(http.MethodDelete, http.MethodOptions)

	return nil
}

func (h *handler) read(w http.ResponseWriter, r *http.Request) error {
	n, err := h.getNode(w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get node error")
	}

	root := jsonast.NewRootAstPart("document")

	err = h.r.Render(root, n)
	if err != nil {
		httputil.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to render node content to JSON",
		)

		return errors.Wrap(err, "failed to render node content to JSON")
	}

	childNodes, err := h.getChildNodes(w, n)
	if err != nil {
		return errors.Wrap(err, "failed to get child node paths")
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	err = httputil.WriteResponse(
		w,
		http.StatusOK,
		response{
			ID:         n.Header.ID,
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
	body, err := httputil.Body[struct {
		Name string `json:"name" binding:"required"`
	}](w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get body")
	}

	n, err := h.getNode(w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get node error")
	}

	child, err := n.AddSub(h.gs.Graph, body.Name)
	if err != nil {
		httputil.WriteError(
			w, http.StatusInternalServerError, "failed to create node",
		)

		return errors.Wrap(err, "failed to create node")
	}

	root := jsonast.NewRootAstPart("document")

	err = h.r.Render(root, child)
	if err != nil {
		httputil.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to render node content to JSON",
		)

		return errors.Wrap(err, "failed to render node content to JSON")
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	err = httputil.WriteResponse(
		w,
		http.StatusOK,
		response{
			ID:     child.Header.ID,
			Name:   child.Name,
			Path:   child.Path,
			Length: len(strings.Split(n.Content, "\n")),
			AST:    root,
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to write response")
	}

	return nil
}

func (h *handler) delete(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE")

	if r.Method == http.MethodOptions {
		return nil
	}

	n, err := h.getNode(w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get node error")
	}

	err = h.gs.Graph.Delete(n)
	if err != nil {
		httputil.WriteError(
			w, http.StatusInternalServerError, "failed to delete node",
		)

		return errors.Wrap(err, "failed to delete node")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (h *handler) getChildNodes(
	w http.ResponseWriter,
	n *graph.Node,
) ([]*response, error) {
	children, err := n.GetLeafs(h.gs.Graph)
	if err != nil {
		httputil.WriteError(
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
			httputil.WriteError(
				w,
				http.StatusInternalServerError,
				"failed to render child node",
			)

			return nil, errors.Wrap(err, "failed to render child node")
		}

		childNodes = append(
			childNodes,
			&response{
				ID:     child.Header.ID,
				Name:   child.Name,
				Path:   child.Path,
				Length: len(strings.Split(child.Content, "\n")),
				AST:    root,
			},
		)
	}

	return childNodes, nil
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
		n, err = h.gs.Graph.Root()
		if err != nil {
			httputil.WriteError(
				w,
				http.StatusInternalServerError,
				"failed to get node",
			)

			return nil, errors.Wrap(err, "failed to write error")
		}
	} else {
		n, err = h.gs.Graph.GetByPath(pathutil.NodePath(path))
		if err != nil {
			httputil.WriteError(
				w,
				http.StatusInternalServerError,
				"failed to get node",
			)

			return nil, errors.Wrap(err, "failed to write error")
		}
	}

	return n, nil
}