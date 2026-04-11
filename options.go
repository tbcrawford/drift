package drift

import (
	"fmt"

	"github.com/charmbracelet/colorprofile"
)

// Algorithm selects the diff algorithm to use.
type Algorithm int

const (
	// Myers is an O(ND) algorithm; fastest for small edit distances.
	Myers Algorithm = iota
	// Patience selects unique-line anchors; better for refactored code.
	Patience
	// Histogram is Git's preferred algorithm; frequency-aware anchor selection.
	Histogram
	// Auto selects Myers or Histogram based on file size and line-frequency analysis.
	Auto
)

// Option is a functional option that configures a diff operation.
type Option func(*config)

// diffConfig holds options that affect the diff algorithm and hunk construction.
type diffConfig struct {
	algorithm    Algorithm
	contextLines int
}

// renderConfig holds options that affect terminal rendering.
type renderConfig struct {
	noColor            bool
	colorProfile       colorprofile.Profile
	hasProfile         bool // true when colorProfile was explicitly set via WithColorProfile
	isDark             bool
	hasIsDark          bool // true when isDark was explicitly set via WithIsDark
	lang               string
	theme              string
	split              bool
	lineNumbers        bool
	lineDiffStyle      bool
	wordDiff           bool
	termWidth          int
	themeResolved      func(string)
	hunkHeaderRenderer func(newStart int, codeFragment string, noColor bool) string
	gutterMiddleSep    string
	gutterRightBorder  string
}

// config holds all configuration for Diff and Render operations.
// Options that affect only diffing are in diff; rendering options are in render.
type config struct {
	diff   diffConfig
	render renderConfig
}

// defaultConfig returns a config with production-ready defaults.
func defaultConfig() *config {
	return &config{
		diff: diffConfig{
			algorithm:    Auto,
			contextLines: 3,
		},
		render: renderConfig{
			lineNumbers:   true,
			lineDiffStyle: true,
			wordDiff:      true,
		},
	}
}

// WithAlgorithm sets the diff algorithm (Myers, Patience, Histogram, or Auto).
func WithAlgorithm(a Algorithm) Option {
	return func(c *config) { c.diff.algorithm = a }
}

// WithContext sets the number of unchanged context lines surrounding each hunk.
// Default is 3, matching git diff -U3.
func WithContext(n int) Option {
	return func(c *config) { c.diff.contextLines = n }
}

// WithNoColor disables all ANSI color output. Also respected when NO_COLOR
// environment variable is set.
func WithNoColor() Option {
	return func(c *config) { c.render.noColor = true }
}

// WithColorProfile sets the terminal color profile used for rendering.
// Use this to preserve ANSI colors when rendering to a buffer that will later
// be written to a real terminal (e.g., through a pager). Without this, a
// bytes.Buffer destination causes resolveProfile to return NoTTY and emit
// plain text. Detect the profile from the real output file before buffering:
//
//	if f, ok := streams.Out.(*os.File); ok {
//	    opts = append(opts, drift.WithColorProfile(colorprofile.Detect(f, os.Environ())))
//	}
func WithColorProfile(p colorprofile.Profile) Option {
	return func(c *config) {
		c.render.colorProfile = p
		c.render.hasProfile = true
	}
}

// WithIsDark sets whether the terminal has a dark background, bypassing the
// runtime OSC 11 terminal query. Use this when the background has already been
// detected (e.g. from the real TTY before rendering to a buffer) so that
// concurrent or buffered render calls do not issue additional terminal queries.
//
//	if f, ok := streams.Out.(*os.File); ok {
//	    opts = append(opts, drift.WithIsDark(lipgloss.HasDarkBackground(os.Stdin, f)))
//	}
func WithIsDark(dark bool) Option {
	return func(c *config) {
		c.render.isDark = dark
		c.render.hasIsDark = true
	}
}

// WithLang overrides the language used for Chroma syntax highlighting.
// Use Chroma language names (e.g., "go", "python", "javascript").
func WithLang(lang string) Option {
	return func(c *config) { c.render.lang = lang }
}

// WithTheme sets the Chroma theme for syntax highlighting.
// Use Chroma style names (e.g., "monokai", "github", "dracula").
func WithTheme(theme string) Option {
	return func(c *config) { c.render.theme = theme }
}

// WithThemeResolved registers a callback invoked with the resolved Chroma theme
// name after style selection (including OSC 4 best-match on supported Unix TTYs).
func WithThemeResolved(fn func(string)) Option {
	return func(c *config) { c.render.themeResolved = fn }
}

// WithSplit enables side-by-side split diff rendering.
// The output is rendered as two equal-width panels (old on the left,
// new on the right) joined by a space + │ (light vertical, TUI box stroke) separator.
func WithSplit() Option {
	return func(c *config) { c.render.split = true }
}

// WithLineNumbers toggles old/new gutter columns before the diff prefix (unified)
// or inside each split panel. The default is true.
func WithLineNumbers(v bool) Option {
	return func(c *config) { c.render.lineNumbers = v }
}

// WithoutLineNumbers disables line-number gutters (same as WithLineNumbers(false)).
func WithoutLineNumbers() Option {
	return func(c *config) { c.render.lineNumbers = false }
}

// WithLineDiffStyle toggles theme-derived full-line backgrounds on added and removed
// lines in unified and split output. The default is true when using Render.
func WithLineDiffStyle(v bool) Option {
	return func(c *config) { c.render.lineDiffStyle = v }
}

// WithWordDiff toggles word-level intra-line highlights for paired delete/insert
// lines in unified and split output. The default is true when using Render.
func WithWordDiff(v bool) Option {
	return func(c *config) { c.render.wordDiff = v }
}

// WithTermWidth sets a fixed terminal width for split diff rendering.
// Use this to produce deterministic output in tests or non-TTY environments
// where the terminal width cannot be detected automatically. A value of 0
// (the default) falls back to automatic detection (80 columns for non-TTYs).
func WithTermWidth(w int) Option {
	return func(c *config) { c.render.termWidth = w }
}

// WithHunkHeaderRenderer sets a custom function to render hunk headers.
// When fn returns a non-empty string for a given (newStart, codeFragment, noColor)
// triple, that string is written verbatim instead of the standard
// "@@ -old,n +new,n @@" format. Returning "" falls back to the standard format.
// Passing nil clears any previously set renderer.
func WithHunkHeaderRenderer(fn func(newStart int, codeFragment string, noColor bool) string) Option {
	return func(c *config) { c.render.hunkHeaderRenderer = fn }
}

// WithGutterSeparators configures the gutter column separator strings.
// middleSep is placed between the old and new line-number columns.
// rightBorder is placed after the new line-number column and before line content in unified mode.
// Both default to "" which preserves the built-in defaults (" │" middle, "" right border).
// This is used internally by chrome themes; most callers do not need to set this directly.
func WithGutterSeparators(middleSep, rightBorder string) Option {
	return func(c *config) {
		c.render.gutterMiddleSep = middleSep
		c.render.gutterRightBorder = rightBorder
	}
}

// validate checks that all config fields hold valid values.
// It is called by Diff() before any diff work begins.
func (c *config) validate() error {
	if c.diff.contextLines < 0 {
		return fmt.Errorf("drift: WithContext value must be non-negative, got %d", c.diff.contextLines)
	}
	return nil
}
