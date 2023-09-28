package graphhttp

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"rat/graph"
	"rat/graph/services"
	"rat/graph/util"
	pathutil "rat/graph/util/path"
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

	graphRouter.HandleFunc("/search/", httputil.Wrap(h.log, h.search)).
		Methods(http.MethodPost)

	idRe := regexp.MustCompile(
		`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
	)

	graphRouter.HandleFunc(fmt.Sprintf(
		"/move/{id:%s}", idRe.String()), httputil.Wrap(h.log, h.move),
	).
		Methods(http.MethodPost)

	err := nodeshttp.RegisterRoutes(graphRouter, h.log, gs)
	if err != nil {
		return errors.Wrap(err, "failed to register nodes routes")
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

	type response struct {
		Results []string `json:"results"`
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	err = httputil.WriteResponse(
		w,
		http.StatusOK,
		response{
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
		httputil.WriteError(
			w, http.StatusBadRequest, "invalid node id",
		)

		return errors.Wrap(err, "failed to parse node id")
	}

	body, err := httputil.Body[struct {
		NewPath string `json:"newPath" binding:"required"`
	}](w, r)
	if err != nil {
		return errors.Wrap(err, "failed to get body")
	}

	err = h.gs.Graph.Move(id, pathutil.NodePath(body.NewPath))
	if err != nil {
		httputil.WriteError(
			w, http.StatusInternalServerError, "failed to move node",
		)

		return errors.Wrap(err, "failed to move node")
	}

	n, err := h.gs.Graph.GetByID(id)
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
