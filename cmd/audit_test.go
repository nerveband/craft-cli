package cmd

import "testing"

func TestRunAgentDXAuditMeetsTarget(t *testing.T) {
	result := runAgentDXAudit()

	if result.Max != 85 {
		t.Fatalf("Max = %d, want 85", result.Max)
	}
	if result.Score != 85 {
		t.Fatalf("Score = %d, want 85; failures = %#v", result.Score, result.Failures)
	}
	if result.Grade != "agent-native" {
		t.Fatalf("Grade = %q, want agent-native", result.Grade)
	}
	if len(result.Categories) != 15 {
		t.Fatalf("Categories = %d, want 15", len(result.Categories))
	}
}

func TestRunAgentDXAuditReportsNoFailures(t *testing.T) {
	result := runAgentDXAudit()

	if len(result.Failures) != 0 {
		t.Fatalf("expected no failures for 85/85 audit, got %#v", result.Failures)
	}
}
