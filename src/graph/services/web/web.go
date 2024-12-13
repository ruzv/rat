package web

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"rat/graph/services"
	"rat/graph/services/api/httputil"
	"rat/logr"
)

var _ services.Service = (*Web)(nil)

// Config defines configuration parameters for the api server.
type Config struct {
	Port         int                      `yaml:"port" validate:"nonzero"`
	APIAuthority string                   `yaml:"apiAuthority" validate:"nonzero"` //nolint:lll
	Timeouts     *httputil.ServerTimeouts `yaml:"timeouts"`
}

// Web is a rat service that server the Rat web app.
type Web struct {
	log    *logr.LogR
	config *Config
	server *http.Server
}

type handler struct {
	log          *logr.LogR
	wsc          fs.FS
	apiAuthority string
}

// New creates a new API server service.
func New(
	config *Config,
	log *logr.LogR,
	webStaticContent fs.FS,
) (*Web, error) {
	log = log.Prefix("web")

	r, err := newRouter(log, webStaticContent, config.APIAuthority)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create router")
	}

	timeouts := config.Timeouts.FillDefaults()

	return &Web{
		log:    log,
		config: config,
		server: &http.Server{
			Handler:      r,
			Addr:         fmt.Sprintf(":%d", config.Port),
			WriteTimeout: timeouts.Write,
			ReadTimeout:  timeouts.Read,
			IdleTimeout:  timeouts.Idle,
		},
	}, nil
}

// Run runs the web app server.
func (web *Web) Run() error {
	web.log.Infof("web available on: http://localhost:%d", web.config.Port)

	start := time.Now()

	err := web.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "listen and serve error")
	}

	web.log.Infof("uptime: %s", time.Since(start).String())

	return nil
}

// Stop stops the web app server.
func (web *Web) Stop(ctx context.Context) error {
	web.log.Infof("shutting down")

	err := web.server.Shutdown(ctx) // trigger exit
	if err != nil {
		return errors.Wrap(err, "failed to shutdown server")
	}

	return nil
}

// newRouter creates a new router and registers static file server routes
// on given router.
func newRouter(
	log *logr.LogR, webStaticContent fs.FS, apiAuthority string,
) (*mux.Router, error) {
	router := mux.NewRouter()

	h := &handler{
		log:          log,
		wsc:          webStaticContent,
		apiAuthority: apiAuthority,
	}

	router.PathPrefix("/view/{path:.*}").
		HandlerFunc(
			httputil.Wrap(h.serveFile("index.html"), h.log, "view-node"),
		).
		Methods(http.MethodGet)

	router.PathPrefix("/view").
		HandlerFunc(
			httputil.Wrap(h.serveFile("index.html"), h.log, "view-root"),
		).
		Methods(http.MethodGet)

	router.PathPrefix("/favicon.png").
		HandlerFunc(
			httputil.Wrap(h.serveFile("favicon.png"), h.log, "favicon"),
		).
		Methods(http.MethodGet)

	router.PathPrefix("/api-authority").
		HandlerFunc(
			httputil.Wrap(h.getAPIAuthority, h.log, "favicon"),
		).
		Methods(http.MethodGet)

	staticContent, err := fs.Sub(webStaticContent, "static")
	if err != nil {
		return nil, errors.Wrap(
			err,
			"failed to sub static content dir from embed",
		)
	}

	router.PathPrefix("/static/").
		Handler(http.StripPrefix(
			"/static/", http.FileServer(http.FS(staticContent)),
		))

	h.logServedContent()

	return router, nil
}

func (h *handler) serveFile(
	name string,
) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, _ *http.Request) error {
		file, err := h.wsc.Open(name)
		if err != nil {
			return errors.Wrap(err, "failed to open index.html")
		}

		defer file.Close() //nolint:errcheck // ignore.

		contentType := mime.TypeByExtension(filepath.Ext(name))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)

		_, err = io.Copy(w, file)
		if err != nil {
			return errors.Wrap(err, "failed to copy index.html")
		}

		return nil
	}
}

func (h *handler) getAPIAuthority(
	w http.ResponseWriter, _ *http.Request,
) error {
	err := httputil.WriteResponse(w, http.StatusOK, struct {
		Authority string `json:"authority"`
	}{
		Authority: h.apiAuthority,
	})
	if err != nil {
		return errors.Wrap(err, "failed to write response")
	}

	return nil
}

func (h *handler) logServedContent() {
	lg := h.log.Group(logr.LogLevelInfo)
	defer lg.Close()

	lg.Log("serving static content")

	err := fs.WalkDir(
		h.wsc,
		".",
		func(path string, _ fs.DirEntry, err error) error {
			if err != nil {
				lg.Log("  failed to walk %s", path)
			} else {
				lg.Log("  %s", path)
			}

			return nil
		},
	)
	if err != nil {
		h.log.Errorf("failed to walk served content: %v", err)
	}
}
