package shared

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"

	"rat/config"
	"rat/graph"
	"rat/graph/pathcache"
	"rat/graph/render"
	"rat/graph/render/templ"
	"rat/graph/singlefile"
	"rat/graph/sync"
	"rat/graph/util/path"

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
	Graph     graph.Provider
	Templates *templ.TemplateStore
	Renderer  *render.Renderer
	Syncer    *sync.Syncer
}

// NewServices creates a new Services.
func NewServices(
	graphConf *config.GraphConfig,
	templateFS fs.FS,
) (*Services, error) {
	p := singlefile.NewSingleFile(graphConf.Name, graphConf.Path)

	pc := pathcache.NewPathCache(p)

	ts, err := templ.FileTemplateStore(templateFS)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create default template store")
	}

	s, err := sync.NewSyncer(graphConf.Path, graphConf.Sync)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create syncer")
	}

	s.Start()

	return &Services{
		Graph:     pc,
		Templates: ts,
		Renderer:  render.NewRenderer(ts, pc),
		Syncer:    s,
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
func Body[T any](
	w http.ResponseWriter,
	r *http.Request,
) (T, error) { //nolint:ireturn
	defer r.Body.Close()

	var body, empty T

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "failed to decode body")

		return empty, errors.Wrap(err, "failed to decode body")
	}

	return body, nil
}
