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

var (
	commentContent string
	commentJSON    string
	commentStdin   bool
)

var commentsAddCmd = &cobra.Command{
	Use:   "add [block-id]",
	Short: "Add a comment to a block",
	Long: `Add a comment to a specific block.

Examples:
  craft comments add BLOCK_ID --content "This needs review"
  craft comments add BLOCK_ID --content "LGTM" --format json
  craft comments add BLOCK_ID --content "Note" --dry-run
  craft comments add --json '{"comments":[{"blockId":"BLOCK_ID","content":"Note"}]}' --dry-run`,
	Args: func(cmd *cobra.Command, args []string) error {
		if commentJSON != "" || commentStdin {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if commentJSON != "" || commentStdin {
			payload, err := readRawPayload(commentJSON, commentStdin)
			if err != nil {
				return err
			}
			comments, ok := payloadArray(payload, "comments")
			if !ok || len(comments) == 0 {
				return fmt.Errorf("raw comments add payload must include non-empty \"comments\" array")
			}
			if isDryRun() {
				return dryRunOutput("add comments", map[string]interface{}{"payload": payload, "count": len(comments)})
			}
			client, err := getAPIClient()
			if err != nil {
				return err
			}
			result, err := client.AddCommentsRaw(payload)
			if err != nil {
				return err
			}
			return outputJSON(result)
		}

		blockID := args[0]
		if commentContent == "" {
			return fmt.Errorf("--content is required")
		}

		if isDryRun() {
			return dryRunOutput("add comment", map[string]interface{}{
				"block_id": blockID, "content": commentContent,
			})
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
		if isJSONFormat(format) {
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
	commentsAddCmd.Flags().StringVar(&commentJSON, "json", "", "Raw REST add comments payload JSON")
	commentsAddCmd.Flags().BoolVar(&commentStdin, "stdin", false, "Read raw REST add comments payload from stdin")
}
