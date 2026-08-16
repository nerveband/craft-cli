package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	moveTargetFolder   string
	moveTargetLocation string
	moveJSON           string
	moveStdin          bool
)

var moveCmd = &cobra.Command{
	Use:   "move <document-id...>",
	Short: "Move one or more documents to a folder or location",
	Long: `Move documents to a different folder or special location.
Multiple IDs are sent in a single API request.

Locations:
  unsorted   - Move to unsorted documents
  trash      - Move to trash

Examples:
  craft move abc123 --to-folder def456        # Move to folder
  craft move abc123 def456 --to-folder xyz789 # Batch move
  craft move abc123 --to-location unsorted    # Move to unsorted
  craft move abc123 --to-location trash       # Move to trash
  craft move --json '{"documentIds":["abc123"],"destination":"unsorted"}' --dry-run`,
	Args: func(cmd *cobra.Command, args []string) error {
		if moveJSON != "" || moveStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if moveJSON != "" || moveStdin {
			payload, err := readRawPayload(moveJSON, moveStdin)
			if err != nil {
				return err
			}
			return runMoveRaw(payload)
		}

		if moveTargetFolder == "" && moveTargetLocation == "" {
			return fmt.Errorf("either --to-folder or --to-location is required")
		}
		for _, id := range args {
			if err := validateResourceID(id, "document-id"); err != nil {
				return err
			}
		}
		if moveTargetFolder != "" {
			if err := validateResourceID(moveTargetFolder, "folder-id"); err != nil {
				return err
			}
		}

		if isDryRun() {
			target := map[string]interface{}{}
			if len(args) == 1 {
				target["id"] = args[0]
			} else {
				target["ids"] = args
				target["count"] = len(args)
			}
			if moveTargetFolder != "" {
				target["destination_folder"] = moveTargetFolder
			} else {
				target["destination_location"] = moveTargetLocation
			}
			return dryRunOutput("move", target)
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.MoveDocuments(args, moveTargetFolder, moveTargetLocation); err != nil {
			return err
		}

		if !isQuiet() {
			destination := moveTargetLocation
			if moveTargetFolder != "" {
				destination = "folder " + moveTargetFolder
			}
			if len(args) == 1 {
				if moveTargetFolder != "" {
					fmt.Printf("Document %s moved to folder %s\n", args[0], moveTargetFolder)
				} else {
					fmt.Printf("Document %s moved to %s\n", args[0], moveTargetLocation)
				}
			} else {
				fmt.Printf("%d documents moved to %s\n", len(args), destination)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(moveCmd)
	moveCmd.Flags().StringVar(&moveTargetFolder, "to-folder", "", "Target folder ID")
	moveCmd.Flags().StringVar(&moveTargetLocation, "to-location", "", "Target location (unsorted, trash)")
	moveCmd.Flags().StringVar(&moveJSON, "json", "", "Raw REST move documents payload JSON")
	moveCmd.Flags().BoolVar(&moveStdin, "stdin", false, "Read raw REST move documents payload from stdin")
}

func runMoveRaw(payload map[string]interface{}) error {
	ids, ok := payloadArray(payload, "documentIds")
	if !ok {
		if id := payloadString(payload, "id", "documentId"); id != "" {
			ids = []interface{}{id}
			ok = true
		}
	}
	if !ok || len(ids) == 0 {
		return fmt.Errorf("raw move payload must include non-empty \"documentIds\" array or \"id\"")
	}
	folderID := payloadString(payload, "folderId", "folderID", "destinationFolderId", "destination_folder")
	location := payloadString(payload, "location", "destination", "destinationLocation", "destination_location")
	if folderID == "" && location == "" {
		return fmt.Errorf("raw move payload must include folderId or location/destination")
	}
	if isDryRun() {
		return dryRunOutput("move documents", map[string]interface{}{"payload": payload, "count": len(ids)})
	}
	client, err := getAPIClient()
	if err != nil {
		return err
	}
	docIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		docID, ok := id.(string)
		if !ok || docID == "" {
			return fmt.Errorf("documentIds must contain strings")
		}
		docIDs = append(docIDs, docID)
	}
	if err := client.MoveDocuments(docIDs, folderID, location); err != nil {
		return err
	}
	return outputJSON(map[string]interface{}{"moved": len(docIDs)})
}
