package viewhttp

import (
	"io/fs"
	"net/http"
	"text/template"

	hUtil "private/rat/handler"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type handler struct {
	// viewTemplate *template.Template
	embeds fs.FS
}

// creates a new Handler.
func newHandler(embeds fs.FS) (*handler, error) {
	return &handler{
		embeds: embeds,
		// viewTemplate: t,
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

	viewRouter.HandleFunc("/", hUtil.Wrap(h.read)).Methods(http.MethodGet)

	return nil
}

// -------------------------------------------------------------------------- //
// READ
// -------------------------------------------------------------------------- //

func (h *handler) read(w http.ResponseWriter, r *http.Request) error {
	path := r.URL.Query().Get("node")

	// w.Header().Add("Access-Control-Allow-Origin", "*")

	t, err := h.template(w)
	if err != nil {
		return errors.Wrap(err, "failed to get template")
	}

	w.WriteHeader(http.StatusOK)

	err = t.Execute(
		w,
		map[string]string{"path": path, "api_base": ""},
	)
	if err != nil {
		hUtil.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to execute template",
		)

		return errors.Wrap(err, "failed to execute template")
	}

	return nil
}

func (h *handler) template(w http.ResponseWriter) (*template.Template, error) {
	t, err := template.ParseFS(h.embeds, "index.html")
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse index.html")
	}

	return t, nil
}
