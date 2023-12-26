package graphhttp

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/services"
	"rat/graph/util"
	pathutil "rat/graph/util/path"
	"rat/handler/fileshttp"
	"rat/handler/graphhttp/nodeshttp"
	"rat/handler/httputil"
	"rat/logr"
)

type handler struct {
	log *logr.LogR

	gs *services.GraphServices
}

// RegisterRoutes registers graph routes on given router.
func RegisterRoutes(
	router *mux.Router, log *logr.LogR, gs *services.GraphServices,
) error {
	h := &handler{
		log: log.Prefix("graphhttp"),
		gs:  gs,
	}

	graphRouter := router.PathPrefix("/graph").Subrouter()

	graphRouter.HandleFunc(
		"/search",
		httputil.Wrap(
			httputil.WrapOptions(
				h.search,
				[]string{http.MethodPost},
				[]string{"Content-Type"},
			),
			h.log,
			"search",
		),
	).Methods(http.MethodPost, http.MethodOptions)

	graphRouter.HandleFunc(
		"/move/{id:.+}",
		httputil.Wrap(
			httputil.WrapOptions(
				h.move,
				[]string{http.MethodPost},
				[]string{"Content-Type"},
			),
			h.log,
			"move",
		),
	).Methods(http.MethodPost, http.MethodOptions)

	err := nodeshttp.RegisterRoutes(graphRouter, h.log, gs)
	if err != nil {
		return errors.Wrap(err, "failed to register nodes routes")
	}

	err = fileshttp.RegisterRoutes(graphRouter, h.log, gs)
	if err != nil {
		return errors.Wrap(err, "failed to register fileshttp routes")
	}

	return nil
}

func (h *handler) search(w http.ResponseWriter, r *http.Request) error {
	body, err := httputil.Body[struct {
		Query string `json:"query"`
	}](w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get body")
	}

	res, err := h.gs.Index.Search(body.Query)
	if err != nil {
		httputil.WriteError(
			w, http.StatusInternalServerError, "failed to search",
		)

		return errors.Wrap(err, "failed to search index")
	}

	err = httputil.WriteResponse(
		w,
		http.StatusOK,
		struct {
			Results []string `json:"results"`
		}{
			Results: util.Map(
				res,
				func(n *graph.Node) string {
					return string(n.Path)
				},
			),
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to write response")
	}

	return nil
}

func (h *handler) move(w http.ResponseWriter, r *http.Request) error {
	id, err := uuid.FromString(mux.Vars(r)["id"])
	if err != nil {
		return httputil.Error(
			http.StatusBadRequest,
			errors.Wrap(err, "failed to parse node id"),
		)
	}

	body, err := httputil.Body[struct {
		NewPath string `json:"newPath" binding:"required"`
	}](w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get body")
	}

	err = h.gs.Provider.Move(id, pathutil.NodePath(body.NewPath))
	if err != nil {
		httputil.WriteError(
			w, http.StatusInternalServerError, "failed to move node",
		)

		return errors.Wrap(err, "failed to move node")
	}

	n, err := h.gs.Provider.GetByID(id)
	if err != nil {
		httputil.WriteError(
			w, http.StatusInternalServerError, "failed to get node",
		)
	}

	err = httputil.WriteResponse(w, http.StatusOK, n)
	if err != nil {
		return errors.Wrap(err, "failed to write response")
	}

	return nil
}
