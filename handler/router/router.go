package router

import (
	"bytes"
	"io/fs"
	"net/http"
	"time"

	"private/rat/config"
	"private/rat/handler/nodeshttp"
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

	router.Use(GetAccessLoggerMW())

	err := nodeshttp.RegisterRoutes(conf, router)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register node routes")
	}

	err = viewhttp.RegisterRoutes(router, embeds)
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
				log.Info("route %-7s %s %s", "", path, err.Error())
			}

			for _, m := range methods {
				log.Info("route %-7s %s", m, path)
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

type BufferResponseWriter struct {
	Code int
	Body bytes.Buffer
}

// MINUS25

var _ http.ResponseWriter = (*BufferResponseWriter)(nil)

func (w *BufferResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *BufferResponseWriter) Write(b []byte) (int, error) {
	n, err := w.Body.Write(b)
	if err != nil {
		return 0, errors.Wrap(err, "failed to write to buffer")
	}

	return n, nil
}

func (w *BufferResponseWriter) WriteHeader(code int) {
	w.Code = code
}

func (w *BufferResponseWriter) Flush(out http.ResponseWriter) error {
	if w.Code != 0 {
		out.WriteHeader(w.Code)
	}

	_, err := w.Body.WriteTo(out)

	return errors.Wrap(err, "failed to write body")
}

func GetAccessLoggerMW() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startT := time.Now()
			b := &BufferResponseWriter{}

			next.ServeHTTP(b, r)

			log.Infof(
				"rat access %s %d %.5fs %-6d %s",
				r.Method,
				b.Code,
				time.Since(startT).Seconds(),
				b.Body.Len(),
				r.URL.Path,
			)

			err := b.Flush(w)
			if err != nil {
				log.Error("failure to flush response", err)
			}
		})
	}
}
