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

func TestReplaceSectionByDecoratedHeading(t *testing.T) {
	md := "# Title\n\n<callout>## Your Overall Investment</callout>\n\nOld pricing.\n\n## Next\n\nOther.\n"

	out, err := replaceSectionByHeading(md, "Your Overall Investment", "New pricing")
	if err != nil {
		t.Fatalf("replaceSectionByHeading() error = %v", err)
	}
	if want := "<callout>## Your Overall Investment</callout>\n\nNew pricing"; !contains(out, want) {
		t.Fatalf("expected output to contain %q, got:\n%s", want, out)
	}
	if contains(out, "Old pricing") {
		t.Fatalf("expected old decorated section removed, got:\n%s", out)
	}
	if !contains(out, "## Next") {
		t.Fatalf("expected subsequent section preserved, got:\n%s", out)
	}
}

func TestReplaceSectionReplacementWithDecoratedHeading(t *testing.T) {
	md := "# Title\n\n## Pricing\n\nOld pricing.\n\n## Next\n\nOther.\n"

	out, err := replaceSectionByHeading(md, "Pricing", "<callout>## New Pricing</callout>\n\nNew pricing")
	if err != nil {
		t.Fatalf("replaceSectionByHeading() error = %v", err)
	}
	if want := "<callout>## New Pricing</callout>\n\nNew pricing"; !contains(out, want) {
		t.Fatalf("expected output to contain %q, got:\n%s", want, out)
	}
	if contains(out, "## Pricing\n\n<callout>## New Pricing</callout>") {
		t.Fatalf("expected decorated replacement heading not to be wrapped with original heading, got:\n%s", out)
	}
}
