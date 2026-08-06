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

// Identity is who a verified Supabase access token says the caller is.
type Identity struct {
	AuthUserID string
	Email      string
}

// VerifyToken checks a Supabase access token and returns its identity. It is an
// injection point so tests can verify the middleware without reaching Supabase.
type VerifyToken func(ctx context.Context, token string) (Identity, error)

// caller is what authenticate attaches to the request context: the verified
// identity, plus the local account — which is nil until the caller has onboarded.
type caller struct {
	Identity
	User *store.User
}

type ctxKey int

const callerKey ctxKey = 0

// authenticate rejects requests without a valid token, then resolves the local
// user (leaving it nil for an authenticated-but-not-onboarded caller).
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

		c := caller{Identity: id}
		user, err := s.store.UserByAuthID(r.Context(), id.AuthUserID)
		switch {
		case errors.Is(err, store.ErrNotFound):
			// Authenticated but not onboarded — User stays nil.
		case err != nil:
			s.log.Error("resolve user", "err", err)
			writeError(w, http.StatusInternalServerError, "could not resolve user")
			return
		default:
			c.User = &user
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

// me tells the frontend whether this caller has a local account yet, so it can
// route them to onboarding or into the app.
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	c, _ := callerFrom(r.Context())
	if c.User == nil {
		writeJSON(w, http.StatusOK, map[string]any{"onboarded": false, "email": c.Email})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"onboarded": true, "user": c.User})
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

	return func(_ context.Context, raw string) (Identity, error) {
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
			return Identity{}, err
		}
		if claims.Subject == "" {
			return Identity{}, errors.New("token has no subject")
		}
		return Identity{AuthUserID: claims.Subject, Email: claims.Email}, nil
	}, nil
}
