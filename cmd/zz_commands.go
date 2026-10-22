package cmd

import (
	"github.com/spf13/cobra"
	"strings"
)

func canonicalCommandPath(c *cobra.Command) string {
	path := strings.TrimPrefix(c.CommandPath(), "craft ")
	if strings.HasPrefix(path, "documents ") {
		tail := strings.TrimPrefix(path, "documents ")
		switch tail {
		case "list", "get", "create", "update", "delete", "move", "search", "clear":
			return tail
		}
	}
	return path
}

func installDocumentAliases() {
	for _, source := range []*cobra.Command{listCmd, getCmd, createCmd, updateCmd, deleteCmd, moveCmd, searchCmd, clearCmd} {
		alias := &cobra.Command{Use: source.Use, Short: source.Short, Long: source.Long, Example: source.Example, Args: source.Args, RunE: source.RunE, PreRunE: source.PreRunE}
		alias.Flags().AddFlagSet(source.LocalNonPersistentFlags())
		documentsCmd.AddCommand(alias)
	}
}

func init() { installDocumentAliases(); installExamples() }
