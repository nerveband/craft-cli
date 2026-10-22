package cmd

import (
	"github.com/spf13/cobra"
	"time"
)

var responseBackend = "rest"
var activeCommand *cobra.Command
var maxItems int
var maxChars int

func addOutputMetadata(value interface{}) interface{} {
	if activeCommand == nil || activeCommand.Name() == "schema" || getOutputOnly() != "" || transformExpr != "" || outputFormat == "raw" {
		return value
	}
	effect, ok := declaredEffect(activeCommand)
	if !ok {
		return value
	}
	source := responseSource
	if effect.Scope == "local" {
		source = "local"
		responseBackend = "local"
	}
	meta := map[string]interface{}{"scope": effect.Scope, "source": source, "backend": responseBackend}
	if source == "local" {
		meta["scope"] = "local"
	}
	if effect.Effects == "read_only" {
		for _, name := range []string{"max-depth", "depth"} {
			if activeCommand.Flags().Lookup(name) != nil {
				depth, _ := activeCommand.Flags().GetInt(name)
				meta["depth_limit"] = depth
			}
		}
	}
	if !responseSnapshot.IsZero() {
		meta["snapshot_at"] = responseSnapshot.Format(time.RFC3339)
		meta["age_seconds"] = int64(time.Since(responseSnapshot).Seconds())
	}
	obj, ok := value.(map[string]interface{})
	if !ok {
		return value
	}
	// Upstream writes remain lossless; read output has explicit bounded defaults.
	if effect.Effects == "read_only" && effect.Scope == "remote" {
		limit := maxItems
		if flag := activeCommand.Flags().Lookup("limit"); flag != nil {
			limit = 0
		} // Native limits include their own metadata.
		trimmedItems, trimmedChars := 0, 0
		var bound func(interface{})
		bound = func(value interface{}) {
			switch v := value.(type) {
			case map[string]interface{}:
				for key, item := range v {
					if key == "_metadata" || key == "_provenance" {
						continue
					}
					if text, ok := item.(string); ok && maxChars > 0 && (key == "markdown" || key == "content" || key == "rawCode") {
						runes := []rune(text)
						if len(runes) > maxChars {
							trimmedChars += len(runes) - maxChars
							v[key] = string(runes[:maxChars])
						}
					}
					if items, ok := item.([]interface{}); ok && limit > 0 && (key == "items" || key == "content" || key == "elements" || key == "views") {
						if len(items) > limit {
							trimmedItems += len(items) - limit
							v[key] = items[:limit]
						}
					}
					bound(v[key])
				}
			case []interface{}:
				for _, item := range v {
					bound(item)
				}
			}
		}
		bound(obj)
		if trimmedItems > 0 {
			meta["truncated_items"] = trimmedItems
			meta["full_output_hint"] = "Use --max-items 0 and --limit 0 where supported"
		}
		if trimmedChars > 0 {
			meta["truncated_characters"] = trimmedChars
			meta["full_content_hint"] = "Use --max-chars 0 and --max-depth -1 (or --depth -1)"
		}

	}
	obj["_provenance"] = meta
	return obj
}
func init() {
	rootCmd.PersistentFlags().IntVar(&maxItems, "max-items", 50, "Maximum records for read commands without --limit (0 = all; raw output is unchanged)")
	rootCmd.PersistentFlags().IntVar(&maxChars, "max-chars", 20000, "Maximum document text characters (0 = all; raw output is unchanged)")
}
