package cli

import (
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/omjogani/shape-hill/api/internal/hills"
)

func (a *app) authCmd() *cobra.Command {
	auth := parent(&cobra.Command{Use: "auth", Short: "Check how the CLI is authenticated"})
	auth.AddCommand(&cobra.Command{
		Use:     "status",
		Short:   "Show which account SHAPEHILL_TOKEN belongs to",
		Example: `  shapehill auth status`,
		Args:    exactArgs(0, "no arguments"),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := a.client()
			if err != nil {
				return err
			}
			user, err := c.me(cmd.Context())
			if err != nil {
				return err
			}
			return a.emit(map[string]any{"user": user, "api": a.apiURL}, func(w io.Writer) {
				printOK(w, "Signed in as %s %s", bold("@"+user.Username), faint("("+user.Email+")"))
				fmt.Fprintf(w, "  %s %s\n", faint("api"), a.apiURL)
			})
		},
	})
	return auth
}

func (a *app) hillCmd() *cobra.Command {
	hill := parent(&cobra.Command{Use: "hill", Aliases: []string{"hills"}, Short: "Create, inspect and embed hills"})
	hill.AddCommand(a.hillListCmd(), a.hillCreateCmd(), a.hillShowCmd(), a.hillUpdateCmd(), a.hillEmbedCmd())
	return hill
}

func (a *app) hillListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your hills",
		Example: `  shapehill hill list
  shapehill hill ls --json`,
		Args: exactArgs(0, "no arguments"),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := a.client()
			if err != nil {
				return err
			}
			owned, err := c.listHills(cmd.Context())
			if err != nil {
				return err
			}
			return a.emit(owned, func(w io.Writer) {
				if len(owned) == 0 {
					fmt.Fprintf(w, "No hills yet. Create one with %s\n", accent("shapehill hill create <slug> --title <title>"))
					return
				}
				rows := make([][]cell, len(owned))
				for i, h := range owned {
					rows[i] = []cell{painted(h.Slug, accent), plain(h.Title), painted(visibility(h), faint)}
				}
				table(w, []string{"SLUG", "TITLE", "VISIBILITY"}, rows)
			})
		},
	}
}

func (a *app) hillCreateCmd() *cobra.Command {
	var title, description string
	var public, ifNotExists bool

	cmd := &cobra.Command{
		Use:     "create <slug>",
		Aliases: []string{"new"},
		Short:   "Create a hill and print its embed snippet",
		Example: `  shapehill hill create launch-v2 --title "Launch v2" --public
  shapehill hill create launch-v2 --title "Launch v2" --public --if-not-exists --json`,
		Args: exactArgs(1, "exactly one argument: the slug"),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			if !hills.ValidSlug(slug) {
				return usageErr("slug must be lowercase letters, numbers and hyphens, e.g. launch-v2")
			}
			if title == "" {
				return usageErr("--title is required")
			}
			c, err := a.client()
			if err != nil {
				return err
			}

			created := true
			hill, err := c.createHill(cmd.Context(), slug, title, description, public)
			var api *apiError
			if errors.As(err, &api) && api.Status == http.StatusConflict && ifNotExists {
				existing, getErr := c.hill(cmd.Context(), slug)
				if errors.As(getErr, &api) && api.Status == http.StatusNotFound {
					return &exitError{code: exitConflict, msg: "slug " + slug + " is taken by another account"}
				}
				hill, err, created = existing.Hill, getErr, false
			}
			if err != nil {
				return err
			}

			embed := a.embedFor(hill, "")
			return a.emit(map[string]any{"hill": hill, "created": created, "embed": embed}, func(w io.Writer) {
				if created {
					printOK(w, "Created hill %s %s", accent(hill.Slug), faint("("+visibility(hill)+")"))
				} else {
					printOK(w, "Hill %s already exists %s", accent(hill.Slug), faint("("+visibility(hill)+")"))
				}
				if !hill.IsPublic {
					printWarn(w, "It's private, so the embed shows a placeholder. Run %s", accent("shapehill hill update "+hill.Slug+" --public"))
				}
				fmt.Fprintf(w, "\n%s\n%s\n", faint("Embed it in your story:"), embed.Markdown)
			})
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "hill title (required)")
	cmd.Flags().StringVar(&description, "description", "", "what the hill tracks")
	cmd.Flags().BoolVar(&public, "public", false, "make the hill and its embed public")
	cmd.Flags().BoolVar(&ifNotExists, "if-not-exists", false, "succeed with the existing hill if you already own this slug")
	return cmd
}

func (a *app) hillShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "show <slug>",
		Aliases: []string{"view"},
		Short:   "Show a hill and where each scope sits",
		Example: `  shapehill hill show launch-v2
  shapehill hill show launch-v2 --json`,
		Args:              exactArgs(1, "exactly one argument: the slug"),
		ValidArgsFunction: a.completeArgs(false),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := a.client()
			if err != nil {
				return err
			}
			found, err := c.hill(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			now := time.Now()
			type scopeView struct {
				hills.Scope
				Phase   string
				Stalled bool
			}
			views := make([]scopeView, len(found.Scopes))
			for i, s := range found.Scopes {
				views[i] = scopeView{Scope: s, Phase: s.Phase(), Stalled: s.Stalled(found.Hill, now)}
			}

			return a.emit(map[string]any{"hill": found.Hill, "scopes": views}, func(w io.Writer) {
				fmt.Fprintf(w, "%s  %s %s\n\n", bold(found.Hill.Title), accent(found.Hill.Slug), faint("· "+visibility(found.Hill)))
				if len(views) == 0 {
					fmt.Fprintf(w, "No scopes yet. Add one with %s\n", accent("shapehill scope add "+found.Hill.Slug+" <title>"))
					return
				}
				rows := make([][]cell, len(views))
				for i, v := range views {
					phase := v.Phase
					if v.Stalled {
						phase += " · stalled"
					}
					rows[i] = []cell{
						{strings.Repeat("─", trackWidth), track(v.Scope, v.Stalled)},
						plain(fmt.Sprintf("%3d", v.Position)),
						{phase, paintPhase(v.Scope, v.Stalled)},
						painted(v.Title, bold),
						plain(v.Note),
						painted(ago(v.MovedAt, now), faint),
					}
				}
				table(w, []string{"HILL", "POS", "PHASE", "SCOPE", "NOTE", "MOVED"}, rows)
			})
		},
	}
}

