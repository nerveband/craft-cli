package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Run local proof-of-behavior checks",
	Long:  "Run local validation checks for schema generation, skill packaging, and basic command contracts.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		schema := buildSchema(rootCmd)
		checks := []map[string]interface{}{
			{"name": "schema", "ok": len(schema.Subcommands) > 0 && schema.SchemaVersion != ""},
			{"name": "skill", "ok": fileExists("SKILL.md")},
			{"name": "agent_dx_audit", "ok": findCobraCommand(rootCmd, "audit", "agent-dx") != nil},
		}
		ok := true
		for _, check := range checks {
			if passed, _ := check["ok"].(bool); !passed {
				ok = false
			}
		}
		return outputJSON(map[string]interface{}{"ok": ok, "checks": checks})
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

// validateResourceID checks an ID for agent hallucination patterns.
func validateResourceID(id, label string) error {
	if id == "" {
		return fmt.Errorf("%s cannot be empty", label)
	}
	if strings.Contains(id, "..") {
		return fmt.Errorf("invalid %s: path traversal detected", label)
	}
	lower := strings.ToLower(id)
	if strings.Contains(lower, "%2e") || strings.Contains(lower, "%2f") {
		return fmt.Errorf("invalid %s: encoded path characters detected", label)
	}
	if strings.ContainsAny(id, "?#&=") {
		return fmt.Errorf("invalid %s: query parameters not allowed in IDs", label)
	}
	for _, r := range id {
		if unicode.IsControl(r) {
			return fmt.Errorf("invalid %s: control characters not allowed", label)
		}
	}
	return nil
}

func validateOutputPath(path string, allowOutsideCWD bool) error {
	if path == "" || allowOutsideCWD {
		return nil
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to determine current directory: %w", err)
	}
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return fmt.Errorf("invalid current directory: %w", err)
	}
	rel, err := filepath.Rel(absCwd, absPath)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("invalid output path: outside current directory (use --allow-outside-cwd to override)")
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
