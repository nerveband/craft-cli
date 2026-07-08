package cmd

import "testing"

func TestValidateResourceID(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"abc123", true},
		{"9773E4A5-A5B0-4817-833B-FE11C4A57679", true},
		{"../../../etc/passwd", false},
		{"doc123?admin=true", false},
		{"doc123#fragment", false},
		{"..%2f..%2fetc%2fpasswd", false},
		{"doc\x00id", false},
		{"doc\nid", false},
		{"", false},
		{"a", true},
		{"hello&world=1", false},
		{"normal-id-with-dashes", true},
		{"8cdffe2e-f858-a063-ea91-bb5b27609783", true},
	}
	for _, tt := range tests {
		err := validateResourceID(tt.input, "test-id")
		if tt.valid && err != nil {
			t.Errorf("validateResourceID(%q) = error %v, want nil", tt.input, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("validateResourceID(%q) = nil, want error", tt.input)
		}
	}
}

func TestValidateOutputPath(t *testing.T) {
	t.Run("inside cwd", func(t *testing.T) {
		if err := validateOutputPath("out.md", false); err != nil {
			t.Fatalf("validateOutputPath() error = %v", err)
		}
	})

	t.Run("outside cwd rejected", func(t *testing.T) {
		if err := validateOutputPath("../out.md", false); err == nil {
			t.Fatal("expected outside cwd path to be rejected")
		}
	})

	t.Run("outside cwd allowed with override", func(t *testing.T) {
		if err := validateOutputPath("../out.md", true); err != nil {
			t.Fatalf("validateOutputPath() with override error = %v", err)
		}
	})
}
