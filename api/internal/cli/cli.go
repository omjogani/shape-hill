// Package cli is the shapehill command line: a thin client over the HTTP API,
// built to be driven by people and AI agents alike.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	defaultAPI = "https://shape-hill.onrender.com"
	defaultWeb = "https://shape-hill.vercel.app"
)

// Exit codes are part of the contract agents script against.
const (
	exitOK       = 0
	exitFailure  = 1
	exitUsage    = 2
	exitNotFound = 3
	exitAuth     = 4
	exitConflict = 5
)

type exitError struct {
	code int
	msg  string
}

func (e *exitError) Error() string { return e.msg }

func usageErr(format string, args ...any) error {
	return &exitError{code: exitUsage, msg: fmt.Sprintf(format, args...)}
}

type app struct {
	stdout, stderr io.Writer
	getenv         func(string) string

	jsonOut bool
	apiURL  string
	webURL  string
}

// Version is stamped at build time:
//
//	go build -ldflags "-X github.com/omjogani/shape-hill/internal/cli.Version=v0.1.0" ./cmd/shapehill
var Version = ""

func version() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

// Run executes the CLI and returns the process exit code.
func Run(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	// Checked before parsing so help and flag errors are uncoloured too.
	if slices.Contains(args, "--no-color") {
		color.NoColor = true
	}

	a := &app{stdout: stdout, stderr: stderr, getenv: getenv}
	root := a.rootCmd()
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	cmd, err := root.ExecuteContextC(ctx)
	if err == nil {
		return exitOK
	}
	code := exitCode(err)
	if a.jsonOut {
		_ = json.NewEncoder(stderr).Encode(map[string]any{"error": err.Error(), "code": code})
		return code
	}
	fmt.Fprintf(stderr, "%s %s\n", failure("error:"), err)
	if hint := a.hint(cmd, err, code); hint != "" {
		fmt.Fprintf(stderr, "%s %s\n", faint("hint:"), hint)
	}
	return code
}

func (a *app) hint(cmd *cobra.Command, err error, code int) string {
	var api *apiError
	var transport *transportError
	switch {
	case errors.As(err, &transport):
		return "check the API is up and SHAPEHILL_API_URL is right (using " + a.apiURL + ")"
	case errors.As(err, &api) && api.Status == http.StatusUnauthorized:
		return "SHAPEHILL_TOKEN was rejected; it may be revoked. Create a new one at " + a.envOr("SHAPEHILL_WEB_URL", defaultWeb) + "/app/tokens"
	case code == exitUsage && cmd != nil && !strings.Contains(err.Error(), "usage:"):
		return fmt.Sprintf("run %s for usage", accent(cmd.CommandPath()+" --help"))
	}
	return ""
}

func exitCode(err error) int {
	var exit *exitError
	var api *apiError
	var transport *transportError
	switch {
	case errors.As(err, &exit):
		return exit.code
	case errors.As(err, &transport):
		return exitFailure
	case errors.As(err, &api):
		switch api.Status {
		case http.StatusBadRequest:
			return exitUsage
		case http.StatusUnauthorized, http.StatusForbidden:
			return exitAuth
		case http.StatusNotFound:
			return exitNotFound
		case http.StatusConflict:
			return exitConflict
		}
		return exitFailure
	default:
		// Anything else comes from cobra itself: bad flags, unknown commands.
		return exitUsage
	}
}

func init() {
	cobra.AddTemplateFunc("heading", heading)
	cobra.AddTemplateFunc("accent", accent)
	cobra.AddTemplateFunc("faint", faint)
}

