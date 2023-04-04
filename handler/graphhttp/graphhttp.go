package graphhttp

import (
	"net/http"

	"private/rat/graph"
	"private/rat/handler/graphhttp/nodeshttp"
	"private/rat/handler/shared"

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
		func(d int, n *graph.Node) bool {
			paths = append(paths, string(n.Path))

			return true
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
