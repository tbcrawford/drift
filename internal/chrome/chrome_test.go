package chrome_test

import (
	"strings"
	"testing"

	"github.com/tbcrawford/drift/internal/chrome"
)

func TestDriftTheme_colored(t *testing.T) {
	theme := chrome.DriftTheme{}
	out := theme.RenderFileHeader("foo.go", false, 80)
	if !strings.Contains(out, "▸") {
		t.Errorf("expected chevron '▸' in colored output, got: %q", out)
	}
	if !strings.Contains(out, "foo.go") {
		t.Errorf("expected filename 'foo.go' in colored output, got: %q", out)
	}
	if !strings.Contains(out, "─") {
		t.Errorf("expected rule character '─' in colored output, got: %q", out)
	}
}

func TestDriftTheme_noColor(t *testing.T) {
	theme := chrome.DriftTheme{}
	out := theme.RenderFileHeader("foo.go", true, 80)
	if !strings.Contains(out, "▸ foo.go") {
		t.Errorf("expected '▸ foo.go' in no-color output, got: %q", out)
	}
	if !strings.Contains(out, "---") {
		t.Errorf("expected plain dashes '---' in no-color output, got: %q", out)
	}
}

func TestDeltaTheme_colored(t *testing.T) {
	theme := chrome.DeltaTheme{}
	out := theme.RenderFileHeader("foo.go", false, 80)
	if !strings.Contains(out, "┌") {
		t.Errorf("expected box corner '┌' in colored output, got: %q", out)
	}
	if !strings.Contains(out, "foo.go") {
		t.Errorf("expected filename 'foo.go' in colored output, got: %q", out)
	}
	if !strings.Contains(out, "┘") {
		t.Errorf("expected box corner '┘' in colored output, got: %q", out)
	}
}

func TestDeltaTheme_noColor(t *testing.T) {
	theme := chrome.DeltaTheme{}
	out := theme.RenderFileHeader("foo.go", true, 80)
	if !strings.HasPrefix(out, "+--") {
		t.Errorf("expected plain ASCII box starting with '+--', got: %q", out)
	}
}

func TestParseChromeTheme(t *testing.T) {
	tests := []struct {
		name      string
		wantName  string
		wantError bool
	}{
		{"drift", "drift", false},
		{"", "drift", false},
		{"delta", "delta", false},
		{"bogus", "", true},
		{"DRIFT", "drift", false},
		{"Delta", "delta", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme, err := chrome.ParseChromeTheme(tt.name)
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseChromeTheme(%q): expected error, got nil", tt.name)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseChromeTheme(%q): unexpected error: %v", tt.name, err)
				return
			}
			if theme.Name() != tt.wantName {
				t.Errorf("ParseChromeTheme(%q).Name() = %q, want %q", tt.name, theme.Name(), tt.wantName)
			}
		})
	}
}

func TestDriftTheme_zeroWidth(t *testing.T) {
	theme := chrome.DriftTheme{}
	// Must not panic; falls back to 80 columns.
	out := theme.RenderFileHeader("foo.go", true, 0)
	if len(out) == 0 {
		t.Error("expected non-empty output with termWidth=0")
	}
	// Should contain 80 dashes (plain fallback)
	if !strings.Contains(out, strings.Repeat("-", 80)) {
		t.Errorf("expected 80 dashes in fallback output, got: %q", out)
	}
}
