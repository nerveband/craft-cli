package cmd

import (
	"strings"
	"testing"

	"github.com/ashrafali/craft-cli/internal/models"
)

func TestRenderBlockCraftDoesNotDoubleWrapCallout(t *testing.T) {
	block := models.Block{
		Type:        "text",
		Decorations: []string{"callout"},
		Markdown:    "<callout>## Your Overall Investment</callout>",
	}

	var sb strings.Builder
	renderBlockCraft(&sb, &block, 0)

	got := sb.String()
	if strings.Count(got, "<callout>") != 1 || strings.Count(got, "</callout>") != 1 {
		t.Fatalf("rendered callout should be wrapped exactly once, got %q", got)
	}
}

func TestRenderBlockCraftDoesNotDoubleWrapQuote(t *testing.T) {
	block := models.Block{
		Type:        "text",
		Decorations: []string{"quote"},
		Markdown:    "> Existing quote",
	}

	var sb strings.Builder
	renderBlockCraft(&sb, &block, 0)

	got := sb.String()
	if strings.Count(got, ">") != 1 {
		t.Fatalf("rendered quote should be prefixed exactly once, got %q", got)
	}
}
