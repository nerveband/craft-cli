package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandSchema describes a CLI command for machine consumption
type CommandSchema struct {
	Effect               *CommandEffect  `json:"contract,omitempty"`
	SchemaVersion        string          `json:"schema_version,omitempty"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	Usage                string          `json:"usage"`
	Flags                []FlagSchema    `json:"flags,omitempty"`
	Subcommands          []CommandSchema `json:"subcommands,omitempty"`
	Examples             []string        `json:"examples,omitempty"`
	Safety               *SafetyInfo     `json:"safety,omitempty"`
	Backends             []string        `json:"backends,omitempty"`
	RequiredCapabilities []string        `json:"required_capabilities,omitempty"`
	OptionalCapabilities []string        `json:"optional_capabilities,omitempty"`
}

// FlagSchema describes a command flag
type FlagSchema struct {
	Name     string `json:"name"`
	Short    string `json:"short,omitempty"`
	Type     string `json:"type"`
	Default  string `json:"default,omitempty"`
	Required bool   `json:"required"`
	Desc     string `json:"description"`
}

// SafetyInfo describes the safety characteristics of a command
type SafetyInfo struct {
	ReadOnly    bool `json:"readonly"`
	Destructive bool `json:"destructive"`
	Idempotent  bool `json:"idempotent"`
	DryRun      bool `json:"supports_dry_run"`
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Output machine-readable CLI schema as JSON",
	Long: `Output a structured JSON manifest of all craft-cli commands,
flags, types, and safety metadata. Designed for AI agent introspection.

An agent can call 'craft schema' once to discover all available
commands without parsing --help text.

Examples:
  craft schema                    # Full schema as JSON
  craft schema --command list     # Schema for a specific command
  craft schema --commands-only    # Just command names and descriptions`,
	RunE: func(cmd *cobra.Command, args []string) error {
		commandFilter, _ := cmd.Flags().GetString("command")
		if commandFilter == "" && len(args) > 0 {
			commandFilter = strings.Join(args, " ")
		}
		if commandFilter != "" {
			found, rest, err := rootCmd.Find(strings.Fields(commandFilter))
			if err != nil || found == rootCmd || len(rest) > 0 {
				return fmt.Errorf("unknown command: %s", commandFilter)
			}
		}
		commandsOnly, _ := cmd.Flags().GetBool("commands-only")

		legacy, _ := cmd.Flags().GetBool("legacy")
		if !legacy && !commandsOnly {
			return outputSchemaJSON(cliSpec(commandFilter))
		}
		schema := buildSchema(rootCmd)

		if commandFilter != "" {
			found, rest, err := rootCmd.Find(strings.Fields(commandFilter))
			if err != nil || len(rest) > 0 || found == rootCmd {
				return fmt.Errorf("unknown command: %s", commandFilter)
			}
			return outputSchemaJSON(buildSchema(found))
		}

		if commandsOnly {
			type briefCmd struct {
				Name string `json:"name"`
				Desc string `json:"description"`
			}
			var cmds []briefCmd
			for _, sc := range schema.Subcommands {
				cmds = append(cmds, briefCmd{Name: sc.Name, Desc: sc.Description})
			}
			return outputSchemaJSON(cmds)
		}

		return outputSchemaJSON(schema)
	},
}

func init() {
	rootCmd.AddCommand(schemaCmd)
	schemaCmd.Flags().Bool("legacy", false, "Emit the pre-CLI-Spec nested manifest for compatibility")
	schemaCmd.Flags().String("command", "", "Show schema for a specific command only")
	schemaCmd.Flags().Bool("commands-only", false, "Output only command names and descriptions")
}

