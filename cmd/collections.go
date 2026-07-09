package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
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
	collectionDocumentID   string
	collectionSchemaFormat string
	collectionItemDepth    int
	collectionItemTitle    string
	collectionItemProps    string
	collectionAllowNew     bool
	collectionItemID       string
	collectionJSON         string
	collectionStdin        bool
	collectionName         string
	collectionViewID       string
	collectionViewName     string
	collectionViewType     string
	collectionProperties   []string
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
		if format == FormatJSON {
			return outputJSON(collections)
		}
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

var collectionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a collection",
	Long: `Create a Craft collection.

Use --json or --stdin for raw REST payloads. Use --backend mcp to route a
collection creation command through Craft MCP when richer property flags are
needed by the current MCP server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if backendName == "mcp" {
			command := "collections create"
			if collectionName != "" {
				command += " --name " + quoteMCPArg(collectionName)
			}
			for _, property := range collectionProperties {
				command += " " + propertyFlagToMCP(property)
			}
			if collectionJSON != "" {
				command += " --json " + quoteMCPArg(collectionJSON)
			}
			if collectionStdin {
				return fmt.Errorf("--stdin is not supported with --backend mcp for collections create; use --json")
			}
			return runMCPCollectionCommand("collections.create", command, "craft_write")
		}

		payload, err := readRawPayload(collectionJSON, collectionStdin)
		if err != nil {
			if collectionName == "" {
				return fmt.Errorf("collections create requires --json/--stdin, or --name with --backend mcp")
			}
			payload = map[string]interface{}{"name": collectionName}
		}
		if isDryRun() {
			return dryRunOutput("create collection", map[string]interface{}{"backend": "rest", "payload": payload})
		}
		client, err := getAPIClient()
		if err != nil {
			return err
		}
		result, err := client.CreateCollectionRaw(payload)
		if err != nil {
			return err
		}
		return outputJSON(result)
	},
}

var collectionsRenameCmd = &cobra.Command{
	Use:   "rename [collection-id]",
	Short: "Rename a collection through Craft MCP",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if collectionName == "" {
			return fmt.Errorf("--name is required")
		}
		command := "collections rename --collection " + quoteMCPArg(args[0]) + " --name " + quoteMCPArg(collectionName)
		return runMCPCollectionCommand("collections.rename", command, "craft_write")
	},
}

var collectionsSchemaUpdateCmd = &cobra.Command{
	Use:   "update [collection-id]",
	Short: "Update collection schema",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		payload, err := readRawPayload(collectionJSON, collectionStdin)
		if err != nil {
			return err
		}
		if isDryRun() {
			return dryRunOutput("update collection schema", map[string]interface{}{"backend": "rest", "collection_id": args[0], "payload": payload})
		}
		client, err := getAPIClient()
		if err != nil {
			return err
		}
		result, err := client.UpdateCollectionSchemaRaw(args[0], payload)
		if err != nil {
			return err
		}
		return outputJSON(result)
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
		if format == FormatJSON {
			return outputJSON(items)
		}
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
  craft collections add COLLECTION_ID --title "Item" --properties '{"Tag":"new-value"}' --allow-new-options
  craft collections add COLLECTION_ID --json '{"items":[{"title":"New Item"}]}' --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if collectionJSON != "" || collectionStdin {
			payload, err := readRawPayload(collectionJSON, collectionStdin)
			if err != nil {
				return err
			}
			items, ok := payloadArray(payload, "items")
			if !ok || len(items) == 0 {
				return fmt.Errorf("raw collection add payload must include non-empty \"items\" array")
			}
			if isDryRun() {
				return dryRunOutput("add collection items", map[string]interface{}{"collection_id": args[0], "payload": payload, "count": len(items)})
			}
			client, err := getAPIClient()
			if err != nil {
				return err
			}
			result, err := client.AddCollectionItemsRaw(args[0], payload)
			if err != nil {
				return err
			}
			return outputJSON(result)
		}
		if backendName == "mcp" {
			if collectionItemTitle == "" && len(collectionProperties) == 0 {
				return fmt.Errorf("--title or --property is required")
			}
			command := "collections items-add --collection " + quoteMCPArg(args[0])
			if collectionItemTitle != "" {
				command += " --title " + quoteMCPArg(collectionItemTitle)
			}
			for _, property := range collectionProperties {
				command += " " + propertyFlagToMCP(property)
			}
			return runMCPCollectionCommand("collections.items-add", command, "craft_write")
		}

		if isDryRun() {
			if collectionItemTitle == "" {
				return fmt.Errorf("--title is required")
			}
			return dryRunOutput("add collection item", map[string]interface{}{
				"collection_id": args[0], "title": collectionItemTitle,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		collectionID := args[0]
		if collectionItemTitle == "" {
			return fmt.Errorf("--title is required")
		}

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
  craft collections update COLLECTION_ID --item ITEM_ID --properties '{"Tag":"new"}' --allow-new-options
  craft collections update COLLECTION_ID --json '{"itemsToUpdate":[{"id":"ITEM_ID","properties":{"Status":"Done"}}]}' --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if collectionJSON != "" || collectionStdin {
			payload, err := readRawPayload(collectionJSON, collectionStdin)
			if err != nil {
				return err
			}
			items, ok := payloadArray(payload, "itemsToUpdate")
			if !ok || len(items) == 0 {
				return fmt.Errorf("raw collection update payload must include non-empty \"itemsToUpdate\" array")
			}
			if isDryRun() {
				return dryRunOutput("update collection items", map[string]interface{}{"collection_id": args[0], "payload": payload, "count": len(items)})
			}
			client, err := getAPIClient()
			if err != nil {
				return err
			}
			return client.UpdateCollectionItemsRaw(args[0], payload)
		}
		if backendName == "mcp" {
			if collectionItemID == "" {
				return fmt.Errorf("--item is required")
			}
			if len(collectionProperties) == 0 && collectionItemProps == "" {
				return fmt.Errorf("--property or --properties is required")
			}
			command := "collections items-update --collection " + quoteMCPArg(args[0]) + " --id " + quoteMCPArg(collectionItemID)
			if collectionItemProps != "" {
				command += " --properties " + quoteMCPArg(collectionItemProps)
			}
			for _, property := range collectionProperties {
				command += " " + propertyFlagToMCP(property)
			}
			return runMCPCollectionCommand("collections.items-update", command, "craft_write")
		}

		if isDryRun() {
			if collectionItemID == "" {
				return fmt.Errorf("--item is required")
			}
			if collectionItemProps == "" {
				return fmt.Errorf("--properties is required")
			}
			return dryRunOutput("update collection item", map[string]interface{}{
				"collection_id": args[0], "item_id": collectionItemID,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		collectionID := args[0]
		if collectionItemID == "" {
			return fmt.Errorf("--item is required")
		}
		if collectionItemProps == "" {
			return fmt.Errorf("--properties is required")
		}

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
	Long: `Delete an item from a collection by its ID.

