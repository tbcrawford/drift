# drift

**drift** is a Go library and CLI for line-level text diffs (Myers, Patience, Histogram) with Chroma syntax highlighting and unified or side-by-side terminal output—similar in spirit to GitHub’s PR diff view, for your terminal or your own tools.

In a color-capable TTY, output uses ANSI styles; when stdout is piped or you pass `--no-color` / `NO_COLOR`, output is plain text.

## Installation

```bash
go get github.com/tbcrawford/drift/drift@latest
go install github.com/tbcrawford/drift/cmd/drift@latest
```

## CLI usage

The `drift` command compares two files, stdin pairs, or `--from` / `--to` strings and prints a pretty diff to stdout (exit code `1` when there are differences, `0` when identical).

```bash
drift old.go new.go
drift --algorithm histogram --split file_a.txt file_b.txt
drift --no-color --lang go ./old.go ./new.go
drift ./internal/foo.go   # inside a git repo: diff working tree vs HEAD
```

With a single path inside a Git worktree, `drift` compares the file on disk to the version at `HEAD` (same idea as `git diff HEAD -- path`). See `drift --help` for all flags.

Common flags: `--algorithm` (`myers`, `patience`, `histogram`), `--split` (side-by-side), `--theme`, `--lang`, `--no-color`, `--context`, `--from` / `--to`.

**Syntax theme:** When `--theme` is omitted on a Unix TTY with color output, `drift` queries the terminal palette (OSC 4) and picks the closest registered Chroma style. Piped stdout, `NO_COLOR`, or non-Unix environments use the usual dark/light defaults (`monokai` / `github`). The hidden flag `--show-theme` prints `drift: resolved syntax theme: <name>` to stderr so you can see which style was chosen.

## Library — functional API

```go
package main

import (
	"fmt"
	"os"

	"github.com/tbcrawford/drift/drift"
)

func main() {
	old, newText := "a\nb\n", "a\nc\n"
	result, err := drift.Diff(old, newText, drift.WithAlgorithm(drift.Myers))
	if err != nil {
		panic(err)
	}
	if err := drift.Render(result, os.Stdout, drift.WithNoColor(), drift.WithLang("go")); err != nil {
		panic(err)
	}
}
```

Disable optional visual enhancements individually:

```go
// Line numbers are on by default; hide gutters:
drift.Render(result, os.Stdout, drift.WithoutLineNumbers())

// Word-level highlights and full-line diff styling are on by default; disable:
drift.Render(result, os.Stdout, drift.WithWordDiff(false), drift.WithLineDiffStyle(false))
```

## Library — builder API

```go
b := drift.New().Algorithm(drift.Patience).Theme("github").NoColor()
result, err := b.Diff(old, newText)
if err != nil {
	panic(err)
}
if err := b.Render(result, os.Stdout); err != nil {
	panic(err)
}
```

## Rendering

- **Unified** (default): classic `---` / `+++` / `@@` hunks, one column.
- **Line numbers**: old/new gutters appear before the `+`/`-`/space prefix by default. Opt out with `drift.WithLineNumbers(false)`, `drift.WithoutLineNumbers()`, or the CLI flag `--no-line-numbers`.
- **Split**: pass `drift.WithSplit()` to `Render` (or `b.Split()` on the builder) for two panels separated by ` │ `.
- **Word diff** (`WithWordDiff`): word-level intra-line highlights on paired delete/insert lines (default: on). Disable with `drift.WithWordDiff(false)`.
- **Full-line diff style** (`WithLineDiffStyle`): theme-derived full-line background colour on added and removed lines (default: on). Disable with `drift.WithLineDiffStyle(false)`.
- **Theme** (`WithTheme`): Chroma style name (e.g. `github`, `monokai`).
- **Language** (`WithLang`): Chroma lexer name (e.g. `go`, `python`) for highlighting.

```go
drift.Render(result, w, drift.WithSplit(), drift.WithTheme("dracula"), drift.WithLang("go"))
```

```go
// Default: gutters on. Hide gutters:
drift.Render(result, w, drift.WithoutLineNumbers())
```

## License

This project is licensed under the MIT License — see [LICENSE](LICENSE).
