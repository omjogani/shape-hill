package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

// fakeVerify stands in for Supabase: it treats the token string as the auth
// user id, so a test picks an identity by choosing its bearer token. "bad" fails.
func fakeVerify(_ context.Context, token string) (Identity, error) {
	if token == "bad" {
		return Identity{}, errors.New("bad token")
	}
	return Identity{AuthUserID: token, Email: "caller@example.com"}, nil
}

func getWithToken(t *testing.T, url, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func TestMeRejectsMissingToken(t *testing.T) {
	srv, _, _ := testServer(t)

	if resp := getWithToken(t, srv.URL+"/api/me", ""); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no token = %d, want 401", resp.StatusCode)
	}
}

func TestMeRejectsInvalidToken(t *testing.T) {
	srv, _, _ := testServer(t)

	if resp := getWithToken(t, srv.URL+"/api/me", "bad"); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("bad token = %d, want 401", resp.StatusCode)
	}
}

func TestMeReportsNotOnboarded(t *testing.T) {
	srv, _, _ := testServer(t)

	// A valid token whose sub matches no users row: authenticated, not onboarded.
	// Any well-formed uuid works — CreateUser leaves auth_user_id NULL, so none match.
	resp := getWithToken(t, srv.URL+"/api/me", "00000000-0000-0000-0000-0000000000ab")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("valid token = %d, want 200", resp.StatusCode)
	}
	var body struct {
		Onboarded bool `json:"onboarded"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Onboarded {
		t.Fatal("a caller with no users row must report onboarded=false")
	}
}
