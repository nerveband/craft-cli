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
	Use:   "move [document-id]",
	Short: "Move a document to a folder or location",
	Long: `Move a document to a different folder or special location.

Locations:
  unsorted   - Move to unsorted documents
  trash      - Move to trash

Examples:
  craft move abc123 --to-folder def456    # Move to folder
  craft move abc123 --to-location unsorted # Move to unsorted
  craft move abc123 --to-location trash    # Move to trash
  craft move --json '{"documentIds":["abc123"],"destination":"unsorted"}' --dry-run`,
	Args: func(cmd *cobra.Command, args []string) error {
		if moveJSON != "" || moveStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
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
		docID := args[0]
		if err := validateResourceID(docID, "document-id"); err != nil {
			return err
		}
		if moveTargetFolder != "" {
			if err := validateResourceID(moveTargetFolder, "folder-id"); err != nil {
				return err
			}
		}

		if isDryRun() {
			target := map[string]interface{}{"id": docID}
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

		if err := client.MoveDocument(docID, moveTargetFolder, moveTargetLocation); err != nil {
			return err
		}

		if !isQuiet() {
			if moveTargetFolder != "" {
				fmt.Printf("Document %s moved to folder %s\n", docID, moveTargetFolder)
			} else {
				fmt.Printf("Document %s moved to %s\n", docID, moveTargetLocation)
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
	for _, id := range ids {
		docID, ok := id.(string)
		if !ok || docID == "" {
			return fmt.Errorf("documentIds must contain strings")
		}
		if err := validateResourceID(docID, "document-id"); err != nil {
			return err
		}
		if folderID != "" {
			if err := validateResourceID(folderID, "folder-id"); err != nil {
				return err
			}
		}
		if err := client.MoveDocument(docID, folderID, location); err != nil {
			return err
		}
	}
	return outputJSON(map[string]interface{}{"moved": len(ids)})
}
