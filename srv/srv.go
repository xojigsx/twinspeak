// Package srv provides the HTTP server implementation for Twinspeak.
package srv

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"jig.sx/twinspeak/pkg/session"
)

// Server represents the HTTP server with session management.
type Server struct {
	Store *session.Store
	mux   *chi.Mux
}

// New creates a new server instance with configured routes.
func New() *Server {
	s := &Server{
		Store: session.NewStore(),
		mux:   chi.NewRouter(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.Use(middleware.Logger)
	s.mux.Use(middleware.Recoverer)

	s.mux.Get("/healthz", s.handleHealth)
	s.mux.Get("/v1/speak", s.handleSpeakWS)
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// handleSpeakWS is implemented in ws.go
