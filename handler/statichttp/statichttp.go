package statichttp

import (
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var log = logging.MustGetLogger("statichttp")

// RegisterRoutes registers the routes for the static http handler.
func RegisterRoutes(router *mux.Router, embeds fs.FS) error {
	log.Info("seting up statics serving")

	statics, err := fs.Sub(embeds, "static")
	if err != nil {
		return errors.Wrap(err, "failed to get statics sub fs")
	}

	err = fs.WalkDir(
		statics,
		".",
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			log.Info(path)

			return nil
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to walk static files")
	}

	router.PathPrefix("/static/").
		Handler(http.StripPrefix("/static/", http.FileServer(http.FS(statics))))

	log.Notice("serving statics")

	return nil
}
