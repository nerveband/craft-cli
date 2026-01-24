package cmd

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

// Local commands work via craftdocs:// URL scheme (macOS only)
// These don't require API configuration

var (
	localSpaceID string
	localTitle   string
	localContent string
	localFolder  string
)

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Local Craft app commands (macOS only)",
	Long: `Commands that interact directly with the Craft app via URL schemes.
These commands work without API configuration but require the Craft app to be installed.
Currently only supported on macOS.`,
}

var localOpenCmd = &cobra.Command{
	Use:   "open <blockId>",
	Short: "Open a document in Craft app",
	Long:  "Open a specific document by its block ID in the Craft app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkMacOS(); err != nil {
			return err
		}
		blockID := args[0]
		if localSpaceID == "" {
			return fmt.Errorf("--space-id is required")
		}
		uri := fmt.Sprintf("craftdocs://open?spaceId=%s&blockId=%s",
			url.QueryEscape(localSpaceID),
			url.QueryEscape(blockID))
		return openURL(uri)
	},
}

var localOpenSpaceCmd = &cobra.Command{
	Use:   "space",
	Short: "Open a space in Craft app",
	Long:  "Open a specific space in the Craft app",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkMacOS(); err != nil {
			return err
		}
		if localSpaceID == "" {
			return fmt.Errorf("--space-id is required")
		}
		tab, _ := cmd.Flags().GetString("tab")
		uri := fmt.Sprintf("craftdocs://openspace?spaceId=%s", url.QueryEscape(localSpaceID))
		if tab != "" {
			uri += "&tab=" + url.QueryEscape(tab)
		}
		return openURL(uri)
	},
}

var localNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new document in Craft app",
	Long:  "Create a new document in the currently open space or a specified space",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkMacOS(); err != nil {
			return err
		}

		// Simple new doc in current space
		if localSpaceID == "" && localTitle == "" && localContent == "" {
			return openURL("craftdocs://createnewdocument")
		}

		// Create with params
		if localSpaceID == "" {
			return fmt.Errorf("--space-id is required when specifying title or content")
		}

		uri := fmt.Sprintf("craftdocs://createdocument?spaceId=%s&folderId=%s",
			url.QueryEscape(localSpaceID),
			url.QueryEscape(localFolder))

		if localTitle != "" {
			uri += "&title=" + url.QueryEscape(localTitle)
		}
		if localContent != "" {
			uri += "&content=" + url.QueryEscape(localContent)
		}

		return openURL(uri)
	},
}

var localAppendCmd = &cobra.Command{
	Use:   "append <blockId> <content>",
	Short: "Append content to a document",
	Long:  "Append content to an existing document by its block ID",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkMacOS(); err != nil {
			return err
		}
		blockID := args[0]
		content := args[1]

		if localSpaceID == "" {
			return fmt.Errorf("--space-id is required")
		}

		// Use a large index to append at the end
		uri := fmt.Sprintf("craftdocs://createblock?spaceId=%s&parentBlockId=%s&content=%s&index=999999",
			url.QueryEscape(localSpaceID),
			url.QueryEscape(blockID),
			url.QueryEscape(content))

		return openURL(uri)
	},
}

var localSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Open search in Craft app",
	Long:  "Open the search interface in Craft with a pre-filled query",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkMacOS(); err != nil {
			return err
		}
		query := args[0]

		if localSpaceID == "" {
			return fmt.Errorf("--space-id is required")
		}

		uri := fmt.Sprintf("craftdocs://opensearch?spaceId=%s&query=%s",
			url.QueryEscape(localSpaceID),
			url.QueryEscape(query))

		return openURL(uri)
	},
}

var localTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "Open today's daily note",
	Long:  "Open today's daily note in the Craft app",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkMacOS(); err != nil {
			return err
		}
		if localSpaceID == "" {
			return fmt.Errorf("--space-id is required")
		}
		uri := fmt.Sprintf("craftdocs://openByQuery?query=today&spaceId=%s", url.QueryEscape(localSpaceID))
		return openURL(uri)
	},
}

var localYesterdayCmd = &cobra.Command{
	Use:   "yesterday",
	Short: "Open yesterday's daily note",
	Long:  "Open yesterday's daily note in the Craft app",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkMacOS(); err != nil {
			return err
		}
		if localSpaceID == "" {
			return fmt.Errorf("--space-id is required")
		}
		uri := fmt.Sprintf("craftdocs://openByQuery?query=yesterday&spaceId=%s", url.QueryEscape(localSpaceID))
		return openURL(uri)
	},
}

var localTomorrowCmd = &cobra.Command{
	Use:   "tomorrow",
	Short: "Open tomorrow's daily note",
	Long:  "Open tomorrow's daily note in the Craft app",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkMacOS(); err != nil {
			return err
		}
		if localSpaceID == "" {
			return fmt.Errorf("--space-id is required")
		}
		uri := fmt.Sprintf("craftdocs://openByQuery?query=tomorrow&spaceId=%s", url.QueryEscape(localSpaceID))
		return openURL(uri)
	},
}

func init() {
	rootCmd.AddCommand(localCmd)

	// Add persistent flags for space ID
	localCmd.PersistentFlags().StringVar(&localSpaceID, "space-id", "", "Craft space ID (get from Copy Deeplink)")

	// Open command
	localCmd.AddCommand(localOpenCmd)

	// Open space command
	localCmd.AddCommand(localOpenSpaceCmd)
	localOpenSpaceCmd.Flags().String("tab", "", "Tab to open (calendar, search, documents)")

	// New document command
	localCmd.AddCommand(localNewCmd)
	localNewCmd.Flags().StringVar(&localTitle, "title", "", "Document title")
	localNewCmd.Flags().StringVar(&localContent, "content", "", "Document content (markdown)")
	localNewCmd.Flags().StringVar(&localFolder, "folder", "", "Folder ID to create in")

	// Append command
	localCmd.AddCommand(localAppendCmd)

	// Search command
	localCmd.AddCommand(localSearchCmd)

	// Daily notes commands
	localCmd.AddCommand(localTodayCmd)
	localCmd.AddCommand(localYesterdayCmd)
	localCmd.AddCommand(localTomorrowCmd)
}

func checkMacOS() error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("local commands are only supported on macOS (current: %s)", runtime.GOOS)
	}
	return nil
}

func openURL(uri string) error {
	cmd := exec.Command("open", uri)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open URL: %w", err)
	}
	return nil
}
