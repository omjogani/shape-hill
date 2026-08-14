package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/omjogani/shape-hill/internal/store"
)

// Slugs live in URLs and get typed by hand: lowercase words joined by single hyphens.
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (s *Server) listHills(w http.ResponseWriter, r *http.Request) {
	hills, err := s.store.ListHillsByOwner(r.Context(), ownerID(r))
	if err != nil {
		s.log.Error("list hills", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load hills")
		return
	}
	writeJSON(w, http.StatusOK, hills)
}

func (s *Server) createHill(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Slug        string `json:"slug"`
		Title       string `json:"title"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
	}
	if !decode(w, r, &body) {
		return
	}
	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if !slugPattern.MatchString(body.Slug) {
		writeError(w, http.StatusBadRequest, "slug must be lowercase letters, numbers and hyphens")
		return
	}

	hill, err := s.store.CreateHill(r.Context(), ownerID(r), body.Slug, body.Title, body.Description, body.IsPublic)
	if errors.Is(err, store.ErrSlugTaken) {
		writeError(w, http.StatusConflict, "that slug is already taken")
		return
	}
	if err != nil {
		s.log.Error("create hill", "err", err)
		writeError(w, http.StatusInternalServerError, "could not create hill")
		return
	}
	writeJSON(w, http.StatusCreated, hill)
}

func (s *Server) getHill(w http.ResponseWriter, r *http.Request) {
	hill, err := s.store.HillBySlug(r.Context(), r.PathValue("slug"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "hill not found")
		return
	}
	if err != nil {
		s.log.Error("get hill", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load hill")
		return
	}

	if hill.OwnerID != ownerID(r) {
		writeError(w, http.StatusNotFound, "hill not found")
		return
	}

	scopes, err := s.store.ScopesForHill(r.Context(), hill.ID)
	if err != nil {
		s.log.Error("list scopes", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load scopes")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"hill": hill, "scopes": scopes})
}

func (s *Server) getPublicHill(w http.ResponseWriter, r *http.Request) {
	hill, err := s.store.HillBySlug(r.Context(), r.PathValue("slug"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "hill not found")
		return
	}
	if err != nil {
		s.log.Error("get public hill", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load hill")
		return
	}
	if !hill.IsPublic {
		writeError(w, http.StatusNotFound, "hill not found")
		return
	}

	scopes, err := s.store.ScopesForHill(r.Context(), hill.ID)
	if err != nil {
		s.log.Error("list scopes for public hill", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load scopes")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"hill": hill, "scopes": scopes})
}

func (s *Server) updateHill(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title        *string `json:"title"`
		IsPublic     *bool   `json:"is_public"`
		TrackStalled *bool   `json:"track_stalled"`
	}
	if !decode(w, r, &body) {
		return
	}
	if body.Title == nil && body.IsPublic == nil && body.TrackStalled == nil {
		writeError(w, http.StatusBadRequest, "nothing to update")
		return
	}
	if body.Title != nil && *body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	hill, err := s.store.UpdateHill(r.Context(), r.PathValue("slug"), ownerID(r), body.Title, body.IsPublic, body.TrackStalled)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "hill not found")
		return
	}
	if err != nil {
		s.log.Error("update hill", "err", err)
		writeError(w, http.StatusInternalServerError, "could not update hill")
		return
	}
	writeJSON(w, http.StatusOK, hill)
}

func (s *Server) createScope(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Color       string `json:"color"`
		SortOrder   int16  `json:"sort_order"`
	}
	if !decode(w, r, &body) {
		return
	}
	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if body.Color == "" {
		body.Color = "#2F4C64"
	}

	hill, err := s.store.HillBySlug(r.Context(), r.PathValue("slug"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "hill not found")
		return
	}
	if err != nil {
		s.log.Error("get hill", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load hill")
		return
	}
	if hill.OwnerID != ownerID(r) {
		writeError(w, http.StatusNotFound, "hill not found")
		return
	}

	scope, err := s.store.CreateScope(r.Context(), hill.ID, body.Title, body.Description, body.Color, body.SortOrder)
	if err != nil {
		s.log.Error("create scope", "err", err)
		writeError(w, http.StatusInternalServerError, "could not create scope")
		return
	}
	writeJSON(w, http.StatusCreated, scope)
}

func (s *Server) scopeSnapshots(w http.ResponseWriter, r *http.Request) {
	snapshots, err := s.store.SnapshotsForScope(r.Context(), r.PathValue("id"), ownerID(r))
	if err != nil {
		s.log.Error("list snapshots", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load snapshots")
		return
	}
	writeJSON(w, http.StatusOK, snapshots)
}

func (s *Server) publicScopeSnapshots(w http.ResponseWriter, r *http.Request) {
	snapshots, err := s.store.SnapshotsForPublicScope(r.Context(), r.PathValue("id"))
	if err != nil {
		s.log.Error("list public snapshots", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load snapshots")
		return
	}
	writeJSON(w, http.StatusOK, snapshots)
}

func (s *Server) updateScope(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
		Color string `json:"color"`
	}
	if !decode(w, r, &body) {
		return
	}
	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if body.Color == "" {
		body.Color = "#2F4C64"
	}

	err := s.store.UpdateScope(r.Context(), r.PathValue("id"), body.Title, body.Color, ownerID(r))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "scope not found")
		return
	}
	if err != nil {
		s.log.Error("update scope", "err", err)
		writeError(w, http.StatusInternalServerError, "could not update scope")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) deleteScope(w http.ResponseWriter, r *http.Request) {
	err := s.store.ArchiveScope(r.Context(), r.PathValue("id"), ownerID(r))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "scope not found")
		return
	}
	if err != nil {
		s.log.Error("delete scope", "err", err)
		writeError(w, http.StatusInternalServerError, "could not delete scope")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) moveScope(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Position *int16 `json:"position"`
		Note     string `json:"note"`
	}
	if !decode(w, r, &body) {
		return
	}
	if body.Position == nil {
		writeError(w, http.StatusBadRequest, "position is required")
		return
	}
	if *body.Position < 0 || *body.Position > 100 {
		writeError(w, http.StatusBadRequest, "position must be between 0 and 100")
		return
	}

	err := s.store.MoveScope(r.Context(), r.PathValue("id"), *body.Position, body.Note, ownerID(r))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "scope not found")
		return
	}
	if err != nil {
		s.log.Error("move scope", "err", err)
		writeError(w, http.StatusInternalServerError, "could not move scope")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decode(w http.ResponseWriter, r *http.Request, into any) bool {
	if err := json.NewDecoder(r.Body).Decode(into); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}
