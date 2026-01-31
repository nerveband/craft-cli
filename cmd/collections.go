package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var collectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "Manage collections",
	Long: `Manage Craft collections (databases) - list, view schema, and manage items.

Examples:
  craft collections list                                  # List all collections
  craft collections list --document ID                    # List collections in document
  craft collections schema COLLECTION_ID                  # Get collection schema
  craft collections items COLLECTION_ID                   # List items in collection
  craft collections add COLLECTION_ID --title "New Item"  # Add item
  craft collections update COLLECTION_ID --item ITEM_ID --properties '{"key":"value"}'
  craft collections delete COLLECTION_ID --item ITEM_ID   # Delete item`,
}

var (
	collectionDocumentID    string
	collectionSchemaFormat  string
	collectionItemDepth     int
	collectionItemTitle     string
	collectionItemProps     string
	collectionAllowNew      bool
	collectionItemID        string
)

var collectionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all collections",
	Long: `List all collections in your Craft space.

Use --document to filter collections by document ID.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		collections, err := client.GetCollections(collectionDocumentID)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		return outputCollections(collections.Items, format)
	},
}

var collectionsSchemaCmd = &cobra.Command{
	Use:   "schema [collection-id]",
	Short: "Get collection schema",
	Long: `Get the schema for a collection. Output is always JSON.

Schema formats:
  schema  - Standard schema format (default)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		collectionID := args[0]
		schema, err := client.GetCollectionSchema(collectionID, collectionSchemaFormat)
		if err != nil {
			return err
		}

		return outputJSON(schema)
	},
}

