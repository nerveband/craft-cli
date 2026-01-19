package cmd

import (
	"fmt"
	"os"

	"github.com/ashrafali/craft-cli/internal/api"
	"github.com/ashrafali/craft-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	apiURL        string
	outputFormat  string
	cfgManager    *config.Manager
	version       = "1.0.0"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "craft",
	Short: "Craft CLI - Interact with Craft Documents API",
	Long: `A command-line interface for interacting with Craft Documents.
Fast, token-efficient, and built for LLM/agent integration.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "Craft API URL (overrides config)")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "", "Output format (json, table, markdown)")
}

func initConfig() {
	var err error
	cfgManager, err = config.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(3)
	}
}

// getAPIClient returns a configured API client
func getAPIClient() (*api.Client, error) {
	url := apiURL
	if url == "" {
		cfg, err := cfgManager.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		url = cfg.APIURL
	}

	if url == "" {
		return nil, fmt.Errorf("no API URL configured. Run 'craft config set-api <url>' first")
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
