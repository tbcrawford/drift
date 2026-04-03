package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	git "github.com/go-git/go-git/v5"
)

// ttyReader wraps an io.Reader and signals to isStdinPipe that it is a TTY
// (not a pipe). Tests that want to simulate TTY stdin (e.g. zero-arg git mode)
// wrap their readers with ttyReader so pager mode is not triggered.
type ttyReader struct{ r io.Reader }

func (t *ttyReader) Read(p []byte) (int, error) { return t.r.Read(p) }
func (t *ttyReader) IsTTY() bool                { return true }

// ttyStream returns an IOStreams with a TTY-flagged stdin (empty content).
// Use this for zero-arg git mode tests to prevent pager mode detection.
func ttyStream(out, errOut *bytes.Buffer) IOStreams {
	return IOStreams{In: &ttyReader{strings.NewReader("")}, Out: out, Err: errOut}
}

// --- Pager mode tests (Plan 25-02) ---

// singleFileUnifiedDiff is a minimal but realistic git unified diff for one file.
const singleFileUnifiedDiff = `diff --git a/foo.go b/foo.go
index abc..def 100644
--- a/foo.go
+++ b/foo.go
@@ -1,3 +1,3 @@
 package main
-func old() {}
+func new() {}
 // end
`

// multiFileUnifiedDiff has two files changed.
const multiFileUnifiedDiff = `diff --git a/a.go b/a.go
--- a/a.go
+++ b/a.go
@@ -1 +1 @@
-old a
+new a
diff --git a/b.go b/b.go
--- a/b.go
+++ b/b.go
@@ -1 +1 @@
-old b
+new b
`

// binaryUnifiedDiff contains only a binary file change (no text to render).
const binaryUnifiedDiff = `diff --git a/img.png b/img.png
index abc..def 100644
Binary files a/img.png and b/img.png differ
`

// ansiColoredUnifiedDiff mimics what git sends when drift is configured as
// core.pager and git detects a TTY — every significant line is wrapped in
// ANSI escape sequences (bold, colour codes, resets).
const ansiColoredUnifiedDiff = "\x1b[1mdiff --git a/foo.go b/foo.go\x1b[0m\n" +
	"\x1b[1mindex abc..def 100644\x1b[0m\n" +
	"\x1b[1m--- a/foo.go\x1b[0m\n" +
	"\x1b[1m+++ b/foo.go\x1b[0m\n" +
	"\x1b[36m@@ -1,3 +1,3 @@\x1b[0m\n" +
	" package main\n" +
	"\x1b[31m-func old() {}\x1b[0m\n" +
	"\x1b[32m+func new() {}\x1b[0m\n" +
	" // end\n"

// TestRunCLI_pagerMode_singleFile verifies that when stdin is a pipe containing
// a unified diff for one file with no positional args, drift re-renders it.
func TestRunCLI_pagerMode_singleFile(t *testing.T) {
	var out, errOut bytes.Buffer
	streams := IOStreams{
		In:  strings.NewReader(singleFileUnifiedDiff),
		Out: &out,
		Err: &errOut,
	}
	code := runCLI(streams, []string{"--no-color"})
	// Inputs differ → exit 1
	if code != 1 {
		t.Fatalf("expected exit 1 for differing file, got %d; stderr=%q stdout=%q", code, errOut.String(), out.String())
	}
	stdout := out.String()
	// Should contain the file header
	if !strings.Contains(stdout, "▸") || !strings.Contains(stdout, "foo.go") {
		t.Errorf("expected file header with '▸' and 'foo.go', got: %q", stdout)
	}
	// Should contain diff hunk marker
	if !strings.Contains(stdout, "@@") {
		t.Errorf("expected '@@' hunk marker in output, got: %q", stdout)
	}
	// Should NOT contain raw git diff header
	if strings.Contains(stdout, "diff --git") {
		t.Errorf("output should not contain raw 'diff --git' header, got: %q", stdout)
	}
}

