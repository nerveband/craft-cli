package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display the current version of Craft CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("craft-cli version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
