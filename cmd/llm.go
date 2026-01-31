package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type llmFlagSpec struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Type      string `json:"type"`
	Default   string `json:"default"`
	Usage     string `json:"usage"`
}

type llmCommandSpec struct {
	Use     string           `json:"use"`
	Short   string           `json:"short,omitempty"`
	Long    string           `json:"long,omitempty"`
	Aliases []string         `json:"aliases,omitempty"`
	Flags   []llmFlagSpec    `json:"flags,omitempty"`
	Sub     []llmCommandSpec `json:"subcommands,omitempty"`
}

type llmSpec struct {
	Tool      string           `json:"tool"`
	Version   string           `json:"version"`
	Generated string           `json:"generatedAt"`
	Global    []llmFlagSpec    `json:"globalFlags"`
	Commands  []llmCommandSpec `json:"commands"`
	Notes     []string         `json:"notes"`
}

func flagsToSpec(fs *pflag.FlagSet) []llmFlagSpec {
	var specs []llmFlagSpec
	fs.VisitAll(func(f *pflag.Flag) {
		specs = append(specs, llmFlagSpec{
			Name:      f.Name,
			Shorthand: f.Shorthand,
			Type:      f.Value.Type(),
			Default:   f.DefValue,
			Usage:     f.Usage,
		})
	})
	return specs
}

func commandToSpec(c *cobra.Command) llmCommandSpec {
	spec := llmCommandSpec{
		Use:     c.Use,
		Short:   c.Short,
		Long:    c.Long,
		Aliases: c.Aliases,
		Flags:   flagsToSpec(c.Flags()),
	}

	for _, sc := range c.Commands() {
		if !sc.IsAvailableCommand() || sc.Name() == "help" {
			continue
		}
		spec.Sub = append(spec.Sub, commandToSpec(sc))
	}
	return spec
}

var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Machine-readable command reference",
	Long: `Outputs a JSON schema-like description of commands, flags, and semantics.

Intended for LLMs/agents and scripting tools.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		spec := llmSpec{
			Tool:      "craft",
			Version:   version,
			Generated: time.Now().UTC().Format(time.RFC3339),
			Global:    flagsToSpec(rootCmd.PersistentFlags()),
			Notes: []string{
				"Default output is JSON. Use --format compact (legacy JSON), table, or markdown for human output where supported.",
				"craft delete is a soft-delete to trash (DELETE /documents).",
				"craft clear deletes all content blocks in a document (cannot be undone without a backup).",
				"craft update supports --mode append|replace and auto-chunks large markdown inserts.",
			},
		}

		for _, c := range rootCmd.Commands() {
			if !c.IsAvailableCommand() || c.Name() == "help" {
				continue
			}
			spec.Commands = append(spec.Commands, commandToSpec(c))
		}

		return outputJSON(spec)
	},
}

func init() {
	rootCmd.AddCommand(llmCmd)
}
