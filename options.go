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

// config holds all diff configuration. Unexported — callers use Option functions.
type config struct {
	algorithm    Algorithm
	contextLines int
	noColor      bool
	lang         string
	theme        string
	split        bool
}

// defaultConfig returns a config with production-ready defaults.
func defaultConfig() *config {
	return &config{
		algorithm:    Myers,
		contextLines: 3,
	}
}

// WithAlgorithm sets the diff algorithm (Myers, Patience, or Histogram).
func WithAlgorithm(a Algorithm) Option {
	return func(c *config) { c.algorithm = a }
}

// WithContext sets the number of unchanged context lines surrounding each hunk.
// Default is 3, matching git diff -U3.
func WithContext(n int) Option {
	return func(c *config) { c.contextLines = n }
}

// WithNoColor disables all ANSI color output. Also respected when NO_COLOR
// environment variable is set.
func WithNoColor() Option {
	return func(c *config) { c.noColor = true }
}

// WithLang overrides the language used for Chroma syntax highlighting.
// Use Chroma language names (e.g., "go", "python", "javascript").
func WithLang(lang string) Option {
	return func(c *config) { c.lang = lang }
}

// WithTheme sets the Chroma theme for syntax highlighting.
// Use Chroma style names (e.g., "monokai", "github", "dracula").
func WithTheme(theme string) Option {
	return func(c *config) { c.theme = theme }
}

// WithSplit enables side-by-side split diff rendering.
// The output is rendered as two equal-width panels (old on the left,
// new on the right) joined by a " │ " separator.
func WithSplit() Option {
	return func(c *config) { c.split = true }
}
