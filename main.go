/*
** Zabbix
** Copyright (C) 2001-2023 Zabbix SIA
**
** This program is free software; you can redistribute it and/or modify
** it under the terms of the GNU General Public License as published by
** the Free Software Foundation; either version 2 of the License, or
** (at your option) any later version.
**
** This program is distributed in the hope that it will be useful,
** but WITHOUT ANY WARRANTY; without even the implied warranty of
** MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
** GNU General Public License for more details.
**
** You should have received a copy of the GNU General Public License
** along with this program; if not, write to the Free Software
** Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
**/

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"rat/args"
	"rat/config"
	"rat/graph/services"
	"rat/handler/router"
	"rat/logr"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

// Rat describes the rat server.
type Rat struct {
	log *logr.LogR
	gs  *services.GraphServices
	s   *http.Server
}

// NewRat creates a new rat server.
func NewRat(cmdArgs *args.Args) (*Rat, error) {
	conf, err := config.Load(cmdArgs.ConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	log := logr.NewLogR(os.Stdout, "rat", conf.LogLevel)

	gs, err := services.NewGraphServices(log, conf.Graph)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create graph services")
	}

	r, err := router.NewRouter(log, gs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new router")
	}

	return &Rat{
		log: log.Prefix("server"),
		gs:  gs,
		s: &http.Server{
			Handler:      r,
			Addr:         fmt.Sprintf(":%d", conf.Port),
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		},
	}, nil
}

// Serve starts the rat server. Blocks.
func (r *Rat) Serve() {
	for {
		r.log.Infof("serving on %s", r.s.Addr)

		err := r.s.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}

			r.log.Errorf("failed to listen and server: %s", err.Error())
		}
	}
}

// Stop stops the rat server.
func (r *Rat) Stop() error {
	r.log.Infof("stopping rat server")

	err := r.s.Shutdown(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to stop server")
	}

	r.log.Infof("stopping rat services")

	err = r.gs.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close graph services")
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}

func run() error {
	cmdArgs, ok := args.Load()
	if !ok {
		return nil
	}

	rat, err := NewRat(cmdArgs)
	if err != nil {
		return errors.Wrap(err, "failed to create rat")
	}

	go rat.Serve()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT)

	<-quit

	defer close(quit)

	err = rat.Stop()
	if err != nil {
		return errors.Wrap(err, "failed to stop rat")
	}

	rat.log.Infof("bye")

	return nil
}
