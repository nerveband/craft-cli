package cmd

import (
	localindex "github.com/ashrafali/craft-cli/internal/index"
	"github.com/spf13/cobra"
	"os"
)

var cacheCmd = &cobra.Command{Use: "cache", Short: "Inspect or remove the current connection's local index"}
var cacheStatusCmd = &cobra.Command{Use: "status", Short: "Show local index completeness and freshness without fetching Craft", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
	path, err := indexPath()
	if err != nil {
		return err
	}
	s, err := localindex.Load(path)
	if err != nil {
		return err
	}
	return outputJSON(map[string]interface{}{"documents": len(s.Entries), "complete": s.Complete, "updated_at": s.UpdatedAt, "path": path})
}}
var cacheClearCmd = &cobra.Command{Use: "clear", Short: "Delete the current connection's private local index", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
	path, err := indexPath()
	if err != nil {
		return err
	}
	if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return outputJSON(map[string]string{"status": "cleared"})
}}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheStatusCmd, cacheClearCmd)
	commandEffects["cache status"] = CommandEffect{"read_only", "local", false}
	commandEffects["cache clear"] = CommandEffect{"non_idempotent", "local", true}
	workflowExamples["cache status"] = "craft cache status"
	workflowExamples["cache clear"] = "craft cache clear --dry-run"
}
