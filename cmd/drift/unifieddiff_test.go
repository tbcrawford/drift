package main

import (
	"strings"
	"testing"
)

// TestParseUnifiedDiff verifies parseUnifiedDiff handles all git diff format variants.
func TestParseUnifiedDiff(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		files, err := parseUnifiedDiff(strings.NewReader(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 0 {
			t.Fatalf("expected 0 files for empty input, got %d", len(files))
		}
	})

	t.Run("git metadata only no diff lines", func(t *testing.T) {
		input := `diff --git a/go.sum b/go.sum
index abc..def 100644
`
		files, err := parseUnifiedDiff(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// No --- / +++ lines → no reconstructed content; entry may be produced
		// but with empty contents. What matters: no panic. We just check no error.
		_ = files
	})

	t.Run("single file one hunk", func(t *testing.T) {
		input := `diff --git a/main.go b/main.go
index abc..def 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,3 @@
 package main
-func old() {}
+func new() {}
 // end
`
		files, err := parseUnifiedDiff(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		f := files[0]
		if f.Name != "main.go" {
			t.Errorf("expected Name=main.go, got %q", f.Name)
		}
		if !strings.Contains(f.OldContent, "func old() {}") {
			t.Errorf("OldContent should contain 'func old() {}', got: %q", f.OldContent)
		}
		if strings.Contains(f.OldContent, "func new() {}") {
			t.Errorf("OldContent should not contain 'func new() {}', got: %q", f.OldContent)
		}
		if !strings.Contains(f.NewContent, "func new() {}") {
			t.Errorf("NewContent should contain 'func new() {}', got: %q", f.NewContent)
		}
		if strings.Contains(f.NewContent, "func old() {}") {
			t.Errorf("NewContent should not contain 'func old() {}', got: %q", f.NewContent)
		}
		// Context lines appear in both
		if !strings.Contains(f.OldContent, "package main") {
			t.Errorf("OldContent should contain context line 'package main', got: %q", f.OldContent)
		}
		if !strings.Contains(f.NewContent, "package main") {
			t.Errorf("NewContent should contain context line 'package main', got: %q", f.NewContent)
		}
	})

	t.Run("multi-file diff three files", func(t *testing.T) {
		input := `diff --git a/a.go b/a.go
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
diff --git a/c.go b/c.go
--- a/c.go
+++ b/c.go
@@ -1 +1 @@
-old c
+new c
`
		files, err := parseUnifiedDiff(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 3 {
			t.Fatalf("expected 3 files, got %d", len(files))
		}
		for i, name := range []string{"a.go", "b.go", "c.go"} {
			if files[i].Name != name {
				t.Errorf("files[%d].Name = %q, want %q", i, files[i].Name, name)
			}
		}
		// Verify order and contents
		if !strings.Contains(files[0].OldContent, "old a") {
			t.Errorf("files[0].OldContent missing 'old a': %q", files[0].OldContent)
		}
		if !strings.Contains(files[2].NewContent, "new c") {
			t.Errorf("files[2].NewContent missing 'new c': %q", files[2].NewContent)
		}
	})

	t.Run("newly added file", func(t *testing.T) {
		input := `diff --git a/new.go b/new.go
new file mode 100644
--- /dev/null
+++ b/new.go
@@ -0,0 +1,2 @@
+package main
+// added
`
		files, err := parseUnifiedDiff(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		f := files[0]
		if f.OldContent != "" {
			t.Errorf("new file: OldContent should be empty, got %q", f.OldContent)
		}
		if !strings.Contains(f.NewContent, "package main") {
			t.Errorf("new file: NewContent missing 'package main', got %q", f.NewContent)
		}
		if !strings.Contains(f.NewContent, "// added") {
			t.Errorf("new file: NewContent missing '// added', got %q", f.NewContent)
		}
	})

	t.Run("deleted file", func(t *testing.T) {
		input := `diff --git a/old.go b/old.go
deleted file mode 100644
--- a/old.go
+++ /dev/null
@@ -1,2 +0,0 @@
-package main
-// removed
`
		files, err := parseUnifiedDiff(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		f := files[0]
		if f.NewContent != "" {
			t.Errorf("deleted file: NewContent should be empty, got %q", f.NewContent)
		}
		if !strings.Contains(f.OldContent, "package main") {
			t.Errorf("deleted file: OldContent missing 'package main', got %q", f.OldContent)
		}
		if !strings.Contains(f.OldContent, "// removed") {
			t.Errorf("deleted file: OldContent missing '// removed', got %q", f.OldContent)
		}
	})

	t.Run("binary file", func(t *testing.T) {
		input := `diff --git a/image.png b/image.png
index abc..def 100644
Binary files a/image.png and b/image.png differ
`
		files, err := parseUnifiedDiff(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file for binary, got %d", len(files))
		}
		f := files[0]
		if !f.IsBinary {
			t.Errorf("expected IsBinary=true for binary file diff")
		}
		if f.OldContent != "" {
			t.Errorf("binary file: OldContent should be empty, got %q", f.OldContent)
		}
		if f.NewContent != "" {
			t.Errorf("binary file: NewContent should be empty, got %q", f.NewContent)
		}
	})

	t.Run("renamed file uses new name", func(t *testing.T) {
		input := `diff --git a/old_name.go b/new_name.go
similarity index 80%
rename from old_name.go
rename to new_name.go
--- a/old_name.go
+++ b/new_name.go
@@ -1,2 +1,2 @@
 package main
-// old
+// new
`
		files, err := parseUnifiedDiff(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		f := files[0]
		// Name derived from NewName (strip b/ prefix from b/new_name.go)
		if f.Name != "new_name.go" {
			t.Errorf("expected Name=new_name.go, got %q", f.Name)
		}
		if f.OldName != "a/old_name.go" {
			t.Errorf("expected OldName=a/old_name.go, got %q", f.OldName)
		}
		if f.NewName != "b/new_name.go" {
			t.Errorf("expected NewName=b/new_name.go, got %q", f.NewName)
		}
	})

	t.Run("multiple hunks in one file", func(t *testing.T) {
		input := `diff --git a/multi.go b/multi.go
--- a/multi.go
+++ b/multi.go
@@ -1,3 +1,3 @@
 line1
-oldA
+newA
 line3
@@ -10,3 +10,3 @@
 line10
-oldB
+newB
 line12
`
		files, err := parseUnifiedDiff(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file for multiple hunks, got %d", len(files))
		}
		f := files[0]
		if !strings.Contains(f.OldContent, "oldA") {
			t.Errorf("OldContent missing 'oldA': %q", f.OldContent)
		}
		if !strings.Contains(f.OldContent, "oldB") {
			t.Errorf("OldContent missing 'oldB': %q", f.OldContent)
		}
		if !strings.Contains(f.NewContent, "newA") {
			t.Errorf("NewContent missing 'newA': %q", f.NewContent)
		}
		if !strings.Contains(f.NewContent, "newB") {
			t.Errorf("NewContent missing 'newB': %q", f.NewContent)
		}
	})

	// Regression test: git sends ANSI-colored output when stdout is a TTY (i.e.
	// when drift runs as core.pager). Before the fix, strings.HasPrefix checks
	// like `line == "diff --git ..."` never matched because every line was
	// prefixed with escape sequences, causing parseUnifiedDiff to return 0 files.
	t.Run("ANSI-colored git output", func(t *testing.T) {
		// Mimic what `git diff --color=always` produces: escape sequences wrap
		// every significant line, content lines are plain.
		input := "\x1b[1mdiff --git a/main.go b/main.go\x1b[0m\n" +
			"\x1b[1mindex abc..def 100644\x1b[0m\n" +
			"\x1b[1m--- a/main.go\x1b[0m\n" +
			"\x1b[1m+++ b/main.go\x1b[0m\n" +
			"\x1b[36m@@ -1,3 +1,3 @@\x1b[0m\n" +
			" package main\n" +
			"\x1b[31m-func old() {}\x1b[0m\n" +
			"\x1b[32m+func new() {}\x1b[0m\n" +
			" // end\n"
		files, err := parseUnifiedDiff(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d (ANSI escape sequences may have broken prefix matching)", len(files))
		}
		f := files[0]
		if f.Name != "main.go" {
			t.Errorf("expected Name=main.go, got %q", f.Name)
		}
		if !strings.Contains(f.OldContent, "func old() {}") {
			t.Errorf("OldContent should contain 'func old() {}', got: %q", f.OldContent)
		}
		if !strings.Contains(f.NewContent, "func new() {}") {
			t.Errorf("NewContent should contain 'func new() {}', got: %q", f.NewContent)
		}
	})

	t.Run("no newline at end of file marker skipped", func(t *testing.T) {
		input := `diff --git a/noeol.txt b/noeol.txt
--- a/noeol.txt
+++ b/noeol.txt
@@ -1 +1 @@
-old
\ No newline at end of file
+new
\ No newline at end of file
`
		files, err := parseUnifiedDiff(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		f := files[0]
		// The backslash line should not appear in content
		if strings.Contains(f.OldContent, "No newline") {
			t.Errorf("OldContent should not contain 'No newline' marker: %q", f.OldContent)
		}
		if strings.Contains(f.NewContent, "No newline") {
			t.Errorf("NewContent should not contain 'No newline' marker: %q", f.NewContent)
		}
	})
}

// TestParseUnifiedDiff_CodeFragment verifies that parseUnifiedDiff correctly
// extracts the code_fragment from git @@ hunk header lines.
func TestParseUnifiedDiff_CodeFragment(t *testing.T) {
	t.Run("single hunk with code_fragment", func(t *testing.T) {
		input := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -12,7 +12,9 @@ func ParseOptions(args []string)
 context
-old line
+new line
 context
`
		files, err := parseUnifiedDiff(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		f := files[0]
		if len(f.Hunks) != 1 {
			t.Fatalf("expected 1 hunk, got %d", len(f.Hunks))
		}
		if f.Hunks[0].CodeFragment != "func ParseOptions(args []string)" {
			t.Errorf("CodeFragment = %q, want %q", f.Hunks[0].CodeFragment, "func ParseOptions(args []string)")
		}
		if f.Hunks[0].NewStart != 12 {
			t.Errorf("NewStart = %d, want 12", f.Hunks[0].NewStart)
		}
	})

	t.Run("hunk without code_fragment", func(t *testing.T) {
		input := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,3 @@
 context
-old
+new
 context
`
		files, err := parseUnifiedDiff(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 || len(files[0].Hunks) != 1 {
			t.Fatalf("expected 1 file with 1 hunk")
		}
		if files[0].Hunks[0].CodeFragment != "" {
			t.Errorf("expected empty CodeFragment, got %q", files[0].Hunks[0].CodeFragment)
		}
	})

	t.Run("multi-hunk with different code_fragments", func(t *testing.T) {
		input := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,3 @@ func Alpha()
 ctx
-old a
+new a
 ctx
@@ -20,3 +20,3 @@ func Beta()
 ctx
-old b
+new b
 ctx
`
		files, err := parseUnifiedDiff(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		f := files[0]
		if len(f.Hunks) != 2 {
			t.Fatalf("expected 2 hunks, got %d", len(f.Hunks))
		}
		if f.Hunks[0].CodeFragment != "func Alpha()" {
			t.Errorf("Hunks[0].CodeFragment = %q, want %q", f.Hunks[0].CodeFragment, "func Alpha()")
		}
		if f.Hunks[1].CodeFragment != "func Beta()" {
			t.Errorf("Hunks[1].CodeFragment = %q, want %q", f.Hunks[1].CodeFragment, "func Beta()")
		}
	})

	t.Run("ANSI-colored @@ line with code_fragment", func(t *testing.T) {
		// ansi.Strip cleans ANSI before the switch; verify code_fragment still extracted.
		input := "diff --git a/main.go b/main.go\n" +
			"--- a/main.go\n" +
			"+++ b/main.go\n" +
			"\x1b[36m@@ -5,3 +5,3 @@ func Foo()\x1b[0m\n" +
			" ctx\n" +
			"-old\n" +
			"+new\n"
		files, err := parseUnifiedDiff(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 || len(files[0].Hunks) != 1 {
			t.Fatalf("expected 1 file with 1 hunk")
		}
		if files[0].Hunks[0].CodeFragment != "func Foo()" {
			t.Errorf("CodeFragment = %q, want %q", files[0].Hunks[0].CodeFragment, "func Foo()")
		}
	})

	t.Run("whitespace-only code_fragment treated as empty", func(t *testing.T) {
		input := "diff --git a/main.go b/main.go\n" +
			"--- a/main.go\n" +
			"+++ b/main.go\n" +
			"@@ -1,3 +1,3 @@  \n" + // trailing spaces after @@
			" ctx\n" +
			"-old\n" +
			"+new\n"
		files, err := parseUnifiedDiff(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 || len(files[0].Hunks) != 1 {
			t.Fatalf("expected 1 file with 1 hunk")
		}
		if files[0].Hunks[0].CodeFragment != "" {
			t.Errorf("expected empty CodeFragment for whitespace-only, got %q", files[0].Hunks[0].CodeFragment)
		}
	})
}
