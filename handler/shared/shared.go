package shared

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"

	"private/rat/config"
	"private/rat/graph"
	"private/rat/graph/filesystem"
	"private/rat/graph/pathcache"
	"private/rat/graph/render"
	"private/rat/graph/render/templ"
	"private/rat/graph/util/path"

	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var log = logging.MustGetLogger("handler-utils")

type (
	// MuxHandlerFunc is a handler function for mux.
	MuxHandlerFunc func(http.ResponseWriter, *http.Request)
	// RatHandlerFunc is a handler function.
	RatHandlerFunc func(http.ResponseWriter, *http.Request) error
)

// Services contains services that are shared between handlers.
type Services struct {
	Graph      graph.Provider
	Templates  *templ.TemplateStore
	Renderer   *render.Renderer
	templateFS fs.FS
}

// NewServices creates a new Services.
func NewServices(
	graphConf *config.GraphConfig,
	templateFS fs.FS,
) (*Services, error) {
	p, err := filesystem.NewFileSystem(graphConf.Name, graphConf.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create net fs")
	}

	pc := pathcache.NewPathCache(p)

	ts, err := templ.FileTemplateStore(templateFS)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create default template store")
	}

	return &Services{
		Graph:      pc,
		Templates:  ts,
		Renderer:   render.NewRenderer(ts, pc),
		templateFS: templateFS,
	}, nil
}

// GetNode returns a node by path. If path is empty, the root node is returned.
func (ss *Services) GetNode(path path.NodePath) (*graph.Node, error) {
	if path == "" {
		n, err := ss.Graph.Root()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get root node")
		}

		return n, nil
	}

	n, err := ss.Graph.GetByPath(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get node by path")
	}

	return n, nil
}

// ReloadTemplatesMW reloads the template store before each request.
func (ss *Services) ReloadTemplatesMW(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ts, err := templ.FileTemplateStore(ss.templateFS)
			if err != nil {
				log.Errorf("failed to reload templates: %v", err)
			} else {
				ss.Templates = ts
			}

			next.ServeHTTP(w, r)
		},
	)
}

// Wrap wraps a RatHandlerFunc to be used with mux.
func Wrap(f RatHandlerFunc) MuxHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			log.Errorf("handler error: %v", err)
		}
	}
}

// WriteResponse writes a response to the response writer.
func WriteResponse(w http.ResponseWriter, code int, body any) error {
	b := &bytes.Buffer{}

	err := json.NewEncoder(b).Encode(body)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to encode body")

		return errors.Wrap(err, "failed to encode body")
	}

	w.WriteHeader(code)
	w.Write(b.Bytes()) //nolint:errcheck

	return nil
}

// WriteError writes an error to the response.
func WriteError(w http.ResponseWriter, code int, message string) {
	WriteResponse( //nolint:errcheck
		w,
		code,
		struct {
			Code  int    `json:"code"`
			Error string `json:"error"`
		}{
			Code:  code,
			Error: message,
		},
	)
}

// Body reads the requests body as a specified struct.
func Body[T any](w http.ResponseWriter, r *http.Request) (T, error) { //nolint:ireturn,lll
	defer r.Body.Close()

	var body, empty T

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "failed to decode body")

		return empty, errors.Wrap(err, "failed to decode body")
	}

	return body, nil
}
