package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/omjogani/shape-hill/internal/account"
	"github.com/omjogani/shape-hill/internal/hills"
	"github.com/omjogani/shape-hill/internal/server"
	"github.com/omjogani/shape-hill/internal/store"
)

type env map[string]string

func (e env) get(key string) string { return e[key] }

type result struct {
	code   int
	stdout string
	stderr string
}

func (r result) decode(t *testing.T, into any) {
	t.Helper()
	if err := json.Unmarshal([]byte(r.stdout), into); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, r.stdout)
	}
}

func run(e env, args ...string) result {
	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), args, e.get, &stdout, &stderr)
	return result{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func unique() string { return strconv.FormatInt(time.Now().UnixNano(), 36) }

// testEnv starts the real API over the local database and returns an env whose
// token belongs to a fresh user.
func testEnv(t *testing.T) env {
	t.Helper()

	db, err := store.New(context.Background(), "postgres://postgres:postgres@localhost:5432/shapehill?sslmode=disable")
	if err != nil {
		t.Skipf("no local database (docker compose up -d): %v", err)
	}
	t.Cleanup(db.Close)

	noSessions := func(context.Context, string) (account.AuthUser, error) {
		return account.AuthUser{}, errors.New("sessions not used in cli tests")
	}
	srv := httptest.NewServer(server.New(db, slog.New(slog.NewTextHandler(io.Discard, nil)), noSessions))
	t.Cleanup(srv.Close)

	id := unique()
	user, err := db.CreateUser(context.Background(), "cli-"+id+"@example.com", "cli"+id, "CLI Test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.DeleteUser(context.Background(), user.ID) })

	raw, hint, hash := account.NewAPIToken()
	if _, err := db.CreateAPIToken(context.Background(), user.ID, "cli test", hint, hash); err != nil {
		t.Fatal(err)
	}

	return env{"SHAPEHILL_TOKEN": raw, "SHAPEHILL_API_URL": srv.URL, "SHAPEHILL_WEB_URL": "https://web.test"}
}

func TestCreateHillPrintsEmbedAndIsRetryable(t *testing.T) {
	e := testEnv(t)
	slug := "cli-" + unique()

	first := run(e, "hill", "create", slug, "--title", "Launch v2", "--public", "--json")
	if first.code != exitOK {
		t.Fatalf("create = %d: %s", first.code, first.stderr)
	}
	var created struct {
		Hill    hills.Hill
		Created bool
		Embed   embed
	}
	first.decode(t, &created)
	if !created.Created || created.Hill.Slug != slug || !created.Hill.IsPublic {
		t.Fatalf("unexpected create result: %+v", created)
	}
	if want := "](https://web.test/" + slug + "/view)"; !strings.HasSuffix(created.Embed.Markdown, want) {
		t.Errorf("embed markdown %q should link to the public view", created.Embed.Markdown)
	}

	if again := run(e, "hill", "create", slug, "--title", "Launch v2"); again.code != exitConflict {
		t.Fatalf("duplicate create = %d, want %d", again.code, exitConflict)
	}

	retry := run(e, "hill", "create", slug, "--title", "Launch v2", "--if-not-exists", "--json")
	if retry.code != exitOK {
		t.Fatalf("create --if-not-exists = %d: %s", retry.code, retry.stderr)
	}
	retry.decode(t, &created)
	if created.Created {
		t.Error("--if-not-exists on an existing hill must report created=false")
	}
}

func TestScopeLifecycle(t *testing.T) {
	e := testEnv(t)
	slug := "cli-" + unique()
	if r := run(e, "hill", "create", slug, "--title", "Lifecycle"); r.code != exitOK {
		t.Fatalf("create hill = %d: %s", r.code, r.stderr)
	}

	if r := run(e, "scope", "add", slug, "API integration", "--color", "#336699"); r.code != exitOK {
		t.Fatalf("add scope = %d: %s", r.code, r.stderr)
	}
	if r := run(e, "scope", "add", slug, "api integration"); r.code != exitConflict {
		t.Fatalf("duplicate title = %d, want %d", r.code, exitConflict)
	}
	if r := run(e, "scope", "add", slug, "API integration", "--if-not-exists"); r.code != exitOK {
		t.Fatalf("add --if-not-exists = %d: %s", r.code, r.stderr)
	}

	var move struct {
		From, To int
		Moved    bool
	}
	moved := run(e, "scope", "move", slug, "api INTEGRATION", "62", "--note", "wiring retries", "--json")
	if moved.code != exitOK {
		t.Fatalf("move = %d: %s", moved.code, moved.stderr)
	}
	moved.decode(t, &move)
	if !move.Moved || move.From != 0 || move.To != 62 {
		t.Fatalf("unexpected move: %+v", move)
	}

	repeat := run(e, "scope", "move", slug, "API integration", "62", "--note", "wiring retries", "--json")
	repeat.decode(t, &move)
	if repeat.code != exitOK || move.Moved {
		t.Fatalf("repeating the latest move must be a no-op, got code %d %+v", repeat.code, move)
	}

	var shown struct {
		Scopes []struct {
			Title    string
			Position int
			Phase    string
		}
	}
	show := run(e, "hill", "show", slug, "--json")
	show.decode(t, &shown)
	if len(shown.Scopes) != 1 || shown.Scopes[0].Position != 62 || shown.Scopes[0].Phase != "downhill" {
		t.Fatalf("show should hold one downhill scope at 62, got %+v", shown.Scopes)
	}

	var log struct{ Snapshots []hills.Snapshot }
	run(e, "scope", "log", slug, "API integration", "--json").decode(t, &log)
	if len(log.Snapshots) != 1 || log.Snapshots[0].Note != "wiring retries" {
		t.Fatalf("log should hold exactly the one recorded move, got %+v", log.Snapshots)
	}
}

func TestHillUpdateAndEmbed(t *testing.T) {
	e := testEnv(t)
	slug := "cli-" + unique()
	run(e, "hill", "create", slug, "--title", "Private first")

	private := run(e, "hill", "embed", slug, "--format", "url")
	if private.code != exitOK || !strings.Contains(private.stderr, "private") {
		t.Fatalf("embed of a private hill should warn on stderr, got %d %q", private.code, private.stderr)
	}

	if r := run(e, "hill", "update", slug, "--public"); r.code != exitOK {
		t.Fatalf("update --public = %d: %s", r.code, r.stderr)
	}

	html := run(e, "hill", "embed", slug, "--style", "github", "--format", "html")
	if html.code != exitOK || !strings.Contains(html.stdout, `<img src="`) || !strings.Contains(html.stdout, "style=github") {
		t.Fatalf("html embed = %d %q", html.code, html.stdout)
	}
	if html.stderr != "" {
		t.Errorf("a public hill should embed without a warning, got %q", html.stderr)
	}
}

func TestExitCodes(t *testing.T) {
	e := testEnv(t)
	slug := "cli-" + unique()
	run(e, "hill", "create", slug, "--title", "Codes")

	tests := []struct {
		name string
		env  env
		args []string
		want int
	}{
		{"no token", env{"SHAPEHILL_API_URL": e["SHAPEHILL_API_URL"]}, []string{"hill", "list"}, exitAuth},
		{"revoked or unknown token", env{"SHAPEHILL_API_URL": e["SHAPEHILL_API_URL"], "SHAPEHILL_TOKEN": "shk_nope"}, []string{"hill", "list"}, exitAuth},
		{"unknown hill", e, []string{"hill", "show", "no-such-hill-" + unique()}, exitNotFound},
		{"unknown scope", e, []string{"scope", "move", slug, "Nope", "10"}, exitNotFound},
		{"position out of range", e, []string{"scope", "move", slug, "Nope", "101"}, exitUsage},
		{"bad slug", e, []string{"hill", "create", "Bad Slug", "--title", "x"}, exitUsage},
		{"missing title", e, []string{"hill", "create", "fine-slug"}, exitUsage},
		{"unknown flag", e, []string{"hill", "list", "--nope"}, exitUsage},
		{"wrong arg count", e, []string{"scope", "move", slug}, exitUsage},
		{"bad colour", e, []string{"scope", "add", slug, "X", "--color", "red"}, exitUsage},
		{"unreachable api", env{"SHAPEHILL_API_URL": "http://127.0.0.1:1", "SHAPEHILL_TOKEN": "shk_x"}, []string{"hill", "list"}, exitFailure},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if r := run(tc.env, tc.args...); r.code != tc.want {
				t.Fatalf("exit = %d, want %d (stderr: %s)", r.code, tc.want, r.stderr)
			}
		})
	}
}

