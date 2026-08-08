package server

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fakeVerify stands in for Supabase: it treats the token string as the auth
// user id and derives a matching email, so a test picks an identity by choosing
// its bearer token. "bad" fails verification.
func fakeVerify(_ context.Context, token string) (AuthUser, error) {
	if token == "bad" {
		return AuthUser{}, errors.New("bad token")
	}
	return AuthUser{ID: token, Email: token + "@example.com"}, nil
}

// newAuthID is a fresh uuid to use as a bearer token (fakeVerify treats it as sub).
func newAuthID(t *testing.T) string {
	t.Helper()
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

var idCounter atomic.Uint64

// shortID is a run-unique suffix for usernames (nanotime guards across reruns,
// the counter guards within one).
func shortID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36) + strconv.FormatUint(idCounter.Add(1), 36)
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

func postWithToken(t *testing.T, url, token string, body any) (*http.Response, map[string]any) {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(payload)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })

	var decoded map[string]any
	if resp.StatusCode != http.StatusNoContent {
		_ = json.NewDecoder(resp.Body).Decode(&decoded)
	}
	return resp, decoded
}

func TestMeRejectsMissingToken(t *testing.T) {
	srv, _, _, _ := testServer(t)

	if resp := getWithToken(t, srv.URL+"/api/me", ""); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no token = %d, want 401", resp.StatusCode)
	}
}

func TestMeRejectsInvalidToken(t *testing.T) {
	srv, _, _, _ := testServer(t)

	if resp := getWithToken(t, srv.URL+"/api/me", "bad"); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("bad token = %d, want 401", resp.StatusCode)
	}
}

func TestMeReportsNotOnboarded(t *testing.T) {
	srv, _, _, _ := testServer(t)

	resp := getWithToken(t, srv.URL+"/api/me", newAuthID(t))
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

func TestOnboardCreatesAccount(t *testing.T) {
	srv, db, _, _ := testServer(t)
	token := newAuthID(t)

	resp, body := postWithToken(t, srv.URL+"/api/onboard", token, map[string]any{"username": "alice" + shortID()})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("onboard = %d, want 201", resp.StatusCode)
	}
	id, _ := body["ID"].(string)
	if id == "" {
		t.Fatal("onboard should return the created user")
	}
	t.Cleanup(func() { _ = db.DeleteUser(context.Background(), id) })

	me := getWithToken(t, srv.URL+"/api/me", token)
	var decoded struct {
		Onboarded bool `json:"onboarded"`
	}
	_ = json.NewDecoder(me.Body).Decode(&decoded)
	if !decoded.Onboarded {
		t.Fatal("after onboarding, /api/me must report onboarded=true")
	}
}

func TestOnboardLinksByEmail(t *testing.T) {
	srv, db, _, _ := testServer(t)
	token := newAuthID(t)

	// A pre-auth row already exists for this token's email (auth_user_id still NULL).
	existing, err := db.CreateUser(context.Background(), token+"@example.com", "legacy"+shortID(), "Legacy")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.DeleteUser(context.Background(), existing.ID) })

	resp, body := postWithToken(t, srv.URL+"/api/onboard", token, map[string]any{"username": "newname" + shortID()})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("onboard(link) = %d, want 201", resp.StatusCode)
	}
	if body["ID"] != existing.ID {
		t.Fatalf("link-by-email should adopt existing row %s, got %v", existing.ID, body["ID"])
	}
	if body["Username"] != existing.Username {
		t.Fatalf("linking keeps existing username %q, got %v", existing.Username, body["Username"])
	}
}

func TestOnboardRejectsDuplicateUsername(t *testing.T) {
	srv, db, _, _ := testServer(t)
	taken := "dup" + shortID()

	first, err := db.CreateUser(context.Background(), newAuthID(t)+"@example.com", taken, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.DeleteUser(context.Background(), first.ID) })

	// Different email (so no link), same username → conflict.
	resp, _ := postWithToken(t, srv.URL+"/api/onboard", newAuthID(t), map[string]any{"username": taken})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate username = %d, want 409", resp.StatusCode)
	}
}

func TestOnboardRejectsWhenAlreadyOnboarded(t *testing.T) {
	srv, db, _, _ := testServer(t)
	token := newAuthID(t)

	resp, body := postWithToken(t, srv.URL+"/api/onboard", token, map[string]any{"username": "once" + shortID()})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("first onboard = %d, want 201", resp.StatusCode)
	}
	t.Cleanup(func() { _ = db.DeleteUser(context.Background(), body["ID"].(string)) })

	again, _ := postWithToken(t, srv.URL+"/api/onboard", token, map[string]any{"username": "twice" + shortID()})
	if again.StatusCode != http.StatusConflict {
		t.Fatalf("second onboard = %d, want 409", again.StatusCode)
	}
}

func TestOnboardRejectsBadUsername(t *testing.T) {
	srv, _, _, _ := testServer(t)

	resp, _ := postWithToken(t, srv.URL+"/api/onboard", newAuthID(t), map[string]any{"username": "ab"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("short username = %d, want 400", resp.StatusCode)
	}
}
