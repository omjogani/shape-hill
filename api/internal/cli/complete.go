package cli

import (
	"github.com/spf13/cobra"
)

// Errors give no suggestions rather than breaking the shell.
func (a *app) completeArgs(withScope bool) cobra.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		c, err := a.client()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		switch {
		case len(args) == 0:
			owned, err := c.listHills(cmd.Context())
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			slugs := make([]string, len(owned))
			for i, h := range owned {
				slugs[i] = h.Slug + "\t" + h.Title
			}
			return slugs, cobra.ShellCompDirectiveNoFileComp
		case len(args) == 1 && withScope:
			found, err := c.hill(cmd.Context(), args[0])
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			titles := make([]string, len(found.Scopes))
			for i, s := range found.Scopes {
				titles[i] = s.Title + "\t" + s.Phase()
			}
			return titles, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

func fixedCompletion(values ...string) cobra.CompletionFunc {
	return func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return values, cobra.ShellCompDirectiveNoFileComp
	}
}
