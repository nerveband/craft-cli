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
		if format == FormatJSON {
			return outputJSON(folders)
		}
		return outputFolders(folders.Items, format)
	},
}

var (
	folderParentID string
	folderJSON     string
	folderStdin    bool
)

var foldersCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new folder",
	Long:  "Create a new folder in your Craft space",
	Args: func(cmd *cobra.Command, args []string) error {
		if folderJSON != "" || folderStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if folderJSON != "" || folderStdin {
			payload, err := readRawPayload(folderJSON, folderStdin)
			if err != nil {
				return err
			}
			name := payloadString(payload, "name")
			if name == "" {
				return fmt.Errorf("raw folder create payload must include \"name\"")
			}
			parentID := payloadString(payload, "parentId", "parentID", "parent")
			if isDryRun() {
				return dryRunOutput("create folder", map[string]interface{}{"payload": payload})
			}
			client, err := getAPIClient()
			if err != nil {
				return err
			}
			folder, err := client.CreateFolder(name, parentID)
			if err != nil {
				return err
			}
			if isQuiet() {
				fmt.Println(folder.ID)
				return nil
			}
			return outputFolder(folder, getOutputFormat())
		}

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
	folderTargetID    string
	folderIconOffset  int
	folderDeleteJSON  string
	folderDeleteStdin bool
)

var foldersMoveCmd = &cobra.Command{
	Use:   "move [folder-id]",
	Short: "Move a folder to a new parent",
	Long: `Move a folder to a different parent folder.

Examples:
  craft folders move abc123 --to def456   # Move to another folder
  craft folders move abc123 --to root     # Move to root level`,
	Args: func(cmd *cobra.Command, args []string) error {
		if folderJSON != "" || folderStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if folderJSON != "" || folderStdin {
			payload, err := readRawPayload(folderJSON, folderStdin)
			if err != nil {
				return err
			}
			ids, ok := payloadArray(payload, "folderIds")
			if !ok {
				if id := payloadString(payload, "id", "folderId"); id != "" {
					ids = []interface{}{id}
					ok = true
				}
			}
			if !ok || len(ids) == 0 {
				return fmt.Errorf("raw folder move payload must include non-empty \"folderIds\" array or \"id\"")
			}
			targetID := payloadString(payload, "parentFolderId", "parentId", "to", "destination")
			if targetID == "root" {
				targetID = ""
			}
			if isDryRun() {
				return dryRunOutput("move folders", map[string]interface{}{"payload": payload, "count": len(ids)})
			}
			client, err := getAPIClient()
			if err != nil {
				return err
			}
			for _, id := range ids {
				folderID, ok := id.(string)
				if !ok || folderID == "" {
					return fmt.Errorf("folderIds must contain strings")
				}
				if err := client.MoveFolder(folderID, targetID); err != nil {
					return err
				}
			}
			return outputJSON(map[string]interface{}{"moved": len(ids)})
		}

		if folderTargetID == "" {
			return fmt.Errorf("--to is required")
		}

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
	Args: func(cmd *cobra.Command, args []string) error {
		if folderDeleteJSON != "" || folderDeleteStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if folderDeleteJSON != "" || folderDeleteStdin {
			payload, err := readRawPayload(folderDeleteJSON, folderDeleteStdin)
			if err != nil {
				return err
			}
			ids, ok := payloadArray(payload, "folderIds")
			if !ok {
				if id := payloadString(payload, "id", "folderId"); id != "" {
					ids = []interface{}{id}
					ok = true
				}
			}
			if !ok || len(ids) == 0 {
				return fmt.Errorf("raw folder delete payload must include non-empty \"folderIds\" array or \"id\"")
			}
			if isDryRun() {
				return dryRunOutput("delete folders", map[string]interface{}{"payload": payload, "count": len(ids), "destructive": true})
			}
			client, err := getAPIClient()
			if err != nil {
				return err
			}
			for _, id := range ids {
				folderID, ok := id.(string)
				if !ok || folderID == "" {
					return fmt.Errorf("folderIds must contain strings")
				}
				if err := client.DeleteFolder(folderID); err != nil {
					return err
				}
			}
			return outputJSON(map[string]interface{}{"deleted": len(ids)})
		}

		if isDryRun() {
			return dryRunOutput("delete folder", map[string]interface{}{
				"id": args[0], "destructive": true,
			})
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

var foldersExploreIconsCmd = &cobra.Command{
	Use:   "explore-icons [query]",
	Short: "Search Craft folder icons through MCP",
	Long:  "Search Craft folder icons through the Craft MCP server.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		command := "folders explore-icons " + quoteMCPArg(args[0])
		if folderIconOffset > 0 {
			command += fmt.Sprintf(" --offset %d", folderIconOffset)
		}
		return runMCPReadCommand(command)
	},
}

func init() {
	rootCmd.AddCommand(foldersCmd)

	foldersCmd.AddCommand(foldersListCmd)

	foldersCmd.AddCommand(foldersCreateCmd)
	foldersCreateCmd.Flags().StringVar(&folderParentID, "parent", "", "Parent folder ID (optional)")
	foldersCreateCmd.Flags().StringVar(&folderJSON, "json", "", "Raw folder create payload JSON")
	foldersCreateCmd.Flags().BoolVar(&folderStdin, "stdin", false, "Read raw folder create payload from stdin")

	foldersCmd.AddCommand(foldersMoveCmd)
	foldersMoveCmd.Flags().StringVar(&folderTargetID, "to", "", "Target parent folder ID (use 'root' for root level)")
	foldersMoveCmd.Flags().StringVar(&folderJSON, "json", "", "Raw folder move payload JSON")
	foldersMoveCmd.Flags().BoolVar(&folderStdin, "stdin", false, "Read raw folder move payload from stdin")

	foldersCmd.AddCommand(foldersDeleteCmd)
	foldersDeleteCmd.Flags().StringVar(&folderDeleteJSON, "json", "", "Raw folder delete payload JSON")
	foldersDeleteCmd.Flags().BoolVar(&folderDeleteStdin, "stdin", false, "Read raw folder delete payload from stdin")

	foldersCmd.AddCommand(foldersExploreIconsCmd)
	foldersExploreIconsCmd.Flags().IntVar(&folderIconOffset, "offset", 0, "Result offset for pagination")
}

// outputFolders prints folders in the specified format
func outputFolders(folders []models.Folder, format string) error {
	switch format {
	case FormatCompact:
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
