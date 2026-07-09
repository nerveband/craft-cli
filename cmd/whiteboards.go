package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var whiteboardsCmd = &cobra.Command{
	Use:   "whiteboards",
	Short: "Manage whiteboards (experimental)",
	Long: `Manage Craft whiteboards - create, get elements, add, update, and delete elements.

Whiteboards use Excalidraw-format elements.

Examples:
  craft whiteboards create PAGE_ID                    # Create whiteboard in page
  craft whiteboards get WHITEBOARD_ID                 # Get elements
  craft whiteboards add WHITEBOARD_ID --json '[...]'  # Add elements
  craft whiteboards update WHITEBOARD_ID --json '[...]' # Update elements
  craft whiteboards delete WHITEBOARD_ID --ids "id1,id2" # Delete elements`,
}

var whiteboardElementsCmd = &cobra.Command{
	Use:   "elements",
	Short: "Read whiteboard elements through Craft MCP",
}

var whiteboardElementsGetCmd = &cobra.Command{
	Use:   "get [whiteboard-id]",
	Short: "Get whiteboard elements through Craft MCP",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateResourceID(args[0], "whiteboard-id"); err != nil {
			return err
		}
		command := "whiteboards elements get --whiteboard " + quoteMCPArg(args[0])
		if isDryRun() {
			return dryRunOutput("whiteboards.elements.get", map[string]interface{}{
				"backend":      "mcp",
				"tool":         "craft_read",
				"command":      command,
				"capabilities": []string{"mcp", "whiteboards.read"},
			})
		}
		return runMCPReadCommand(command)
	},
}

var (
	whiteboardCreateJSON  string
	whiteboardCreateStdin bool
	whiteboardDeleteJSON  string
	whiteboardDeleteStdin bool
)

