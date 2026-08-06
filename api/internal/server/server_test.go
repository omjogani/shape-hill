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

func testServer(t *testing.T) (*httptest.Server, *store.Store, store.User) {
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

	unique := strconv.FormatInt(time.Now().UnixNano(), 36)
	user, err := db.CreateUser(context.Background(), "srv-"+unique+"@example.com", "srv-"+unique, "Server Test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.DeleteUser(context.Background(), user.ID) })

	return srv, db, user
}

func post(t *testing.T, url string, body any) (*http.Response, map[string]any) {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(url, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })

	var decoded map[string]any
	if resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
			t.Fatalf("decode response from %s: %v", url, err)
		}
	}
	return resp, decoded
}

func patch(t *testing.T, url string, body any) *http.Response {
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
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func TestHealth(t *testing.T) {
	srv, _, _ := testServer(t)

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
	srv, _, user := testServer(t)

	_, hill := post(t, srv.URL+"/api/hills", map[string]any{
		"owner_id": user.ID, "slug": "billing-v2", "title": "Billing v2", "is_public": true,
	})
	slug := hill["Slug"].(string)

	_, scope := post(t, srv.URL+"/api/hills/"+slug+"/scopes", map[string]any{
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
	moved, _ := post(t, srv.URL+"/api/scopes/"+scopeID+"/positions", map[string]any{
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
	srv, _, user := testServer(t)

	_, hill := post(t, srv.URL+"/api/hills", map[string]any{
		"owner_id": user.ID, "slug": "secret-roadmap", "title": "Secret roadmap", "is_public": false,
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
	srv, _, user := testServer(t)

	_, hill := post(t, srv.URL+"/api/hills", map[string]any{
		"owner_id": user.ID, "slug": "billing-v2", "title": "Billing v2", "is_public": false,
	})
	slug := hill["Slug"].(string)

	patch(t, srv.URL+"/api/hills/"+slug, map[string]any{"is_public": true})

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
	srv, _, _ := testServer(t)

	resp, err := http.Get(srv.URL + "/hill/nosuchhill.svg")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown slug = %d, want 404", resp.StatusCode)
	}
}

func TestMoveScopeRejectsPositionOutOfRange(t *testing.T) {
	srv, _, user := testServer(t)

	_, hill := post(t, srv.URL+"/api/hills", map[string]any{
		"owner_id": user.ID, "slug": "billing-v2", "title": "Billing v2", "is_public": true,
	})
	_, scope := post(t, srv.URL+"/api/hills/"+hill["Slug"].(string)+"/scopes", map[string]any{"title": "Refunds"})

	resp, _ := post(t, srv.URL+"/api/scopes/"+scope["ID"].(string)+"/positions", map[string]any{"position": 140})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("position 140 = %d, want 400", resp.StatusCode)
	}
}

func TestCreateHillRequiresTitle(t *testing.T) {
	srv, _, user := testServer(t)

	resp, _ := post(t, srv.URL+"/api/hills", map[string]any{"owner_id": user.ID})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("hill without a title = %d, want 400", resp.StatusCode)
	}
}