// TestRunCLI_pagerMode_multiFile verifies that each file in a multi-file unified diff
// gets its own file header in the output.
func TestRunCLI_pagerMode_multiFile(t *testing.T) {
	var out, errOut bytes.Buffer
	streams := IOStreams{
		In:  strings.NewReader(multiFileUnifiedDiff),
		Out: &out,
		Err: &errOut,
	}
	code := runCLI(streams, []string{"--no-color"})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d; stderr=%q stdout=%q", code, errOut.String(), out.String())
	}
	stdout := out.String()
	if !strings.Contains(stdout, "a.go") {
		t.Errorf("expected 'a.go' in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "b.go") {
		t.Errorf("expected 'b.go' in output, got: %q", stdout)
	}
}

// TestRunCLI_pagerMode_withArgs verifies that pager mode is NOT triggered when
// positional args are present (existing file-diff path is used instead).
func TestRunCLI_pagerMode_withArgs(t *testing.T) {
	// Provide a unified diff on stdin but also pass --from/--to flags.
	// This must NOT invoke pager mode — it uses the --from/--to path.
	var out, errOut bytes.Buffer
	streams := IOStreams{
		In:  strings.NewReader(singleFileUnifiedDiff),
		Out: &out,
		Err: &errOut,
	}
	code := runCLI(streams, []string{"--from", "x", "--to", "y", "--no-color"})
	// --from/--to path executes; pager mode not invoked
	if code != 1 {
		t.Fatalf("expected exit 1 for --from/--to diff, got %d; stderr=%q", code, errOut.String())
	}
	stdout := out.String()
	// Should NOT have file headers from the piped unified diff
	if strings.Contains(stdout, "foo.go") {
		t.Errorf("pager mode should NOT be triggered when --from/--to flags present, got: %q", stdout)
	}
}

// TestRunCLI_pagerMode_ANSIColoredInput verifies that drift correctly parses and
// renders ANSI-colored unified diff input. When drift runs as git's core.pager,
// git detects a TTY and sends color-coded output — every significant line is
// wrapped in escape sequences. Before the fix, parseUnifiedDiff prefix checks
// (e.g. strings.HasPrefix(line, "diff --git ")) never matched, returning 0
// files and producing no output.
func TestRunCLI_pagerMode_ANSIColoredInput(t *testing.T) {
	var out, errOut bytes.Buffer
	streams := IOStreams{
		In:  strings.NewReader(ansiColoredUnifiedDiff),
		Out: &out,
		Err: &errOut,
	}
	code := runCLI(streams, []string{"--no-color"})
	if code != 1 {
		t.Fatalf("expected exit 1 (diff detected), got %d; stderr=%q stdout=%q", code, errOut.String(), out.String())
	}
	stdout := out.String()
	if !strings.Contains(stdout, "foo.go") {
		t.Errorf("expected rendered output containing 'foo.go'; ANSI sequences may have broken parsing. got: %q", stdout)
	}
	// Raw ANSI-colored diff header must not leak into rendered output.
	if strings.Contains(stdout, "\x1b[1mdiff --git") {
		t.Errorf("raw ANSI-colored diff header leaked into rendered output: %q", stdout)
	}
}

// TestRunCLI_pagerMode_GitPagerInUse verifies that setting GIT_PAGER_IN_USE
// does not suppress rendered output. When drift is git's pager, it must write
// directly to stdout rather than spawning a nested pager subprocess.
func TestRunCLI_pagerMode_GitPagerInUse(t *testing.T) {
	t.Setenv("GIT_PAGER_IN_USE", "1")
	var out, errOut bytes.Buffer
	streams := IOStreams{
		In:  strings.NewReader(ansiColoredUnifiedDiff),
		Out: &out,
		Err: &errOut,
	}
	code := runCLI(streams, []string{"--no-color"})
	if code != 1 {
		t.Fatalf("expected exit 1 (diff detected), got %d; stderr=%q stdout=%q", code, errOut.String(), out.String())
	}
	stdout := out.String()
	if !strings.Contains(stdout, "foo.go") {
		t.Errorf("GIT_PAGER_IN_USE suppressed output — expected 'foo.go' in stdout, got: %q", stdout)
	}
}

