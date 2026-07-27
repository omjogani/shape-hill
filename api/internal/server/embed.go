package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/omjogani/shape-hill/internal/hillchart"
	"github.com/omjogani/shape-hill/internal/store"
)

// embed serves the image that gets pasted into tickets and READMEs. It is fetched
// server-side by image proxies, not by a browser we control: no cookies, no auth
func (s *Server) embed(w http.ResponseWriter, r *http.Request) {
	slug, ok := strings.CutSuffix(r.PathValue("file"), ".svg")
	if !ok {
		http.NotFound(w, r)
		return
	}

	hill, ok := s.loadHill(w, r, slug)
	if !ok {
		return
	}

	style := hillchart.Style(r.URL.Query().Get("style"))

	if !hill.IsPublic {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
		w.Write(hillchart.RenderPrivate(style))
		return
	}

	scopes, ok := s.loadScopes(w, r, hill.ID)
	if !ok {
		return
	}

	// ETag so a proxy can be told "nothing moved" without us drawing anything.
	etag := fmt.Sprintf(`W/"%d-%s"`, store.LastMovedOn(hill, scopes).UnixNano(), style)
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age=60, must-revalidate")
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Write(hillchart.Render(chartOf(hill, scopes, style)))
}

func (s *Server) loadHill(w http.ResponseWriter, r *http.Request, slug string) (store.Hill, bool) {
	hill, err := s.store.HillBySlug(r.Context(), slug)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return store.Hill{}, false
	}
	if err != nil {
		s.log.Error("load hill for embed", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return store.Hill{}, false
	}
	return hill, true
}

func (s *Server) loadScopes(w http.ResponseWriter, r *http.Request, hillID string) ([]store.Scope, bool) {
	scopes, err := s.store.ScopesForHill(r.Context(), hillID)
	if err != nil {
		s.log.Error("load scopes for embed", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return nil, false
	}
	return scopes, true
}

func chartOf(hill store.Hill, scopes []store.Scope, style hillchart.Style) hillchart.Chart {
	chart := hillchart.Chart{Title: hill.Title, Style: style}
	for _, scope := range scopes {
		chart.Dots = append(chart.Dots, hillchart.Dot{
			Label:    scope.Title,
			Note:     scope.Note,
			Color:    scope.Color,
			Position: scope.Position,
			Stalled:  time.Since(scope.MovedAt) > stalledAfter && scope.Position < 100,
		})
	}
	return chart
}
