package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	deleteJSON  string
	deleteStdin bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete <document-id>",
	Short: "Move a document to trash",
	Long: `Soft-delete a document by moving it to Craft trash.

This uses DELETE /documents (you can restore via documents/move).

Use --dry-run to preview what would be deleted without making changes.

Examples:
  craft delete abc123
  craft delete --json '{"documentIds":["abc123"]}' --dry-run
  craft delete abc123 --dry-run    # Preview without deleting
  craft delete abc123 -q           # Silent delete`,
	Args: func(cmd *cobra.Command, args []string) error {
		if deleteJSON != "" || deleteStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if deleteJSON != "" || deleteStdin {
			payload, err := readRawPayload(deleteJSON, deleteStdin)
			if err != nil {
				return err
			}
			return runDeleteRaw(payload)
		}

		docID := args[0]
		if err := validateResourceID(docID, "document-id"); err != nil {
			return err
		}

		if isDryRun() {
			client, err := getAPIClient()
			if err != nil {
				return err
			}
			doc, err := client.GetDocument(docID)
			if err != nil {
				return fmt.Errorf("document not found: %s", docID)
			}
			return dryRunOutput("delete", map[string]interface{}{
				"id": doc.ID, "title": doc.Title, "reversible": true,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.DeleteDocument(docID); err != nil {
			return err
		}

		outputDeleted(docID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVar(&deleteJSON, "json", "", "Raw REST delete documents payload JSON")
	deleteCmd.Flags().BoolVar(&deleteStdin, "stdin", false, "Read raw REST delete documents payload from stdin")
}

func runDeleteRaw(payload map[string]interface{}) error {
	ids, ok := payloadArray(payload, "documentIds")
	if !ok {
		if id := payloadString(payload, "id", "documentId"); id != "" {
			ids = []interface{}{id}
			ok = true
		}
	}
	if !ok || len(ids) == 0 {
		return fmt.Errorf("raw delete payload must include non-empty \"documentIds\" array or \"id\"")
	}
	if isDryRun() {
		return dryRunOutput("delete documents", map[string]interface{}{
			"payload":    payload,
			"count":      len(ids),
			"reversible": true,
		})
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
		if err := client.DeleteDocument(docID); err != nil {
			return err
		}
	}
	return outputJSON(map[string]interface{}{"deleted": len(ids)})
}
