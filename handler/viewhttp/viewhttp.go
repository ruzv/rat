package viewhttp

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"rat/handler/httputil"
	"rat/logr"
)

type handler struct {
	log *logr.LogR

	// gs *services.GraphServices
}

var pathRe = regexp.MustCompile(
	`[[:alnum:]]+(?:-(?:[[:alnum:]]+))*(?:\/[[:alnum:]]+(?:-(?:[[:alnum:]]+))*)*`, //nolint:lll
)

// RegisterRoutes registers graph routes on given router.
func RegisterRoutes(
	router *mux.Router, log *logr.LogR,
	// gs *services.GraphServices,
) error {
	h := &handler{
		log: log.Prefix("webhttp"),
	}

	router.PathPrefix(
		fmt.Sprintf("/view/{path:%s}", pathRe.String()),
	).
		HandlerFunc(httputil.Wrap(h.log, h.web)).
		Methods(http.MethodGet)

	router.PathPrefix("/static/").
		Handler(
			http.StripPrefix(
				"/static/",
				http.FileServer(http.Dir("web/build/static")),
			),
		)

	return nil
}

func (h *handler) web(w http.ResponseWriter, r *http.Request) error {
	http.ServeFile(w, r, "web/build/index.html")

	return nil
}
