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

