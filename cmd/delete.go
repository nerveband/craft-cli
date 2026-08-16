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
	Use:   "delete <document-id...>",
	Short: "Move one or more documents to trash",
	Long: `Soft-delete documents by moving them to Craft trash.

This uses DELETE /documents (you can restore via documents/move).
Multiple IDs are sent in a single API request.

Use --dry-run to preview what would be deleted without making changes.

Examples:
  craft delete abc123
  craft delete abc123 def456 ghi789   # Batch delete in one request
  craft delete --json '{"documentIds":["abc123"]}' --dry-run
  craft delete abc123 --dry-run       # Preview without deleting
  craft delete abc123 -q              # Silent delete`,
	Args: func(cmd *cobra.Command, args []string) error {
		if deleteJSON != "" || deleteStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if deleteJSON != "" || deleteStdin {
			payload, err := readRawPayload(deleteJSON, deleteStdin)
			if err != nil {
				return err
			}
			return runDeleteRaw(payload)
		}

		for _, id := range args {
			if err := validateResourceID(id, "document-id"); err != nil {
				return err
			}
		}

		if isDryRun() {
			client, err := getAPIClient()
			if err != nil {
				return err
			}
			if len(args) == 1 {
				doc, err := client.GetDocument(args[0])
				if err != nil {
					return fmt.Errorf("document not found: %s", args[0])
				}
				return dryRunOutput("delete", map[string]interface{}{
					"id": doc.ID, "title": doc.Title, "reversible": true,
				})
			}
			targets := make([]map[string]interface{}, 0, len(args))
			for _, id := range args {
				doc, err := client.GetDocument(id)
				if err != nil {
					return fmt.Errorf("document not found: %s", id)
				}
				targets = append(targets, map[string]interface{}{"id": doc.ID, "title": doc.Title})
			}
			return dryRunOutput("delete", map[string]interface{}{
				"documents": targets, "count": len(targets), "reversible": true,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.DeleteDocuments(args); err != nil {
			return err
		}

		if len(args) == 1 {
			outputDeleted(args[0])
			return nil
		}
		return outputJSON(map[string]interface{}{"deleted": len(args), "ids": args})
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
	docIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		docID, ok := id.(string)
		if !ok || docID == "" {
			return fmt.Errorf("documentIds must contain strings")
		}
		docIDs = append(docIDs, docID)
	}
	if err := client.DeleteDocuments(docIDs); err != nil {
		return err
	}
	return outputJSON(map[string]interface{}{"deleted": len(docIDs)})
}
