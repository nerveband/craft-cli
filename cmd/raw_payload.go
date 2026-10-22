package cmd

import (
	"fmt"
	"strings"
)

func parseJSONObject(input string) (map[string]interface{}, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty JSON payload")
	}
	var payload map[string]interface{}
	if err := decodeInputJSON([]byte(input), &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON object: %w", err)
	}
	if payload == nil {
		return nil, fmt.Errorf("expected a JSON object")
	}
	return payload, nil
}

func readRawPayload(jsonPayload string, stdin bool) (map[string]interface{}, error) {
	if jsonPayload != "" && stdin {
		return nil, fmt.Errorf("--json and --stdin are mutually exclusive")
	}
	if stdin {
		input, err := readStdinString()
		if err != nil {
			return nil, err
		}
		return parseJSONObject(input)
	}
	return parseJSONObject(jsonPayload)
}

func payloadString(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key].(string); ok {
			return value
		}
	}
	return ""
}

func payloadArray(payload map[string]interface{}, key string) ([]interface{}, bool) {
	items, ok := payload[key].([]interface{})
	return items, ok
}
