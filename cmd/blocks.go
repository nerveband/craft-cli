package cmd

import (
	"fmt"
	"strings"

	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

var blocksCmd = &cobra.Command{
	Use:   "blocks",
	Short: "Manage document blocks",
	Long: `Manage blocks within Craft documents - get, add, update, delete, and move blocks.

Examples:
  craft blocks get BLOCK_ID                           # Get a specific block
  craft blocks add PAGE_ID --markdown "Hello"         # Add block at end
  craft blocks add PAGE_ID --markdown "Hi" --pos start # Add at start
  craft blocks add --sibling ID --pos before -m "..."  # Add before sibling
  craft blocks update BLOCK_ID --markdown "New text"  # Update block content
  craft blocks delete BLOCK_ID                        # Delete a block
  craft blocks move BLOCK_ID --to PAGE_ID --pos end   # Move block`,
}

var blocksGetCmd = &cobra.Command{
	Use:   "get [block-id]",
	Short: "Get a specific block",
	Long:  "Retrieve a specific block by ID with its content and children",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		blockID := args[0]
		block, err := client.GetBlock(blockID)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		switch format {
		case FormatStructured, "json":
			return outputJSON(block)
		case FormatCraft:
			var sb strings.Builder
			renderBlockCraft(&sb, block, 0)
			fmt.Print(sb.String())
			return nil
		case FormatRich:
			var sb strings.Builder
			renderBlockRich(&sb, block, 0)
			fmt.Print(sb.String())
			return nil
		default:
			fmt.Println(block.Markdown)
			return nil
		}
	},
}

var (
	blockMarkdown   string
	blockPosition   string
	blockSiblingID  string
	blockTargetPage string
)

var blocksAddCmd = &cobra.Command{
	Use:   "add [page-id]",
	Short: "Add a block to a document",
	Long: `Add a new block to a document at a specified position.

Positions:
  start  - Add at the beginning of the page
  end    - Add at the end of the page (default)
  before - Add before a sibling block (requires --sibling)
  after  - Add after a sibling block (requires --sibling)

Examples:
  craft blocks add PAGE_ID --markdown "Hello world"
  craft blocks add PAGE_ID --markdown "# Header" --position start
  craft blocks add --sibling BLOCK_ID --position before --markdown "Insert here"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if blockMarkdown == "" {
			return fmt.Errorf("--markdown is required")
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		var block *models.Block

		// Check if using sibling positioning
		if blockSiblingID != "" {
			if blockPosition != "before" && blockPosition != "after" {
				return fmt.Errorf("--position must be 'before' or 'after' when using --sibling")
			}
			block, err = client.AddBlockRelative(blockSiblingID, blockMarkdown, blockPosition)
		} else {
			if len(args) == 0 {
				return fmt.Errorf("page-id is required when not using --sibling")
			}
			pageID := args[0]
			if blockPosition == "" {
				blockPosition = "end"
			}
			block, err = client.AddBlock(pageID, blockMarkdown, blockPosition)
		}

		if err != nil {
			return err
		}

		if isQuiet() {
			fmt.Println(block.ID)
			return nil
		}

		format := getOutputFormat()
		if format == "json" {
			return outputJSON(block)
		}
		fmt.Printf("Block created: %s\n", block.ID)
		return nil
	},
}

var blocksUpdateCmd = &cobra.Command{
	Use:   "update [block-id]",
	Short: "Update a block's content",
	Long:  "Update the markdown content of an existing block",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if blockMarkdown == "" {
			return fmt.Errorf("--markdown is required")
		}

		if isDryRun() {
			fmt.Printf("[dry-run] Would update block %s with: %s\n", args[0], blockMarkdown)
			return nil
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		blockID := args[0]
		if err := client.UpdateBlockMarkdown(blockID, blockMarkdown); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Block %s updated\n", blockID)
		}
		return nil
	},
}

var blocksDeleteCmd = &cobra.Command{
	Use:   "delete [block-id]",
	Short: "Delete a block",
	Long:  "Delete a specific block from a document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isDryRun() {
			fmt.Printf("[dry-run] Would delete block: %s\n", args[0])
			return nil
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		blockID := args[0]
		if err := client.DeleteBlock(blockID); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Block %s deleted\n", blockID)
		}
		return nil
	},
}

var blocksMoveCmd = &cobra.Command{
	Use:   "move [block-id]",
	Short: "Move a block to a new location",
	Long: `Move a block to a different page or position.

Examples:
  craft blocks move BLOCK_ID --to PAGE_ID --position end
  craft blocks move BLOCK_ID --to PAGE_ID --position start`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if blockTargetPage == "" {
			return fmt.Errorf("--to is required")
		}
		if blockPosition == "" {
			blockPosition = "end"
		}

		if isDryRun() {
			fmt.Printf("[dry-run] Would move block %s to %s at position %s\n",
				args[0], blockTargetPage, blockPosition)
			return nil
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		blockID := args[0]
		if err := client.MoveBlock(blockID, blockTargetPage, blockPosition); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Block %s moved to %s\n", blockID, blockTargetPage)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(blocksCmd)

	blocksCmd.AddCommand(blocksGetCmd)

	blocksCmd.AddCommand(blocksAddCmd)
	blocksAddCmd.Flags().StringVarP(&blockMarkdown, "markdown", "m", "", "Markdown content for the block")
	blocksAddCmd.Flags().StringVarP(&blockPosition, "position", "p", "end", "Position: start, end, before, after")
	blocksAddCmd.Flags().StringVar(&blockSiblingID, "sibling", "", "Sibling block ID for relative positioning")
	blocksAddCmd.MarkFlagRequired("markdown")

	blocksCmd.AddCommand(blocksUpdateCmd)
	blocksUpdateCmd.Flags().StringVarP(&blockMarkdown, "markdown", "m", "", "New markdown content")
	blocksUpdateCmd.MarkFlagRequired("markdown")

	blocksCmd.AddCommand(blocksDeleteCmd)

	blocksCmd.AddCommand(blocksMoveCmd)
	blocksMoveCmd.Flags().StringVar(&blockTargetPage, "to", "", "Target page ID")
	blocksMoveCmd.Flags().StringVarP(&blockPosition, "position", "p", "end", "Position: start, end")
	blocksMoveCmd.MarkFlagRequired("to")
}
