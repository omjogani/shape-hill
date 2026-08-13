package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/omjogani/shape-hill/internal/store"
)

func testServer(t *testing.T) (*httptest.Server, *store.Store, store.User, string) {
	t.Helper()

	url := "postgres://postgres:postgres@localhost:5432/shapehill?sslmode=disable"
	db, err := store.New(context.Background(), url)
	if err != nil {
		t.Skipf("no local database (docker compose up -d): %v", err)
	}
	t.Cleanup(db.Close)

	quiet := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := httptest.NewServer(New(db, quiet, fakeVerify))
	t.Cleanup(srv.Close)

	token := newAuthID(t)
	unique := strconv.FormatInt(time.Now().UnixNano(), 36)
	user, err := db.OnboardUser(context.Background(), token, token+"@example.com", "srv"+unique, "Server Test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.DeleteUser(context.Background(), user.ID) })

	return srv, db, user, token
}

func patchWithToken(t *testing.T, url, token string, body any) *http.Response {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPatch, url, strings.NewReader(string(payload)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func TestHealth(t *testing.T) {
	srv, _, _, _ := testServer(t)

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("healthz = %d, want 200", resp.StatusCode)
	}
}

func TestEmbedServesSVGAndHonoursETag(t *testing.T) {
	srv, _, _, token := testServer(t)

	_, hill := postWithToken(t, srv.URL+"/api/hills", token, map[string]any{
		"slug": "billing-v2", "title": "Billing v2", "is_public": true,
	})
	slug := hill["Slug"].(string)

	_, scope := postWithToken(t, srv.URL+"/api/hills/"+slug+"/scopes", token, map[string]any{
		"title": "Card on file", "color": "#2F4C64",
	})
	scopeID := scope["ID"].(string)

	resp, err := http.Get(srv.URL + "/hill/" + slug + ".svg")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("embed = %d, want 200", resp.StatusCode)
	}
	if contentType := resp.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "image/svg+xml") {
		t.Errorf("Content-Type = %q, want image/svg+xml", contentType)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Card on file") {
		t.Error("the rendered chart should name its scope")
	}

	etag := resp.Header.Get("ETag")
	if etag == "" {
		t.Fatal("embed must send an ETag: it is what lets an image proxy skip a redraw")
	}

	// An unchanged hill must answer 304, or every proxy refetch redraws it.
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/hill/"+slug+".svg", nil)
	req.Header.Set("If-None-Match", etag)
	cached, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer cached.Body.Close()
	if cached.StatusCode != http.StatusNotModified {
		t.Fatalf("unchanged hill = %d, want 304 Not Modified", cached.StatusCode)
	}

	// Moving a dot must change the ETag, or the ticket never updates.
	moved, _ := postWithToken(t, srv.URL+"/api/scopes/"+scopeID+"/positions", token, map[string]any{
		"position": 62, "note": "wiring the settings screen",
	})
	if moved.StatusCode != http.StatusNoContent {
		t.Fatalf("move scope = %d, want 204", moved.StatusCode)
	}

	after, err := http.Get(srv.URL + "/hill/" + slug + ".svg")
	if err != nil {
		t.Fatal(err)
	}
	defer after.Body.Close()

	if after.Header.Get("ETag") == etag {
		t.Fatal("the ETag must change once a dot moves, or embeds would serve a stale chart forever")
	}
	updated, _ := io.ReadAll(after.Body)
	if !strings.Contains(string(updated), "wiring the settings screen") {
		t.Error("the redrawn chart should carry the new note")
	}
}

func TestEmbedHidesPrivateHills(t *testing.T) {
	srv, _, _, token := testServer(t)

	_, hill := postWithToken(t, srv.URL+"/api/hills", token, map[string]any{
		"slug": "secret-roadmap", "title": "Secret roadmap", "is_public": false,
	})
	slug := hill["Slug"].(string)

	resp, err := http.Get(srv.URL + "/hill/" + slug + ".svg")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("private hill embed = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "private") {
		t.Error("a private hill's embed should say it is private")
	}
	if strings.Contains(string(body), "Secret roadmap") {
		t.Error("a private hill's title must not leak in its embed")
	}
}

