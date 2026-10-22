package cmd

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"strings"
	"testing"
)

func TestAllRunnableCommandsHaveCLIContractsAndExamples(t *testing.T) {
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		if c.Runnable() {
			if c != rootCmd && c.Example == "" {
				t.Errorf("missing example %s", c.CommandPath())
			}
			if len(cliOutputSchema(c)) == 0 || len(cliInputSchema(c)) == 0 {
				t.Errorf("missing contract %s", c.CommandPath())
			}
		}
		for _, child := range c.Commands() {
			walk(child)
		}
	}
	walk(rootCmd)
	for _, name := range []string{"list", "get", "search", "create", "update", "delete", "move", "clear"} {
		c, _, err := rootCmd.Find([]string{"documents", name})
		if err != nil || canonicalCommandPath(c) != name {
			t.Fatal(name, err)
		}
		e, ok := declaredEffect(c)
		if !ok || e != commandEffects[name] {
			t.Fatal("alias effect drift", name)
		}
	}
	raw, err := json.Marshal(cliSpec(""))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "stdout_schema") || !strings.Contains(string(raw), "input_schema") {
		t.Fatal("missing contracts")
	}
}
