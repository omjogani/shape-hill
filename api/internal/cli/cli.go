// Package cli is the shapehill command line, a thin client over the HTTP API.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	defaultAPI = "https://shape-hill.onrender.com"
	defaultWeb = "https://shape-hill.vercel.app"
)

type app struct {
	stdout, stderr io.Writer
	getenv         func(string) string

	jsonOut bool
	apiURL  string
	webURL  string
}

// Run returns the process exit code.
func Run(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	// Before parsing, so help and flag errors are uncoloured too.
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

func (a *app) rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:              "shapehill",
		Short:            "Shape Up hill charts from the terminal",
		Long:             renderRootLong(),
		Example:          strings.TrimRight(rootExamples, "\n"),
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

// parent makes unknown subcommands fail instead of printing help and exiting 0.
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

func exactArgs(n int, names string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return usageErr("%s takes %s\n  usage: %s", cmd.CommandPath(), names, cmd.UseLine())
		}
		return nil
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

// Env only: a --token flag would leak into shell history.
func (a *app) client() (*client, error) {
	a.resolveConfig()
	token := a.getenv("SHAPEHILL_TOKEN")
	if token == "" {
		return nil, &exitError{code: exitAuth, msg: "SHAPEHILL_TOKEN is not set; create a token at " + a.webURL + "/app/tokens"}
	}
	return newClient(a.apiURL, token), nil
}

func (a *app) emit(v any, human func(w io.Writer)) error {
	if a.jsonOut {
		enc := json.NewEncoder(a.stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	human(a.stdout)
	return nil
}
