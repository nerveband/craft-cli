package cmd

import "testing"

func TestIsJSONFormatAcceptsCompact(t *testing.T) {
	if !isJSONFormat(FormatJSON) {
		t.Fatalf("expected json to be a JSON format")
	}
	if !isJSONFormat(FormatCompact) {
		t.Fatalf("expected compact to be a JSON format")
	}
	if isJSONFormat(FormatTable) {
		t.Fatalf("expected table to not be a JSON format")
	}
}
