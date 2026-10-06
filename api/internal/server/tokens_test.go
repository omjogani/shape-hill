package server

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func deleteWithToken(t *testing.T, url, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func mintToken(t *testing.T, base, session string) (raw, id string) {
	t.Helper()
	resp, body := postWithToken(t, base+"/api/tokens", session, map[string]any{"name": "cli"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create token = %d, want 201", resp.StatusCode)
	}
	raw, _ = body["token"].(string)
	meta, _ := body["api_token"].(map[string]any)
	id, _ = meta["ID"].(string)
	if raw == "" || id == "" {
		t.Fatalf("create token should return the raw token and its record, got %v", body)
	}
	return raw, id
}

func TestAPITokenActsAsItsOwner(t *testing.T) {
	srv, _, user, session := testServer(t)
	raw, _ := mintToken(t, srv.URL, session)

	resp := getWithToken(t, srv.URL+"/api/me", raw)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/api/me with token = %d, want 200", resp.StatusCode)
	}
	var me struct {
		Onboarded bool `json:"onboarded"`
		User      struct{ ID string }
	}
	_ = json.NewDecoder(resp.Body).Decode(&me)
	if !me.Onboarded || me.User.ID != user.ID {
		t.Fatalf("token should resolve to its owner %s, got %+v", user.ID, me)
	}

	created, _ := postWithToken(t, srv.URL+"/api/hills", raw, map[string]any{
		"slug": "cli-" + shortID(), "title": "From the CLI",
	})
	if created.StatusCode != http.StatusCreated {
		t.Fatalf("create hill with token = %d, want 201", created.StatusCode)
	}
}

func TestAPITokenListRecordsUseAndHidesSecret(t *testing.T) {
	srv, _, _, session := testServer(t)
	raw, id := mintToken(t, srv.URL, session)
	getWithToken(t, srv.URL+"/api/hills", raw)

	resp := getWithToken(t, srv.URL+"/api/tokens", session)
	var tokens []map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&tokens)
	if len(tokens) != 1 || tokens[0]["ID"] != id {
		t.Fatalf("list should hold the one minted token, got %v", tokens)
	}
	if tokens[0]["LastUsedAt"] == nil {
		t.Error("using a token should record last_used_at")
	}
	for _, v := range tokens[0] {
		if v == raw {
			t.Fatal("the raw token must never be listed")
		}
	}
}

func TestAPITokenStopsWorkingOnceDeleted(t *testing.T) {
	srv, _, _, session := testServer(t)
	raw, id := mintToken(t, srv.URL, session)

	if resp := deleteWithToken(t, srv.URL+"/api/tokens/"+id, session); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete token = %d, want 204", resp.StatusCode)
	}
	if resp := getWithToken(t, srv.URL+"/api/hills", raw); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("deleted token = %d, want 401", resp.StatusCode)
	}
}

func TestUnknownAPITokenIsRejected(t *testing.T) {
	srv, _, _, _ := testServer(t)

	if resp := getWithToken(t, srv.URL+"/api/hills", "shk_notarealtoken"); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unknown token = %d, want 401", resp.StatusCode)
	}
}

func TestAPITokenCannotManageTokens(t *testing.T) {
	srv, _, _, session := testServer(t)
	raw, id := mintToken(t, srv.URL, session)

	if resp, _ := postWithToken(t, srv.URL+"/api/tokens", raw, map[string]any{"name": "escalate"}); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("mint with token = %d, want 403", resp.StatusCode)
	}
	if resp := getWithToken(t, srv.URL+"/api/tokens", raw); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("list with token = %d, want 403", resp.StatusCode)
	}
	if resp := deleteWithToken(t, srv.URL+"/api/tokens/"+id, raw); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("delete with token = %d, want 403", resp.StatusCode)
	}
}

func TestTokensAreOwnerScoped(t *testing.T) {
	srv, db, _, session := testServer(t)
	_, id := mintToken(t, srv.URL, session)

	other := newAuthID(t)
	resp, body := postWithToken(t, srv.URL+"/api/onboard", other, map[string]any{"username": "other" + shortID()})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("onboard other = %d, want 201", resp.StatusCode)
	}
	otherID := body["ID"].(string)
	t.Cleanup(func() { _ = db.DeleteUser(context.Background(), otherID) })

	if resp := deleteWithToken(t, srv.URL+"/api/tokens/"+id, other); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("deleting another user's token = %d, want 404", resp.StatusCode)
	}
}

func TestCreateTokenRequiresName(t *testing.T) {
	srv, _, _, session := testServer(t)

	if resp, _ := postWithToken(t, srv.URL+"/api/tokens", session, map[string]any{"name": "  "}); resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("blank name = %d, want 400", resp.StatusCode)
	}
}
