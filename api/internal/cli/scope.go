package cli

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/omjogani/shape-hill/internal/hills"
)

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func (a *app) scopeCmd() *cobra.Command {
	scope := parent(&cobra.Command{
		Use:     "scope",
		Aliases: []string{"scopes"},
		Short:   "Add scopes to a hill and move them",
		Long: `Add scopes to a hill and move them.

A scope is referenced by its ID or its exact title (case-insensitive).`,
	})
	scope.AddCommand(a.scopeAddCmd(), a.scopeMoveCmd(), a.scopeLogCmd())
	return scope
}

func (a *app) scopeAddCmd() *cobra.Command {
	var color, description string
	var ifNotExists bool

	cmd := &cobra.Command{
		Use:               "add <hill> <title>",
		Short:             "Add a scope to a hill; it starts at position 0",
		ValidArgsFunction: a.completeArgs(false),
		Example:           `  shapehill scope add launch-v2 "API integration" --color "#2F4C64"`,
		Args:              exactArgs(2, "two arguments: the hill slug and the scope title"),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, title := args[0], strings.TrimSpace(args[1])
			if title == "" {
				return usageErr("scope title cannot be empty")
			}
			if color != "" && !colorPattern.MatchString(color) {
				return usageErr("--color must be a hex colour like #2F4C64")
			}
			c, err := a.client()
			if err != nil {
				return err
			}
			found, err := c.hill(cmd.Context(), slug)
			if err != nil {
				return err
			}

			// Titles double as references, so they stay unique within a hill.
			if existing, ok := byTitle(found.Scopes, title); ok {
				if !ifNotExists {
					return &exitError{code: exitConflict, msg: fmt.Sprintf("%s already has a scope titled %q (%s)", slug, existing.Title, existing.ID)}
				}
				return a.emit(map[string]any{"scope": existing, "created": false}, func(w io.Writer) {
					printOK(w, "%s already exists at %d %s", bold(existing.Title), existing.Position, faint(existing.ID))
				})
			}

			scope, err := c.addScope(cmd.Context(), slug, title, description, color, len(found.Scopes))
			if err != nil {
				return err
			}
			return a.emit(map[string]any{"scope": scope, "created": true}, func(w io.Writer) {
				printOK(w, "Added %s to %s at 0 %s", bold(scope.Title), accent(slug), faint(scope.ID))
			})
		},
	}
	cmd.Flags().StringVar(&color, "color", "", "dot colour as hex, e.g. #2F4C64")
	cmd.Flags().StringVar(&description, "description", "", "what the scope covers")
	cmd.Flags().BoolVar(&ifNotExists, "if-not-exists", false, "succeed with the existing scope if the title is taken")
	return cmd
}

func (a *app) scopeMoveCmd() *cobra.Command {
	var note string

	cmd := &cobra.Command{
		Use:               "move <hill> <scope> <position>",
		Aliases:           []string{"mv"},
		Short:             "Move a scope to a position from 0 to 100, recording a snapshot",
		ValidArgsFunction: a.completeArgs(true),
		Long: `Move a scope to a position from 0 to 100. Each move records a snapshot,
so add a --note saying why it moved.

  0-49 uphill (figuring it out), 50-99 downhill (executing; 50 is the top), 100 done

Repeating the latest move (same position and note) is a no-op, so it is safe to retry.`,
		Example: `  shapehill scope move launch-v2 "API integration" 40 --note "auth approach settled, retries unknown"`,
		Args:    exactArgs(3, "three arguments: the hill slug, the scope and the position"),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, ref := args[0], args[1]
			parsed, err := strconv.ParseInt(args[2], 10, 16)
			if err != nil || !hills.ValidPosition(int16(parsed)) {
				return usageErr("position must be a whole number from 0 to 100")
			}
			position := int16(parsed)
			note = strings.TrimSpace(note)

			c, err := a.client()
			if err != nil {
				return err
			}
			scope, err := a.findScope(cmd.Context(), c, slug, ref)
			if err != nil {
				return err
			}

			moved := scope.Position != position || scope.Note != note
			if moved {
				if err := c.moveScope(cmd.Context(), scope.ID, position, note); err != nil {
					return err
				}
			}

			result := map[string]any{
				"scope_id": scope.ID, "title": scope.Title,
				"from": scope.Position, "to": position, "note": note, "moved": moved,
			}
			return a.emit(result, func(w io.Writer) {
				if !moved {
					printOK(w, "%s is already at %d with that note; nothing recorded", bold(scope.Title), position)
					return
				}
				after := scope
				after.Position = position
				printOK(w, "Moved %s %d → %s %s", bold(scope.Title), scope.Position, bold(fmt.Sprint(position)), paintPhase(after, false))
				fmt.Fprintf(w, "  %s\n", track(after, false))
			})
		},
	}
	cmd.Flags().StringVar(&note, "note", "", "why it moved; shown on the chart and in the log")
	return cmd
}

