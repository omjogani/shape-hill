package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/omjogani/shape-hill/internal/store"
)

// A scope that hasn't moved in this long is stalled, the signal the chart exists to show.
const stalledAfter = 7 * 24 * time.Hour

type Server struct {
	store  *store.Store
	log    *slog.Logger
	mux    *http.ServeMux
	verify VerifyToken
}

func New(st *store.Store, log *slog.Logger, verify VerifyToken) *Server {
	server := &Server{store: st, log: log, mux: http.NewServeMux(), verify: verify}

	server.mux.HandleFunc("GET /healthz", server.health)
	server.mux.HandleFunc("GET /hill/{file}", server.embed)
	server.mux.HandleFunc("GET /api/public/hills/{slug}", server.getPublicHill)
	server.mux.HandleFunc("GET /api/public/scopes/{id}/positions", server.publicScopeSnapshots)
	server.mux.HandleFunc("GET /api/me", server.authenticate(server.currentUser))
	server.mux.HandleFunc("POST /api/onboard", server.authenticate(server.onboard))
	server.mux.HandleFunc("GET /api/hills", server.authed(server.listHills))
	server.mux.HandleFunc("POST /api/hills", server.authed(server.createHill))
	server.mux.HandleFunc("GET /api/hills/{slug}", server.authed(server.getHill))
	server.mux.HandleFunc("PATCH /api/hills/{slug}", server.authed(server.updateHill))
	server.mux.HandleFunc("POST /api/hills/{slug}/scopes", server.authed(server.createScope))
	server.mux.HandleFunc("PATCH /api/scopes/{id}", server.authed(server.updateScope))
	server.mux.HandleFunc("DELETE /api/scopes/{id}", server.authed(server.deleteScope))
	server.mux.HandleFunc("GET /api/scopes/{id}/positions", server.authed(server.scopeSnapshots))
	server.mux.HandleFunc("POST /api/scopes/{id}/positions", server.authed(server.moveScope))

	return server
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