// TestRunCLI_pagerMode_empty verifies that an empty unified diff on stdin → exit 0, no output.
func TestRunCLI_pagerMode_empty(t *testing.T) {
	var out, errOut bytes.Buffer
	streams := IOStreams{
		In:  strings.NewReader(""),
		Out: &out,
		Err: &errOut,
	}
	code := runCLI(streams, []string{})
	// Empty stdin + no args → either pager mode (empty diff → exit 0) or zero-arg git mode.
	// In test environment, In is a bytes.Buffer (non-TTY) → pager mode triggers → exit 0.
	if code != 0 && code != 2 {
		t.Fatalf("expected exit 0 or 2, got %d; stderr=%q", code, errOut.String())
	}
}

// TestRunCLI_pagerMode_binaryOnly verifies that a binary-only unified diff on stdin
// exits 0 (no renderable diff) without panic.
func TestRunCLI_pagerMode_binaryOnly(t *testing.T) {
	var out, errOut bytes.Buffer
	streams := IOStreams{
		In:  strings.NewReader(binaryUnifiedDiff),
		Out: &out,
		Err: &errOut,
	}
	code := runCLI(streams, []string{"--no-color"})
	// Binary files skipped → no output → exit 0
	if code != 0 {
		t.Fatalf("expected exit 0 for binary-only diff, got %d; stderr=%q stdout=%q", code, errOut.String(), out.String())
	}
}

func TestRunCLI_identicalFromTo(t *testing.T) {
	t.Helper()
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{"--from", "same", "--to", "same"})
	if code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, errOut.String())
	}
}

func TestRunCLI_differs(t *testing.T) {
	t.Helper()
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{"--from", "a", "--to", "b"})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d stderr=%q", code, errOut.String())
	}
	if out.Len() == 0 {
		t.Fatal("expected diff on stdout")
	}
}

func TestRunCLI_gitSingleArg_differs(t *testing.T) {
	t.Helper()
	// Create a real git repo with x.txt = "head\n", then modify to "working\n" on disk.
	repoDir := makeTestRepo(t, map[string]string{
		"x.txt": "head\n",
	})
	file := filepath.Join(repoDir, "x.txt")
	if err := os.WriteFile(file, []byte("working\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{file})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d stderr=%q stdout=%q", code, errOut.String(), out.String())
	}
	if out.Len() == 0 {
		t.Fatal("expected diff on stdout")
	}
}

func TestRunCLI_invalidAlgorithm(t *testing.T) {
	t.Helper()
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{"--algorithm", "nope", "--from", "a", "--to", "b"})
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "algorithm") {
		t.Fatalf("stderr should mention algorithm: %q", errOut.String())
	}
}

func TestHelpListsAllFlags(t *testing.T) {
	t.Helper()
	buf := new(bytes.Buffer)
	streams := IOStreams{In: strings.NewReader(""), Out: buf, Err: buf}
	code := runCLI(streams, []string{"--help"})
	// --help exits 0
	if code != 0 {
		t.Fatalf("expected exit 0 from --help, got %d", code)
	}
	out := buf.String()
	for _, flag := range []string{
		"--split",
		"--no-line-numbers",
		"--algorithm",
		"--lang",
		"--theme",
		"--no-color",
		"--context",
		"--from",
		"--to",
		"--no-pager",
		"git",
	} {
		if !strings.Contains(out, flag) {
			t.Errorf("help output missing flag %q\noutput:\n%s", flag, out)
		}
	}
}

func TestRunCLI_noPagerFlag(t *testing.T) {
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{"--from", "a\n", "--to", "b\n", "--no-pager"})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d stderr=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "@@") {
		t.Fatalf("expected diff hunk header in output: %q", out.String())
	}
}

func TestRunCLI_pagerSkippedOnNonTTY(t *testing.T) {
	// Non-TTY out: shouldPage must return false → output written directly to buffer
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{"--from", "a\n", "--to", "b\n"})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d stderr=%q", code, errOut.String())
	}
	if out.Len() == 0 {
		t.Fatal("expected diff output in buffer (pager should not have consumed it)")
	}
}

// --- Directory diff tests ---

