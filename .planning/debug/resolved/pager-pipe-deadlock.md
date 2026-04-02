---
status: resolved
trigger: "drift --split deadlocks after quitting the pager with fatal goroutine deadlock in writeThroughPager"
created: 2026-04-02T00:00:00Z
updated: 2026-04-02T00:00:00Z
symptoms_prefilled: true
goal: find_and_fix
---

## Current Focus

hypothesis: When the user quits the pager early, the pager process exits and closes the read end of the io.Pipe. The next write to pw (the write end) blocks forever in io.(*pipe).write because Go's pipe.write uses a select with a done channel — it blocks waiting for a reader that will never come. buf.WriteTo(pagerWriter) in writeThroughPager is the blocking call. cleanup() is never reached because the write itself hangs.
test: Read pager.go and main.go startPager/writeThroughPager to confirm the exact blocking path
expecting: Confirmed that pw.Write blocks when pr is closed (no reader), and that cleanup() is never called because WriteTo never returns
next_action: DONE — fix committed

## Symptoms

expected: drift exits cleanly after the user quits the pager (e.g. pressing q in less)
actual: After quitting the pager, drift panics with "fatal error: all goroutines are asleep - deadlock!" The goroutine is stuck in io.(*pipe).write inside bytes.(*Buffer).WriteTo inside writeThroughPager.
errors: |
  fatal error: all goroutines are asleep - deadlock!

  goroutine 1 [select]:
  io.(*pipe).write(0x619b6c63c120, {0x619b6dc80000, 0x15e79f, 0x1a0000})
          /opt/homebrew/Cellar/go/1.26.1/libexec/src/io/pipe.go:86 +0x198
  io.(*PipeWriter).Write(0x105493b00?, {0x619b6dc80000?, 0x619b6c770001?, 0x619b6c770000?})
          /opt/homebrew/Cellar/go/1.26.1/libexec/src/io/pipe.go:161 +0x24
  bytes.(*Buffer).WriteTo(0x619b6c37e210, {0x105515bc0?, 0x619b6c63c120?})
          /opt/homebrew/Cellar/go/1.26.1/libexec/src/bytes/buffer.go:279 +0x70
  main.writeThroughPager(0x619b6c37e210, 0x619b6c3d06c0)
          /Users/tylercrawford/dev/github/drift/cmd/drift/main.go:253 +0x1b0
  main.runRoot(0x619b6c3d06c0)
          /Users/tylercrawford/dev/github/drift/cmd/drift/main.go:296 +0x2a0
reproduction: Run `drift --split` from inside a git repo with changed files (or `drift <directory> --split`). Scroll through the pager output, then quit (press q). The deadlock occurs intermittently — more likely with large repos/many changed files.
started: Likely introduced when io.Pipe-based pager was introduced.

## Eliminated

(none yet)

## Evidence

- timestamp: 2026-04-02T00:00:00Z
  checked: cmd/drift/pager.go startPager (lines 71-91)
  found: |
    io.Pipe() creates pr (PipeReader) and pw (PipeWriter).
    cmd.Stdin = pr — pager process reads from pr.
    cleanup() closes pw, then pr, then cmd.Wait().
    The cleanup goroutine is NOT started automatically — it runs only after WriteTo returns.
  implication: If the pager process exits early (user presses q), the OS closes the read end of the pipe from the pager's perspective. However, pr is still open (Go side). Go's io.Pipe is an in-process pipe, not an OS pipe — so when the pager subprocess exits, it does NOT close pr on the Go side. The write to pw will block because pr still has a reader goroutine... wait, actually there is NO goroutine copying pr→something in startPager. The pipe is pr→cmd.Stdin directly via OS. When less exits, the OS fd for the read end closes. But pw.Write() goes through Go's in-process io.Pipe, not an OS pipe. So pw.Write blocks in select waiting for a reader — but pr's OS side is gone.

- timestamp: 2026-04-02T00:00:00Z
  checked: Go stdlib io.Pipe behavior when pager exits
  found: |
    io.Pipe() is a pure in-memory synchronous pipe. When cmd.Stdin = pr, exec.Cmd copies from pr to the process stdin via an OS pipe internally. When the subprocess (less) exits, exec.Cmd's internal goroutine stops reading from pr. At that point pr.Close() would be needed to unblock pw.Write(). But cleanup() (which calls pw.Close() and pr.Close()) is never reached because buf.WriteTo(pagerWriter) never returns — it's blocked on pw.Write().
  implication: Classic deadlock: WriteTo blocks waiting for a reader on the pipe; cleanup() that would close pw is after WriteTo; pr is never closed from the reader side because exec.Cmd's goroutine exited when pager quit. Both sides are stuck.

- timestamp: 2026-04-02T00:00:00Z
  checked: writeThroughPager in main.go (lines 240-260)
  found: |
    pagerWriter, cleanup, pErr := startPager(resolvePager(), opts.streams)
    // ...
    _, _ = buf.WriteTo(pagerWriter)  // ← BLOCKS if pager exits early
    cleanup()                         // ← never reached
  implication: The fix must either (a) run WriteTo in a goroutine and use a done channel to cancel it when pager exits, (b) close pw from a goroutine that monitors cmd.Wait() so the write unblocks with io.ErrClosedPipe, or (c) change startPager to launch a goroutine that monitors the pager process and proactively closes pw when it exits.

## Resolution

root_cause: |
  startPager() returns a WriteCloser (pw) and a cleanup() func. writeThroughPager calls buf.WriteTo(pw) synchronously, then calls cleanup(). When the user quits the pager (less exits), exec.Cmd's internal goroutine stops reading from the io.Pipe's read end (pr), but pr is never closed. pw.Write() then blocks forever in Go's in-process pipe select, because the read end still appears open (pr.Close() hasn't been called). cleanup() — which would call pw.Close() and pr.Close() — is unreachable because it's sequenced after WriteTo. Result: single-goroutine deadlock.

fix: |
  In startPager, launch a goroutine that calls cmd.Wait() and then closes pr. This causes pw.Write() to receive io.ErrClosedPipe (or io.EOF) when the pager exits, unblocking WriteTo. In writeThroughPager, treat io.ErrClosedPipe / syscall.EPIPE errors from WriteTo as a normal "user quit" condition (not an error). The cleanup func should no longer need to call cmd.Wait() since the goroutine handles it, but it should still close pw and drain the wait.

verification: |
  - go build ./cmd/drift/ succeeds
  - All 292 tests pass (up from 290; 2 new tests added)
  - TestPagerStartEarlyExit specifically exercises early-exit path using head -1 as pager;
    it times out in 5s if deadlock regresses
  - TestIsPipeClosedErr verifies the helper correctly identifies OS broken-pipe errors
files_changed: [cmd/drift/pager.go, cmd/drift/main.go, cmd/drift/pager_test.go]
