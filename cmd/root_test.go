package cmd

import (
	"errors"
	"testing"
)

func TestCategorizeErrorTimeout(t *testing.T) {
	err := errors.New(`request failed: Post "https://connect.craft.do/links/x/api/v1/blocks": context deadline exceeded (Client.Timeout exceeded while awaiting headers)`)
	if got := categorizeError(err); got != "API_TIMEOUT" {
		t.Fatalf("categorizeError() = %q, want API_TIMEOUT", got)
	}
}

func TestErrorHintTimeoutWarnsWriteMayHaveLanded(t *testing.T) {
	hint := errorHint("API_TIMEOUT")
	for _, want := range []string{"may still have applied", "repeated reads", "avoid duplicates", "retryable"} {
		if !contains(hint, want) {
			t.Fatalf("hint %q does not contain %q", hint, want)
		}
	}
}