Examples:
  craft collections delete COLLECTION_ID --item ITEM_ID
  craft collections delete COLLECTION_ID --json '{"idsToDelete":["ITEM_ID"]}' --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if collectionJSON != "" || collectionStdin {
			payload, err := readRawPayload(collectionJSON, collectionStdin)
			if err != nil {
				return err
			}
			ids, ok := payloadArray(payload, "idsToDelete")
			if !ok || len(ids) == 0 {
				return fmt.Errorf("raw collection delete payload must include non-empty \"idsToDelete\" array")
			}
			if isDryRun() {
				return dryRunOutput("delete collection items", map[string]interface{}{"collection_id": args[0], "payload": payload, "count": len(ids), "destructive": true})
			}
			client, err := getAPIClient()
			if err != nil {
				return err
			}
			return client.DeleteCollectionItemsRaw(args[0], payload)
		}

		if isDryRun() {
			if collectionItemID == "" {
				return fmt.Errorf("--item is required")
			}
			return dryRunOutput("delete collection item", map[string]interface{}{
				"collection_id": args[0], "item_id": collectionItemID, "destructive": true,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		collectionID := args[0]
		if collectionItemID == "" {
			return fmt.Errorf("--item is required")
		}
		if err := client.DeleteCollectionItem(collectionID, collectionItemID); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Item %s deleted from collection %s\n", collectionItemID, collectionID)
		}
		return nil
	},
}

var collectionsViewsCmd = &cobra.Command{
	Use:   "views",
	Short: "Manage collection views through Craft MCP",
	Long: `Manage collection views through Craft MCP.

The captured REST docs do not expose stable collection view endpoints, while
Craft MCP exposes the collection view command surface. These commands use
craft_read for list and craft_write for create/update/delete.`,
}

var collectionsViewsListCmd = &cobra.Command{
	Use:   "list [collection-id]",
	Short: "List collection views",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMCPCollectionCommand("collections.views.list", "collections views list "+quoteMCPArg(args[0]), "craft_read")
	},
}

var collectionsViewsCreateCmd = &cobra.Command{
	Use:   "create [collection-id]",
	Short: "Create a collection view",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		command := "collections views create " + quoteMCPArg(args[0])
		if collectionViewName != "" {
			command += " --name " + quoteMCPArg(collectionViewName)
		}
		if collectionViewType != "" {
			command += " --type " + quoteMCPArg(collectionViewType)
		}
		if collectionJSON != "" {
			command += " --json " + quoteMCPArg(collectionJSON)
		}
		if collectionStdin {
			return fmt.Errorf("--stdin is not supported for MCP collection view commands; use --json")
		}
		return runMCPCollectionCommand("collections.views.create", command, "craft_write")
	},
}

