package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"os"
	"strings"
)

// CommandEffect declarations are keyed by complete canonical paths. Do not infer
// permission or retry safety from a leaf name. Unknown commands fail closed.
// Destructive means the command can lose data (deletes, clears, resets, schema
// changes that drop properties). Only destructive commands require --yes; edits,
// moves and renames run freely and still support --dry-run. `update` is
// destructive only in --mode replace, which cmd/update.go gates itself.
type CommandEffect struct {
	Effects     string `json:"effects"`
	Scope       string `json:"scope"`
	Destructive bool   `json:"destructive"`
}

var commandEffects = map[string]CommandEffect{
	"":                            {"read_only", "local", false},
	"schema":                      {"read_only", "local", false},
	"audit":                       {"read_only", "local", false},
	"audit agent-dx":              {"read_only", "local", false},
	"batch":                       {"non_idempotent", "remote", false},
	"blocks":                      {"read_only", "remote", false},
	"blocks add":                  {"non_idempotent", "remote", false},
	"blocks delete":               {"non_idempotent", "remote", true},
	"blocks explore-themes":       {"read_only", "remote", false},
	"blocks explore-washi":        {"read_only", "remote", false},
	"blocks get":                  {"read_only", "remote", false},
	"blocks move":                 {"non_idempotent", "remote", false},
	"blocks revert":               {"non_idempotent", "remote", false},
	"blocks search-unsplash":      {"read_only", "remote", false},
	"blocks update":               {"non_idempotent", "remote", false},
	"clear":                       {"non_idempotent", "remote", true},
	"collections":                 {"read_only", "remote", false},
	"collections active-view":     {"read_only", "remote", false},
	"collections active-view set": {"non_idempotent", "remote", false},
	"collections add":             {"non_idempotent", "remote", false},
	"collections create":          {"non_idempotent", "remote", false},
	"collections delete":          {"non_idempotent", "remote", true},
	"collections items":           {"read_only", "remote", false},
	"collections list":            {"read_only", "remote", false},
	"collections rename":          {"non_idempotent", "remote", false},
	"collections schema":          {"read_only", "remote", false},
	"collections schema update":   {"non_idempotent", "remote", true},
	"collections update":          {"non_idempotent", "remote", false},
	"collections views":           {"read_only", "remote", false},
	"collections views create":    {"non_idempotent", "remote", false},
	"collections views delete":    {"non_idempotent", "remote", true},
	"collections views list":      {"read_only", "remote", false},
	"collections views update":    {"non_idempotent", "remote", false},
	"comments":                    {"read_only", "remote", false},
	"comments add":                {"non_idempotent", "remote", false},
	"completion":                  {"read_only", "local", false},
	"config":                      {"read_only", "local", false},
	"config add":                  {"non_idempotent", "local", false},
	"config add-mcp":              {"non_idempotent", "local", false},
	"config add-rest":             {"non_idempotent", "local", false},
	"config list":                 {"read_only", "local", false},
	"config remove":               {"non_idempotent", "local", true},
	"config reset":                {"non_idempotent", "local", true},
	"config use":                  {"non_idempotent", "local", false},
	"connection":                  {"read_only", "remote", false},
	"create":                      {"non_idempotent", "remote", false},
	"delete":                      {"non_idempotent", "remote", true},
	"docs":                        {"read_only", "remote", false},
	"documents":                   {"read_only", "remote", false},
	"documents resolve-link":      {"read_only", "remote", false},
	"feedback":                    {"non_idempotent", "local", false},
	"folders":                     {"read_only", "remote", false},
	"folders create":              {"non_idempotent", "remote", false},
	"folders delete":              {"non_idempotent", "remote", true},
	"folders explore-icons":       {"read_only", "remote", false},
	"folders list":                {"read_only", "remote", false},
	"folders move":                {"non_idempotent", "remote", false},
	"get":                         {"read_only", "remote", false},
	"images":                      {"read_only", "remote", false},
	"images view":                 {"read_only", "remote", false},
	"info":                        {"read_only", "remote", false},
	"limits":                      {"read_only", "local", false},
	"list":                        {"read_only", "remote", false},
	"llm":                         {"read_only", "local", false},
	"llm styles":                  {"read_only", "local", false},
	"local":                       {"read_only", "remote", false},
	"local append":                {"non_idempotent", "local", false},
	"local new":                   {"non_idempotent", "local", false},
	"local open":                  {"non_idempotent", "local", false},
	"local search":                {"non_idempotent", "local", false},
	"local space":                 {"non_idempotent", "local", false},
	"local today":                 {"non_idempotent", "local", false},
	"local tomorrow":              {"non_idempotent", "local", false},
	"local yesterday":             {"non_idempotent", "local", false},
	"mcp":                         {"read_only", "remote", false},
	"mcp batch":                   {"non_idempotent", "remote", false},
	"mcp call":                    {"non_idempotent", "remote", false},
	"mcp edit-review":             {"read_only", "remote", false},
	"mcp initialize":              {"read_only", "remote", false},
	"mcp read-resource":           {"read_only", "remote", false},
	"mcp resources":               {"read_only", "remote", false},
	"mcp tools":                   {"read_only", "remote", false},
	"move":                        {"non_idempotent", "remote", false},
	"profiles":                    {"read_only", "local", false},
	"profiles add-mcp":            {"non_idempotent", "local", false},
	"profiles add-rest":           {"non_idempotent", "local", false},
	"profiles capabilities":       {"read_only", "local", false},
	"profiles list":               {"read_only", "local", false},
	"profiles remove":             {"non_idempotent", "local", true},
	"profiles show":               {"read_only", "local", false},
	"profiles test":               {"read_only", "remote", false},
	"profiles use":                {"non_idempotent", "local", false},
	"search":                      {"read_only", "remote", false},
	"setup":                       {"non_idempotent", "local", false},
	"skill-path":                  {"read_only", "local", false},
	"tasks":                       {"read_only", "remote", false},
	"tasks add":                   {"non_idempotent", "remote", false},
	"tasks delete":                {"non_idempotent", "remote", true},
	"tasks list":                  {"read_only", "remote", false},
	"tasks update":                {"non_idempotent", "remote", false},
	"update":                      {"non_idempotent", "remote", false},
	"upgrade":                     {"non_idempotent", "local", false},
	"upload":                      {"non_idempotent", "remote", false},
	"validate":                    {"read_only", "local", false},
	"version":                     {"read_only", "local", false},
	"whiteboards":                 {"read_only", "remote", false},
	"whiteboards add":             {"non_idempotent", "remote", false},
	"whiteboards create":          {"non_idempotent", "remote", false},
	"whiteboards delete":          {"non_idempotent", "remote", true},
	"whiteboards elements":        {"read_only", "remote", false},
	"whiteboards elements get":    {"read_only", "remote", false},
	"whiteboards get":             {"read_only", "remote", false},
	"whiteboards update":          {"non_idempotent", "remote", false},
}