func TestJSONErrorsAreMachineReadable(t *testing.T) {
	r := run(env{}, "hill", "list", "--json")

	var problem struct {
		Error string
		Code  int
	}
	if err := json.Unmarshal([]byte(r.stderr), &problem); err != nil {
		t.Fatalf("--json errors must be JSON on stderr, got %q", r.stderr)
	}
	if problem.Code != exitAuth || r.code != exitAuth || problem.Error == "" {
		t.Fatalf("unexpected error payload %+v (exit %d)", problem, r.code)
	}
}

func TestResolveScope(t *testing.T) {
	scopes := []hills.Scope{
		{ID: "a1", Title: "API"},
		{ID: "b2", Title: "UI"},
		{ID: "c3", Title: "ui"},
	}

	if s, err := resolveScope(scopes, "h", "a1"); err != nil || s.ID != "a1" {
		t.Errorf("by ID = %v, %v", s, err)
	}
	if s, err := resolveScope(scopes, "h", " api "); err != nil || s.ID != "a1" {
		t.Errorf("by title = %v, %v", s, err)
	}
	if _, err := resolveScope(scopes, "h", "UI"); exitCode(err) != exitUsage {
		t.Errorf("an ambiguous title must be refused, got %v", err)
	}
	if _, err := resolveScope(scopes, "h", "Docs"); exitCode(err) != exitNotFound {
		t.Errorf("an unknown title must be not-found, got %v", err)
	}
}