func TestEmbedTogglesWithVisibility(t *testing.T) {
	srv, _, _, token := testServer(t)

	_, hill := postWithToken(t, srv.URL+"/api/hills", token, map[string]any{
		"slug": "billing-v2", "title": "Billing v2", "is_public": false,
	})
	slug := hill["Slug"].(string)

	patchWithToken(t, srv.URL+"/api/hills/"+slug, token, map[string]any{"is_public": true})

	resp, err := http.Get(srv.URL + "/hill/" + slug + ".svg")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(body), "private") {
		t.Error("after being made public, the embed should render the chart, not the private card")
	}
}

func TestEmbedUnknownSlug(t *testing.T) {
	srv, _, _, _ := testServer(t)

	resp, err := http.Get(srv.URL + "/hill/nosuchhill.svg")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown slug = %d, want 404", resp.StatusCode)
	}
}

func TestPublicHillServesPublicChart(t *testing.T) {
	srv, _, _, token := testServer(t)

	_, hill := postWithToken(t, srv.URL+"/api/hills", token, map[string]any{
		"slug": "billing-v2", "title": "Billing v2", "is_public": true,
	})
	slug := hill["Slug"].(string)

	_, scope := postWithToken(t, srv.URL+"/api/hills/"+slug+"/scopes", token, map[string]any{
		"title": "Card on file", "color": "#2F4C64",
	})
	scopeID := scope["ID"].(string)
	postWithToken(t, srv.URL+"/api/scopes/"+scopeID+"/positions", token, map[string]any{
		"position": 40, "note": "wiring it up",
	})

	resp := getWithToken(t, srv.URL+"/api/public/hills/"+slug, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("public hill = %d, want 200", resp.StatusCode)
	}
	var body struct {
		Hill   map[string]any
		Scopes []map[string]any
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Hill["Title"] != "Billing v2" {
		t.Fatalf("public hill title = %v, want Billing v2", body.Hill["Title"])
	}
	if len(body.Scopes) != 1 || body.Scopes[0]["Title"] != "Card on file" {
		t.Fatalf("public hill scopes = %v, want one scope named Card on file", body.Scopes)
	}

	snaps := getWithToken(t, srv.URL+"/api/public/scopes/"+scopeID+"/positions", "")
	if snaps.StatusCode != http.StatusOK {
		t.Fatalf("public snapshots = %d, want 200", snaps.StatusCode)
	}
	var snapshots []map[string]any
	if err := json.NewDecoder(snaps.Body).Decode(&snapshots); err != nil {
		t.Fatal(err)
	}
	if len(snapshots) != 1 || snapshots[0]["Note"] != "wiring it up" {
		t.Fatalf("public snapshots = %v, want one snapshot noting wiring it up", snapshots)
	}
}

func TestPublicHillHidesPrivate(t *testing.T) {
	srv, _, _, token := testServer(t)

	_, hill := postWithToken(t, srv.URL+"/api/hills", token, map[string]any{
		"slug": "secret-roadmap", "title": "Secret roadmap", "is_public": false,
	})
	slug := hill["Slug"].(string)
	_, scope := postWithToken(t, srv.URL+"/api/hills/"+slug+"/scopes", token, map[string]any{"title": "Secret scope"})
	scopeID := scope["ID"].(string)
	postWithToken(t, srv.URL+"/api/scopes/"+scopeID+"/positions", token, map[string]any{
		"position": 30, "note": "should stay hidden",
	})

	if resp := getWithToken(t, srv.URL+"/api/public/hills/"+slug, ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("public GET of private hill = %d, want 404", resp.StatusCode)
	}

	snaps := getWithToken(t, srv.URL+"/api/public/scopes/"+scopeID+"/positions", "")
	if snaps.StatusCode != http.StatusOK {
		t.Fatalf("public snapshots for private hill = %d, want 200 with an empty list", snaps.StatusCode)
	}
	var snapshots []map[string]any
	_ = json.NewDecoder(snaps.Body).Decode(&snapshots)
	if len(snapshots) != 0 {
		t.Fatalf("public snapshots for a private hill = %v, want none", snapshots)
	}
}

func TestPublicHillUnknownSlug(t *testing.T) {
	srv, _, _, _ := testServer(t)

	resp := getWithToken(t, srv.URL+"/api/public/hills/nosuchhill", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("public GET of unknown slug = %d, want 404", resp.StatusCode)
	}
}

func TestMoveScopeRejectsPositionOutOfRange(t *testing.T) {
	srv, _, _, token := testServer(t)

	_, hill := postWithToken(t, srv.URL+"/api/hills", token, map[string]any{
		"slug": "billing-v2", "title": "Billing v2", "is_public": true,
	})
	_, scope := postWithToken(t, srv.URL+"/api/hills/"+hill["Slug"].(string)+"/scopes", token, map[string]any{"title": "Refunds"})

	resp, _ := postWithToken(t, srv.URL+"/api/scopes/"+scope["ID"].(string)+"/positions", token, map[string]any{"position": 140})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("position 140 = %d, want 400", resp.StatusCode)
	}
}