func (a *app) rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "shapehill",
		Short: "Shape Up hill charts from the terminal",
		// Built per run, after colour is decided, so the headings can be styled.
		Long: fmt.Sprintf(`Shape Up hill charts from the terminal: create a hill, embed it in a story,
then move its scopes as the work progresses.

%s
  0-49    %s    figuring out the approach; unknowns remain
  50-99   %s  approach settled (50 is the top); executing
  100     %s

%s
  SHAPEHILL_TOKEN    personal access token, from <web>/app/tokens
  SHAPEHILL_API_URL  API origin (default %s)
  SHAPEHILL_WEB_URL  web origin (default %s)

%s 0 ok · 1 failure · 2 invalid input · 3 not found · 4 auth · 5 conflict`,
			heading("Positions (0-100)"),
			phaseColors["uphill"]("uphill"), phaseColors["downhill"]("downhill"), phaseColors["done"]("done"),
			heading("Environment"), defaultAPI, defaultWeb,
			heading("Exit codes")),
		Example: `  export SHAPEHILL_TOKEN=shk_...
  shapehill hill create launch-v2 --title "Launch v2" --public
  shapehill scope add launch-v2 "API integration"
  shapehill scope move launch-v2 "API integration" 40 --note "auth approach settled"
  shapehill hill show launch-v2`,
		Version:          version(),
		SilenceUsage:     true,
		SilenceErrors:    true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) { a.resolveConfig() },
	}
	parent(root)
	root.SetUsageTemplate(usageTemplate)
	root.SetVersionTemplate("shapehill {{.Version}}\n")
	root.PersistentFlags().BoolVar(&a.jsonOut, "json", false, "print machine-readable JSON")
	root.PersistentFlags().StringVar(&a.apiURL, "api", "", "API origin (overrides SHAPEHILL_API_URL)")
	root.PersistentFlags().Bool("no-color", false, "disable colour (also: NO_COLOR=1)")
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return usageErr("%v", err) })

	root.AddGroup(
		&cobra.Group{ID: "charts", Title: "Hill charts"},
		&cobra.Group{ID: "setup", Title: "Setup"},
	)
	for _, cmd := range []*cobra.Command{a.hillCmd(), a.scopeCmd()} {
		cmd.GroupID = "charts"
		root.AddCommand(cmd)
	}
	for _, cmd := range []*cobra.Command{a.authCmd(), versionCmd()} {
		cmd.GroupID = "setup"
		root.AddCommand(cmd)
	}
	root.AddCommand(agentsTopic())
	root.SetHelpCommandGroupID("setup")
	root.SetCompletionCommandGroupID("setup")
	return root
}

// parent makes a command that only groups others fail on an unknown
// subcommand instead of printing help and exiting 0.
func parent(cmd *cobra.Command) *cobra.Command {
	cmd.Args = cobra.ArbitraryArgs
	cmd.SuggestionsMinimumDistance = 2
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		msg := fmt.Sprintf("unknown command %q for %q", args[0], cmd.CommandPath())
		if suggestions := cmd.SuggestionsFor(args[0]); len(suggestions) > 0 {
			msg += "; did you mean " + strings.Join(suggestions, " or ") + "?"
		}
		return usageErr("%s", msg)
	}
	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Args:  exactArgs(0, "no arguments"),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "shapehill", version())
		},
	}
}

func agentsTopic() *cobra.Command {
	return &cobra.Command{
		Use:   "agents",
		Short: "How AI agents should drive shapehill",
		Long: `Driving shapehill from an AI agent

  - Pass --json. Results go to stdout; errors go to stderr as {"error", "code"}.
  - Branch on the exit code: 0 ok, 1 failure (retry), 2 invalid input (fix the
    command), 3 not found, 4 auth (stop and ask the user), 5 conflict.
  - Retries are safe: "hill create" and "scope add" take --if-not-exists, and
    repeating the latest "scope move" records nothing.
  - Refer to scopes by exact title or ID; "hill show <slug> --json" lists both.
  - Always pass --note on "scope move" saying what changed. The note is the
    snapshot people read on the chart.
  - Choose positions by certainty, not by percent done: below 50 while unknowns
    remain, 50 once the approach is settled, 100 when shipped.`,
	}
}

func (a *app) resolveConfig() {
	if a.apiURL == "" {
		a.apiURL = a.envOr("SHAPEHILL_API_URL", defaultAPI)
	}
	a.webURL = a.envOr("SHAPEHILL_WEB_URL", defaultWeb)
}

func (a *app) envOr(key, fallback string) string {
	if v := a.getenv(key); v != "" {
		return v
	}
	return fallback
}

// The token is read from the environment only: a flag would leak it into shell
// history and process listings.
func (a *app) client() (*client, error) {
	a.resolveConfig()
	token := a.getenv("SHAPEHILL_TOKEN")
	if token == "" {
		return nil, &exitError{code: exitAuth, msg: "SHAPEHILL_TOKEN is not set; create a token at " + a.webURL + "/app/tokens"}
	}
	return newClient(a.apiURL, token), nil
}

// emit prints v as JSON under --json, otherwise runs the human printer.
func (a *app) emit(v any, human func(w io.Writer)) error {
	if a.jsonOut {
		enc := json.NewEncoder(a.stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	human(a.stdout)
	return nil
}

func exactArgs(n int, names string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return usageErr("%s takes %s\n  usage: %s", cmd.CommandPath(), names, cmd.UseLine())
		}
		return nil
	}
}
