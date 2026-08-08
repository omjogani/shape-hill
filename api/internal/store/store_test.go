package store

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"
)

// testStore connects to the local database. Without one, the integration tests
// skip rather than fail: `go test ./...` still works on a machine with no Docker.
func testStore(t *testing.T) *Store {
	t.Helper()

	url := "postgres://postgres:postgres@localhost:5432/shapehill?sslmode=disable"
	st, err := New(context.Background(), url)
	if err != nil {
		t.Skipf("no local database (docker compose up -d): %v", err)
	}
	t.Cleanup(st.Close)
	return st
}

// testUser creates a user and removes it afterwards, taking its hills, scopes and
// positions with it through the cascade.
func testUser(t *testing.T, st *Store) User {
	t.Helper()

	unique := strconv.FormatInt(time.Now().UnixNano(), 36)
	user, err := st.CreateUser(context.Background(),
		"test-"+unique+"@example.com", "user-"+unique, "Test User")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	t.Cleanup(func() {
		_, _ = st.pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1::uuid`, user.ID)
	})
	return user
}

func TestCreateHillRejectsDuplicateSlug(t *testing.T) {
	st := testStore(t)
	ctx := context.Background()
	user := testUser(t, st)

	first, err := st.CreateHill(ctx, user.ID, "billing-v2", "Billing v2", "", true)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := st.CreateHill(ctx, user.ID, "billing-v2", "Billing v2 again", "", true); !errors.Is(err, ErrSlugTaken) {
		t.Fatalf("want ErrSlugTaken for a duplicate slug, got %v", err)
	}

	found, err := st.HillBySlug(ctx, "billing-v2")
	if err != nil {
		t.Fatal(err)
	}
	if found.ID != first.ID || found.Title != "Billing v2" {
		t.Fatalf("HillBySlug returned the wrong hill: %+v", found)
	}
}

func TestListHillsIncludesCreatedHill(t *testing.T) {
	st := testStore(t)
	ctx := context.Background()
	user := testUser(t, st)

	hill, err := st.CreateHill(ctx, user.ID, "list-hills", "Billing v2", "", true)
	if err != nil {
		t.Fatal(err)
	}

	hills, err := st.ListHillsByOwner(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, h := range hills {
		if h.ID == hill.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("ListHillsByOwner omitted the created hill %s", hill.Slug)
	}
}

func TestHillBySlugReportsNotFound(t *testing.T) {
	st := testStore(t)

	_, err := st.HillBySlug(context.Background(), "does-not-exist")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound for an unknown slug, got %v", err)
	}
}

func TestUnmovedScopeSitsAtZero(t *testing.T) {
	st := testStore(t)
	ctx := context.Background()
	user := testUser(t, st)

	hill, err := st.CreateHill(ctx, user.ID, "unmoved-scope", "Billing v2", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.CreateScope(ctx, hill.ID, "Card on file", "", "#2F4C64", 1); err != nil {
		t.Fatal(err)
	}

	scopes, err := st.ScopesForHill(ctx, hill.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(scopes) != 1 {
		t.Fatalf("want 1 scope, got %d", len(scopes))
	}
	if scopes[0].Position != 0 {
		t.Errorf("a scope nobody has moved should sit at 0, got %d", scopes[0].Position)
	}
}

func TestMoveScopeAppendsAndLatestWins(t *testing.T) {
	st := testStore(t)
	ctx := context.Background()
	user := testUser(t, st)

	hill, err := st.CreateHill(ctx, user.ID, "move-scope", "Billing v2", "", true)
	if err != nil {
		t.Fatal(err)
	}
	scope, err := st.CreateScope(ctx, hill.ID, "Card on file", "", "#2F4C64", 1)
	if err != nil {
		t.Fatal(err)
	}

	for _, move := range []struct {
		position int16
		note     string
	}{{8, "spiking"}, {45, "approach proven"}, {95, "shipped"}} {
		if err := st.MoveScope(ctx, scope.ID, move.position, move.note, user.ID); err != nil {
			t.Fatal(err)
		}
	}

	scopes, err := st.ScopesForHill(ctx, hill.ID)
	if err != nil {
		t.Fatal(err)
	}
	if scopes[0].Position != 95 || scopes[0].Note != "shipped" {
		t.Fatalf("the newest position should win, got %d %q", scopes[0].Position, scopes[0].Note)
	}

	var history int
	err = st.pool.QueryRow(ctx,
		`SELECT count(*) FROM scope_positions WHERE scope_id = $1::uuid`, scope.ID).Scan(&history)
	if err != nil {
		t.Fatal(err)
	}
	if history != 3 {
		t.Errorf("moves must append, not overwrite: want 3 rows of history, got %d", history)
	}
}

func TestMoveUnknownScopeReportsNotFound(t *testing.T) {
	st := testStore(t)

	zero := "00000000-0000-0000-0000-000000000000"
	err := st.MoveScope(context.Background(), zero, 50, "", zero)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound moving a scope that does not exist, got %v", err)
	}
}

func TestLastMovedOnTracksTheNewestMovement(t *testing.T) {
	hill := Hill{UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	older := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	newest := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	got := LastMovedOn(hill, []Scope{{MovedAt: older}, {MovedAt: newest}, {MovedAt: older}})
	if !got.Equal(newest) {
		t.Errorf("LastMovedOn = %v, want the newest movement %v", got, newest)
	}

	if got := LastMovedOn(hill, nil); !got.Equal(hill.UpdatedAt) {
		t.Errorf("with no scopes it should fall back to the hill itself, got %v", got)
	}
}