func declaredEffect(cmd *cobra.Command) (CommandEffect, bool) {
	e, ok := commandEffects[strings.TrimPrefix(cmd.CommandPath(), "craft ")]
	if cmd.CommandPath() == "craft" {
		e, ok = commandEffects[""]
	}
	return e, ok
}

func stdinIsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

var previewComplete = fmt.Errorf("local preview complete")

func commandPreflight(cmd *cobra.Command, args []string) error {
	if requestTimeout <= 0 {
		return fmt.Errorf("--timeout must be positive")
	}
	switch getOutputFormat() {
	case "json", "compact", "jsonl", "yaml", "raw", "table", "markdown", "rich", "craft", "structured":
	default:
		return fmt.Errorf("invalid --format %q; valid: json, compact, jsonl, yaml, raw, table, markdown, rich, craft, structured", getOutputFormat())
	}
	switch backendName {
	case "auto", "rest", "mcp", "local":
	default:
		return fmt.Errorf("invalid --backend %q; valid: auto, rest, mcp, local", backendName)
	}
	switch dataSource {
	case "", "live":
	case "local", "auto":
		return fmt.Errorf("--data-source %q is unavailable: this CLI has no local data index; use live", dataSource)
	default:
		return fmt.Errorf("invalid --data-source %q; supported: live", dataSource)
	}
	if deliverTarget != "" && deliverTarget != "stdout" && !strings.HasPrefix(deliverTarget, "file:") {
		return fmt.Errorf("invalid --deliver %q; supported: stdout, file:<path>", deliverTarget)
	}
	if strings.HasPrefix(deliverTarget, "file:") {
		path := strings.TrimPrefix(deliverTarget, "file:")
		if path == "" {
			return fmt.Errorf("--deliver file: requires a path")
		}
		if err := validateOutputPath(path, allowOutsideCWD); err != nil {
			return err
		}
		if _, err := os.Stat(path); err == nil && !yesFlag {
			return fmt.Errorf("output exists; use --yes to replace %s", path)
		}
	}
	e, ok := declaredEffect(cmd)
	if !ok {
		return fmt.Errorf("command %s has no declared effects", cmd.CommandPath())
	}
	if e.Scope == "local" && e.Effects != "read_only" && isDryRun() {
		if err := dryRunOutput(cmd.CommandPath(), map[string]interface{}{"scope": "local", "arguments": redactedLocalArgs(cmd, args), "reversible": false}); err != nil {
			return err
		}
		return previewComplete
	}
	if e.Destructive && !isDryRun() && !yesFlag {
		return newCLIError("CONFIRMATION_REQUIRED", "operation changes or deletes existing state; review --dry-run and rerun with --yes")
	}
	return nil
}

func redactedLocalArgs(cmd *cobra.Command, args []string) []string {
	out := append([]string(nil), args...)
	if cmd.CommandPath() == "craft config add" && len(out) > 1 {
		out[1] = redactedURL(out[1])
	}
	return out
}

func selectedMCPProfile() bool {
	if backendName != "auto" || cfgManager == nil {
		return false
	}
	if profileName != "" {
		profile, err := cfgManager.GetProfile(profileName)
		return err == nil && profile.TypeOrDefault() == "mcp"
	}
	profile, err := cfgManager.GetActiveProfile()
	return err == nil && profile.TypeOrDefault() == "mcp"
}
