// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package http

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hubinix/gameplay/pkg/logger"
	"io"
	"net/http"
	"time"
)

// Server is the interface for the healthz server
type Server interface {
	Run(context.Context, int) error
	Ready()
	NotReady()
}

type server struct {
	ready bool
	log   logger.Logger
}

// NewServer returns a new healthz server
func NewServer(log logger.Logger) Server {
	return &server{
		log: log,
	}
}

// Ready sets a ready state for the endpoint handlers
func (s *server) Ready() {
	s.ready = true
}

// NotReady sets a not ready state for the endpoint handlers
func (s *server) NotReady() {
	s.ready = false
}

// Run starts a net/http server with a healthz endpoint
func (s *server) Run(ctx context.Context, port int) error {
	router := http.NewServeMux()
	router.Handle("/healthz", s.healthz())
	router.Handle("/join", s.join())
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r)
	})
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	doneCh := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			s.log.Info("Healthz server is shutting down")
			shutdownCtx, cancel := context.WithTimeout(
				context.Background(),
				time.Second*5,
			)
			defer cancel()
			srv.Shutdown(shutdownCtx) // nolint: errcheck
		case <-doneCh:
		}
	}()

	s.log.Infof("Healthz server is listening on %s", srv.Addr)
	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		s.log.Errorf("Healthz server error: %s", err)
	}
	close(doneCh)
	return err
}

// healthz is a health endpoint handler
func (s *server) healthz() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var status int
		if s.ready {
			status = http.StatusOK
		} else {
			status = http.StatusServiceUnavailable
		}
		w.WriteHeader(status)
	})
}

func decodeJsonFromRead(read io.Reader, v interface{}) error {
	return json.NewDecoder(read).Decode(v)
}

// healthz is a health endpoint handler
func (s *server) join() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := make(map[string]string)
		if e := decodeJsonFromRead(r.Body, a); e != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {

			var status int

			if s.ready {
				status = http.StatusOK
			} else {
				status = http.StatusServiceUnavailable
			}
			w.WriteHeader(status)
		}
	})
}