var collectionsItemsCmd = &cobra.Command{
	Use:   "items [collection-id]",
	Short: "List items in a collection",
	Long: `List all items in a collection.

Use --depth to control the depth of nested content returned.
Default depth is -1 (no limit).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		collectionID := args[0]
		items, err := client.GetCollectionItems(collectionID, collectionItemDepth)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		return outputCollectionItems(items.Items, format)
	},
}

var collectionsAddCmd = &cobra.Command{
	Use:   "add [collection-id]",
	Short: "Add an item to a collection",
	Long: `Add a new item to a collection.

Examples:
  craft collections add COLLECTION_ID --title "New Item"
  craft collections add COLLECTION_ID --title "Item" --properties '{"Status":"Active","Priority":"High"}'
  craft collections add COLLECTION_ID --title "Item" --properties '{"Tag":"new-value"}' --allow-new-options`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isDryRun() {
			fmt.Printf("[dry-run] Would add item to collection %s: %s\n", args[0], collectionItemTitle)
			return nil
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		collectionID := args[0]

		var props map[string]interface{}
		if collectionItemProps != "" {
			if err := json.Unmarshal([]byte(collectionItemProps), &props); err != nil {
				return fmt.Errorf("invalid --properties JSON: %w", err)
			}
		}

		result, err := client.AddCollectionItem(collectionID, collectionItemTitle, props, collectionAllowNew)
		if err != nil {
			return err
		}

		if isQuiet() {
			if len(result.Items) > 0 {
				fmt.Println(result.Items[0].ID)
			}
			return nil
		}

		format := getOutputFormat()
		if isJSONFormat(format) {
			return outputJSON(result)
		}

		if len(result.Items) > 0 {
			fmt.Printf("Item added: %s (ID: %s)\n", result.Items[0].Title, result.Items[0].ID)
		} else {
			fmt.Println("Item added")
		}
		return nil
	},
}

var collectionsUpdateCmd = &cobra.Command{
	Use:   "update [collection-id]",
	Short: "Update an item in a collection",
	Long: `Update an existing item in a collection.

Examples:
  craft collections update COLLECTION_ID --item ITEM_ID --properties '{"Status":"Done"}'
  craft collections update COLLECTION_ID --item ITEM_ID --properties '{"Tag":"new"}' --allow-new-options`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isDryRun() {
			fmt.Printf("[dry-run] Would update item %s in collection %s\n", collectionItemID, args[0])
			return nil
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		collectionID := args[0]

		var props map[string]interface{}
		if err := json.Unmarshal([]byte(collectionItemProps), &props); err != nil {
			return fmt.Errorf("invalid --properties JSON: %w", err)
		}

		if err := client.UpdateCollectionItem(collectionID, collectionItemID, props, collectionAllowNew); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Item %s updated in collection %s\n", collectionItemID, collectionID)
		}
		return nil
	},
}

var collectionsDeleteCmd = &cobra.Command{
	Use:   "delete [collection-id]",
	Short: "Delete an item from a collection",
	Long:  "Delete an item from a collection by its ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isDryRun() {
			fmt.Printf("[dry-run] Would delete item %s from collection %s\n", collectionItemID, args[0])
			return nil
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		collectionID := args[0]
		if err := client.DeleteCollectionItem(collectionID, collectionItemID); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Item %s deleted from collection %s\n", collectionItemID, collectionID)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(collectionsCmd)

	collectionsCmd.AddCommand(collectionsListCmd)
	collectionsListCmd.Flags().StringVar(&collectionDocumentID, "document", "", "Filter by document ID")

	collectionsCmd.AddCommand(collectionsSchemaCmd)
	collectionsSchemaCmd.Flags().StringVar(&collectionSchemaFormat, "schema-format", "schema", "Schema format (default: schema)")

	collectionsCmd.AddCommand(collectionsItemsCmd)
	collectionsItemsCmd.Flags().IntVar(&collectionItemDepth, "depth", -1, "Max depth of nested content (-1 for no limit)")

	collectionsCmd.AddCommand(collectionsAddCmd)
	collectionsAddCmd.Flags().StringVar(&collectionItemTitle, "title", "", "Item title (required)")
	collectionsAddCmd.Flags().StringVar(&collectionItemProps, "properties", "", "Item properties as JSON string")
	collectionsAddCmd.Flags().BoolVar(&collectionAllowNew, "allow-new-options", false, "Allow creating new options for select properties")
	collectionsAddCmd.MarkFlagRequired("title")

	collectionsCmd.AddCommand(collectionsUpdateCmd)
	collectionsUpdateCmd.Flags().StringVar(&collectionItemID, "item", "", "Item ID to update (required)")
	collectionsUpdateCmd.Flags().StringVar(&collectionItemProps, "properties", "", "Item properties as JSON string (required)")
	collectionsUpdateCmd.Flags().BoolVar(&collectionAllowNew, "allow-new-options", false, "Allow creating new options for select properties")
	collectionsUpdateCmd.MarkFlagRequired("item")
	collectionsUpdateCmd.MarkFlagRequired("properties")

	collectionsCmd.AddCommand(collectionsDeleteCmd)
	collectionsDeleteCmd.Flags().StringVar(&collectionItemID, "item", "", "Item ID to delete (required)")
	collectionsDeleteCmd.MarkFlagRequired("item")
}

// outputCollections prints collections in the specified format
func outputCollections(collections []models.Collection, format string) error {
	switch format {
	case FormatJSON, FormatCompact:
		return outputJSON(collections)
	case "table":
		return outputCollectionsTable(collections)
	case "markdown":
		return outputCollectionsMarkdown(collections)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// outputCollectionsTable prints collections as a table
func outputCollectionsTable(collections []models.Collection) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !hasNoHeaders() {
		fmt.Fprintln(w, "ID\tNAME\tITEMS\tDOCUMENT")
		fmt.Fprintln(w, "---\t----\t-----\t--------")
	}

	for _, c := range collections {
		docID := c.DocumentID
		if docID == "" {
			docID = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", c.ID, c.Name, c.ItemCount, docID)
	}

	return w.Flush()
}

// outputCollectionsMarkdown prints collections as markdown
func outputCollectionsMarkdown(collections []models.Collection) error {
	fmt.Println("# Collections")
	for _, c := range collections {
		fmt.Printf("## %s\n", c.Name)
		fmt.Printf("- **ID**: %s\n", c.ID)
		fmt.Printf("- **Items**: %d\n", c.ItemCount)
		if c.DocumentID != "" {
			fmt.Printf("- **Document**: %s\n", c.DocumentID)
		}
		fmt.Println()
	}
	return nil
}

// outputCollectionItems prints collection items in the specified format
func outputCollectionItems(items []models.CollectionItem, format string) error {
	switch format {
	case FormatJSON, FormatCompact:
		return outputJSON(items)
	case "table":
		return outputCollectionItemsTable(items)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// outputCollectionItemsTable prints collection items as a table
func outputCollectionItemsTable(items []models.CollectionItem) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !hasNoHeaders() {
		fmt.Fprintln(w, "ID\tTITLE\tPROPERTIES")
		fmt.Fprintln(w, "---\t-----\t----------")
	}

	for _, item := range items {
		title := item.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}

		propCount := len(item.Properties)
		propSummary := fmt.Sprintf("%d properties", propCount)
		if propCount == 0 {
			propSummary = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", item.ID, title, propSummary)
	}

	return w.Flush()
}
