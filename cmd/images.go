package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var imageURL string

var imagesCmd = &cobra.Command{
	Use:   "images",
	Short: "View image metadata through Craft MCP",
	Long:  "View image metadata through Craft MCP. This wraps the MCP images view command.",
}

var imagesViewCmd = &cobra.Command{
	Use:   "view [url]",
	Short: "View an image URL through Craft MCP",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			imageURL = args[0]
		}
		if imageURL == "" {
			return fmt.Errorf("image URL is required")
		}
		command := "images view --url " + quoteMCPArg(imageURL)
		if isDryRun() {
			return dryRunOutput("images view", map[string]interface{}{
				"backend":      "mcp",
				"tool":         "craft_read",
				"command":      command,
				"capabilities": []string{"mcp", "images.view"},
			})
		}
		return runMCPReadCommand(command)
	},
}

func init() {
	rootCmd.AddCommand(imagesCmd)
	imagesCmd.AddCommand(imagesViewCmd)
	imagesViewCmd.Flags().StringVar(&imageURL, "url", "", "Image URL")
}
