package viewhttp

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"rat/graph"
	"rat/graph/render/templ"
	"rat/graph/util"
	pathutil "rat/graph/util/path"
	"rat/handler/shared"

	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var log = logging.MustGetLogger("nodeshttp")

type handler struct {
	ss *shared.Services
}

// RegisterRoutes registers view routes on given router.
func RegisterRoutes(
	router *mux.Router,
	ss *shared.Services,
) error {
	h := &handler{
		ss: ss,
	}

	viewRouter := router.PathPrefix("/view").
		Subrouter().
		StrictSlash(true)

	viewRouter.HandleFunc("/", shared.Wrap(h.rootRedirect)).
		Methods(http.MethodGet)

	viewRouter.HandleFunc("/{path:.+}", shared.Wrap(h.view)).
		Methods(http.MethodGet)

	return nil
}

func (h *handler) rootRedirect(w http.ResponseWriter, r *http.Request) error {
	n, err := h.ss.Graph.Root()
	if err != nil {
		h.writeError(
			w,
			&templ.ErrorPageTemplData{
				Code:    http.StatusInternalServerError,
				Message: "something went wrong while looking for root node",
				Cause:   err.Error(),
			},
		)

		return errors.Wrap(err, "failed to get root node")
	}

	path, err := url.JoinPath("/view", string(n.Path))
	if err != nil {
		h.writeError(
			w,
			&templ.ErrorPageTemplData{
				Code:    http.StatusInternalServerError,
				Message: "something went wrong while preparing root node path",
				Cause:   err.Error(),
			},
		)

		return errors.Wrap(err, "failed to get root node path")
	}

	http.Redirect(w, r, path, http.StatusFound)

	return nil
}

func (h *handler) view(w http.ResponseWriter, r *http.Request) error {
	n, err := h.ss.GetNode(pathutil.NodePath(mux.Vars(r)["path"]))
	if err != nil {
		if !errors.Is(err, graph.ErrNodeNotFound) {
			h.writeError(
				w,
				&templ.ErrorPageTemplData{
					Code:    http.StatusInternalServerError,
					Message: "something went wrong while looking for node",
					Cause:   err.Error(),
				},
			)

			return errors.Wrap(err, "failed to get node")
		}

		h.writeError(
			w,
			&templ.ErrorPageTemplData{
				Code:    http.StatusNotFound,
				Message: "node not found",
			},
		)

		return nil
	}

	leafs, err := n.GetLeafs(h.ss.Graph)
	if err != nil {
		h.writeError(
			w,
			&templ.ErrorPageTemplData{
				Code:    http.StatusInternalServerError,
				Message: "something went wrong while looking for leafs",
				Cause:   err.Error(),
			},
		)

		return errors.Wrap(err, "failed to get leafs")
	}

	err = h.writeIndex(
		w,
		&templ.IndexTemplData{
			ID:      n.ID.String(),
			Name:    n.Name,
			Path:    string(n.Path),
			Content: h.ss.Renderer.Render(n),
			Leafs: util.Map(
				leafs,
				func(l *graph.Node) *templ.IndexTemplLeafData {
					return &templ.IndexTemplLeafData{
						Content: h.ss.Renderer.Render(l),
						Path:    string(l.Path),
					}
				},
			),
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to render page")
	}

	return nil
}

func (h *handler) writeError(
	w http.ResponseWriter, data *templ.ErrorPageTemplData,
) {
	b := &bytes.Buffer{}

	err := h.ss.Templates.ErrorPage(b, data)
	if err != nil {
		shared.WriteError(
			w,
			http.StatusInternalServerError,
			fmt.Sprintf(
				"an error (%s) occurred while rendering error "+
					"page with the following data: %v",
				err.Error(),
				data,
			),
		)

		log.Errorf("failed to render error page: %s", err.Error())

		return
	}

	w.WriteHeader(data.Code)

	_, err = w.Write(b.Bytes())
	if err != nil {
		shared.WriteError(
			w,
			http.StatusInternalServerError,
			fmt.Sprintf(
				"an error (%s) occurred while rendering error "+
					"page with the following data: %v",
				err.Error(),
				data,
			),
		)

		log.Errorf("failed to write error page: %s", err.Error())
	}
}

func (h *handler) writeIndex(
	w http.ResponseWriter, data *templ.IndexTemplData,
) error {
	b := &bytes.Buffer{}

	err := h.ss.Templates.Index(b, data)
	if err != nil {
		h.writeError(
			w,
			&templ.ErrorPageTemplData{
				Code:    http.StatusInternalServerError,
				Message: "something went wrong while rendering page",
				Cause:   err.Error(),
			},
		)

		return errors.Wrap(err, "failed to render index page")
	}

	w.WriteHeader(http.StatusOK)

	_, err = w.Write(b.Bytes())
	if err != nil {
		h.writeError(
			w,
			&templ.ErrorPageTemplData{
				Code:    http.StatusInternalServerError,
				Message: "something went wrong while rendering page",
				Cause:   err.Error(),
			},
		)

		return errors.Wrap(err, "failed to write index page")
	}

	return nil
}
