package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ashrafali/craft-cli/internal/api"
	"github.com/ashrafali/craft-cli/internal/config"
	"github.com/spf13/cobra"
)

// Exit codes for scripting
const (
	ExitSuccess     = 0
	ExitUserError   = 1
	ExitAPIError    = 2
	ExitConfigError = 3
)

var (
	apiURL        string
	apiKey        string
	mcpURL        string
	profileName   string
	backendName   string
	outputFormat  string
	deliverTarget string
	transformExpr string
	dataSource    string
	cfgManager    *config.Manager
	version       = "1.11.4"

	// Global flags for LLM/scripting friendliness
	quietMode      bool
	jsonErrors     bool
	outputOnly     string
	noHeaders      bool
	rawOutput      bool
	idOnly         bool
	dryRun         bool
	yesFlag        bool
	requestSchema  bool
	responseSchema bool
)

type cliCodeError struct {
	Code    string
	Message string
}

func (e *cliCodeError) Error() string {
	return e.Message
}

func newCLIError(code, message string) error {
	return &cliCodeError{Code: code, Message: message}
}

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "craft",
	Short: "Craft CLI - Interact with Craft Documents API",
	Long: `A command-line interface for interacting with Craft Documents.
Fast, token-efficient, and built for LLM/agent integration.

Output is JSON by default for easy parsing. Use --format for alternatives.
Use --quiet to suppress status messages for cleaner piping.
Use --json-errors for machine-readable error output.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip update check for upgrade, version, and help commands
		cmdName := cmd.Name()
		if cmdName == "upgrade" || cmdName == "version" || cmdName == "help" || cmdName == "completion" {
			return
		}
		// Check for updates in background (non-blocking)
		go notifyUpdateAvailable()
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		handleError(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Set custom usage template with documentation footer
	rootCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}

Documentation:
  Full docs:        https://github.com/nerveband/craft-cli
  Report issues:    https://github.com/nerveband/craft-cli/issues
  Craft API docs:   https://connect.craft.do/api-docs
`)

	// API and format flags
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "Craft API URL (overrides config)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for authentication (overrides config)")
	rootCmd.PersistentFlags().StringVar(&mcpURL, "mcp-url", "", "Craft MCP URL (overrides CRAFT_MCP_URL)")
	rootCmd.PersistentFlags().StringVar(&profileName, "profile", "", "Named profile to use for this command")
	rootCmd.PersistentFlags().StringVar(&backendName, "backend", "auto", "Backend preference: auto, rest, mcp, or local")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "", "Output format (json, compact=legacy JSON, table, markdown, raw, jsonl, yaml)")
	rootCmd.PersistentFlags().StringVar(&deliverTarget, "deliver", "", "Deliver output to stdout or file:<path> (atomic file writes)")
	rootCmd.PersistentFlags().StringVar(&transformExpr, "transform", "", "Extract a field/projection from structured output")
	rootCmd.PersistentFlags().StringVar(&dataSource, "data-source", "", "Data source preference: local, live, or auto")

	// LLM/scripting friendly flags
	rootCmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false, "Suppress status messages, output data only")
	rootCmd.PersistentFlags().BoolVar(&jsonErrors, "json-errors", false, "Output errors as JSON")
	rootCmd.PersistentFlags().StringVar(&outputOnly, "output-only", "", "Output only specified field (e.g., id, title)")
	rootCmd.PersistentFlags().BoolVar(&noHeaders, "no-headers", false, "Omit headers in table output")
	rootCmd.PersistentFlags().BoolVar(&rawOutput, "raw", false, "Output raw content without formatting")
	rootCmd.PersistentFlags().BoolVar(&idOnly, "id-only", false, "Output only document IDs (shorthand for --output-only id)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would happen without making changes")
	rootCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompts")
	rootCmd.PersistentFlags().BoolVar(&requestSchema, "request-schema", false, "Print request JSON Schema for supported commands")
	rootCmd.PersistentFlags().BoolVar(&responseSchema, "response-schema", false, "Print response JSON Schema for supported commands")
}

