package cmd

import "testing"

func TestReplaceSectionByHeading(t *testing.T) {
	md := "# Title\n\n## Overview\n\nOld overview.\n\n### Details\n\nOld details.\n\n## Next\n\nOther.\n"

	out, err := replaceSectionByHeading(md, "Overview", "New overview\n\n- Item")
	if err != nil {
		t.Fatalf("replaceSectionByHeading() error = %v", err)
	}
	if want := "## Overview\n\nNew overview"; !contains(out, want) {
		t.Fatalf("expected output to contain %q, got:\n%s", want, out)
	}
	if contains(out, "Old overview") {
		t.Fatalf("expected old section removed, got:\n%s", out)
	}
	if !contains(out, "## Next") {
		t.Fatalf("expected subsequent sections preserved, got:\n%s", out)
	}
}
