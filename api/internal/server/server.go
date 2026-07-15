package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/omjogani/shape-hill/internal/store"
)

// A scope that hasn't moved in this long is stalled — the signal the chart exists to show.
const stalledAfter = 7 * 24 * time.Hour

type Server struct {
	store *store.Store
	log   *slog.Logger
	mux   *http.ServeMux
}

func New(st *store.Store, log *slog.Logger) *Server {
	s := &Server{store: st, log: log, mux: http.NewServeMux()}

	s.mux.HandleFunc("GET /healthz", s.health)
	s.mux.HandleFunc("GET /h/{file}", s.embed)
	s.mux.HandleFunc("POST /api/users", s.createUser)
	s.mux.HandleFunc("POST /api/hills", s.createHill)
	s.mux.HandleFunc("GET /api/hills/{slug}", s.getHill)
	s.mux.HandleFunc("POST /api/hills/{slug}/scopes", s.createScope)
	s.mux.HandleFunc("POST /api/scopes/{id}/positions", s.moveScope)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			s.log.Error("panic serving request", "path", r.URL.Path, "err", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}()

	start := time.Now()
	s.mux.ServeHTTP(w, r)
	s.log.Info("request", "method", r.Method, "path", r.URL.Path, "took", time.Since(start))
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		return
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