func initConfig() {
	var err error
	cfgManager, err = config.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(ExitConfigError)
	}
}

// getAPIClient returns a configured API client
func getAPIClient() (*api.Client, error) {
	url := apiURL
	key := apiKey
	if profileName != "" && url == "" {
		profile, err := cfgManager.GetProfile(profileName)
		if err != nil {
			return nil, err
		}
		if profile.TypeOrDefault() != "rest" {
			return nil, fmt.Errorf("profile '%s' is %s, not rest. Use an MCP command or select a REST profile", profileName, profile.TypeOrDefault())
		}
		url = profile.URL
		if key == "" {
			key = profile.APIKey
		}
	}
	if url == "" {
		var err error
		url, err = cfgManager.GetActiveURL()
		if err != nil {
			// Check if this is first run and offer setup
			if checkFirstRun() {
				// User went through setup, try again
				url, err = cfgManager.GetActiveURL()
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	}

	// Get API key: flag > config > empty
	if key == "" {
		var err error
		if profileName == "" {
			key, err = cfgManager.GetActiveAPIKey()
			if err != nil {
				key = ""
			}
		}
	}

	if key != "" {
		return api.NewClientWithKey(url, key), nil
	}
	return api.NewClient(url), nil
}

// getOutputFormat returns the output format to use
func getOutputFormat() string {
	if outputFormat != "" {
		return outputFormat
	}

	cfg, err := cfgManager.Load()
	if err != nil {
		return "json"
	}

	if cfg.DefaultFormat != "" {
		return cfg.DefaultFormat
	}

	return "json"
}

// printStatus prints a status message (respects --quiet)
func printStatus(format string, args ...interface{}) {
	if !quietMode {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

// handleError handles errors with appropriate exit codes and formatting
func handleError(err error) {
	if jsonErrors {
		code := categorizeError(err)
		errObj := map[string]interface{}{
			"error": err.Error(),
			"code":  code,
		}
		if hint := errorHint(code); hint != "" {
			errObj["hint"] = hint
		}
		if apiErr, ok := err.(*api.APIError); ok {
			errObj["status"] = apiErr.StatusCode
		}
		json.NewEncoder(os.Stderr).Encode(errObj)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if hint := errorHint(categorizeError(err)); hint != "" {
			fmt.Fprintf(os.Stderr, "Hint: %s\n", hint)
		}
	}

	switch categorizeError(err) {
	case "CONFIG_ERROR":
		os.Exit(ExitConfigError)
	case "API_ERROR", "API_TIMEOUT":
		os.Exit(ExitAPIError)
	default:
		os.Exit(ExitUserError)
	}
}

// categorizeError returns an error category for JSON output
func categorizeError(err error) string {
	if coded, ok := err.(*cliCodeError); ok {
		return coded.Code
	}
	if apiErr, ok := err.(*api.APIError); ok {
		switch apiErr.StatusCode {
		case 401:
			return "AUTH_ERROR"
		case 403:
			return "PERMISSION_DENIED"
		case 404:
			return "NOT_FOUND"
		case 413:
			return "PAYLOAD_TOO_LARGE"
		case 429:
			return "RATE_LIMIT"
		default:
			if apiErr.StatusCode >= 500 {
				return "API_ERROR"
			}
			// Fall back to string matching below.
		}
	}

	errStr := strings.ToLower(err.Error())
	switch {
	case contains(errStr, "no active profile"), contains(errStr, "config"):
		return "CONFIG_ERROR"
	case contains(errStr, "authentication"), contains(errStr, "unauthorized"):
		return "AUTH_ERROR"
	case contains(errStr, "permission denied"):
		return "PERMISSION_DENIED"
	case contains(errStr, "not found"):
		return "NOT_FOUND"
	case contains(errStr, "rate limit"):
		return "RATE_LIMIT"
	case contains(errStr, "context deadline exceeded"), contains(errStr, "client.timeout"), contains(errStr, "timeout exceeded"), contains(errStr, "i/o timeout"):
		return "API_TIMEOUT"
	case contains(errStr, "request entity too large"), contains(errStr, "entity too large"), contains(errStr, "payload too large"), contains(errStr, "413"):
		return "PAYLOAD_TOO_LARGE"
	case contains(errStr, "server"), contains(errStr, "500"):
		return "API_ERROR"
	default:
		return "USER_ERROR"
	}
}

func errorHint(code string) string {
	switch code {
	case "PAYLOAD_TOO_LARGE":
		return "Reduce payload size or use chunking (craft update --chunk-bytes 20000). (not retryable)"
	case "PERMISSION_DENIED":
		return "Check link permissions in Craft. Use 'craft info --test-permissions'. (not retryable)"
	case "AUTH_ERROR":
		return "Check API key. Use --api-key flag or 'craft config add'. (not retryable)"
	case "CONFIG_ERROR":
		return "Run 'craft config list' or 'craft setup' to reconfigure. (not retryable)"
	case "NOT_FOUND":
		return "Check the ID is correct. Use 'craft list --id-only' to find valid IDs. (not retryable)"
	case "RATE_LIMIT":
		return "Wait and retry. The API limits request frequency. (retryable)"
	case "API_ERROR":
		return "Server error. Retry in a few seconds. If persistent, check Craft status. (retryable)"
	case "API_TIMEOUT":
		return "Network timeout while waiting for Craft. For writes, the server may still have applied the change; verify with repeated reads or search before retrying to avoid duplicates. For reads, wait and retry. (retryable)"
	case "CAPABILITY_UNAVAILABLE":
		return "This operation requires MCP-only capabilities. Ask the user for a Craft MCP URL if none is configured, then run 'craft config add-mcp <name> --mcp-url URL', verify with 'craft profiles test <name>', and retry with --profile <name> or --backend mcp. (not retryable)"
	case "BACKEND_REQUIRED":
		return "Select a compatible backend/profile. For MCP-only features, use --backend mcp with --mcp-url/CRAFT_MCP_URL or a separate MCP profile created by 'craft config add-mcp'. (not retryable)"
	case "READ_UNAVAILABLE":
		return "This connection appears write-only or lacks read scope. Switch to a read-capable profile. (not retryable)"
	case "WRITE_UNAVAILABLE":
		return "This connection lacks write permission. Switch to a write-capable profile. (not retryable)"
	case "IMAGE_ASSET_UNAVAILABLE":
		return "Validate the image URL with 'craft images view URL'. If it is expired, auth-gated, or a Craft r.craft.do short link that cannot resolve, export the image to a local file and insert it with 'craft upload FILE --page PAGE_ID' or 'craft upload FILE --sibling BLOCK_ID --position after'. (not retryable)"
	case "USER_ERROR":
		return "Check command usage with --help. (not retryable)"
	default:
		return ""
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// isQuiet returns whether quiet mode is enabled
func isQuiet() bool {
	return quietMode
}

// dryRunOutput prints structured dry-run info.
// If JSON mode, outputs JSON to stdout. Otherwise prints human prose.
func dryRunOutput(action string, target map[string]interface{}) error {
	if getOutputFormat() == "json" || jsonErrors {
		result := map[string]interface{}{
			"dry_run": true,
			"action":  action,
			"target":  target,
		}
		return outputJSON(result)
	}
	fmt.Printf("[dry-run] Would %s", action)
	if id, ok := target["id"]; ok {
		fmt.Printf(" %v", id)
	}
	if title, ok := target["title"]; ok {
		fmt.Printf(" (%v)", title)
	}
	fmt.Println()
	return nil
}

// isDryRun returns whether dry-run mode is enabled
func isDryRun() bool {
	return dryRun
}

// getOutputOnly returns the field to output (if specified)
func getOutputOnly() string {
	if idOnly {
		return "id"
	}
	return outputOnly
}

// hasNoHeaders returns whether to omit table headers
func hasNoHeaders() bool {
	return noHeaders
}

// isRawOutput returns whether raw output is requested
func isRawOutput() bool {
	return rawOutput
}