func TestHelpAndDiscovery(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
		out  string
	}{
		{"root help", []string{"--help"}, exitOK, "Hill charts"},
		{"bare command shows help", []string{}, exitOK, "Positions (0-100)"},
		{"group help", []string{"scope"}, exitOK, "move"},
		{"agents topic", []string{"help", "agents"}, exitOK, "--if-not-exists"},
		{"version", []string{"version"}, exitOK, "shapehill "},
		{"version flag", []string{"--version"}, exitOK, "shapehill "},
		{"alias", []string{"scope", "mv", "--help"}, exitOK, "<position>"},
		{"unknown subcommand fails", []string{"hill", "frobnicate"}, exitUsage, ""},
		{"unknown command fails", []string{"hils"}, exitUsage, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := run(env{}, tc.args...)
			if r.code != tc.want {
				t.Fatalf("exit = %d, want %d (stderr: %s)", r.code, tc.want, r.stderr)
			}
			if !strings.Contains(r.stdout, tc.out) {
				t.Errorf("stdout should mention %q, got:\n%s", tc.out, r.stdout)
			}
		})
	}
}

func TestUnknownSubcommandSuggests(t *testing.T) {
	r := run(env{}, "scope", "mvoe")
	if r.code != exitUsage || !strings.Contains(r.stderr, "did you mean move") {
		t.Fatalf("a typo should suggest the command, got %d %q", r.code, r.stderr)
	}
}

func TestOutputIsPlainWhenNotATerminal(t *testing.T) {
	r := run(env{}, "--help")
	if strings.Contains(r.stdout, "\x1b[") {
		t.Fatal("help written to a pipe must not contain colour codes")
	}
}
