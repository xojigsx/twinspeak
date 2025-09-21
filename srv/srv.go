package srv

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"jig.sx/twinspeak/pkg/session"
)

type Server struct {
	Store *session.Store
	mux   *chi.Mux
}

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

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// handleSpeakWS is implemented in ws.go
