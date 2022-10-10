package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"private/rat/args"
	"private/rat/config"
	"private/rat/handler/router"

	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

//go:embed embed
var binaryEmbeds embed.FS

var log = logging.MustGetLogger("rat")

func main() {
	err := run()
	if err != nil {
		log.Error(err)
		panic(err)
	}
}

func run() error {
	cmdArgs, ok := args.Load()
	if !ok {
		return nil
	}

	server, err := setupServer(cmdArgs)
	if err != nil {
		return errors.Wrap(err, "failed to setup server")
	}

	go startServer(server)

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT)

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	for q := range quit {
		if q != syscall.SIGINT {
			log.Warning("received unexpected signal", q)

			continue
		}

		if ctx == nil {
			err = stopServer(server)
			if err != nil {
				close(quit)

				return errors.Wrap(err, "failed to stop server")
			}

			ctx, cancel = context.WithCancel(context.Background())

			go func() {
				select {
				case <-time.After(1 * time.Second):
					defer func() {
						cancel()

						ctx = nil
					}()

					server, err = setupServer(cmdArgs)
					if err != nil {
						log.Error("failed to setup server", err)
						close(quit)

						return
					}

					go startServer(server)

				case <-ctx.Done():
					log.Warning("restart cancelled")
					close(quit)
				}
			}()
		} else {
			cancel()
			ctx = nil
		}
	}

	return nil
}

func startServer(server *http.Server) {
	log.Notice("starting rat server")

	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("failed to listen and server", err)

		return
	}
}

func setupServer(cmdArgs *args.Args) (*http.Server, error) {
	initLogger()

	embeds, err := getEmbeds(cmdArgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get embeds")
	}

	conf, err := config.Load(cmdArgs.ConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	log.Infof("have config")

	r, err := router.New(conf, embeds)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new router")
	}

	log.Infof("have router")

	server := http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%d", conf.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Notice("server setup on port", conf.Port)

	return &server, nil
}

func stopServer(server *http.Server) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to shutdown server")
	}

	log.Infof("rat server stopped")

	return nil
}

func initLogger() {
	logging.SetBackend(
		logging.NewBackendFormatter(
			logging.NewLogBackend(os.Stdout, "", 0),
			logging.MustStringFormatter(
				`%{color}%{time:15:04:05.0000} `+
					`%{level:.4s} %{id:03x} %{module}.%{longfunc} `+
					`▶%{color:reset} %{message}`,
			),
		),
	)
}

func getEmbeds(cmdArgs *args.Args) (fs.FS, error) {
	var (
		embeds fs.FS
		err    error
	)

	if cmdArgs.Embed {
		log.Notice("using binary embedded file system")

		embeds, err = fs.Sub(&binaryEmbeds, "embed")
		if err != nil {
			return nil, errors.Wrap(err, "failed to get embeds")
		}
	} else {
		log.Notice("using file system")

		embeds = os.DirFS("embed")
	}

	log.Notice("embedded files")

	err = fs.WalkDir(
		embeds,
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
		return nil, errors.Wrap(err, "failed to walk static files")
	}

	return embeds, nil
}
