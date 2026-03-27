# Architecture

This document defines the architectural vision for `terrasort`. All contributors
and agents must follow these guidelines. The goal is not merely a working CLI — it
is a reference-quality Go CLI that is readable, testable, extensible, and
maintainable by anyone familiar with idiomatic Go.

---

## Guiding principles

**Single responsibility.** Every function, type, and package does one thing. If a
function's purpose cannot be described in a single sentence, it needs to be
decomposed. Cognitive complexity should be low enough that intent is apparent from
name and signature alone.

**No global state.** Nothing meaningful lives at the package level. No `init()`
registers commands, binds flags, or mutates shared state. Everything is
constructed, injected, and passed explicitly.

**Interfaces at boundaries.** Anything with more than one implementation — or that
needs to be replaced in tests — is expressed as an interface. Depend on the
interface, not the concrete type. Keep interfaces small.

**Testability by construction.** Every meaningful function must be unit-testable
without spawning a subprocess or touching the real filesystem. If a function cannot
be tested without mocking `os.Stdout`, its design is wrong. I/O, config, and
terminal access are always injected.

**Explicit over implicit.** No magic. No reflection-based wiring. Dependencies are
passed as parameters. The data flow through any command execution must be traceable
from `main()` to the leaf function by reading the call graph.

**Errors carry context.** Always wrap errors with enough context that the call site
is visible in the message. Never swallow errors. Never fail silently or fall back to
a default when an explicit value was expected.

---

## Structure

**Strict dependency direction.** `cmd/` → `internal/cli/` → `internal/` packages.
Business logic packages (`sorter`, `diff`, `highlight`, etc.) know nothing about
cobra, `IOStreams`, or config. They are pure functions and small types with explicit
inputs and outputs — the easiest things in the codebase to test and reason about.

**CLI wiring is thin.** Cobra commands parse arguments and resolve options. They do
not contain business logic. The business logic layer in `internal/` does the actual
work.

**All output flows through `IOStreams`.** No code below `main.go` writes to
`os.Stdout` or reads from `os.Stdin` directly. An `IOStreams` struct is constructed
at startup and injected throughout. This is what makes every command testable with
a buffer swap.

**Commands follow a Flags → Options → run() lifecycle.** A `Flags` struct holds raw
parsed values. A `toOptions()` (or `resolve*Options`) method resolves them against
config into a fully populated `Options` struct — all decisions made before execution
begins. A `run()` function orchestrates the work by calling into the business logic
layer. A `validate()` function checks invariants with no side effects.

**Config is lazy-loaded.** The `AppContext` carries config as a loader function, not
a pre-loaded value. Commands that never need config pay no I/O cost.

---

## Package dependency map

```
cmd/terrasort/
  └── internal/cli/
        ├── internal/appcontext/    (AppContext: IOStreams + lazy Config)
        ├── internal/config/        (YAML config loading and merging)
        ├── internal/sorter/        (HCL sorting engine — pure)
        ├── internal/diff/          (diff rendering — pure output logic)
        ├── internal/highlight/     (syntax highlighting — pure)
        ├── internal/glob/          (file discovery — pure)
        ├── internal/walker/        (filesystem traversal — pure)
        ├── internal/iostreams/     (I/O abstraction)
        └── internal/theme/         (interactive theme picker)
```

Business logic packages must not import `internal/cli/` or `internal/appcontext/`.
The import graph is a DAG with no cycles.

---

## Command lifecycle (canonical example)

```go
// 1. Flags — cobra-parsed raw values
type SomeFlags struct {
    Output string
}

// 2. Options — all decisions resolved before execution
type someOptions struct {
    output  string
    streams *iostreams.IOStreams
    cfg     *config.Config
}

// 3. Resolve: Flags + AppContext → Options (all I/O decisions made here)
func resolveSomeOptions(ctx *appcontext.AppContext, flags *SomeFlags, args []string) (*someOptions, error) {
    cfg, err := ctx.Config()
    if err != nil {
        return nil, fmt.Errorf("loading config: %w", err)
    }
    return &someOptions{
        output:  flags.Output,
        streams: ctx.Streams,
        cfg:     cfg,
    }, nil
}

// 4. Validate: pure invariant check, no side effects
func validateSomeOptions(opts *someOptions) error {
    if opts.output == "" {
        return fmt.Errorf("output is required")
    }
    return nil
}

// 5. Run: orchestrates business logic — thin, readable, no parsing
func runSome(opts *someOptions) error {
    if err := validateSomeOptions(opts); err != nil {
        return err
    }
    result, err := somelibrary.DoWork(opts.input)
    if err != nil {
        return fmt.Errorf("doing work: %w", err)
    }
    fmt.Fprintln(opts.streams.Out, result)
    return nil
}
```

`RunE` is always a two-liner:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    opts, err := resolveSomeOptions(ctx, flags, args)
    if err != nil {
        return err
    }
    return runSome(opts)
},
```

---

## What to avoid

- `init()` for anything other than blank imports
- Package-level variables holding mutable state
- Calling `os.Stdout`, `os.Stdin`, or `os.Exit` below `main.go`
- Business logic inside cobra `RunE` functions
- Importing `internal/cli/` from `internal/` packages
- Functions that parse, resolve, execute, and render all in one body
- Errors returned without wrapping or context
- Silent fallbacks on invalid or missing input
