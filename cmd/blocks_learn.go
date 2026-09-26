package cmd

import (
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	c := &cobra.Command{Use: "learn [topic...]", Short: "Read Craft MCP block-shape guidance", Example: "  craft blocks learn pages tables", RunE: func(cmd *cobra.Command, args []string) error {
		parts := []string{"blocks learn"}
		for _, topic := range args {
			parts = append(parts, quoteMCPArg(topic))
		}
		return runMCPReadCommand(strings.Join(parts, " "))
	}}
	blocksCmd.AddCommand(c)
	commandEffects["blocks learn"] = CommandEffect{"read_only", "remote", false}
}
