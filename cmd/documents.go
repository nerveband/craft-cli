package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var documentsCmd = &cobra.Command{
	Use:   "documents",
	Short: "Document helper commands",
	Long: `Document helper commands for capabilities not exposed by the top-level list/get/create commands.

Examples:
  craft documents resolve-link "https://www.craft.do/s/..." --mcp-url https://mcp.craft.do/links/.../mcp`,
}

var documentsResolveLinkCmd = &cobra.Command{
	Use:   "resolve-link <craft-url>",
	Short: "Resolve a Craft app or web URL to a root block ID",
	Long: `Resolve a Craft app or web URL to the rootBlockId used by blocks and write commands.

This uses Craft MCP because Craft links may contain a raw document ID that differs from
the rootBlockId expected by block-level commands.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMCPReadCommand(fmt.Sprintf("documents resolve-link %s", quoteMCPArg(args[0])))
	},
}

func init() {
	rootCmd.AddCommand(documentsCmd)
	documentsCmd.AddCommand(documentsResolveLinkCmd)
}
