package drift

// Algorithm selects the diff algorithm to use.
type Algorithm int

const (
	// Myers is the default O(ND) algorithm; fastest for small edit distances.
	Myers Algorithm = iota
	// Patience selects unique-line anchors; better for refactored code.
	Patience
	// Histogram is Git's preferred algorithm; frequency-aware anchor selection.
	Histogram
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
	noColor       bool
	lang          string
	theme         string
	split         bool
	lineNumbers   bool
	lineDiffStyle bool
	wordDiff      bool
	themeResolved func(string)
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
			algorithm:    Myers,
			contextLines: 3,
		},
		render: renderConfig{
			lineNumbers:   true,
			lineDiffStyle: true,
			wordDiff:      true,
		},
	}
}

// WithAlgorithm sets the diff algorithm (Myers, Patience, or Histogram).
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
