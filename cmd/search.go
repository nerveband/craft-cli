package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ashrafali/craft-cli/internal/api"
	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	searchDocument      string
	searchRegex         string
	searchCaseSensitive bool
	searchContext        int
	searchLocation      string
	searchFolder        string
	searchMetadata      bool
	searchCreatedAfter  string
	searchCreatedBefore string
	searchModifiedAfter string
	searchModifiedBefore string
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for documents or blocks",
	Long: `Search for documents matching a query, or search within a document for blocks.

Document search (default):
  craft search "meeting notes"
  craft search --regex "TODO|FIXME"
  craft search "project" --location daily_notes
  craft search "budget" --folder <folder-id> --metadata
  craft search --created-after 2024-01-01 --modified-before 2024-12-31 "report"

Block search (with --document):
  craft search --document <doc-id> "keyword"
  craft search --document <doc-id> --regex "pattern" --case-sensitive
  craft search --document <doc-id> "term" --context 10`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		format := getOutputFormat()

		// Block search mode: --document is set
		if searchDocument != "" {
			return runBlockSearch(client, args, format)
		}

		// Document search mode (default)
		return runDocumentSearch(client, args, format)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVar(&searchDocument, "document", "", "Document/block ID for block-level search")
	searchCmd.Flags().StringVar(&searchRegex, "regex", "", "RE2 regex pattern")
	searchCmd.Flags().BoolVar(&searchCaseSensitive, "case-sensitive", false, "Case-sensitive matching (block search only)")
	searchCmd.Flags().IntVar(&searchContext, "context", 5, "Number of surrounding blocks to include (block search only)")
	searchCmd.Flags().StringVar(&searchLocation, "location", "", "Filter by location: unsorted, trash, templates, daily_notes")
	searchCmd.Flags().StringVar(&searchFolder, "folder", "", "Filter by folder ID")
	searchCmd.Flags().BoolVar(&searchMetadata, "metadata", false, "Include document metadata in results")
	searchCmd.Flags().StringVar(&searchCreatedAfter, "created-after", "", "Filter: created on or after date (YYYY-MM-DD)")
	searchCmd.Flags().StringVar(&searchCreatedBefore, "created-before", "", "Filter: created on or before date (YYYY-MM-DD)")
	searchCmd.Flags().StringVar(&searchModifiedAfter, "modified-after", "", "Filter: modified on or after date (YYYY-MM-DD)")
	searchCmd.Flags().StringVar(&searchModifiedBefore, "modified-before", "", "Filter: modified on or before date (YYYY-MM-DD)")
}

// runBlockSearch executes a block-level search within a document.
func runBlockSearch(client *api.Client, args []string, format string) error {
	// Determine the search pattern: positional arg or --regex
	pattern := ""
	if len(args) > 0 {
		pattern = args[0]
	}
	if searchRegex != "" {
		pattern = searchRegex
	}
	if pattern == "" {
		return fmt.Errorf("block search requires a query argument or --regex pattern")
	}

	result, err := client.SearchBlocks(searchDocument, pattern, searchCaseSensitive, searchContext, searchContext)
	if err != nil {
		return err
	}

	if len(result.Items) == 0 {
		printStatus("No matching blocks found\n")
		// Still output empty result in the requested format
		return outputBlockSearchResults(result.Items, format)
	}

	printStatus("Found %d matching block(s)\n", len(result.Items))
	return outputBlockSearchResults(result.Items, format)
}

// runDocumentSearch executes an advanced document search.
func runDocumentSearch(client *api.Client, args []string, format string) error {
	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	// If no query and no regex, require at least one
	if query == "" && searchRegex == "" {
		return fmt.Errorf("search requires a query argument or --regex pattern")
	}

	opts := api.SearchOptions{
		Regexps:             searchRegex,
		Location:            searchLocation,
		FolderIDs:           searchFolder,
		FetchMetadata:       searchMetadata,
		CreatedDateGte:      searchCreatedAfter,
		CreatedDateLte:      searchCreatedBefore,
		LastModifiedDateGte: searchModifiedAfter,
		LastModifiedDateLte: searchModifiedBefore,
	}

	result, err := client.SearchDocumentsAdvanced(query, opts)
	if err != nil {
		return err
	}

	if len(result.Items) == 0 {
		printStatus("No matching documents found\n")
	} else {
		printStatus("Found %d result(s)\n", len(result.Items))
	}

	return outputSearchResults(result.Items, format)
}

// --- Block search output functions ---

// outputBlockSearchResults dispatches block search results to the appropriate formatter.
func outputBlockSearchResults(items []models.BlockSearchResult, format string) error {
	switch format {
	case FormatJSON, FormatCompact:
		return outputJSON(items)
	case "table":
		return outputBlockSearchTable(items)
	case "markdown":
		return outputBlockSearchMarkdown(items)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// outputBlockSearchTable prints block search results as a table with BLOCK_ID, MATCH, PATH columns.
func outputBlockSearchTable(items []models.BlockSearchResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !hasNoHeaders() {
		fmt.Fprintln(w, "BLOCK_ID\tMATCH\tPATH")
		fmt.Fprintln(w, "--------\t-----\t----")
	}

	for _, item := range items {
		match := strings.ReplaceAll(item.Markdown, "\n", " ")
		if len(match) > 60 {
			match = match[:57] + "..."
		}

		// Build path from PageBlockPath entries
		path := buildBlockPath(item.PageBlockPath)
		if len(path) > 40 {
			path = path[:37] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", item.BlockID, match, path)
	}

	return w.Flush()
}

// outputBlockSearchMarkdown prints block search results in a rich markdown format
// showing path, before context, the match (highlighted), and after context.
func outputBlockSearchMarkdown(items []models.BlockSearchResult) error {
	fmt.Println("# Block Search Results")
	fmt.Printf("*%d match(es)*\n\n", len(items))

	for i, item := range items {
		// Path breadcrumb
		path := buildBlockPath(item.PageBlockPath)
		if path != "" {
			fmt.Printf("## Match %d: %s\n", i+1, path)
		} else {
			fmt.Printf("## Match %d\n", i+1)
		}
		fmt.Printf("**Block ID**: `%s`\n\n", item.BlockID)

		// Before context
		if len(item.BeforeBlocks) > 0 {
			fmt.Println("*Context before:*")
			for _, b := range item.BeforeBlocks {
				fmt.Printf("> %s\n", b.Markdown)
			}
			fmt.Println()
		}

		// The matching block (highlighted)
		fmt.Printf("**>>> %s <<<**\n\n", item.Markdown)

		// After context
		if len(item.AfterBlocks) > 0 {
			fmt.Println("*Context after:*")
			for _, b := range item.AfterBlocks {
				fmt.Printf("> %s\n", b.Markdown)
			}
			fmt.Println()
		}

		fmt.Println("---")
		fmt.Println()
	}

	return nil
}

// buildBlockPath constructs a breadcrumb path from PageBlockPath entries.
func buildBlockPath(entries []models.PageBlockEntry) string {
	if len(entries) == 0 {
		return ""
	}

	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		label := entry.Content
		if label == "" {
			label = entry.ID
		}
		parts = append(parts, label)
	}
	return strings.Join(parts, " > ")
}
