package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var foldersCmd = &cobra.Command{
	Use:   "folders",
	Short: "Manage folders",
	Long: `Manage Craft folders - list, create, move, and delete folders.

Examples:
  craft folders list                           # List all folders
  craft folders list --format json             # List as JSON
  craft folders create "New Folder"            # Create a folder
  craft folders create "Subfolder" --parent ID # Create nested folder
  craft folders move ID --to PARENT_ID         # Move folder
  craft folders delete ID                      # Delete folder`,
}

var foldersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all folders",
	Long:  "List all folders in your Craft space",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		folders, err := client.GetFolders()
		if err != nil {
			return err
		}

		format := getOutputFormat()
		return outputFolders(folders.Items, format)
	},
}

var (
	folderParentID string
)

var foldersCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new folder",
	Long:  "Create a new folder in your Craft space",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		name := args[0]
		folder, err := client.CreateFolder(name, folderParentID)
		if err != nil {
			return err
		}

		if isQuiet() {
			fmt.Println(folder.ID)
			return nil
		}

		format := getOutputFormat()
		return outputFolder(folder, format)
	},
}

var (
	folderTargetID string
)

var foldersMoveCmd = &cobra.Command{
	Use:   "move [folder-id]",
	Short: "Move a folder to a new parent",
	Long: `Move a folder to a different parent folder.

Examples:
  craft folders move abc123 --to def456   # Move to another folder
  craft folders move abc123 --to root     # Move to root level`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		folderID := args[0]
		targetID := folderTargetID
		if targetID == "root" {
			targetID = ""
		}

		if err := client.MoveFolder(folderID, targetID); err != nil {
			return err
		}

		if !isQuiet() {
			if targetID == "" {
				fmt.Printf("Folder %s moved to root\n", folderID)
			} else {
				fmt.Printf("Folder %s moved to %s\n", folderID, targetID)
			}
		}
		return nil
	},
}

var foldersDeleteCmd = &cobra.Command{
	Use:   "delete [folder-id]",
	Short: "Delete a folder",
	Long:  "Delete a folder. Contents will be moved to the parent folder.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isDryRun() {
			fmt.Printf("[dry-run] Would delete folder: %s\n", args[0])
			return nil
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		folderID := args[0]
		if err := client.DeleteFolder(folderID); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Folder %s deleted\n", folderID)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(foldersCmd)

	foldersCmd.AddCommand(foldersListCmd)

	foldersCmd.AddCommand(foldersCreateCmd)
	foldersCreateCmd.Flags().StringVar(&folderParentID, "parent", "", "Parent folder ID (optional)")

	foldersCmd.AddCommand(foldersMoveCmd)
	foldersMoveCmd.Flags().StringVar(&folderTargetID, "to", "", "Target parent folder ID (use 'root' for root level)")
	foldersMoveCmd.MarkFlagRequired("to")

	foldersCmd.AddCommand(foldersDeleteCmd)
}

// outputFolders prints folders in the specified format
func outputFolders(folders []models.Folder, format string) error {
	switch format {
	case FormatJSON, FormatCompact:
		return outputJSON(folders)
	case "table":
		return outputFoldersTable(folders)
	case "markdown":
		return outputFoldersMarkdown(folders)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// outputFoldersTable prints folders as a table
func outputFoldersTable(folders []models.Folder) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !hasNoHeaders() {
		fmt.Fprintln(w, "ID\tNAME\tPARENT\tDOCUMENTS")
		fmt.Fprintln(w, "---\t----\t------\t---------")
	}

	for _, f := range folders {
		parent := f.ParentID
		if parent == "" {
			parent = "(root)"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", f.ID, f.Name, parent, f.DocumentCount)
	}

	return w.Flush()
}

// outputFoldersMarkdown prints folders as markdown
func outputFoldersMarkdown(folders []models.Folder) error {
	fmt.Println("# Folders")
	for _, f := range folders {
		fmt.Printf("## %s\n", f.Name)
		fmt.Printf("- **ID**: %s\n", f.ID)
		if f.ParentID != "" {
			fmt.Printf("- **Parent**: %s\n", f.ParentID)
		}
		fmt.Printf("- **Documents**: %d\n", f.DocumentCount)
		fmt.Println()
	}
	return nil
}

// outputFolder prints a single folder
func outputFolder(folder *models.Folder, format string) error {
	switch format {
	case FormatJSON, FormatCompact:
		return outputJSON(folder)
	case "table", "markdown":
		fmt.Printf("Created folder: %s (ID: %s)\n", folder.Name, folder.ID)
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}
