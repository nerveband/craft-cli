package cmd

import (
	"encoding/json"
	"fmt"
	"os"

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
	apiURL      string
	outputFormat string
	cfgManager  *config.Manager
	version     = "1.0.0"

	// Global flags for LLM/scripting friendliness
	quietMode   bool
	jsonErrors  bool
	outputOnly  string
	noHeaders   bool
	rawOutput   bool
	idOnly      bool
	dryRun      bool
)

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
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		handleError(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// API and format flags
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "Craft API URL (overrides config)")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "", "Output format (json, table, markdown)")

	// LLM/scripting friendly flags
	rootCmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false, "Suppress status messages, output data only")
	rootCmd.PersistentFlags().BoolVar(&jsonErrors, "json-errors", false, "Output errors as JSON")
	rootCmd.PersistentFlags().StringVar(&outputOnly, "output-only", "", "Output only specified field (e.g., id, title)")
	rootCmd.PersistentFlags().BoolVar(&noHeaders, "no-headers", false, "Omit headers in table output")
	rootCmd.PersistentFlags().BoolVar(&rawOutput, "raw", false, "Output raw content without formatting")
	rootCmd.PersistentFlags().BoolVar(&idOnly, "id-only", false, "Output only document IDs (shorthand for --output-only id)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would happen without making changes")
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
		errObj := map[string]interface{}{
			"error": err.Error(),
			"code":  categorizeError(err),
		}
		json.NewEncoder(os.Stderr).Encode(errObj)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	switch categorizeError(err) {
	case "CONFIG_ERROR":
		os.Exit(ExitConfigError)
	case "API_ERROR":
		os.Exit(ExitAPIError)
	default:
		os.Exit(ExitUserError)
	}
}

// categorizeError returns an error category for JSON output
func categorizeError(err error) string {
	errStr := err.Error()
	switch {
	case contains(errStr, "no active profile"), contains(errStr, "config"):
		return "CONFIG_ERROR"
	case contains(errStr, "authentication"), contains(errStr, "unauthorized"):
		return "AUTH_ERROR"
	case contains(errStr, "not found"):
		return "NOT_FOUND"
	case contains(errStr, "rate limit"):
		return "RATE_LIMIT"
	case contains(errStr, "server"), contains(errStr, "500"):
		return "API_ERROR"
	default:
		return "USER_ERROR"
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
