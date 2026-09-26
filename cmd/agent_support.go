package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ashrafali/craft-cli/internal/agentdoc"
	"github.com/spf13/cobra"
)

var skillPathCmd = &cobra.Command{
	Use:   "skill-path",
	Short: "Emit bundled skill content and metadata",
	Long:  "Emit the embedded craft-cli skill from any working directory. JSON includes name, source, and content; text output can be redirected to SKILL.md.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isJSONFormat(getOutputFormat()) {
			return outputJSON(map[string]string{"name": "craft-cli", "source": "bundled", "content": agentdoc.Content})
		}
		fmt.Print(agentdoc.Content)
		return nil
	},
}

var feedbackCmd = &cobra.Command{
	Use:   "feedback [message]",
	Short: "Record local CLI feedback",
	Long:  "Record local structured feedback for agent/user friction without sending network requests.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dir := filepath.Join(home, ".craft-cli")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		path := filepath.Join(dir, "feedback.jsonl")
		record := map[string]string{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"message":   args[0],
		}
		data, err := json.Marshal(record)
		if err != nil {
			return err
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.Write(append(data, '\n')); err != nil {
			return err
		}
		return outputJSON(map[string]string{"path": path, "status": "recorded"})
	},
}

func init() {
	rootCmd.AddCommand(skillPathCmd)
	rootCmd.AddCommand(feedbackCmd)
}