var collectionsViewsUpdateCmd = &cobra.Command{
	Use:   "update [collection-id]",
	Short: "Update a collection view",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if collectionViewID == "" {
			return fmt.Errorf("--view is required")
		}
		command := "collections views update " + quoteMCPArg(args[0]) + " --view " + quoteMCPArg(collectionViewID)
		if collectionViewName != "" {
			command += " --name " + quoteMCPArg(collectionViewName)
		}
		if collectionViewType != "" {
			command += " --type " + quoteMCPArg(collectionViewType)
		}
		if collectionJSON != "" {
			command += " --json " + quoteMCPArg(collectionJSON)
		}
		if collectionStdin {
			return fmt.Errorf("--stdin is not supported for MCP collection view commands; use --json")
		}
		return runMCPCollectionCommand("collections.views.update", command, "craft_write")
	},
}

var collectionsViewsDeleteCmd = &cobra.Command{
	Use:   "delete [collection-id]",
	Short: "Delete a collection view",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if collectionViewID == "" {
			return fmt.Errorf("--view is required")
		}
		command := "collections views delete " + quoteMCPArg(args[0]) + " --view " + quoteMCPArg(collectionViewID)
		return runMCPCollectionCommand("collections.views.delete", command, "craft_write")
	},
}

var collectionsActiveViewCmd = &cobra.Command{
	Use:   "active-view",
	Short: "Manage active collection view through Craft MCP",
}

var collectionsActiveViewSetCmd = &cobra.Command{
	Use:   "set [collection-id]",
	Short: "Set active collection view",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if collectionViewID == "" {
			return fmt.Errorf("--view is required")
		}
		command := "collections active-view set " + quoteMCPArg(args[0]) + " --view " + quoteMCPArg(collectionViewID)
		return runMCPCollectionCommand("collections.active-view.set", command, "craft_write")
	},
}

