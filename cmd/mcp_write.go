package cmd

import (
	"encoding/json"
	"fmt"

	craftmcp "github.com/ashrafali/craft-cli/internal/mcp"
)

// completeMCPWrite finishes a craft_write mutation: persists revertInfo when
// requested, then emits the result (with edit-review metadata when diff is set).
func completeMCPWrite(client *craftmcp.Client, result json.RawMessage, saveRevertPath string, diff bool) error {
	if saveRevertPath != "" {
		if err := saveRevertInfo(result, saveRevertPath); err != nil {
			return err
		}
	}
	if diff {
		return outputMCPMutationWithReview(client, result)
	}
	return outputRawJSON(result)
}

// runMCPWriteCommand executes a craft_write command with the standard
// dry-run preview, --yes confirmation gate, and revert/diff handling.
func runMCPWriteCommand(operation, command string, capabilities []string, saveRevertPath string, diff bool) error {
	target := map[string]interface{}{
		"backend":      "mcp",
		"tool":         "craft_write",
		"command":      command,
		"capabilities": capabilities,
	}
	if saveRevertPath != "" {
		target["save_revert"] = saveRevertPath
	}
	if diff {
		target["diff"] = map[string]interface{}{
			"available": false,
			"hint":      "MCP edit-review metadata is captured from real craft_write results; dry-run shows the planned command.",
		}
	}
	if isDryRun() {
		return dryRunOutput(operation, target)
	}
	if !yesFlag {
		return fmt.Errorf("%s uses craft_write; rerun with --yes after reviewing --dry-run", operation)
	}
	client, err := getMCPClient()
	if err != nil {
		return err
	}
	result, err := client.CallTool("craft_write", map[string]interface{}{"command": command})
	if err != nil {
		return enhanceMCPWriteError(command, err)
	}
	return completeMCPWrite(client, result, saveRevertPath, diff)
}
