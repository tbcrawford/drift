// Package drift provides a production-quality text diff library and CLI for Go.
//
// Drift computes line-level diffs between two multi-line strings using the Myers,
// Patience, or Histogram algorithms and renders terminal output with Chroma syntax
// highlighting in unified or side-by-side split layouts — the same quality you see in
// GitHub's PR review UI, delivered to your terminal or any [io.Writer].
//
// # Functional API
//
// The simplest usage is two calls:
//
//	result, err := drift.Diff(oldText, newText)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if err := drift.Render(result, os.Stdout, drift.WithLang("go")); err != nil {
//	    log.Fatal(err)
//	}
//
// To include named file headers (like git diff output), use [RenderWithNames]:
//
//	drift.RenderWithNames(result, os.Stdout, "old.go", "new.go", drift.WithLang("go"))
//
// # Builder API
//
// For repeated operations with shared settings, the fluent builder avoids repeating
// options on every call:
//
//	b := drift.New().Algorithm(drift.Histogram).Theme("dracula").Split()
//	result, err := b.Diff(oldText, newText)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if err := b.Render(result, os.Stdout); err != nil {
//	    log.Fatal(err)
//	}
//
// # Diff Options
//
// Options accepted by [Diff]:
//
//   - [WithAlgorithm]: choose Myers (default, O(ND), fastest), Patience (unique-line
//     anchors, better for refactors), or Histogram (Git's default, frequency-aware
//     anchor selection).
//   - [WithContext]: number of unchanged context lines surrounding each hunk (default 3,
//     matching git diff -U3).
//
// # Render Options
//
// Options accepted by [Render] and [RenderWithNames]:
//
//   - [WithNoColor]: disable all ANSI sequences. Also honoured when the NO_COLOR
//     environment variable is set or stdout is not a TTY.
//   - [WithLang]: Chroma lexer name for syntax highlighting (e.g. "go", "python").
//   - [WithTheme]: Chroma style name (e.g. "github", "monokai", "dracula"). When
//     omitted on a Unix TTY, drift queries the terminal palette (OSC 4) and picks the
//     closest registered Chroma style automatically.
//   - [WithThemeResolved]: callback invoked with the resolved theme name after style
//     selection, useful for debugging or display.
//   - [WithSplit]: render as two equal-width panels (old on left, new on right)
//     separated by a │ column; default is unified output.
//   - [WithLineNumbers] / [WithoutLineNumbers]: show or hide old/new line-number gutters
//     (default: visible). The CLI flag is --no-line-numbers.
//   - [WithLineDiffStyle]: toggle theme-derived full-line background colour on added
//     and removed lines (default: enabled).
//   - [WithWordDiff]: toggle word-level intra-line highlights on paired delete/insert
//     lines (default: enabled).
//
// # Git Integration
//
// The drift CLI accepts a single file path inside a Git worktree and compares the
// working-tree version to HEAD, equivalent to git diff HEAD -- path:
//
//	drift ./internal/foo.go
//
// This does not affect the library API; library callers always pass two strings to
// [Diff] directly.
//
// See the examples/ directory for runnable examples.
package drift
