package nodeshttp

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/render"
	"rat/graph/render/jsonast"
	"rat/graph/services"
	"rat/graph/services/auth"
	"rat/graph/services/provider/access"
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

func accessWrapper(
	base graph.Provider,
	log *logr.LogR,
	handler func(
		p graph.Provider, w http.ResponseWriter, r *http.Request,
	) error,
) httputil.RatHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		token, ok := r.Context().Value(auth.AuthTokenCtxKey).(*auth.Token)
		if !ok || token == nil {
			return httputil.Error(
				http.StatusInternalServerError,
				errors.New("failed to get auth token from context"),
			)
		}

		b, err := json.MarshalIndent(token, "", "  ")
		if err != nil {
			return httputil.Error(
				http.StatusInternalServerError,
				errors.Wrap(err, "failed to marshal token"),
			)
		}

		log.Debugf("token:\n%s", string(b))

		err = handler(access.NewProvider(base, log, token.Scopes), w, r)
		if err != nil {
			return err
		}

		return nil
	}
}

// RegisterRoutes registers graph routes on given router.
func RegisterRoutes(
	router *mux.Router, log *logr.LogR, gs *services.GraphServices,
) error {
	log = log.Prefix("nodeshttp")

	h := &handler{
		log: log,
		gs:  gs,
		r:   render.NewJSONRenderer(log, gs.Provider),
	}

	nodeRouter := router.PathPrefix("/node/{path:.*}").Subrouter()

	nodeRouter.HandleFunc(
		"",
		httputil.Wrap(
			httputil.WrapOptions(
				func(w http.ResponseWriter, r *http.Request) error {
					return nil
				},
				[]string{http.MethodGet, http.MethodPost, http.MethodDelete},
				[]string{"Content-Type", "Authorization"},
			),
			log, "read"),
	).Methods(http.MethodOptions)

	nodeRouter.HandleFunc(
		"", httputil.Wrap(accessWrapper(gs.Provider, log, h.read), log, "read"),
	).Methods(http.MethodGet)

	nodeRouter.HandleFunc("", httputil.Wrap(h.create, h.log, "create")).
		Methods(http.MethodPost)

	nodeRouter.HandleFunc("", httputil.Wrap(h.delete, h.log, "delete")).
		Methods(http.MethodDelete)

	return nil
}

func (h *handler) read(
	p graph.Provider, w http.ResponseWriter, r *http.Request,
) error {
	n, err := getNode(p, r)
	if err != nil {
		return httputil.Error(
			http.StatusInternalServerError,
			errors.Wrap(err, "failed to get node"),
		)
	}

	root := jsonast.NewRootAstPart("document")

	render.NewJSONRenderer(h.log, p).Render(root, n, n.Content)

	h.log.Debugf("rendered node %q", n.Path)

	childNodes, err := getChildNodes(p, h.log, w, n)
	if err != nil {
		return errors.Wrap(err, "failed to get child node paths")
	}

	h.log.Debugf("got child nodes for node %q", n.Path)

	err = httputil.WriteResponse(
		w,
		http.StatusOK,
		response{
			ID:         n.Header.ID,
			Name:       n.Name(),
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

	n, err := getNode(h.gs.Provider, r)
	if err != nil {
		return httputil.Error(
			http.StatusInternalServerError,
			errors.Wrap(err, "failed to get node"),
		)
	}

	sub, err := n.AddSub(h.gs.Provider, body.Name)
	if err != nil {
		return httputil.Error(
			http.StatusInternalServerError,
			errors.Wrap(err, "failed to create node"),
		)
	}

	root := jsonast.NewRootAstPart("document")

	h.r.Render(root, sub, sub.Content)

	err = httputil.WriteResponse(
		w,
		http.StatusOK,
		response{
			ID:     sub.Header.ID,
			Name:   sub.Name(),
			Path:   sub.Path,
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
	n, err := getNode(h.gs.Provider, r)
	if err != nil {
		return httputil.Error(
			http.StatusInternalServerError,
			errors.Wrap(err, "failed to get node"),
		)
	}

	err = h.gs.Provider.Delete(n)
	if err != nil {
		httputil.WriteError(
			w, http.StatusInternalServerError, "failed to delete node",
		)

		return errors.Wrap(err, "failed to delete node")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func getChildNodes(
	p graph.Provider,
	log *logr.LogR,
	w http.ResponseWriter,
	n *graph.Node,
) ([]*response, error) {
	children, err := n.GetLeafs(p)
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

		render.NewJSONRenderer(log, p).Render(root, child, child.Content)

		childNodes = append(
			childNodes,
			&response{
				ID:     child.Header.ID,
				Name:   child.Name(),
				Path:   child.Path,
				Length: len(strings.Split(child.Content, "\n")),
				AST:    root,
			},
		)
	}

	return childNodes, nil
}

func getNode(p graph.Provider, r *http.Request) (*graph.Node, error) {
	path := mux.Vars(r)["path"]

	n, err := p.GetByPath(pathutil.NodePath(path))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get node %q", path)
	}

	return n, nil
}
