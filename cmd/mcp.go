package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	craftmcp "github.com/ashrafali/craft-cli/internal/mcp"
	"github.com/spf13/cobra"
)

var (
	mcpToolName      string
	mcpArgsJSON      string
	mcpCommand       string
	mcpResourceURI   string
	mcpMetadataOnly  bool
	mcpBatchFile     string
	mcpBatchTool     string
	mcpBatchCommands []string
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Interact with a Craft MCP server",
	Long: `Interact with a Craft MCP server over Streamable HTTP.

Use --mcp-url or CRAFT_MCP_URL to configure the endpoint.

Examples:
  craft mcp tools --mcp-url https://mcp.craft.do/links/.../mcp
  craft mcp call craft_read --command "connection info"
  craft mcp call craft_write --arguments '{"command":"documents create --title Test"}'`,
}

var mcpInitializeCmd = &cobra.Command{
	Use:   "initialize",
	Short: "Initialize the Craft MCP connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getMCPClient()
		if err != nil {
			return err
		}
		result, err := client.Initialize()
		if err != nil {
			return err
		}
		return outputRawJSON(result)
	},
}

var mcpToolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "List Craft MCP tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getMCPClient()
		if err != nil {
			return err
		}
		result, err := client.ListTools()
		if err != nil {
			return err
		}
		return outputRawJSON(result)
	},
}

var mcpResourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "List Craft MCP resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getMCPClient()
		if err != nil {
			return err
		}
		result, err := client.ListResources()
		if err != nil {
			return err
		}
		return outputRawJSON(result)
	},
}

var mcpCallCmd = &cobra.Command{
	Use:   "call [tool-name]",
	Short: "Call a Craft MCP tool",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			mcpToolName = args[0]
		}
		if mcpToolName == "" {
			return fmt.Errorf("tool name is required")
		}

		arguments := map[string]interface{}{}
		if mcpArgsJSON != "" {
			if err := json.Unmarshal([]byte(mcpArgsJSON), &arguments); err != nil {
				return fmt.Errorf("invalid --arguments JSON: %w", err)
			}
		}
		if mcpCommand != "" {
			arguments["command"] = mcpCommand
		}

		client, err := getMCPClient()
		if err != nil {
			return err
		}
		result, err := client.CallTool(mcpToolName, arguments)
		if err != nil {
			return err
		}
		return outputRawJSON(result)
	},
}

var mcpReadResourceCmd = &cobra.Command{
	Use:   "read-resource [uri]",
	Short: "Read Craft MCP resource metadata or content",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			mcpResourceURI = args[0]
		}
		if mcpResourceURI == "" {
			return fmt.Errorf("resource URI is required")
		}
		client, err := getMCPClient()
		if err != nil {
			return err
		}
		result, err := client.ReadResource(mcpResourceURI)
		if err != nil {
			return err
		}
		if !mcpMetadataOnly {
			return outputRawJSON(result)
		}
		var data map[string]interface{}
		if err := json.Unmarshal(result, &data); err != nil {
			return err
		}
		if contents, ok := data["contents"].([]interface{}); ok {
			summaries := make([]map[string]interface{}, 0, len(contents))
			for _, item := range contents {
				entry, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				summary := map[string]interface{}{}
				for _, key := range []string{"uri", "mimeType", "_meta"} {
					if value, ok := entry[key]; ok {
						summary[key] = value
					}
				}
				if text, ok := entry["text"].(string); ok {
					summary["text_bytes"] = len(text)
				}
				summaries = append(summaries, summary)
			}
			data["contents"] = summaries
		}
		return outputJSON(data)
	},
}

var mcpEditReviewCmd = &cobra.Command{
	Use:   "edit-review",
	Short: "Read Craft MCP edit-review resource metadata",
	Long:  "Read the Craft MCP edit-review UI resource metadata without dumping large resource contents.",
	RunE: func(cmd *cobra.Command, args []string) error {
		mcpResourceURI = "ui://craft/edit-review"
		mcpMetadataOnly = true
		return mcpReadResourceCmd.RunE(cmd, []string{mcpResourceURI})
	},
}

var mcpBatchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Run multiple Craft MCP commands",
	Long: `Run multiple Craft MCP mini-language commands from --command, --file, or --stdin.

Input may be a JSON array of strings, a JSON array of {"tool","command"} objects,
or newline-delimited command strings.

Examples:
  craft mcp batch --command "connection info" --command "documents list --limit 5"
  printf 'connection info\nblocks explore-themes\n' | craft mcp batch --stdin
  craft mcp batch --file ops.json --tool craft_write --yes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMCPBatch(cmd)
	},
}

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Run multiple Craft commands through the selected backend",
	Long: `Run multiple Craft operations. The current implementation routes batch
execution through Craft MCP because MCP exposes the command mini-language used
for reliable multi-command execution.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if backendName != "auto" && backendName != "mcp" {
			return fmt.Errorf("batch currently requires --backend mcp or --backend auto")
		}
		return runMCPBatch(cmd)
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(batchCmd)
	mcpCmd.AddCommand(mcpInitializeCmd)
	mcpCmd.AddCommand(mcpToolsCmd)
	mcpCmd.AddCommand(mcpResourcesCmd)
	mcpCmd.AddCommand(mcpReadResourceCmd)
	mcpCmd.AddCommand(mcpEditReviewCmd)
	mcpCmd.AddCommand(mcpCallCmd)
	mcpCmd.AddCommand(mcpBatchCmd)

	mcpCallCmd.Flags().StringVar(&mcpArgsJSON, "arguments", "", "Tool arguments as JSON")
	mcpCallCmd.Flags().StringVar(&mcpCommand, "command", "", "Shortcut for arguments.command")

	mcpReadResourceCmd.Flags().StringVar(&mcpResourceURI, "uri", "", "MCP resource URI")
	mcpReadResourceCmd.Flags().BoolVar(&mcpMetadataOnly, "metadata-only", false, "Omit large resource contents from output")

	mcpBatchCmd.Flags().StringArrayVar(&mcpBatchCommands, "command", nil, "MCP command string (repeatable)")
	mcpBatchCmd.Flags().StringVar(&mcpBatchFile, "file", "", "Read batch operations from file")
	mcpBatchCmd.Flags().Bool("stdin", false, "Read batch operations from stdin")
	mcpBatchCmd.Flags().StringVar(&mcpBatchTool, "tool", "craft_read", "Default MCP tool: craft_read or craft_write")
	addBatchFlags(batchCmd)
}

func addBatchFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVar(&mcpBatchCommands, "command", nil, "MCP command string (repeatable)")
	cmd.Flags().StringVar(&mcpBatchFile, "file", "", "Read batch operations from file")
	cmd.Flags().Bool("stdin", false, "Read batch operations from stdin")
	cmd.Flags().StringVar(&mcpBatchTool, "tool", "craft_read", "Default MCP tool: craft_read or craft_write")
}

func runMCPBatch(cmd *cobra.Command) error {
	ops, err := readMCPBatchOperations(cmd)
	if err != nil {
		return err
	}
	if len(ops) == 0 {
		return fmt.Errorf("batch requires at least one command")
	}
	if isDryRun() {
		return dryRunOutput("mcp batch", map[string]interface{}{"operations": ops, "count": len(ops), "backend": "mcp"})
	}
	for _, op := range ops {
		if op.Tool == "craft_write" && !yesFlag {
			return fmt.Errorf("mcp batch includes craft_write; rerun with --yes after reviewing --dry-run")
		}
	}
	client, err := getMCPClient()
	if err != nil {
		return err
	}
	results := make([]map[string]interface{}, 0, len(ops))
	for i, op := range ops {
		result, err := client.CallTool(op.Tool, map[string]interface{}{"command": op.Command})
		entry := map[string]interface{}{
			"index":   i,
			"tool":    op.Tool,
			"command": op.Command,
			"ok":      err == nil,
		}
		if err != nil {
			entry["error"] = err.Error()
		} else {
			var data interface{}
			if err := json.Unmarshal(result, &data); err != nil {
				entry["raw"] = string(result)
			} else {
				entry["result"] = data
			}
		}
		results = append(results, entry)
		if err != nil {
			break
		}
	}
	return outputJSON(map[string]interface{}{"items": results, "total": len(results)})
}

