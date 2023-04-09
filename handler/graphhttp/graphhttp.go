package graphhttp

import (
	"fmt"
	"net/http"
	"regexp"

	"private/rat/graph"
	"private/rat/handler/graphhttp/nodeshttp"
	"private/rat/handler/shared"

	pathutil "private/rat/graph/util/path"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type handler struct {
	ss *shared.Services
}

// RegisterRoutes registers graph routes on given router.
func RegisterRoutes(router *mux.Router, ss *shared.Services) error {
	h := &handler{
		ss: ss,
	}

	graphRouter := router.PathPrefix("/graph").Subrouter()

	graphRouter.HandleFunc("/index/", shared.Wrap(h.index)).
		Methods(http.MethodGet)

	idRe := regexp.MustCompile(
		`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
	)

	graphRouter.HandleFunc(
		fmt.Sprintf("/move/{id:%s}", idRe.String()), shared.Wrap(h.move),
	).
		Methods(http.MethodPost)

	err := nodeshttp.RegisterRoutes(graphRouter, ss)
	if err != nil {
		return errors.Wrap(err, "failed to register nodes routes")
	}

	return nil
}

func (h *handler) index(w http.ResponseWriter, _ *http.Request) error {
	root, err := h.ss.Graph.Root()
	if err != nil {
		shared.WriteError(
			w, http.StatusInternalServerError, "failed to get root node",
		)

		return errors.Wrap(err, "failed to get root")
	}

	paths := []string{string(root.Path)}

	err = root.Walk(
		h.ss.Graph,
		func(d int, n *graph.Node) (bool, error) {
			paths = append(paths, string(n.Path))

			return true, nil
		},
	)
	if err != nil {
		shared.WriteError(
			w, http.StatusInternalServerError, "failed to walk graph",
		)

		return errors.Wrap(err, "failed to walk graph")
	}

	err = shared.WriteResponse(w, http.StatusOK, paths)
	if err != nil {
		return errors.Wrap(err, "failed to write response")
	}

	return nil
}

func (h *handler) move(w http.ResponseWriter, r *http.Request) error {
	id, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		shared.WriteError(
			w, http.StatusBadRequest, "invalid node id",
		)

		return errors.Wrap(err, "failed to parse node id")
	}

	body, err := shared.Body[struct {
		NewPath string `json:"new_path" binding:"required"`
	}](w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get body")
	}

	err = h.ss.Graph.Move(id, pathutil.NodePath(body.NewPath))
	if err != nil {
		shared.WriteError(
			w, http.StatusInternalServerError, "failed to move node",
		)

		return errors.Wrap(err, "failed to move node")
	}

	n, err := h.ss.Graph.GetByID(id)
	if err != nil {
		shared.WriteError(
			w, http.StatusInternalServerError, "failed to get node",
		)
	}

	err = shared.WriteResponse(w, http.StatusOK, n)
	if err != nil {
		return errors.Wrap(err, "failed to write response")
	}

	return nil
}