func (a *app) hillUpdateCmd() *cobra.Command {
	var title string
	var public, private, trackStalled, noTrackStalled bool

	cmd := &cobra.Command{
		Use:               "update <slug>",
		Short:             "Rename a hill or change its visibility",
		ValidArgsFunction: a.completeArgs(false),
		Example: `  shapehill hill update launch-v2 --public
  shapehill hill update launch-v2 --title "Launch v2.1" --no-track-stalled`,
		Args: exactArgs(1, "exactly one argument: the slug"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if public && private {
				return usageErr("--public and --private are mutually exclusive")
			}
			if trackStalled && noTrackStalled {
				return usageErr("--track-stalled and --no-track-stalled are mutually exclusive")
			}
			body := map[string]any{}
			if cmd.Flags().Changed("title") {
				if title == "" {
					return usageErr("--title cannot be empty")
				}
				body["title"] = title
			}
			if public || private {
				body["is_public"] = public
			}
			if trackStalled || noTrackStalled {
				body["track_stalled"] = trackStalled
			}
			if len(body) == 0 {
				return usageErr("nothing to update: pass --title, --public/--private or --track-stalled/--no-track-stalled")
			}

			c, err := a.client()
			if err != nil {
				return err
			}
			var hill hills.Hill
			if err := c.do(cmd.Context(), http.MethodPatch, "/api/hills/"+url.PathEscape(args[0]), body, &hill); err != nil {
				return err
			}
			return a.emit(hill, func(w io.Writer) {
				printOK(w, "Updated %s %s %s", accent(hill.Slug), bold(hill.Title), faint("· "+visibility(hill)))
			})
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().BoolVar(&public, "public", false, "make the hill public")
	cmd.Flags().BoolVar(&private, "private", false, "make the hill private")
	cmd.Flags().BoolVar(&trackStalled, "track-stalled", false, "flag scopes that haven't moved in a week")
	cmd.Flags().BoolVar(&noTrackStalled, "no-track-stalled", false, "stop flagging stalled scopes")
	return cmd
}

type embed struct {
	Image    string `json:"image"`
	View     string `json:"view"`
	Markdown string `json:"markdown"`
	HTML     string `json:"html"`
}

func (a *app) embedFor(hill hills.Hill, style string) embed {
	image := a.apiURL + "/hill/" + hill.Slug + ".svg"
	if style != "" {
		image += "?style=" + url.QueryEscape(style)
	}
	view := a.webURL + "/" + hill.Slug + "/view"
	return embed{
		Image:    image,
		View:     view,
		Markdown: fmt.Sprintf("[![%s](%s)](%s)", hill.Title, image, view),
		HTML:     fmt.Sprintf(`<a href="%s"><img src="%s" alt="%s"></a>`, html.EscapeString(view), html.EscapeString(image), html.EscapeString(hill.Title)),
	}
}

func (a *app) hillEmbedCmd() *cobra.Command {
	var style, format string

	cmd := &cobra.Command{
		Use:               "embed <slug>",
		Short:             "Print the snippet that embeds a hill in a story, README or ticket",
		ValidArgsFunction: a.completeArgs(false),
		Example: `  shapehill hill embed launch-v2
  shapehill hill embed launch-v2 --style github --format html`,
		Args: exactArgs(1, "exactly one argument: the slug"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if style != "" && style != "github" {
				return usageErr("--style must be github, or left out for the default paper style")
			}
			if format != "markdown" && format != "html" && format != "url" {
				return usageErr("--format must be markdown, html or url")
			}
			c, err := a.client()
			if err != nil {
				return err
			}
			found, err := c.hill(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			e := a.embedFor(found.Hill, style)
			snippet := map[string]string{"markdown": e.Markdown, "html": e.HTML, "url": e.Image}[format]
			return a.emit(map[string]any{"embed": e, "public": found.Hill.IsPublic}, func(w io.Writer) {
				fmt.Fprintln(w, snippet)
				if !found.Hill.IsPublic {
					printWarn(a.stderr, "%s is private, so the embed shows a placeholder", found.Hill.Slug)
				}
			})
		},
	}
	cmd.Flags().StringVar(&style, "style", "", "image style: github (default: paper)")
	cmd.Flags().StringVar(&format, "format", "markdown", "markdown, html or url")
	_ = cmd.RegisterFlagCompletionFunc("style", fixedCompletion("github"))
	_ = cmd.RegisterFlagCompletionFunc("format", fixedCompletion("markdown", "html", "url"))
	return cmd
}

func visibility(h hills.Hill) string {
	if h.IsPublic {
		return "public"
	}
	return "private"
}
