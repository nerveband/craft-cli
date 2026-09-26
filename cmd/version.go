package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display the current version of Craft CLI",
	RunE: func(cmd *cobra.Command, args []string) error {
		if isJSONFormat(getOutputFormat()) {
			return outputJSON(map[string]string{"version": version})
		}
		fmt.Printf("craft-cli version %s\n", version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
