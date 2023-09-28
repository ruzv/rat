package router

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"rat/graph/services"
	"rat/handler/graphhttp"
	"rat/handler/httputil"
	"rat/handler/viewhttp"
	"rat/logr"
)

// NewRouter creates a new router, loads templates and registers handlers for
// routes.
func NewRouter(
	log *logr.LogR, gs *services.GraphServices, webStaticContent fs.FS,
) (*mux.Router, error) {
	log = log.Prefix("router")
	router := mux.NewRouter()

	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			httputil.WriteError(w, http.StatusNotFound, "not found")
		},
	)

	router.MethodNotAllowedHandler = http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			httputil.WriteError(
				w, http.StatusMethodNotAllowed, "method not allowed",
			)
		},
	)

	router.Use(GetAccessLoggerMW(log, false))

	err := graphhttp.RegisterRoutes(router, log, gs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register graph routes")
	}

	err = viewhttp.RegisterRoutes(router, log, webStaticContent)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register web routes")
	}

	logRoutes(log, router)

	return router, nil
}

func logRoutes(log *logr.LogR, router *mux.Router) {
	lg := log.Group(logr.LogLevelInfo)
	defer lg.Close()

	err := router.Walk(
		func(
			route *mux.Route, _ *mux.Router, _ []*mux.Route,
		) error {
			path, err := route.GetPathTemplate()
			if err != nil {
				return errors.Wrap(err, "failed to get path template")
			}

			methods, err := route.GetMethods()
			if err != nil {
				// route does not have a methods.
				return nil //nolint:nilerr
			}

			for _, m := range methods {
				lg.Log("%-7s %s", m, path)
			}

			return nil
		},
	)
	if err != nil {
		log.Errorf("failed to log routes: %s", err.Error())
	}
}

// GetAccessLoggerMW returns a middleware that logs the access.
func GetAccessLoggerMW(log *logr.LogR, all bool) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b := httputil.NewBufferResponseWriter(w)

			startT := time.Now()

			next.ServeHTTP(b, r)

			if !all && b.Code/100 == 2 {
				return
			}

			log.Infof(
				"rat access %-7s %d %-7fs %s",
				r.Method,
				b.Code,
				time.Since(startT).Seconds(),
				r.URL.Path,
			)
		})
	}
}
