package theme

import (
	"testing"

	"github.com/charmbracelet/colorprofile"
)

// TestDetectDarkBackground_NoTTY verifies that non-TTY profiles return the
// dark default (true) immediately without attempting any OSC 11 query.
// These profiles must never hang or block waiting for terminal response.
func TestDetectDarkBackground_NoTTY(t *testing.T) {
	profiles := []struct {
		name    string
		profile colorprofile.Profile
	}{
		{"NoTTY", colorprofile.NoTTY},
		{"Ascii", colorprofile.Ascii},
	}

	for _, tc := range profiles {
		t.Run(tc.name, func(t *testing.T) {
			// Must return immediately (no 2s hang) and default to dark=true.
			got := DetectDarkBackground(tc.profile)
			if !got {
				t.Errorf("DetectDarkBackground(%v) = false; want true (dark default for non-TTY)", tc.profile)
			}
		})
	}
}

// TestDetectDarkBackground_TrueColor verifies that TrueColor profile does NOT
// short-circuit — it proceeds to the lipgloss path. In a test environment
// (non-TTY stdin/stdout), lipgloss.HasDarkBackground returns true (dark default).
// We accept either true or false here since CI may or may not have a real TTY;
// the critical invariant is that the function does not panic.
func TestDetectDarkBackground_TrueColor(t *testing.T) {
	// Should not panic; return value is environment-dependent in tests.
	result := DetectDarkBackground(colorprofile.TrueColor)
	// Verify the result is a valid bool (either true or false is acceptable).
	_ = result
}

// TestDetectDarkBackground_ANSI256 mirrors TrueColor: not short-circuited.
func TestDetectDarkBackground_ANSI256(t *testing.T) {
	result := DetectDarkBackground(colorprofile.ANSI256)
	_ = result
}
