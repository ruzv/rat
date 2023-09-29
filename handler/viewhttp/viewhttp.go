package viewhttp

import (
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"rat/handler/httputil"
	"rat/logr"
)

var pathRe = regexp.MustCompile(
	`[[:alnum:]]+(?:-(?:[[:alnum:]]+))*(?:\/[[:alnum:]]+(?:-(?:[[:alnum:]]+))*)*`, //nolint:lll // does not make sense to split.
)

type handler struct {
	log *logr.LogR
	wsc fs.FS
}

func (h *handler) logServedContent() {
	lg := h.log.Group(logr.LogLevelInfo)
	defer lg.Close()

	err := fs.WalkDir(
		h.wsc,
		".",
		func(path string, _ fs.DirEntry, err error) error {
			if err != nil {
				lg.Log("failed to walk %s", path)
			} else {
				lg.Log("%s", path)
			}

			return nil
		},
	)
	if err != nil {
		h.log.Errorf("failed to walk served content: %v", err)
	}
}

// RegisterRoutes registers graph routes on given router.
func RegisterRoutes(
	router *mux.Router, log *logr.LogR, webStaticContent fs.FS,
) error {
	log = log.Prefix("viewhttp")

	h := &handler{
		log: log,
		wsc: webStaticContent,
	}

	router.PathPrefix(fmt.Sprintf("/view/{path:%s}", pathRe.String())).
		HandlerFunc(httputil.Wrap(h.log, h.serveFile("index.html"))).
		Methods(http.MethodGet)

	router.PathPrefix("/favicon.png").
		HandlerFunc(httputil.Wrap(h.log, h.serveFile("favicon.png"))).
		Methods(http.MethodGet)

	staticContent, err := fs.Sub(webStaticContent, "static")
	if err != nil {
		return errors.Wrap(err, "failed to sub static content dir from embed")
	}

	router.PathPrefix("/static/").
		Handler(http.StripPrefix(
			"/static/", http.FileServer(http.FS(staticContent)),
		))

	h.logServedContent()

	return nil
}

func (h *handler) serveFile(
	name string,
) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		file, err := h.wsc.Open(name)
		if err != nil {
			return errors.Wrap(err, "failed to open index.html")
		}

		defer file.Close()

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
