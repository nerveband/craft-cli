package cmd

import (
	"fmt"

	"github.com/ashrafali/craft-cli/internal/api"
	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	listFolderID       string
	listLocation       string
	listCreatedAfter   string
	listCreatedBefore  string
	listModifiedAfter  string
	listModifiedBefore string
	listMetadata       bool
	listLimit          int
	listCount          bool
	listFields         string
	listCursor         string
	listOffset         int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all documents",
	Long: `Retrieve and display documents from Craft.

Filters:
  --folder ID              List documents in a specific folder
  --location LOC           List documents in a special location:
                           unsorted, trash, templates, daily_notes
  --created-after DATE     Documents created on or after DATE (YYYY-MM-DD)
  --created-before DATE    Documents created on or before DATE (YYYY-MM-DD)
  --modified-after DATE    Documents modified on or after DATE (YYYY-MM-DD)
  --modified-before DATE   Documents modified on or before DATE (YYYY-MM-DD)
  --metadata               Fetch document metadata (dates, authors)

Examples:
  craft list                                      # List all documents
  craft list --format table                       # List as table
  craft list --folder abc123                      # List documents in folder
  craft list --location unsorted                  # List unsorted documents
  craft list --created-after 2025-01-01           # Created since Jan 2025
  craft list --modified-after 2025-06-01 --metadata  # Recently modified with metadata`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if backendName == "mcp" || listCursor != "" || listOffset > 0 {
			return runMCPListDocuments()
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		useAdvanced := listCreatedAfter != "" || listCreatedBefore != "" ||
			listModifiedAfter != "" || listModifiedBefore != "" || listMetadata

		var result *models.DocumentList
		if useAdvanced {
			opts := api.ListDocumentsOptions{
				FolderID:            listFolderID,
				Location:            listLocation,
				FetchMetadata:       listMetadata,
				CreatedDateGte:      listCreatedAfter,
				CreatedDateLte:      listCreatedBefore,
				LastModifiedDateGte: listModifiedAfter,
				LastModifiedDateLte: listModifiedBefore,
			}
			result, err = client.GetDocumentsAdvanced(opts)
		} else {
			result, err = client.GetDocumentsFiltered(listFolderID, listLocation)
		}
		if err != nil {
			return err
		}

		if listCount {
			return outputCount(result.Total)
		}

		originalReturned := len(result.Items)
		truncated := listLimit > 0 && len(result.Items) > listLimit
		if truncated {
			result.Items = result.Items[:listLimit]
		}

		format := getOutputFormat()
		if listFields != "" {
			items, err := projectItems(result.Items, listFields)
			if err != nil {
				return err
			}
			return outputJSON(payloadWithMetadata(items, result.Total, truncated, originalReturned, len(result.Items), "craft list --limit 0"))
		}
		if format == FormatJSON {
			if truncated {
				return outputJSON(payloadWithMetadata(result.Items, result.Total, true, originalReturned, len(result.Items), "craft list --limit 0"))
			}
			return outputDocumentsPayload(result, format)
		}
		return outputDocuments(result.Items, format)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listFolderID, "folder", "", "Filter by folder ID")
	listCmd.Flags().StringVar(&listLocation, "location", "", "Filter by location: unsorted, trash, templates, daily_notes")
	listCmd.Flags().StringVar(&listCreatedAfter, "created-after", "", "Filter documents created on or after this date (YYYY-MM-DD)")
	listCmd.Flags().StringVar(&listCreatedBefore, "created-before", "", "Filter documents created on or before this date (YYYY-MM-DD)")
	listCmd.Flags().StringVar(&listModifiedAfter, "modified-after", "", "Filter documents modified on or after this date (YYYY-MM-DD)")
	listCmd.Flags().StringVar(&listModifiedBefore, "modified-before", "", "Filter documents modified on or before this date (YYYY-MM-DD)")
	listCmd.Flags().BoolVar(&listMetadata, "metadata", false, "Fetch document metadata (dates, authors)")
	listCmd.Flags().IntVar(&listLimit, "limit", 0, "Maximum number of results to return (0 = all)")
	listCmd.Flags().BoolVar(&listCount, "count", false, "Output only the matching document count")
	listCmd.Flags().StringVar(&listFields, "fields", "", "Comma-separated fields to include in JSON output (e.g. id,title,lastModifiedAt)")
	listCmd.Flags().StringVar(&listCursor, "cursor", "", "MCP pagination cursor")
	listCmd.Flags().IntVar(&listOffset, "offset", 0, "MCP pagination offset")
}

func runMCPListDocuments() error {
	command := "documents list"
	if listLocation != "" {
		command += " --location " + quoteMCPArg(listLocation)
	}
	if listFolderID != "" {
		command += " --folder " + quoteMCPArg(listFolderID)
	}
	if listLimit > 0 {
		command += fmt.Sprintf(" --limit %d", listLimit)
	}
	if listCursor != "" {
		command += " --cursor " + quoteMCPArg(listCursor)
	}
	if listOffset > 0 {
		command += fmt.Sprintf(" --offset %d", listOffset)
	}
	if isDryRun() {
		return dryRunOutput("list documents", map[string]interface{}{
			"backend":      "mcp",
			"tool":         "craft_read",
			"command":      command,
			"capabilities": []string{"mcp", "read", "cursor"},
		})
	}
	return runMCPReadCommand(command)
}
