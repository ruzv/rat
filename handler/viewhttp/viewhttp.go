package viewhttp

import (
	"io/fs"
	"net/http"
	"text/template"

	hUtil "private/rat/handler"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

const viewNodeTemplatePath = "view-node-tmpl.html"

type handler struct {
	embeds fs.FS
}

// RegisterRoutes registers view routes on given router.
func RegisterRoutes(router *mux.Router, embeds fs.FS) error {
	h := &handler{
		embeds: embeds,
	}

	viewRouter := router.PathPrefix("/view").
		Subrouter().
		StrictSlash(true)

	// TODO: REDIRECT TO ROOT NODE VIEW PATH
	viewRouter.HandleFunc("/", hUtil.Wrap(h.read)).Methods(http.MethodGet)

	viewRouter.HandleFunc("/{path:.+}", hUtil.Wrap(h.view)).
		Methods(http.MethodGet)

	return nil
}

func (h *handler) view(w http.ResponseWriter, r *http.Request) error {
	path := mux.Vars(r)["path"]

	tmpl, err := template.ParseFS(h.embeds, viewNodeTemplatePath)
	if err != nil {
		hUtil.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to parse template",
		)
	}

	w.WriteHeader(http.StatusOK)

	err = tmpl.Execute(
		w,
		map[string]string{
			"path": path,
		},
	)
	if err != nil {
		hUtil.WriteError(
			w,
			http.StatusInternalServerError,
			"failed to execute template",
		)
	}

	return nil
}

func (h *handler) read(w http.ResponseWriter, r *http.Request) error {
	path := r.URL.Query().Get("node")

	t, err := h.template()
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

func (h *handler) template() (*template.Template, error) {
	t, err := template.ParseFS(h.embeds, "index.html")
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse index.html")
	}

	return t, nil
}