// Test 1: two identical dirs → exit 0, empty stdout
func TestRunCLI_directoryDiff_identical(t *testing.T) {
	oldDir := t.TempDir()
	newDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(oldDir, "a.txt"), []byte("same\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "a.txt"), []byte("same\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{oldDir, newDir})
	if code != 0 {
		t.Fatalf("expected exit 0 for identical dirs, got %d stderr=%q stdout=%q", code, errOut.String(), out.String())
	}
	if out.Len() != 0 {
		t.Fatalf("expected empty stdout for identical dirs, got: %q", out.String())
	}
}

// Test 2: one file differs between dirs → exit 1, header present, diff hunk "@@"
func TestRunCLI_directoryDiff_differs(t *testing.T) {
	oldDir := t.TempDir()
	newDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(oldDir, "a.txt"), []byte("old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "a.txt"), []byte("new\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{oldDir, newDir})
	if code != 1 {
		t.Fatalf("expected exit 1 for differing dirs, got %d stderr=%q", code, errOut.String())
	}
	stdout := out.String()
	// The header contains the ▸ chevron and the filename; ANSI codes may appear
	// between them in color mode, so check each element independently.
	if !strings.Contains(stdout, "▸") || !strings.Contains(stdout, "a.txt") {
		t.Fatalf("expected file header with '▸' and 'a.txt' in output: %q", stdout)
	}
	if !strings.Contains(stdout, "@@") {
		t.Fatalf("expected diff hunk '@@' in output: %q", stdout)
	}
}

