<div align="center">
  <img src="assets/drift.png" alt="drift" width="600" />
  <br/>
  <br/>

  <a href="https://pkg.go.dev/github.com/tbcrawford/drift"><img src="https://pkg.go.dev/badge/github.com/tbcrawford/drift.svg" alt="Go Reference" /></a>
  <a href="https://github.com/tbcrawford/drift/releases"><img src="https://img.shields.io/github/v/release/tbcrawford/drift" alt="Latest Release" /></a>
  <a href="https://github.com/tbcrawford/drift/actions"><img src="https://img.shields.io/github/actions/workflow/status/tbcrawford/drift/ci.yml?branch=main" alt="CI" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT License" /></a>

  <p>
    A production-quality Go library and CLI for rich, GitHub-style text diffs — right in your terminal.
  </p>
</div>

---

**drift** computes line-level diffs between two texts (or files) using Myers, Patience, or Histogram algorithms, and renders them with full Chroma syntax highlighting, dual line-number gutters, word-level intra-line highlights, and unified or side-by-side layout. In a color-capable TTY, output uses adaptive ANSI styles; when stdout is piped or `--no-color` / `NO_COLOR` is set, output is plain text.

## Contents

- [Installation](#installation)
- [CLI](#cli)
  - [Basic usage](#basic-usage)
  - [Flags](#flags)
  - [Syntax theme resolution](#syntax-theme-resolution)
- [Library](#library)
  - [Functional API](#functional-api)
  - [Builder API](#builder-api)
  - [Render options](#render-options)
- [Algorithms](#algorithms)
- [Requirements](#requirements)
- [License](#license)

---

## Installation

### CLI

```bash
go install github.com/tbcrawford/drift/cmd/drift@latest
```

### Library

```bash
go get github.com/tbcrawford/drift
```

---

## CLI

### Basic usage

```bash
# Compare two files
drift old.go new.go

# Side-by-side layout with histogram algorithm
drift --algorithm histogram --split file_a.txt file_b.txt

# Plain text output, explicit language hint
drift --no-color --lang go ./old.go ./new.go

# Diff working tree vs HEAD (inside a Git repo)
drift ./internal/foo.go
```

With a single path inside a Git worktree, `drift` compares the file on disk to the version at `HEAD` — equivalent to `git diff HEAD -- <path>`.

Exit code `1` when differences exist, `0` when files are identical.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--algorithm` | `myers` | Diff algorithm: `myers`, `patience`, `histogram` |
| `--split` | off | Side-by-side two-panel layout |
| `--theme` | auto | Chroma style name (e.g. `github`, `monokai`, `dracula`) |
| `--lang` | auto | Language hint for syntax highlighting (e.g. `go`, `python`) |
| `--no-color` | off | Disable ANSI color output |
| `--context` | `3` | Number of unchanged context lines shown around each hunk |
| `--no-line-numbers` | off | Hide old/new line number gutters |
| `--from` | — | Diff the given string instead of a file (use with `--to`) |
| `--to` | — | Diff the given string instead of a file (use with `--from`) |

Run `drift --help` for the full flag reference.

### Syntax theme resolution

When `--theme` is omitted on a Unix TTY with color output, `drift` queries the terminal palette via OSC 4 and selects the closest matching registered Chroma style automatically. Piped stdout, `NO_COLOR`, or non-Unix environments fall back to `monokai` (dark) or `github` (light).

The hidden flag `--show-theme` prints the resolved theme name to stderr:

```
drift: resolved syntax theme: github-dark
```

---

## Library

Import path:

```go
import "github.com/tbcrawford/drift"
```

### Functional API

```go
package main

import (
    "os"

    "github.com/tbcrawford/drift"
)

func main() {
    old := "package main\n\nfunc hello() {\n\tfmt.Println(\"hello\")\n}\n"
    new := "package main\n\nfunc hello() {\n\tfmt.Println(\"hello, world\")\n}\n"

    result, err := drift.Diff(old, new, drift.WithAlgorithm(drift.Myers))
    if err != nil {
        panic(err)
    }

    if err := drift.Render(result, os.Stdout, drift.WithLang("go")); err != nil {
        panic(err)
    }
}
```

### Builder API

The builder lets you configure once and reuse across multiple diffs:

```go
b := drift.New().
    Algorithm(drift.Histogram).
    Theme("dracula").
    Lang("go").
    Split()

result, err := b.Diff(old, new)
if err != nil {
    panic(err)
}

if err := b.Render(result, os.Stdout); err != nil {
    panic(err)
}
```

### Render options

Pass any combination of options to `drift.Render` (functional API) or configure them on the builder.

| Option | Builder method | Default | Description |
|--------|---------------|---------|-------------|
| `drift.WithAlgorithm(a)` | `.Algorithm(a)` | `Myers` | Diff algorithm |
| `drift.WithSplit()` | `.Split()` | off | Side-by-side layout |
| `drift.WithTheme(name)` | `.Theme(name)` | auto | Chroma style |
| `drift.WithLang(name)` | `.Lang(name)` | auto | Syntax language |
| `drift.WithNoColor()` | `.NoColor()` | off | Disable ANSI color |
| `drift.WithContext(n)` | `.Context(n)` | `3` | Context lines per hunk |
| `drift.WithLineNumbers(bool)` | `.LineNumbers(bool)` | `true` | Show line gutters |
| `drift.WithoutLineNumbers()` | — | — | Alias: hide gutters |
| `drift.WithWordDiff(bool)` | `.WordDiff(bool)` | `true` | Intra-line word highlights |
| `drift.WithLineDiffStyle(bool)` | `.LineDiffStyle(bool)` | `true` | Full-line add/delete bg |
| `drift.WithTermWidth(n)` | `.TermWidth(n)` | auto | Override terminal width |

**Examples:**

```go
// Plain unified diff — no color, no gutters
drift.Render(result, w,
    drift.WithNoColor(),
    drift.WithoutLineNumbers(),
)

// Side-by-side with a specific theme and language
drift.Render(result, w,
    drift.WithSplit(),
    drift.WithTheme("github-dark"),
    drift.WithLang("python"),
)

// Minimal output — disable word highlights and full-line bg
drift.Render(result, w,
    drift.WithWordDiff(false),
    drift.WithLineDiffStyle(false),
)
```

---

## Algorithms

| Algorithm | Flag value | Characteristics |
|-----------|-----------|-----------------|
| **Myers** | `myers` | O(ND) minimal edit distance. Fastest. Best for small, dense changes. |
| **Patience** | `patience` | Anchors on unique common lines. Better for refactors and moved blocks. |
| **Histogram** | `histogram` | Extends patience by frequency-bucketing rare lines. Git's default since 2011. Best overall for large files with repeated structure. |

---

## Requirements

- **Go 1.21+**
- No cgo required — pure Go, statically linkable

---

## License

MIT — see [LICENSE](LICENSE).
