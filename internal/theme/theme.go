// Package theme provides terminal background detection for drift's syntax
// highlighting theme selection.
package theme

import (
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

// DetectDarkBackground returns true if the terminal has a dark background.
//
// It guards the Lip Gloss v2 OSC 11 terminal query behind a color profile check:
// if the output is not a TTY or colors are disabled, it returns true (dark default)
// immediately without sending any terminal escape sequences. This prevents the
// 2-second timeout that HasDarkBackground would otherwise incur on non-TTY outputs.
func DetectDarkBackground(profile colorprofile.Profile) bool {
	// Skip OSC 11 query for non-TTY and no-color environments.
	// HasDarkBackground defaults to true on error anyway, but we avoid the
	// 2-second timeout path entirely.
	if profile == colorprofile.NoTTY || profile == colorprofile.Ascii {
		return true
	}
	return lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
}
