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
// server-side by image proxies, not by a browser we control: no cookies, no auth,
// and an ETag so a proxy can be told "nothing moved" without us drawing anything.
func (s *Server) embed(w http.ResponseWriter, r *http.Request) {
	// ServeMux wildcards match whole segments, so the route captures "slug.svg"
	// and the extension comes off here.
	slug, ok := strings.CutSuffix(r.PathValue("file"), ".svg")
	if !ok {
		http.NotFound(w, r)
		return
	}

	hill, err := s.store.HillBySlug(r.Context(), slug)
	if errors.Is(err, store.ErrNotFound) || (err == nil && !hill.IsPublic) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		s.log.Error("load hill for embed", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	scopes, err := s.store.ScopesForHill(r.Context(), hill.ID)
	if err != nil {
		s.log.Error("load scopes for embed", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	etag := fmt.Sprintf(`W/"%d"`, store.LastMovedOn(hill, scopes).UnixNano())
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age=60, must-revalidate")
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Write(hillchart.Render(chartOf(hill, scopes)))
}

func chartOf(hill store.Hill, scopes []store.Scope) hillchart.Chart {
	chart := hillchart.Chart{Title: hill.Title}
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
