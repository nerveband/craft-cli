package cmd

import "github.com/spf13/cobra"

var configMigrateCmd = &cobra.Command{Use: "migrate", Short: "Move inline connection URLs and API keys into the private credential store", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
	cfg, err := cfgManager.Load()
	if err != nil {
		return err
	}
	if err = cfgManager.Save(cfg); err != nil {
		return err
	}
	return outputJSON(map[string]interface{}{"status": "migrated", "profiles": len(cfg.Profiles)})
}}

func init() {
	configCmd.AddCommand(configMigrateCmd)
	commandEffects["config migrate"] = CommandEffect{"non_idempotent", "local", false}
	workflowExamples["config migrate"] = "craft config migrate --dry-run"
}
