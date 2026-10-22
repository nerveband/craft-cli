package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

// Share repeated JSON Schema subtrees so whole-CLI discovery stays concise.
// Standalone --cli-response-schema remains self-contained.
func internContract(value interface{}, definitions map[string]interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		out := map[string]interface{}{}
		for key, item := range v {
			if key == "components" {
				continue
			}
			out[key] = internContract(item, definitions)
		}
		data, _ := json.Marshal(out)
		if len(data) > 768 {
			sum := sha256.Sum256(data)
			key := "s" + hex.EncodeToString(sum[:8])
			definitions[key] = out
			return map[string]interface{}{"$ref": "#/extensions/contract_definitions/" + key}
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(v))
		for i, item := range v {
			out[i] = internContract(item, definitions)
		}
		return out
	case string:
		if strings.HasPrefix(v, "#/components/") {
			return strings.Replace(v, "#/components/", "#/extensions/rest_components/", 1)
		}
		return v
	default:
		return value
	}
}
