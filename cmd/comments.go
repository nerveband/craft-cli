package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var commentsCmd = &cobra.Command{
	Use:   "comments",
	Short: "Manage comments (experimental)",
	Long: `Manage comments on Craft blocks (experimental).

Examples:
  craft comments add BLOCK_ID --content "This needs review"
  craft comments add BLOCK_ID --content "LGTM" --format json`,
}

var commentContent string

var commentsAddCmd = &cobra.Command{
	Use:   "add [block-id]",
	Short: "Add a comment to a block",
	Long: `Add a comment to a specific block.

Examples:
  craft comments add BLOCK_ID --content "This needs review"
  craft comments add BLOCK_ID --content "LGTM" --format json
  craft comments add BLOCK_ID --content "Note" --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		blockID := args[0]

		if isDryRun() {
			fmt.Printf("[dry-run] Would add comment to block %s: %s\n", blockID, commentContent)
			return nil
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		result, err := client.AddComment(blockID, commentContent)
		if err != nil {
			return err
		}

		if isQuiet() {
			if len(result.Items) > 0 {
				fmt.Println(result.Items[0].CommentID)
			}
			return nil
		}

		format := getOutputFormat()
		if format == "json" {
			return outputJSON(result)
		}

		if len(result.Items) > 0 {
			fmt.Printf("Comment added: %s\n", result.Items[0].CommentID)
		} else {
			fmt.Println("Comment added")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(commentsCmd)

	commentsCmd.AddCommand(commentsAddCmd)
	commentsAddCmd.Flags().StringVar(&commentContent, "content", "", "Comment content (required)")
	commentsAddCmd.MarkFlagRequired("content")
}
