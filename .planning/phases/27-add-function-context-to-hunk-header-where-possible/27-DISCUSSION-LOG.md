# Phase 27: add function context to hunk header where possible — Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-10
**Phase:** 27-add-function-context-to-hunk-header-where-possible
**Areas discussed:** Regex source, Git pager passthrough, No-match fallback, Activation, Display format

---

## Pre-discussion Research: What Does Delta Do?

Before presenting gray areas, the delta codebase was investigated to answer the user's
specific question: "what does delta use/do here?"

**Finding:** Delta does NOT generate function context. Delta is a formatter/pager for
git diff output. The `@@ -74,15 +75,14 @@ pub fn delta(` text is produced by **git
itself** (via `userdiff.c`). Delta's `hunk_header.rs` extracts the `code_fragment`
(text after the trailing `@@`) using regex `r"@+ ([^@]+)@+(.*\s?)"` and renders it
styled. No language detection, no regex scanning by delta.

Git's `userdiff.c` has a table of per-language regex patterns. This is what actually
produces function context — delta just passes it through.

---

## Regex Source

| Option | Description | Selected |
|--------|-------------|----------|
| Port git's userdiff.c patterns | Battle-tested, matches what users see from `git diff` | ✓ (initially) |
| Write our own patterns | More flexibility, likely to diverge | |
| Library-assisted detection | go-enry + custom regex | |

**User's initial choice:** Port git's userdiff.c patterns

**Revision:** User then changed direction entirely — see "Pivot" section below.

---

## Git Pager Passthrough

| Option | Description | Selected |
|--------|-------------|----------|
| Standalone only — git pager already has it | Phase 27 scoped to standalone drift.Diff() | ✓ (initially) |
| Also handle passthrough in pager mode | Parse code_fragment from incoming @@ lines | |

**User's initial choice:** Standalone only

**Revision:** See "Pivot" section below.

---

## No-Match Fallback

| Option | Description | Selected |
|--------|-------------|----------|
| Empty — just @@ numbers @@ | Matches git's behavior exactly | ✓ |
| Best-effort — nearest block | Try struct/class even if not a function | |
| First non-blank line of hunk | Last-resort fallback | |

**User's choice:** Empty — just `@@ -x,y +a,b @@` with no trailing text.
**Notes:** Matches git's exact behavior.

---

## Activation

| Option | Description | Selected |
|--------|-------------|----------|
| Auto — on when lang is known | Activate only when WithLang() set or detected | ✓ (initially) |
| Always on — opt out | Always on, WithFuncContext(false) to disable | |
| Opt-in — off by default | Must call WithFuncContext(true) | |

**User's initial choice:** Auto — on when lang is known

**Revision:** See "Pivot" section below — activation model changed entirely.

---

## Pivot: The Key Direction Change

After answering the regex-source follow-up, the user provided this clarification:

> "Let's not port userdiff actually. Let's just provide function context when a diff
> is coming from git and default to the existing hunk header when not git. I'd like to
> have something somewhat for free in some contexts and just fallback to the default in
> other contexts."

This changed the entire approach:
- **Before pivot:** Standalone drift.Diff() mode with funcctx package + regex patterns
- **After pivot:** Git pager mode only — parse code_fragment that git already computed

Confirmed with follow-up question: "Git pager only — display what git gives us"

---

## Display Format

| Option | Description | Selected |
|--------|-------------|----------|
| Exactly as git formats it | `@@ -x,y +a,b @@ func_name` (styled) | |
| Extract and style separately | `{line_number}: {func_name}` format | ✓ |
| Agent's discretion | Whatever fits existing rendering | |

**User's choice:** Extract and style separately.
**Notes:** "Something nice like `111: func_name`. We can tweak the exact style as we go."

---

## Agent's Discretion

- Exact styling of `{line_number}: {func_name}` (color, weight, separator).
- Whether to use a new struct field or pass code_fragment inline.
- Trimming heuristics for code_fragment (strip trailing punctuation or keep as-is).

## Deferred Ideas

- Standalone regex-based function context (funcctx package, port git's userdiff.c)
- `WithFuncContext(bool)` option for standalone drift.Diff() control