func buildSchema(cmd *cobra.Command) CommandSchema {
	schema := CommandSchema{
		Name:        cmd.Name(),
		Description: cmd.Short,
		Usage:       cmd.UseLine(),
	}
	if cmd == rootCmd {
		schema.SchemaVersion = "2026-07-08"
	}

	// Collect examples
	if cmd.Example != "" {
		schema.Examples = schemaParseExamples(cmd.Example)
	}

	// Collect local flags
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}
		fs := FlagSchema{
			Name:     "--" + f.Name,
			Type:     f.Value.Type(),
			Desc:     f.Usage,
			Required: len(f.Annotations[cobra.BashCompOneRequiredFlag]) > 0,
		}
		if f.Shorthand != "" {
			fs.Short = "-" + f.Shorthand
		}
		if f.DefValue != "" && f.DefValue != "false" {
			fs.Default = f.DefValue
		}
		schema.Flags = append(schema.Flags, fs)
	})

	// Add safety metadata based on command name
	effect, declared := declaredEffect(cmd)
	if declared {
		schema.Safety = &SafetyInfo{ReadOnly: effect.Effects == "read_only", Destructive: effect.Destructive, Idempotent: effect.Effects == "read_only" || effect.Effects == "idempotent", DryRun: effect.Effects != "read_only"}
		schema.Effect = &effect
	}
	schema.Backends, schema.RequiredCapabilities, schema.OptionalCapabilities = inferCommandCapabilities(cmd)

	// Collect subcommands
	for _, sub := range cmd.Commands() {
		isRootSchemaCommand := cmd.Name() == "craft" && sub.Name() == "schema"
		if sub.IsAvailableCommand() && sub.Name() != "help" && !isRootSchemaCommand {
			subSchema := buildSchema(sub)
			schema.Subcommands = append(schema.Subcommands, subSchema)
		}
	}

	return schema
}

func inferCommandCapabilities(cmd *cobra.Command) ([]string, []string, []string) {
	path := cmd.CommandPath()
	name := cmd.Name()

	if strings.HasPrefix(path, "craft mcp") {
		return []string{"mcp"}, []string{"mcp"}, nil
	}
	if strings.HasPrefix(path, "craft images") {
		return []string{"mcp"}, []string{"mcp", "images.view"}, nil
	}
	if path == "craft batch" {
		return []string{"mcp"}, []string{"mcp", "batch"}, nil
	}
	if strings.HasPrefix(path, "craft audit") {
		return []string{"local"}, []string{"audit"}, nil
	}
	if strings.HasPrefix(path, "craft profiles") {
		return []string{"local"}, []string{"profiles"}, nil
	}
	if strings.HasPrefix(path, "craft collections views") || strings.HasPrefix(path, "craft collections active-view") {
		return []string{"rest", "mcp"}, []string{"collections.views"}, []string{"mcp", "blocks.revert"}
	}
	if path == "craft collections rename" {
		return []string{"mcp"}, []string{"mcp", "write"}, nil
	}

	if strings.HasPrefix(path, "craft whiteboards elements") {
		return []string{"mcp"}, []string{"mcp", "whiteboards.read"}, nil
	}
	if strings.Contains(path, "resolve-link") ||
		strings.Contains(path, "explore-icons") ||
		strings.Contains(path, "explore-themes") ||
		strings.Contains(path, "explore-washi") ||
		strings.Contains(path, "search-unsplash") ||
		strings.Contains(path, "revert") {
		return []string{"mcp"}, []string{"mcp"}, nil
	}

	backends := []string{"rest"}
	var required []string
	var optional []string

	switch name {
	case "list", "get", "search", "connection", "info", "docs", "limits", "version", "schema", "llm", "completion":
		required = []string{"read"}
	case "create", "add", "upload":
		required = []string{"write"}
	case "update", "move":
		required = []string{"write"}
	case "delete", "clear", "remove":
		required = []string{"write", "delete"}
	}

	if strings.HasPrefix(path, "craft blocks") {
		backends = []string{"rest", "mcp"}
		optional = append(optional, "mcp", "blocks.style", "blocks.revert")
	}
	if path == "craft list" {
		backends = []string{"rest", "mcp"}
		optional = append(optional, "cursor")
	}
	if strings.HasPrefix(path, "craft collections") {
		optional = append(optional, "mcp", "collections.views")
	}
	if strings.HasPrefix(path, "craft whiteboards") {
		required = append(required, "whiteboards")
	}
	if strings.HasPrefix(path, "craft local") {
		return []string{"local"}, []string{"craft-app"}, nil
	}

	if len(required) == 0 && len(optional) == 0 {
		return backends, nil, nil
	}
	return backends, uniqueStrings(required), uniqueStrings(optional)
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func schemaParseExamples(examples string) []string {
	var result []string
	for _, line := range strings.Split(examples, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func outputSchemaJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
