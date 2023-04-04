package router

import (
	"io/fs"
	"net/http"
	"time"

	"private/rat/config"
	"private/rat/handler/graphhttp"
	"private/rat/handler/shared"
	"private/rat/handler/statichttp"
	"private/rat/handler/viewhttp"

	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var log = logging.MustGetLogger("router")

// New creates a new router, loads templates and registers handlers for routes.
func New(
	conf *config.Config,
	embeds fs.FS,
) (*mux.Router, error) {
	router := mux.NewRouter()

	templateFS, err := fs.Sub(embeds, "render-templates")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get render-templates sub fs")
	}

	ss, err := shared.NewServices(conf.Graph, templateFS)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create shared services")
	}

	router.Use(
		ss.ReloadTemplatesMW,
		GetAccessLoggerMW(false),
	)

	err = graphhttp.RegisterRoutes(router, ss)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register graph routes")
	}

	err = viewhttp.RegisterRoutes(router, ss)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register view routes")
	}

	err = statichttp.RegisterRoutes(router, embeds)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register static routes")
	}

	err = router.Walk(
		func(
			route *mux.Route, router *mux.Router, ancestors []*mux.Route,
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
				log.Infof("route %-7s %s", m, path)
			}

			return nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk routes")
	}

	log.Notice("registered routes")

	return router, nil
}

// BufferResponseWriter is a wrapper around http.ResponseWriter that proxies
// the data ans stores the status code.
type BufferResponseWriter struct {
	Code int
	w    http.ResponseWriter
}

var _ http.ResponseWriter = (*BufferResponseWriter)(nil)

// Header .
func (w *BufferResponseWriter) Header() http.Header {
	return w.w.Header()
}

// Write .
func (w *BufferResponseWriter) Write(b []byte) (int, error) {
	n, err := w.w.Write(b)
	if err != nil {
		return 0, errors.Wrap(err, "failed to write")
	}

	return n, nil
}

// WriteHeader .
func (w *BufferResponseWriter) WriteHeader(code int) {
	w.Code = code
	w.w.WriteHeader(code)
}

// GetAccessLoggerMW returns a middleware that logs the access.
func GetAccessLoggerMW(all bool) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b := &BufferResponseWriter{
				w: w,
			}

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
