package server

import (
	"errors"
	"net/http"

	"github.com/omjogani/shape-hill/api/internal/account"
)

func (s *Server) listTokens(w http.ResponseWriter, r *http.Request) {
	tokens, err := s.store.ListAPITokens(r.Context(), ownerID(r))
	if err != nil {
		s.log.Error("list api tokens", "err", err)
		writeError(w, http.StatusInternalServerError, "could not load tokens")
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

func (s *Server) createToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if !decode(w, r, &body) {
		return
	}
	name, ok := account.CleanTokenName(body.Name)
	if !ok {
		writeError(w, http.StatusBadRequest, "name is required, up to 64 characters")
		return
	}

	raw, hint, hash := account.NewAPIToken()
	token, err := s.store.CreateAPIToken(r.Context(), ownerID(r), name, hint, hash)
	if err != nil {
		s.log.Error("create api token", "err", err)
		writeError(w, http.StatusInternalServerError, "could not create token")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"token": raw, "api_token": token})
}

func (s *Server) deleteToken(w http.ResponseWriter, r *http.Request) {
	err := s.store.DeleteAPIToken(r.Context(), r.PathValue("id"), ownerID(r))
	if errors.Is(err, account.ErrNotFound) {
		writeError(w, http.StatusNotFound, "token not found")
		return
	}
	if err != nil {
		s.log.Error("delete api token", "err", err)
		writeError(w, http.StatusInternalServerError, "could not delete token")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