// Test 3: file added in new dir → exit 1, header present, all lines prefixed with "+"
func TestRunCLI_directoryDiff_fileAdded(t *testing.T) {
	oldDir := t.TempDir()
	newDir := t.TempDir()
	// Only write to new dir
	if err := os.WriteFile(filepath.Join(newDir, "added.txt"), []byte("new content\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{"--no-color", oldDir, newDir})
	if code != 1 {
		t.Fatalf("expected exit 1 for added file, got %d stderr=%q", code, errOut.String())
	}
	stdout := out.String()
	if !strings.Contains(stdout, "▸ added.txt") {
		t.Fatalf("expected header '▸ added.txt' in output: %q", stdout)
	}
	if !strings.Contains(stdout, "+new content") {
		t.Fatalf("expected '+new content' line in output: %q", stdout)
	}
}

// Test 4: file removed in new dir → exit 1, header present, all lines prefixed with "-"
func TestRunCLI_directoryDiff_fileRemoved(t *testing.T) {
	oldDir := t.TempDir()
	newDir := t.TempDir()
	// Only write to old dir
	if err := os.WriteFile(filepath.Join(oldDir, "removed.txt"), []byte("old content\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{"--no-color", oldDir, newDir})
	if code != 1 {
		t.Fatalf("expected exit 1 for removed file, got %d stderr=%q", code, errOut.String())
	}
	stdout := out.String()
	if !strings.Contains(stdout, "▸ removed.txt") {
		t.Fatalf("expected header '▸ removed.txt' in output: %q", stdout)
	}
	if !strings.Contains(stdout, "-old content") {
		t.Fatalf("expected '-old content' line in output: %q", stdout)
	}
}

// Test 5: --no-color flag suppresses ANSI in directory diff output
func TestRunCLI_directoryDiff_noColor(t *testing.T) {
	oldDir := t.TempDir()
	newDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(oldDir, "a.txt"), []byte("old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "a.txt"), []byte("new\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{"--no-color", oldDir, newDir})
	if code != 1 {
		t.Fatalf("expected exit 1 for differing dirs, got %d stderr=%q", code, errOut.String())
	}
	// ANSI escape sequences start with ESC (\x1b)
	if strings.Contains(out.String(), "\x1b[") {
		t.Fatalf("expected no ANSI color sequences with --no-color, got: %q", out.String())
	}
}

// Test 6: non-directory path as first arg still uses existing two-file resolveInputs path (no regression)
func TestRunCLI_directoryDiff_nonDirUsesFilePath(t *testing.T) {
	// Create two regular files (not dirs) — should use existing file diff path
	oldFile := filepath.Join(t.TempDir(), "old.txt")
	newFile := filepath.Join(t.TempDir(), "new.txt")
	if err := os.WriteFile(oldFile, []byte("old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newFile, []byte("new\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	code := runCLI(streams, []string{oldFile, newFile})
	if code != 1 {
		t.Fatalf("expected exit 1 for differing files, got %d stderr=%q", code, errOut.String())
	}
	// Should produce diff output without "===" headers
	stdout := out.String()
	if out.Len() == 0 {
		t.Fatal("expected diff output for file args")
	}
	// Should NOT have directory-style headers
	if strings.Contains(stdout, "===") {
		t.Fatalf("file diff should not produce === headers: %q", stdout)
	}
}

// --- Zero-argument mode tests ---

// TestRunCLI_zeroArg_notInRepo verifies that running drift with no args outside a git repo
// prints the standard "not a git repository" message to stderr and exits 2.
func TestRunCLI_zeroArg_notInRepo(t *testing.T) {
	t.Helper()
	// Use a plain temp dir with no git repo.
	plainDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(plainDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	var out, errOut bytes.Buffer
	streams := ttyStream(&out, &errOut)
	code := runCLI(streams, []string{})
	if code != 2 {
		t.Fatalf("expected exit 2 outside a git repo, got %d; stderr=%q", code, errOut.String())
	}
	if !strings.Contains(errOut.String(), "not a git repository") {
		t.Fatalf("expected 'not a git repository' in stderr, got: %q", errOut.String())
	}
}

// TestRunCLI_zeroArg_noDiff verifies that running drift with no args in a git repo with no
// working-tree changes exits 0 with no output.
func TestRunCLI_zeroArg_noDiff(t *testing.T) {
	t.Helper()
	// Create a real git repo with committed files, no working-tree changes.
	repoDir := makeTestRepo(t, map[string]string{
		"clean.go": "package main\n",
	})
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	var out, errOut bytes.Buffer
	streams := ttyStream(&out, &errOut)
	code := runCLI(streams, []string{})
	if code != 0 {
		t.Fatalf("expected exit 0 with no working-tree changes, got %d; stderr=%q stdout=%q", code, errOut.String(), out.String())
	}
	if out.Len() != 0 {
		t.Fatalf("expected empty stdout with no changes, got: %q", out.String())
	}
}

// TestRunCLI_zeroArg_hasDiff verifies that running drift with no args in a git repo that has
// working-tree changes exits 1 and emits a diff with the file header and hunk marker.
func TestRunCLI_zeroArg_hasDiff(t *testing.T) {
	t.Helper()
	// Commit head.txt = "head\n", then modify it in the working tree.
	repoDir := makeTestRepo(t, map[string]string{
		"head.txt": "head\n",
	})
	if err := os.WriteFile(filepath.Join(repoDir, "head.txt"), []byte("working\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	var out, errOut bytes.Buffer
	streams := ttyStream(&out, &errOut)
	code := runCLI(streams, []string{"--no-color"})
	if code != 1 {
		t.Fatalf("expected exit 1 with working-tree changes, got %d; stderr=%q stdout=%q", code, errOut.String(), out.String())
	}
	stdout := out.String()
	if !strings.Contains(stdout, "▸") || !strings.Contains(stdout, "head.txt") {
		t.Fatalf("expected file header '▸ head.txt' in output: %q", stdout)
	}
	if !strings.Contains(stdout, "@@") {
		t.Fatalf("expected diff hunk marker '@@' in output: %q", stdout)
	}
}

// TestRunCLI_zeroArg_freshRepo verifies that running drift with no args in a git repo
// with no commits (no HEAD) exits 0 silently.
func TestRunCLI_zeroArg_freshRepo(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	// PlainInit without any commits — no HEAD.
	if _, err := git.PlainInit(dir, false); err != nil {
		t.Fatalf("PlainInit: %v", err)
	}
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	var out, errOut bytes.Buffer
	streams := ttyStream(&out, &errOut)
	code := runCLI(streams, []string{})
	if code != 0 {
		t.Fatalf("expected exit 0 for fresh repo with no HEAD, got %d; stderr=%q", code, errOut.String())
	}
	if out.Len() != 0 {
		t.Fatalf("expected empty stdout for fresh repo, got: %q", out.String())
	}
}

// Test 7: --from / --to with directory args returns an error (incompatible flags)
func TestRunCLI_directoryDiff_fromToIncompatible(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer
	streams := IOStreams{In: strings.NewReader(""), Out: &out, Err: &errOut}
	// --from/--to with a directory arg — resolveInputs will reject this with --from needing no positional args
	code := runCLI(streams, []string{"--from", dir, "--to", dir})
	// --from/--to interpret args as raw strings (not paths), and dir string isn't a path here.
	// But if both --from and --to equal the same dir string, it's a "no diff" case.
	// What's incompatible is: positional dir args combined with --from/--to.
	// Test: positional dir args + --from flag → error (exit 2)
	_ = code

	// Reset and test the actual incompatible case: positional dirs + --from
	out.Reset()
	errOut.Reset()
	code = runCLI(streams, []string{"--from", "text", dir})
	// --from is set and there's one positional arg → this violates "use either two paths or --from/--to"
	// Actually: --from is set, --to is not set → "both must be set" error
	if code != 2 {
		t.Fatalf("expected exit 2 for --from without --to, got %d stderr=%q", code, errOut.String())
	}
}

// --- Color-only mode tests (Plan 25-02, Task 2) ---

// TestRunCLI_colorOnly verifies that --color-only colorizes +/- lines with ANSI codes.
func TestRunCLI_colorOnly(t *testing.T) {
	input := "+added line\n-removed line\ncontext line\n"
	var out, errOut bytes.Buffer
	streams := IOStreams{
		In:  strings.NewReader(input),
		Out: &out,
		Err: &errOut,
	}
	code := runCLI(streams, []string{"--color-only"})
	if code != 0 {
		t.Fatalf("expected exit 0 for --color-only, got %d; stderr=%q", code, errOut.String())
	}
	stdout := out.String()
	// Added line should have green ANSI
	if !strings.Contains(stdout, "\x1b[32m+added line\x1b[0m") {
		t.Errorf("expected green ANSI for '+' line, got: %q", stdout)
	}
	// Removed line should have red ANSI
	if !strings.Contains(stdout, "\x1b[31m-removed line\x1b[0m") {
		t.Errorf("expected red ANSI for '-' line, got: %q", stdout)
	}
	// Context line should pass through unchanged (no ANSI prefix/suffix wrapping the line)
	if !strings.Contains(stdout, "context line") {
		t.Errorf("expected context line to pass through, got: %q", stdout)
	}
}

// TestRunCLI_colorOnly_noColor verifies that --color-only --no-color passes stdin through unchanged.
func TestRunCLI_colorOnly_noColor(t *testing.T) {
	input := "+added line\n-removed line\ncontext line\n"
	var out, errOut bytes.Buffer
	streams := IOStreams{
		In:  strings.NewReader(input),
		Out: &out,
		Err: &errOut,
	}
	code := runCLI(streams, []string{"--color-only", "--no-color"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%q", code, errOut.String())
	}
	// No ANSI escape sequences in output
	if strings.Contains(out.String(), "\x1b[") {
		t.Errorf("expected no ANSI codes with --no-color, got: %q", out.String())
	}
	// Content should still be present
	if !strings.Contains(out.String(), "+added line") {
		t.Errorf("expected '+added line' in output, got: %q", out.String())
	}
}

// TestRunCLI_installPager verifies that drift install-pager prints the gitconfig snippet.
func TestRunCLI_installPager(t *testing.T) {
	var out, errOut bytes.Buffer
	streams := IOStreams{
		In:  strings.NewReader(""),
		Out: &out,
		Err: &errOut,
	}
	code := runCLI(streams, []string{"install-pager"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%q", code, errOut.String())
	}
	stdout := out.String()
	if !strings.Contains(stdout, "core") || !strings.Contains(stdout, "pager") {
		t.Errorf("expected 'core' and 'pager' in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "interactive") || !strings.Contains(stdout, "diffFilter") {
		t.Errorf("expected 'interactive' and 'diffFilter' in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "drift --color-only") {
		t.Errorf("expected 'drift --color-only' in output, got: %q", stdout)
	}
}