func init() {
	rootCmd.AddCommand(collectionsCmd)

	collectionsCmd.AddCommand(collectionsListCmd)
	collectionsListCmd.Flags().StringVar(&collectionDocumentID, "document", "", "Filter by document ID")

	collectionsCmd.AddCommand(collectionsSchemaCmd)
	collectionsSchemaCmd.Flags().StringVar(&collectionSchemaFormat, "schema-format", "schema", "Schema format (default: schema)")
	collectionsSchemaCmd.AddCommand(collectionsSchemaUpdateCmd)
	collectionsSchemaUpdateCmd.Flags().StringVar(&collectionJSON, "json", "", "Raw REST update collection schema payload JSON")
	collectionsSchemaUpdateCmd.Flags().BoolVar(&collectionStdin, "stdin", false, "Read raw REST update collection schema payload from stdin")

	collectionsCmd.AddCommand(collectionsCreateCmd)
	collectionsCreateCmd.Flags().StringVar(&collectionName, "name", "", "Collection name")
	collectionsCreateCmd.Flags().StringArrayVar(&collectionProperties, "property", nil, "MCP dynamic property flag as Name=type or --Name=type (repeatable)")
	collectionsCreateCmd.Flags().StringVar(&collectionJSON, "json", "", "Raw REST create collection payload JSON")
	collectionsCreateCmd.Flags().BoolVar(&collectionStdin, "stdin", false, "Read raw REST create collection payload from stdin")

	collectionsCmd.AddCommand(collectionsRenameCmd)
	collectionsRenameCmd.Flags().StringVar(&collectionName, "name", "", "New collection name")

	collectionsCmd.AddCommand(collectionsItemsCmd)
	collectionsItemsCmd.Flags().IntVar(&collectionItemDepth, "depth", -1, "Max depth of nested content (-1 for no limit)")

	collectionsCmd.AddCommand(collectionsAddCmd)
	collectionsAddCmd.Flags().StringVar(&collectionItemTitle, "title", "", "Item title (required)")
	collectionsAddCmd.Flags().StringVar(&collectionItemProps, "properties", "", "Item properties as JSON string")
	collectionsAddCmd.Flags().StringArrayVar(&collectionProperties, "property", nil, "MCP dynamic item property as Name=value or --Name=value (repeatable)")
	collectionsAddCmd.Flags().BoolVar(&collectionAllowNew, "allow-new-options", false, "Allow creating new options for select properties")
	collectionsAddCmd.Flags().StringVar(&collectionJSON, "json", "", "Raw REST add collection items payload JSON")
	collectionsAddCmd.Flags().BoolVar(&collectionStdin, "stdin", false, "Read raw REST add collection items payload from stdin")

	collectionsCmd.AddCommand(collectionsUpdateCmd)
	collectionsUpdateCmd.Flags().StringVar(&collectionItemID, "item", "", "Item ID to update (required)")
	collectionsUpdateCmd.Flags().StringVar(&collectionItemProps, "properties", "", "Item properties as JSON string (required)")
	collectionsUpdateCmd.Flags().StringArrayVar(&collectionProperties, "property", nil, "MCP dynamic item property as Name=value or --Name=value (repeatable)")
	collectionsUpdateCmd.Flags().BoolVar(&collectionAllowNew, "allow-new-options", false, "Allow creating new options for select properties")
	collectionsUpdateCmd.Flags().StringVar(&collectionJSON, "json", "", "Raw REST update collection items payload JSON")
	collectionsUpdateCmd.Flags().BoolVar(&collectionStdin, "stdin", false, "Read raw REST update collection items payload from stdin")

	collectionsCmd.AddCommand(collectionsDeleteCmd)
	collectionsDeleteCmd.Flags().StringVar(&collectionItemID, "item", "", "Item ID to delete (required)")
	collectionsDeleteCmd.Flags().StringVar(&collectionJSON, "json", "", "Raw REST delete collection items payload JSON")
	collectionsDeleteCmd.Flags().BoolVar(&collectionStdin, "stdin", false, "Read raw REST delete collection items payload from stdin")

	collectionsCmd.AddCommand(collectionsViewsCmd)
	collectionsViewsCmd.AddCommand(collectionsViewsListCmd)
	collectionsViewsCmd.AddCommand(collectionsViewsCreateCmd)
	collectionsViewsCreateCmd.Flags().StringVar(&collectionViewName, "name", "", "View name")
	collectionsViewsCreateCmd.Flags().StringVar(&collectionViewType, "type", "", "View type")
	collectionsViewsCreateCmd.Flags().StringVar(&collectionJSON, "json", "", "Raw MCP collection view payload JSON")
	collectionsViewsCreateCmd.Flags().BoolVar(&collectionStdin, "stdin", false, "Read raw MCP collection view payload from stdin")
	collectionsViewsCmd.AddCommand(collectionsViewsUpdateCmd)
	collectionsViewsUpdateCmd.Flags().StringVar(&collectionViewID, "view", "", "View ID")
	collectionsViewsUpdateCmd.Flags().StringVar(&collectionViewName, "name", "", "View name")
	collectionsViewsUpdateCmd.Flags().StringVar(&collectionViewType, "type", "", "View type")
	collectionsViewsUpdateCmd.Flags().StringVar(&collectionJSON, "json", "", "Raw MCP collection view payload JSON")
	collectionsViewsUpdateCmd.Flags().BoolVar(&collectionStdin, "stdin", false, "Read raw MCP collection view payload from stdin")
	collectionsViewsCmd.AddCommand(collectionsViewsDeleteCmd)
	collectionsViewsDeleteCmd.Flags().StringVar(&collectionViewID, "view", "", "View ID")

	collectionsCmd.AddCommand(collectionsActiveViewCmd)
	collectionsActiveViewCmd.AddCommand(collectionsActiveViewSetCmd)
	collectionsActiveViewSetCmd.Flags().StringVar(&collectionViewID, "view", "", "View ID")
}

func propertyFlagToMCP(property string) string {
	property = strings.TrimSpace(property)
	property = strings.TrimPrefix(property, "--")
	if property == "" {
		return ""
	}
	name, value, ok := strings.Cut(property, "=")
	if !ok {
		return "--" + property
	}
	return "--" + quoteMCPDynamicFlagName(name) + " " + quoteMCPArg(value)
}

func quoteMCPDynamicFlagName(name string) string {
	name = strings.TrimSpace(strings.TrimPrefix(name, "--"))
	if name == "" {
		return name
	}
	if strings.ContainsAny(name, " \t\n\r\"'") {
		return quoteMCPArg(name)
	}
	return name
}

func runMCPCollectionCommand(operation, command, tool string) error {
	if isDryRun() {
		return dryRunOutput(operation, map[string]interface{}{
			"backend":      "mcp",
			"tool":         tool,
			"command":      command,
			"capabilities": []string{"mcp", "collections.views"},
		})
	}
	if tool == "craft_write" && !yesFlag {
		return fmt.Errorf("%s uses craft_write; rerun with --yes after reviewing --dry-run", operation)
	}
	client, err := getMCPClient()
	if err != nil {
		return err
	}
	result, err := client.CallTool(tool, map[string]interface{}{"command": command})
	if err != nil {
		return err
	}
	return outputRawJSON(result)
}

// outputCollections prints collections in the specified format
func outputCollections(collections []models.Collection, format string) error {
	switch format {
	case FormatCompact:
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
	case FormatCompact:
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
