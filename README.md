# drift

**drift** is a Go library and CLI for line-level text diffs (Myers, Patience, Histogram) with Chroma syntax highlighting and unified or side-by-side terminal output—similar in spirit to GitHub’s PR diff view, for your terminal or your own tools.

In a color-capable TTY, output uses ANSI styles; when stdout is piped or you pass `--no-color` / `NO_COLOR`, output is plain text.

## Installation

```bash
go get github.com/tylercrawford/drift@latest
go install github.com/tylercrawford/drift/cmd/drift@latest
```

## CLI usage

The `drift` command compares two files, stdin pairs, or `--from` / `--to` strings and prints a pretty diff to stdout (exit code `1` when there are differences, `0` when identical).

```bash
drift old.go new.go
drift --algorithm histogram --split file_a.txt file_b.txt
drift --no-color --lang go ./old.go ./new.go
```

Common flags: `--algorithm` (`myers`, `patience`, `histogram`), `--split` (side-by-side), `--theme`, `--lang`, `--no-color`, `--context`, `--from` / `--to`.

## Library — functional API

```go
package main

import (
	"fmt"
	"os"

	"github.com/tylercrawford/drift"
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
- **Split**: pass `drift.WithSplit()` to `Render` (or `b.Split()` on the builder) for two panels separated by ` │ `.
- **Theme** (`WithTheme`): Chroma style name (e.g. `github`, `monokai`).
- **Language** (`WithLang`): Chroma lexer name (e.g. `go`, `python`) for highlighting.

```go
drift.Render(result, w, drift.WithSplit(), drift.WithTheme("dracula"), drift.WithLang("go"))
```

## License

This project is licensed under the MIT License — see [LICENSE](LICENSE).