func getMCPClient() (*craftmcp.Client, error) {
	url := mcpURL
	if url == "" {
		url = os.Getenv("CRAFT_MCP_URL")
	}
	if url == "" && profileName != "" {
		profile, err := cfgManager.GetProfile(profileName)
		if err != nil {
			return nil, err
		}
		if profile.TypeOrDefault() != "mcp" {
			return nil, fmt.Errorf("profile '%s' is %s, not mcp. Use 'craft config add-mcp <name> --mcp-url URL' or 'craft profiles add-mcp <name> --mcp-url URL', then pass --profile <name>", profileName, profile.TypeOrDefault())
		}
		url = profile.MCPURL
	}
	if url == "" {
		profile, err := cfgManager.GetActiveProfile()
		if err == nil && profile.TypeOrDefault() == "mcp" {
			url = profile.MCPURL
		}
	}
	if url == "" {
		return nil, fmt.Errorf("Craft MCP URL required. Use --mcp-url, set CRAFT_MCP_URL, or create an MCP profile with 'craft config add-mcp <name> --mcp-url URL'")
	}
	return craftmcp.NewClient(url), nil
}

func outputRawJSON(raw json.RawMessage) error {
	var data interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}
	return outputJSON(data)
}

func runMCPReadCommand(command string) error {
	client, err := getMCPClient()
	if err != nil {
		return err
	}
	result, err := client.CallTool("craft_read", map[string]interface{}{"command": command})
	if err != nil {
		return err
	}
	return outputRawJSON(result)
}

func quoteMCPArg(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

type mcpBatchOperation struct {
	Tool    string `json:"tool"`
	Command string `json:"command"`
}

func readMCPBatchOperations(cmd *cobra.Command) ([]mcpBatchOperation, error) {
	input := ""
	stdinFlag, _ := cmd.Flags().GetBool("stdin")
	switch {
	case mcpBatchFile != "":
		data, err := os.ReadFile(mcpBatchFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read batch file: %w", err)
		}
		input = string(data)
	case stdinFlag:
		data, err := readStdinString()
		if err != nil {
			return nil, err
		}
		input = data
	}
	if input != "" {
		return parseMCPBatchInput(input, mcpBatchTool)
	}
	ops := make([]mcpBatchOperation, 0, len(mcpBatchCommands))
	for _, command := range mcpBatchCommands {
		command = strings.TrimSpace(command)
		if command != "" {
			ops = append(ops, mcpBatchOperation{Tool: mcpBatchTool, Command: command})
		}
	}
	return ops, nil
}

func parseMCPBatchInput(input, defaultTool string) ([]mcpBatchOperation, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}
	var raw []interface{}
	if err := json.Unmarshal([]byte(input), &raw); err == nil {
		ops := make([]mcpBatchOperation, 0, len(raw))
		for _, item := range raw {
			switch value := item.(type) {
			case string:
				ops = append(ops, mcpBatchOperation{Tool: defaultTool, Command: value})
			case map[string]interface{}:
				command, _ := value["command"].(string)
				tool, _ := value["tool"].(string)
				if tool == "" {
					tool = defaultTool
				}
				if command == "" {
					return nil, fmt.Errorf("batch operation object missing command")
				}
				ops = append(ops, mcpBatchOperation{Tool: tool, Command: command})
			default:
				return nil, fmt.Errorf("unsupported batch operation type")
			}
		}
		return ops, nil
	}
	var ops []mcpBatchOperation
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ops = append(ops, mcpBatchOperation{Tool: defaultTool, Command: line})
	}
	return ops, nil
}
