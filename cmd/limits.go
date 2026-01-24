package cmd

import "github.com/spf13/cobra"

type limitsInfo struct {
	DefaultChunkBytes int      `json:"defaultChunkBytes"`
	Recommended       []string `json:"recommended"`
	Notes             []string `json:"notes"`
}

var limitsCmd = &cobra.Command{
	Use:   "limits",
	Short: "Show known API/CLI limits",
	Long:  "Show known API/CLI limits and how the CLI mitigates them.",
	RunE: func(cmd *cobra.Command, args []string) error {
		info := limitsInfo{
			DefaultChunkBytes: 30000,
			Recommended: []string{
				"Use craft update --chunk-bytes to tune chunk size if you see PAYLOAD_TOO_LARGE.",
				"Use craft update --mode replace for large edits instead of appending.",
			},
			Notes: []string{
				"Craft insert blocks has a payload limit; large markdown is auto-chunked by the CLI.",
				"craft delete is a soft-delete to trash (DELETE /documents).",
				"craft clear deletes content blocks and cannot be undone without a backup.",
			},
		}
		return outputJSON(info)
	},
}

func init() {
	rootCmd.AddCommand(limitsCmd)
}
