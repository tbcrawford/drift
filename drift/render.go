package drift

import (
	"io"
	"os"

	"github.com/alecthomas/chroma/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/tylercrawford/drift/internal/highlight"
	"github.com/tylercrawford/drift/internal/render"
	"github.com/tylercrawford/drift/internal/terminal"
	"github.com/tylercrawford/drift/internal/theme"
)

// Render writes a unified diff of result to w with Chroma syntax highlighting.
//
// Color profile is detected automatically from w when w is an *os.File.
// For non-file writers (e.g., bytes.Buffer), the profile is treated as NoTTY
// and output is plain text. Use WithNoColor() to explicitly disable colors.
//
// Lexer detection uses the explicit lang option only; no filename or content
// analysis is performed. Use RenderWithNames to enable extension-based detection.
//
// Example:
//
//	result, err := drift.Diff(old, new)
//	if err != nil {
//	    return err
//	}
//	return drift.Render(result, os.Stdout)
func Render(result DiffResult, w io.Writer, opts ...Option) error {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}

	pipeline := buildRenderPipeline(w, cfg, "")

	rcfg := &render.RenderConfig{
		Lang:            cfg.render.lang,
		Lexer:           pipeline.lexer,
		Style:           pipeline.style,
		Formatter:       pipeline.formatter,
		Profile:         pipeline.profile,
		NoColor:         cfg.render.noColor,
		TermWidth:       pipeline.termWidth,
		ShowLineNumbers: cfg.render.lineNumbers,
		IsDark:          pipeline.isDark,
		LineDiffStyle:   cfg.render.lineDiffStyle,
		WordDiff:        cfg.render.wordDiff,
	}

	if cfg.render.split {
		return render.Split(result, pipeline.wrapped, rcfg)
	}
	return render.Unified(result, pipeline.wrapped, rcfg)
}

// RenderWithNames is like Render but includes file path labels in the diff header.
//
// oldName and newName appear in the "--- oldName" and "+++ newName" header lines,
// matching the format produced by git diff. Pass empty strings to use the defaults
// ("a/input" and "b/input").
//
// Lexer detection uses, in order: explicit lang option, oldName extension, then
// falls back to plaintext. Content-based detection is not performed.
func RenderWithNames(result DiffResult, w io.Writer, oldName, newName string, opts ...Option) error {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}

	pipeline := buildRenderPipeline(w, cfg, oldName)

	rcfg := &render.RenderConfig{
		OldName:         oldName,
		NewName:         newName,
		Lang:            cfg.render.lang,
		Lexer:           pipeline.lexer,
		Style:           pipeline.style,
		Formatter:       pipeline.formatter,
		Profile:         pipeline.profile,
		NoColor:         cfg.render.noColor,
		TermWidth:       pipeline.termWidth,
		ShowLineNumbers: cfg.render.lineNumbers,
		IsDark:          pipeline.isDark,
		LineDiffStyle:   cfg.render.lineDiffStyle,
		WordDiff:        cfg.render.wordDiff,
	}

	if cfg.render.split {
		return render.Split(result, pipeline.wrapped, rcfg)
	}
	return render.Unified(result, pipeline.wrapped, rcfg)
}

// renderPipeline holds the resolved rendering dependencies for a single call.
type renderPipeline struct {
	profile   colorprofile.Profile
	isDark    bool
	lexer     chroma.Lexer
	style     *chroma.Style
	formatter chroma.Formatter
	wrapped   io.Writer
	termWidth int
}

// buildRenderPipeline resolves all rendering dependencies from w, cfg, and an
// optional filename for lexer detection. It is shared by Render and RenderWithNames
// to eliminate duplicated setup code.
func buildRenderPipeline(w io.Writer, cfg *config, filename string) renderPipeline {
	profile := resolveProfile(w, cfg)
	isDark := theme.DetectDarkBackground(profile)

	lexer := highlight.DetectLexer(cfg.render.lang, filename, "")
	style := resolveChromaStyle(cfg, profile, w, isDark)
	formatter := highlight.FormatterForProfile(profile)

	// Wrap the writer for automatic ANSI downsampling when it is an *os.File.
	wrapped := colorprofile.NewWriter(w, os.Environ())

	termWidth := render.TerminalWidth(w)

	return renderPipeline{
		profile:   profile,
		isDark:    isDark,
		lexer:     lexer,
		style:     style,
		formatter: formatter,
		wrapped:   wrapped,
		termWidth: termWidth,
	}
}

func autoThemeName(isDark bool) string {
	if isDark {
		return "github-dark"
	}
	return "github"
}

func resolveChromaStyle(cfg *config, profile colorprofile.Profile, w io.Writer, isDark bool) *chroma.Style {
	var style *chroma.Style
	var name string

	if cfg.render.theme != "" {
		style = highlight.SelectTheme(cfg.render.theme, isDark)
		name = cfg.render.theme
	} else if cfg.render.noColor || profile == colorprofile.NoTTY || profile == colorprofile.Ascii {
		style = highlight.SelectTheme("", isDark)
		name = autoThemeName(isDark)
	} else if _, ok := w.(*os.File); ok {
		if palette, err := terminal.QueryANSIPalette(); err == nil && palette != nil && len(palette) > 0 {
			name = highlight.BestMatchTheme(palette)
			style = highlight.SelectTheme(name, isDark)
		} else {
			style = highlight.SelectTheme("", isDark)
			name = autoThemeName(isDark)
		}
	} else {
		style = highlight.SelectTheme("", isDark)
		name = autoThemeName(isDark)
	}

	if cfg.render.themeResolved != nil {
		cfg.render.themeResolved(name)
	}
	return style
}

// resolveProfile determines the terminal color profile for the given writer
// and config. The resolution order is:
//  1. cfg.render.noColor == true or NO_COLOR is set → Ascii (no color)
//  2. w is an *os.File → colorprofile.Detect(w, os.Environ())
//  3. otherwise → NoTTY (non-file writers without NO_COLOR are treated as piped output)
func resolveProfile(w io.Writer, cfg *config) colorprofile.Profile {
	if cfg.render.noColor || os.Getenv("NO_COLOR") != "" {
		return colorprofile.Ascii
	}
	if f, ok := w.(*os.File); ok {
		return colorprofile.Detect(f, os.Environ())
	}
	return colorprofile.NoTTY
}
