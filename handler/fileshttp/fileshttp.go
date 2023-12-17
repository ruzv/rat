package fileshttp

import (
	"context"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"rat/graph/services"
	"rat/handler/httputil"
	"rat/logr"
)

type handler struct {
	log             *logr.LogR
	FileURLResolver *services.FileURLResolver
}

// RegisterRoutes registers graph routes on given router.
func RegisterRoutes(
	router *mux.Router, log *logr.LogR, gs *services.GraphServices,
) error {
	log = log.Prefix("viewhttp")

	h := &handler{
		log:             log,
		FileURLResolver: gs.FileURLResolver,
	}

	router.PathPrefix("/file/{path:.+}").
		HandlerFunc(httputil.Wrap(h.log, h.proxyFile)).
		Methods(http.MethodGet)

	return nil
}

func (h *handler) proxyFile(w http.ResponseWriter, r *http.Request) error {
	source, err := h.FileURLResolver.Resolve(mux.Vars(r)["path"])
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "file not found")

		return errors.Wrap(err, "failed to resolve file")
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		source,
		http.NoBody,
	)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "file not found")

		return errors.Wrap(err, "failed to create source request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "file not found")

		return errors.Wrap(err, "failed to get source file")
	}

	defer resp.Body.Close() //nolint:errcheck // ignore.

	for k, v := range resp.Header {
		w.Header().Set(k, v[0])
	}

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "file not found")

		return errors.Wrap(err, "failed to copy source request response")
	}

	return nil
}
