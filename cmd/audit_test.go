package cmd

import "testing"

// The legacy diagnostic must count its actual checks. A perfect presence score
// is not a release gate; behavior is exercised by contract_v3_test and fixtures.
func TestLegacyAuditAccounting(t *testing.T) {
	result := runAgentDXAudit()
	max, score := 0, 0
	for _, category := range result.Categories {
		max += category.Max
		score += category.Score
	}
	if result.Max != max || result.Score != score || score > max {
		t.Fatalf("invalid accounting: %#v", result)
	}
}
