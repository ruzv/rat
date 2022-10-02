package statichttp

import (
	"io/fs"
	"net/http"

	"private/rat/errors"

	"github.com/gorilla/mux"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("statichttp")

func RegisterRoutes(router *mux.Router, embeds fs.FS) error {
	statics, err := fs.Sub(embeds, "static")
	if err != nil {
		return errors.Wrap(err, "failed to get statics sub fs")
	}

	log.Notice("serving static files")

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

	return nil
}