func TestCreateHillRequiresTitle(t *testing.T) {
	srv, _, _, token := testServer(t)

	resp, _ := postWithToken(t, srv.URL+"/api/hills", token, map[string]any{"slug": "no-title"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("hill without a title = %d, want 400", resp.StatusCode)
	}
}

// TestHillsAreOwnerScoped is the point of the whole feature: one user cannot see
// or touch another's hills or their scopes.
func TestHillsAreOwnerScoped(t *testing.T) {
	srv, db, _, ownerTok := testServer(t)

	intruderTok := newAuthID(t)
	intruder, err := db.OnboardUser(context.Background(), intruderTok, intruderTok+"@example.com", "intruder"+shortID(), "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.DeleteUser(context.Background(), intruder.ID) })

	slug := "owned-" + shortID()
	postWithToken(t, srv.URL+"/api/hills", ownerTok, map[string]any{"slug": slug, "title": "Owned", "is_public": false})
	_, scope := postWithToken(t, srv.URL+"/api/hills/"+slug+"/scopes", ownerTok, map[string]any{"title": "Secret scope"})
	scopeID := scope["ID"].(string)

	if resp := getWithToken(t, srv.URL+"/api/hills/"+slug, intruderTok); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("intruder GET hill = %d, want 404", resp.StatusCode)
	}
	if resp := patchWithToken(t, srv.URL+"/api/hills/"+slug, intruderTok, map[string]any{"title": "Hijacked"}); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("intruder PATCH hill = %d, want 404", resp.StatusCode)
	}
	if resp, _ := postWithToken(t, srv.URL+"/api/scopes/"+scopeID+"/positions", intruderTok, map[string]any{"position": 50}); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("intruder move scope = %d, want 404", resp.StatusCode)
	}

	listResp := getWithToken(t, srv.URL+"/api/hills", intruderTok)
	var hills []map[string]any
	_ = json.NewDecoder(listResp.Body).Decode(&hills)
	for _, h := range hills {
		if h["Slug"] == slug {
			t.Fatal("intruder's hill list must not include another user's hill")
		}
	}
}

// TestApiRequiresOnboardedCaller: a valid token with no account gets 403, not data.
func TestApiRequiresOnboardedCaller(t *testing.T) {
	srv, _, _, _ := testServer(t)

	if resp := getWithToken(t, srv.URL+"/api/hills", newAuthID(t)); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("un-onboarded caller = %d, want 403", resp.StatusCode)
	}
}