func (a *app) scopeLogCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "log <hill> <scope>",
		Aliases:           []string{"history"},
		Short:             "Show a scope's snapshots, newest first",
		ValidArgsFunction: a.completeArgs(true),
		Example:           `  shapehill scope log launch-v2 "API integration"`,
		Args:              exactArgs(2, "two arguments: the hill slug and the scope"),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := a.client()
			if err != nil {
				return err
			}
			scope, err := a.findScope(cmd.Context(), c, args[0], args[1])
			if err != nil {
				return err
			}
			snaps, err := c.snapshots(cmd.Context(), scope.ID)
			if err != nil {
				return err
			}
			return a.emit(map[string]any{"scope": scope, "snapshots": snaps}, func(w io.Writer) {
				if len(snaps) == 0 {
					fmt.Fprintf(w, "%s has never moved\n", bold(scope.Title))
					return
				}
				fmt.Fprintf(w, "%s %s\n\n", bold(scope.Title), faint(scope.ID))
				rows := make([][]cell, len(snaps))
				for i, s := range snaps {
					at := hills.Scope{Position: s.Position}
					rows[i] = []cell{
						painted(s.CreatedAt.Local().Format("2006-01-02 15:04"), faint),
						{strings.Repeat("─", trackWidth), track(at, false)},
						plain(fmt.Sprintf("%3d", s.Position)),
						plain(s.Note),
					}
				}
				table(w, []string{"WHEN", "HILL", "POS", "NOTE"}, rows)
			})
		},
	}
}

func (a *app) findScope(ctx context.Context, c *client, slug, ref string) (hills.Scope, error) {
	found, err := c.hill(ctx, slug)
	if err != nil {
		return hills.Scope{}, err
	}
	return resolveScope(found.Scopes, slug, ref)
}

func resolveScope(scopes []hills.Scope, slug, ref string) (hills.Scope, error) {
	for _, s := range scopes {
		if s.ID == ref {
			return s, nil
		}
	}

	ref = strings.TrimSpace(ref)
	var matches []hills.Scope
	for _, s := range scopes {
		if strings.EqualFold(s.Title, ref) {
			matches = append(matches, s)
		}
	}

	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return hills.Scope{}, &exitError{code: exitNotFound, msg: fmt.Sprintf("no scope %q on %s; scopes: %s", ref, slug, describeScopes(scopes))}
	default:
		return hills.Scope{}, usageErr("%q matches several scopes on %s, use an ID: %s", ref, slug, describeScopes(matches))
	}
}

func byTitle(scopes []hills.Scope, title string) (hills.Scope, bool) {
	for _, s := range scopes {
		if strings.EqualFold(s.Title, title) {
			return s, true
		}
	}
	return hills.Scope{}, false
}

func describeScopes(scopes []hills.Scope) string {
	if len(scopes) == 0 {
		return "(none)"
	}
	parts := make([]string, len(scopes))
	for i, s := range scopes {
		parts[i] = fmt.Sprintf("%q (%s)", s.Title, s.ID)
	}
	return strings.Join(parts, ", ")
}
