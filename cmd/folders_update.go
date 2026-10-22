package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var folderUpdateName, folderUpdateIcon, folderUpdateTint string
var foldersUpdateCmd = &cobra.Command{Use: "update <folder-id>", Short: "Rename a folder or change its icon and tint through MCP", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
	if backendName == "rest" {
		return newCLIError("CAPABILITY_UNAVAILABLE", "folder metadata updates require MCP")
	}
	if err := validateResourceID(args[0], "folder-id"); err != nil {
		return err
	}
	command := "folders update --folder " + quoteMCPArg(args[0])
	changed := false
	for _, item := range []struct{ name, value string }{{"name", folderUpdateName}, {"icon", folderUpdateIcon}, {"tint", folderUpdateTint}} {
		if cmd.Flags().Changed(item.name) {
			command += " --" + item.name + " " + quoteMCPArg(item.value)
			changed = true
		}
	}
	if !changed {
		return fmt.Errorf("provide --name, --icon or --tint")
	}
	if isDryRun() {
		return dryRunOutput("update folder", map[string]interface{}{"backend": "mcp", "tool": "craft_write", "command": command})
	}
	client, err := getMCPClient()
	if err != nil {
		return err
	}
	result, err := client.CallTool("craft_write", map[string]interface{}{"command": command})
	if err != nil {
		return err
	}
	return outputRawJSON(result)
}}

func init() {
	foldersCmd.AddCommand(foldersUpdateCmd)
	foldersUpdateCmd.Flags().StringVar(&folderUpdateName, "name", "", "New folder name")
	foldersUpdateCmd.Flags().StringVar(&folderUpdateIcon, "icon", "", "Emoji or Craft icon name")
	foldersUpdateCmd.Flags().StringVar(&folderUpdateTint, "tint", "", "Icon color as #RRGGBB or light/dark pair")
	commandEffects["folders update"] = CommandEffect{"non_idempotent", "remote", false}
	workflowExamples["folders update"] = "craft folders update folder-123 --name 'Active projects' --dry-run"
}
