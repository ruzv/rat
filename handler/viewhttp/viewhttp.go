package viewhttp

import (
	"html/template"
	"io/fs"
	"net/http"

	"private/rat/handler"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type Handler struct {
	viewTemplate *template.Template
}

// creates a new Handler.
func newHandler(embeds fs.FS) (*Handler, error) {
	t, err := template.ParseFS(embeds, "index.html")
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse index.html")
	}

	return &Handler{
		viewTemplate: t,
	}, nil
}

// RegisterRoutes registers view routes on given router.
func RegisterRoutes(router *mux.Router, embeds fs.FS) error {
	h, err := newHandler(embeds)
	if err != nil {
		return errors.Wrap(err, "failed create new view handler")
	}

	viewRouter := router.PathPrefix("/view").
		Subrouter().
		StrictSlash(true)

	viewRouter.HandleFunc("/", handler.Wrap(h.read)).Methods(http.MethodGet)

	return nil
}

// -------------------------------------------------------------------------- //
// READ
// -------------------------------------------------------------------------- //

func (h *Handler) read(w http.ResponseWriter, r *http.Request) error {
	path := r.URL.Query().Get("node")

	// w.Header().Add("Access-Control-Allow-Origin", "*")

	w.WriteHeader(http.StatusOK)

	err := h.viewTemplate.Execute(
		w,
		map[string]string{"path": path, "api_base": ""},
	)
	if err != nil {
		handler.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to execute template",
		)

		return errors.Wrap(err, "failed to execute template")
	}

	return nil
}