var whiteboardCreateCmd = &cobra.Command{
	Use:   "create PAGE_ID",
	Short: "Create a new whiteboard in a document",
	Long: `Create an empty whiteboard block inside a document page.

Examples:
  craft whiteboards create abc123
  craft whiteboards create abc123 --dry-run
  craft whiteboards create --json '{"pageId":"abc123"}' --dry-run`,
	Args: func(cmd *cobra.Command, args []string) error {
		if whiteboardCreateJSON != "" || whiteboardCreateStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		pageID := ""
		if whiteboardCreateJSON != "" || whiteboardCreateStdin {
			payload, err := readRawPayload(whiteboardCreateJSON, whiteboardCreateStdin)
			if err != nil {
				return err
			}
			pageID = payloadString(payload, "pageId", "pageID", "id")
			if pageID == "" {
				return fmt.Errorf("raw whiteboard create payload must include \"pageId\"")
			}
		} else {
			pageID = args[0]
		}
		if err := validateResourceID(pageID, "page-id"); err != nil {
			return err
		}
		if isDryRun() {
			return dryRunOutput("create whiteboard", map[string]interface{}{
				"page_id": pageID,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		result, err := client.CreateWhiteboard(pageID)
		if err != nil {
			return err
		}

		return outputJSON(result)
	},
}

var whiteboardGetCmd = &cobra.Command{
	Use:   "get WHITEBOARD_ID",
	Short: "Get whiteboard elements",
	Long: `Retrieve all elements from a whiteboard in Excalidraw format.

Examples:
  craft whiteboards get abc123
  craft whiteboards get abc123 --quiet`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateResourceID(args[0], "whiteboard-id"); err != nil {
			return err
		}
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		result, err := client.GetWhiteboardElements(args[0])
		if err != nil {
			return err
		}

		return outputJSON(result)
	},
}

var (
	whiteboardJSON string
	whiteboardIDs  string
)

var whiteboardAddCmd = &cobra.Command{
	Use:   "add WHITEBOARD_ID",
	Short: "Add elements to a whiteboard",
	Long: `Append new elements to a whiteboard. Elements use Excalidraw format.

Examples:
  craft whiteboards add WB_ID --json '[{"type":"rectangle","x":0,"y":0,"width":100,"height":100}]'
  echo '[{"type":"text","x":50,"y":50,"text":"Hello"}]' | craft whiteboards add WB_ID --stdin`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data := whiteboardJSON
		if data == "" {
			stdinFlag, _ := cmd.Flags().GetBool("stdin")
			if stdinFlag {
				d, err := readStdinString()
				if err != nil {
					return err
				}
				data = d
			}
		}
		if data == "" {
			return fmt.Errorf("provide --json or --stdin with element data")
		}

		if isDryRun() {
			return dryRunOutput("add whiteboard elements", map[string]interface{}{
				"whiteboard_id": args[0],
			})
		}

		var elements []map[string]interface{}
		if err := json.Unmarshal([]byte(data), &elements); err != nil {
			// Try single element
			var single map[string]interface{}
			if err2 := json.Unmarshal([]byte(data), &single); err2 != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}
			elements = []map[string]interface{}{single}
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		result, err := client.AddWhiteboardElements(args[0], elements)
		if err != nil {
			return err
		}

		return outputJSON(result)
	},
}

var whiteboardUpdateCmd = &cobra.Command{
	Use:   "update WHITEBOARD_ID",
	Short: "Update whiteboard elements",
	Long: `Update specific elements in a whiteboard. Each element must include an "id" field.

Examples:
  craft whiteboards update WB_ID --json '[{"id":"elem1","x":50,"y":50}]'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data := whiteboardJSON
		if data == "" {
			stdinFlag, _ := cmd.Flags().GetBool("stdin")
			if stdinFlag {
				d, err := readStdinString()
				if err != nil {
					return err
				}
				data = d
			}
		}
		if data == "" {
			return fmt.Errorf("provide --json or --stdin with element data")
		}

		if isDryRun() {
			return dryRunOutput("update whiteboard elements", map[string]interface{}{
				"whiteboard_id": args[0],
			})
		}

		var elements []map[string]interface{}
		if err := json.Unmarshal([]byte(data), &elements); err != nil {
			var single map[string]interface{}
			if err2 := json.Unmarshal([]byte(data), &single); err2 != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}
			elements = []map[string]interface{}{single}
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.UpdateWhiteboardElements(args[0], elements); err != nil {
			return err
		}

		if !quietMode {
			fmt.Println("Whiteboard elements updated successfully")
		}
		return nil
	},
}

var whiteboardDeleteCmd = &cobra.Command{
	Use:   "delete WHITEBOARD_ID",
	Short: "Delete whiteboard elements",
	Long: `Remove specific elements from a whiteboard by their IDs.

Examples:
  craft whiteboards delete WB_ID --ids "elem1,elem2"
  craft whiteboards delete WB_ID --ids "elem1" --dry-run
  craft whiteboards delete WB_ID --json '{"idsToDelete":["elem1"]}' --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var ids []string
		if whiteboardDeleteJSON != "" || whiteboardDeleteStdin {
			payload, err := readRawPayload(whiteboardDeleteJSON, whiteboardDeleteStdin)
			if err != nil {
				return err
			}
			rawIDs, ok := payloadArray(payload, "idsToDelete")
			if !ok {
				rawIDs, ok = payloadArray(payload, "ids")
			}
			if !ok || len(rawIDs) == 0 {
				return fmt.Errorf("raw whiteboard delete payload must include non-empty \"idsToDelete\" array")
			}
			for _, id := range rawIDs {
				value, ok := id.(string)
				if !ok || value == "" {
					return fmt.Errorf("idsToDelete must contain strings")
				}
				ids = append(ids, value)
			}
		} else {
			if whiteboardIDs == "" {
				return fmt.Errorf("--ids is required: comma-separated element IDs to delete")
			}

			ids = strings.Split(whiteboardIDs, ",")
			for i := range ids {
				ids[i] = strings.TrimSpace(ids[i])
			}
		}

		if isDryRun() {
			return dryRunOutput("delete whiteboard elements", map[string]interface{}{
				"whiteboard_id": args[0], "element_ids": ids, "count": len(ids),
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.DeleteWhiteboardElements(args[0], ids); err != nil {
			return err
		}

		if !quietMode {
			fmt.Printf("Deleted %d elements from whiteboard\n", len(ids))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(whiteboardsCmd)

	whiteboardsCmd.AddCommand(whiteboardCreateCmd)
	whiteboardsCmd.AddCommand(whiteboardGetCmd)
	whiteboardsCmd.AddCommand(whiteboardElementsCmd)
	whiteboardElementsCmd.AddCommand(whiteboardElementsGetCmd)
	whiteboardsCmd.AddCommand(whiteboardAddCmd)
	whiteboardsCmd.AddCommand(whiteboardUpdateCmd)
	whiteboardsCmd.AddCommand(whiteboardDeleteCmd)

	whiteboardCreateCmd.Flags().StringVar(&whiteboardCreateJSON, "json", "", "Raw whiteboard create payload JSON")
	whiteboardCreateCmd.Flags().BoolVar(&whiteboardCreateStdin, "stdin", false, "Read raw whiteboard create payload from stdin")

	whiteboardAddCmd.Flags().StringVar(&whiteboardJSON, "json", "", "JSON element data (Excalidraw format)")
	whiteboardAddCmd.Flags().Bool("stdin", false, "Read element data from stdin")

	whiteboardUpdateCmd.Flags().StringVar(&whiteboardJSON, "json", "", "JSON element data with IDs")
	whiteboardUpdateCmd.Flags().Bool("stdin", false, "Read element data from stdin")

	whiteboardDeleteCmd.Flags().StringVar(&whiteboardIDs, "ids", "", "Comma-separated element IDs to delete")
	whiteboardDeleteCmd.Flags().StringVar(&whiteboardDeleteJSON, "json", "", "Raw whiteboard delete payload JSON")
	whiteboardDeleteCmd.Flags().BoolVar(&whiteboardDeleteStdin, "stdin", false, "Read raw whiteboard delete payload from stdin")
}

func readStdinString() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read stdin: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}
