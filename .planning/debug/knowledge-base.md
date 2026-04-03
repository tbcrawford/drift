# GSD Debug Knowledge Base

Resolved debug sessions. Used by `gsd-debugger` to surface known-pattern hypotheses at the start of new investigations.

---

## pager-pipe-deadlock — io.Pipe write blocks forever when user quits pager early
- **Date:** 2026-04-02
- **Error patterns:** deadlock, goroutine, pipe, write, pager, writeThroughPager, io.(*pipe).write, ErrClosedPipe, EPIPE, less, startPager
- **Root cause:** startPager() returned the pipe write end (pw) to the caller but never closed the read end (pr) when the pager subprocess exited early. When the user quits less/more, exec.Cmd's internal goroutine stops draining pr but pr remains open. The next pw.Write() blocks forever in Go's in-process pipe select. cleanup() — which would close pw and pr — was sequenced after buf.WriteTo(pw) and was unreachable.
- **Fix:** (1) startPager launches a background goroutine that calls cmd.Wait() then closes pr, unblocking any in-flight pw.Write() with io.ErrClosedPipe. cleanup() closes pw and waits on a done channel. (2) writeThroughPager captures the WriteTo error and uses isPipeClosedErr() to treat io.ErrClosedPipe and syscall.EPIPE as clean user-quit exits.
- **Files changed:** cmd/drift/pager.go, cmd/drift/main.go, cmd/drift/pager_test.go
---

## git-pager-no-output — git diff shows nothing when drift is configured as core.pager
- **Date:** 2026-04-03
- **Error patterns:** pager, core.pager, git diff, no output, silent exit, parseUnifiedDiff, files=0, GIT_PAGER_IN_USE, ANSI, escape sequences, diff --git
- **Root cause:** Two compounding issues: (1) `streamThroughPager` unconditionally launched a nested `less` even when drift was already running as git's pager, creating a `git→drift→less` chain where the inner less inherited `LESS=FRX` (-F = quit-if-one-screen) and exited silently. (2) Even after fixing (1), `parseUnifiedDiff` returned 0 files because git sends ANSI-colored output to its pager when stdout is a TTY — every line is prefixed with escape sequences like `\x1b[1m`, so `strings.HasPrefix(line, "diff --git ")` never matched.
- **Fix:** (1) In `streamThroughPager` and `writeThroughPager`, check `os.LookupEnv("GIT_PAGER_IN_USE")` — if set (git always sets it, even to empty string), skip launching a nested pager and write directly to stdout. (2) In `parseUnifiedDiff`, call `ansi.Strip(scanner.Text())` (from the already-imported `github.com/charmbracelet/x/ansi` dep) before all prefix checks, stripping escape sequences before parsing.
- **Files changed:** cmd/drift/unifieddiff.go, cmd/drift/main.go
---

