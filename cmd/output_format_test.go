package cmd

import (
	"testing"

	"github.com/ashrafali/craft-cli/internal/config"
)

func TestDefaultFormatIsJson(t *testing.T) {
	oldCfgManager := cfgManager
	oldOutputFormat := outputFormat
	t.Cleanup(func() {
		cfgManager = oldCfgManager
		outputFormat = oldOutputFormat
	})

	t.Setenv("HOME", t.TempDir())
	manager, err := config.NewManager()
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}
	cfgManager = manager
	outputFormat = ""

	if got := getOutputFormat(); got != "json" {
		t.Fatalf("expected default format json, got %s", got)
	}
}

func TestCompactFormatIsSupported(t *testing.T) {
	if !IsValidFormat("compact") {
		t.Fatalf("expected compact to be a valid output format")
	}
}
