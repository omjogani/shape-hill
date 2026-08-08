package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/omjogani/shape-hill/internal/store"
)

type AuthUser struct {
	ID    string
	Email string
}

type VerifyToken func(ctx context.Context, token string) (AuthUser, error)

type caller struct {
	AuthUser
	Account *store.User
}

type ctxKey int

const callerKey ctxKey = 0

func (s *Server) authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearer(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		id, err := s.verify(r.Context(), token)
		if err != nil {
			s.log.Info("token rejected", "err", err)
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		c := caller{AuthUser: id}
		user, err := s.store.UserByAuthID(r.Context(), id.ID)
		switch {
		case errors.Is(err, store.ErrNotFound):
			// Authenticated but not onboarded — Account stays nil.
		case err != nil:
			s.log.Error("resolve user", "err", err)
			writeError(w, http.StatusInternalServerError, "could not resolve user")
			return
		default:
			c.Account = &user
		}

		ctx := context.WithValue(r.Context(), callerKey, c)
		next(w, r.WithContext(ctx))
	}
}

func callerFrom(ctx context.Context) (caller, bool) {
	c, ok := ctx.Value(callerKey).(caller)
	return c, ok
}

func bearer(r *http.Request) (string, bool) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || token == "" {
		return "", false
	}
	return token, true
}

func (s *Server) currentUser(w http.ResponseWriter, r *http.Request) {
	c, _ := callerFrom(r.Context())
	if c.Account == nil {
		writeJSON(w, http.StatusOK, map[string]any{"onboarded": false, "email": c.Email})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"onboarded": true, "user": c.Account})
}

// NewSupabaseVerifier verifies access tokens against the project's public JWKS.
// ponytail: assumes the project signs with asymmetric JWT keys (Supabase's
// default). A legacy HS256 project publishes no JWKS — add the shared secret here
// and verify HS256 if a first login fails on signature.
func NewSupabaseVerifier(ctx context.Context, supabaseURL string) (VerifyToken, error) {
	base := strings.TrimRight(supabaseURL, "/")
	issuer := base + "/auth/v1"

	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{issuer + "/.well-known/jwks.json"})
	if err != nil {
		return nil, fmt.Errorf("load jwks: %w", err)
	}

	return func(_ context.Context, raw string) (AuthUser, error) {
		var claims struct {
			Email string `json:"email"`
			jwt.RegisteredClaims
		}
		if _, err := jwt.ParseWithClaims(raw, &claims, jwks.Keyfunc,
			jwt.WithValidMethods([]string{"ES256", "RS256"}),
			jwt.WithIssuer(issuer),
			jwt.WithAudience("authenticated"),
			jwt.WithExpirationRequired(),
		); err != nil {
			return AuthUser{}, err
		}
		if claims.Subject == "" {
			return AuthUser{}, errors.New("token has no subject")
		}
		return AuthUser{ID: claims.Subject, Email: claims.Email}, nil
	}, nil
}
