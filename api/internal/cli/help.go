package cli

import (
	_ "embed"
	"fmt"
	"runtime/debug"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var (
	//go:embed help/root.tmpl
	rootLongTmpl string
	//go:embed help/root_examples.txt
	rootExamples string
	//go:embed help/agents.txt
	agentsHelp string
	//go:embed help/usage.tmpl
	usageTemplate string
)

var rootLong = template.Must(template.New("root").Funcs(template.FuncMap{
	"heading": heading,
	"phase":   func(name string) string { return phaseColors[name](name) },
}).Parse(rootLongTmpl))

func init() {
	cobra.AddTemplateFunc("heading", heading)
	cobra.AddTemplateFunc("accent", accent)
	cobra.AddTemplateFunc("faint", faint)
}

// Rendered per run, once colour is decided.
func renderRootLong() string {
	var b strings.Builder
	if err := rootLong.Execute(&b, map[string]string{"API": defaultAPI, "Web": defaultWeb}); err != nil {
		panic(err)
	}
	return strings.TrimRight(b.String(), "\n")
}

// Version is set with -ldflags "-X github.com/omjogani/shape-hill/api/internal/cli.Version=v0.1.0".
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
		Long:  strings.TrimRight(agentsHelp, "\n"),
	}
}
