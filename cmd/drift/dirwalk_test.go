package main

import (
	"os"
	"path/filepath"
	"testing"
)

// --- diffDirectories gitignore filtering tests ---

func TestDiffDirectories_gitignore_skipsIgnoredInOld(t *testing.T) {
	// oldDir has "keep.go" (not ignored) and "dist/app" (ignored).
	// Only keep.go should appear in pairs (as removed, since newDir is empty).
	bin := t.TempDir()
	oldDir := t.TempDir()
	newDir := t.TempDir()
	oldAbs, _ := filepath.Abs(oldDir)

	// Create files in oldDir.
	if err := os.WriteFile(filepath.Join(oldDir, "keep.go"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	distDir := filepath.Join(oldDir, "dist")
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(distDir, "app"), []byte("artifact"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Fake git: oldDir is inside a repo; check-ignore marks dist/app as ignored.
	script := "#!/bin/sh\n" +
		"joined=\"$*\"\n" +
		"case \"$joined\" in\n" +
		"  *rev-parse*--is-inside-work-tree*) echo true; exit 0 ;;\n" +
		"  *rev-parse*--show-toplevel*) echo \"" + oldAbs + "\"; exit 0 ;;\n" +
		"  *check-ignore*) printf 'dist/app\\0'; exit 0 ;;\n" +
		"esac\n" +
		"echo \"fake git: $joined\" >&2; exit 99\n"
	writeFakeGit(t, bin, script)
	prependPath(t, bin)

	pairs, err := diffDirectories(oldDir, newDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range pairs {
		if p.Name == "dist/app" {
			t.Errorf("ignored file dist/app should not appear in pairs")
		}
	}
	found := false
	for _, p := range pairs {
		if p.Name == "keep.go" {
			found = true
		}
	}
	if !found {
		t.Errorf("keep.go should appear in pairs (removed); got: %+v", pairs)
	}
}

func TestDiffDirectories_gitignore_skipsIgnoredInNew(t *testing.T) {
	// newDir has "keep.go" and "dist/app" (ignored); oldDir is empty.
	// dist/app should not appear; keep.go should appear as added.
	bin := t.TempDir()
	oldDir := t.TempDir()
	newDir := t.TempDir()
	newAbs, _ := filepath.Abs(newDir)

	if err := os.WriteFile(filepath.Join(newDir, "keep.go"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	distDir := filepath.Join(newDir, "dist")
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(distDir, "app"), []byte("artifact"), 0o644); err != nil {
		t.Fatal(err)
	}

	// oldDir: not in git repo; newDir: in repo, check-ignore marks dist/app.
	script := "#!/bin/sh\n" +
		"joined=\"$*\"\n" +
		"# oldDir rev-parse returns false (not in repo), newDir returns true.\n" +
		"# We fake by matching the -C arg. Since both dirs are temp, just check\n" +
		"# whether arguments include the newDir absolute path.\n" +
		"case \"$joined\" in\n" +
		"  *rev-parse*--is-inside-work-tree*) echo true; exit 0 ;;\n" +
		"  *rev-parse*--show-toplevel*) echo \"" + newAbs + "\"; exit 0 ;;\n" +
		"  *check-ignore*) printf 'dist/app\\0'; exit 0 ;;\n" +
		"esac\n" +
		"echo \"fake git: $joined\" >&2; exit 99\n"
	writeFakeGit(t, bin, script)
	prependPath(t, bin)

	pairs, err := diffDirectories(oldDir, newDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range pairs {
		if p.Name == "dist/app" {
			t.Errorf("ignored file dist/app should not appear in pairs")
		}
	}
	found := false
	for _, p := range pairs {
		if p.Name == "keep.go" {
			found = true
		}
	}
	if !found {
		t.Errorf("keep.go should appear as added; got: %+v", pairs)
	}
}

func TestDiffDirectories_gitignore_noRepo_walksAll(t *testing.T) {
	// Dirs not in a git repo → all files included (fail-open).
	bin := t.TempDir()
	oldDir := t.TempDir()
	newDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(oldDir, "keep.go"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "keep.go"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Fake git returns false for is-inside-work-tree.
	script := "#!/bin/sh\n" +
		"joined=\"$*\"\n" +
		"case \"$joined\" in\n" +
		"  *rev-parse*--is-inside-work-tree*) echo false; exit 0 ;;\n" +
		"esac\n" +
		"echo \"fake git: $joined\" >&2; exit 99\n"
	writeFakeGit(t, bin, script)
	prependPath(t, bin)

	pairs, err := diffDirectories(oldDir, newDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 1 || pairs[0].Name != "keep.go" {
		t.Errorf("expected keep.go pair; got: %+v", pairs)
	}
}

func TestDiffDirectories_gitignore_gitNotFound_walksAll(t *testing.T) {
	// No git in PATH → fail open, walk all files.
	oldDir := t.TempDir()
	newDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(oldDir, "keep.go"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "keep.go"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Use an empty temp dir as PATH so git is not found.
	t.Setenv("PATH", t.TempDir())

	pairs, err := diffDirectories(oldDir, newDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 1 || pairs[0].Name != "keep.go" {
		t.Errorf("expected keep.go pair; got: %+v", pairs)
	}
}

// TestDiffDirectories covers all 8 behavior cases specified in the plan.
func TestDiffDirectories(t *testing.T) {
	t.Run("empty dirs returns empty slice", func(t *testing.T) {
		oldDir := t.TempDir()
		newDir := t.TempDir()

		pairs, err := diffDirectories(oldDir, newDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pairs) != 0 {
			t.Fatalf("expected 0 pairs, got %d: %v", len(pairs), pairs)
		}
	})

	t.Run("file in old only is removed", func(t *testing.T) {
		oldDir := t.TempDir()
		newDir := t.TempDir()

		if err := os.WriteFile(filepath.Join(oldDir, "only-old.txt"), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}

		pairs, err := diffDirectories(oldDir, newDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pairs) != 1 {
			t.Fatalf("expected 1 pair, got %d", len(pairs))
		}
		fp := pairs[0]
		if fp.Name != "only-old.txt" {
			t.Errorf("Name = %q, want %q", fp.Name, "only-old.txt")
		}
		if fp.OldPath == "" {
			t.Error("OldPath should be non-empty")
		}
		if fp.NewPath != "" {
			t.Errorf("NewPath = %q, want empty", fp.NewPath)
		}
		if !fp.IsRemoved() {
			t.Error("IsRemoved() = false, want true")
		}
		if fp.IsAdded() {
			t.Error("IsAdded() = true, want false")
		}
	})

	t.Run("file in new only is added", func(t *testing.T) {
		oldDir := t.TempDir()
		newDir := t.TempDir()

		if err := os.WriteFile(filepath.Join(newDir, "only-new.txt"), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}

		pairs, err := diffDirectories(oldDir, newDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pairs) != 1 {
			t.Fatalf("expected 1 pair, got %d", len(pairs))
		}
		fp := pairs[0]
		if fp.Name != "only-new.txt" {
			t.Errorf("Name = %q, want %q", fp.Name, "only-new.txt")
		}
		if fp.OldPath != "" {
			t.Errorf("OldPath = %q, want empty", fp.OldPath)
		}
		if fp.NewPath == "" {
			t.Error("NewPath should be non-empty")
		}
		if !fp.IsAdded() {
			t.Error("IsAdded() = false, want true")
		}
		if fp.IsRemoved() {
			t.Error("IsRemoved() = true, want false")
		}
	})

	t.Run("file with same content in both dirs is excluded", func(t *testing.T) {
		oldDir := t.TempDir()
		newDir := t.TempDir()

		content := []byte("identical content")
		if err := os.WriteFile(filepath.Join(oldDir, "same.txt"), content, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(newDir, "same.txt"), content, 0o644); err != nil {
			t.Fatal(err)
		}

		pairs, err := diffDirectories(oldDir, newDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pairs) != 0 {
			t.Fatalf("expected 0 pairs (identical file excluded), got %d: %v", len(pairs), pairs)
		}
	})

	t.Run("file with different content in both dirs is included", func(t *testing.T) {
		oldDir := t.TempDir()
		newDir := t.TempDir()

		if err := os.WriteFile(filepath.Join(oldDir, "changed.txt"), []byte("old content"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(newDir, "changed.txt"), []byte("new content"), 0o644); err != nil {
			t.Fatal(err)
		}

		pairs, err := diffDirectories(oldDir, newDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pairs) != 1 {
			t.Fatalf("expected 1 pair, got %d", len(pairs))
		}
		fp := pairs[0]
		if fp.Name != "changed.txt" {
			t.Errorf("Name = %q, want %q", fp.Name, "changed.txt")
		}
		if fp.OldPath == "" {
			t.Error("OldPath should be non-empty")
		}
		if fp.NewPath == "" {
			t.Error("NewPath should be non-empty")
		}
		if fp.IsAdded() {
			t.Error("IsAdded() = true, want false")
		}
		if fp.IsRemoved() {
			t.Error("IsRemoved() = true, want false")
		}
	})

	t.Run("results are sorted lexicographically by name", func(t *testing.T) {
		oldDir := t.TempDir()
		newDir := t.TempDir()

		// Create files in old only (so they all show up in results)
		for _, name := range []string{"z.txt", "a.txt", "m.txt"} {
			if err := os.WriteFile(filepath.Join(oldDir, name), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		pairs, err := diffDirectories(oldDir, newDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pairs) != 3 {
			t.Fatalf("expected 3 pairs, got %d", len(pairs))
		}
		want := []string{"a.txt", "m.txt", "z.txt"}
		for i, fp := range pairs {
			if fp.Name != want[i] {
				t.Errorf("pairs[%d].Name = %q, want %q", i, fp.Name, want[i])
			}
		}
	})

	t.Run("non-directory path returns error", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "file.txt")
		if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}

		// oldDir is a file — should error
		_, err := diffDirectories(file, dir)
		if err == nil {
			t.Fatal("expected error for file path as oldDir, got nil")
		}

		// newDir is a file — should error
		_, err = diffDirectories(dir, file)
		if err == nil {
			t.Fatal("expected error for file path as newDir, got nil")
		}
	})

	t.Run("nested subdirectory files produce relative name with forward slashes", func(t *testing.T) {
		oldDir := t.TempDir()
		newDir := t.TempDir()

		// Create a nested file in old only
		subDir := filepath.Join(oldDir, "sub")
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(subDir, "file.go"), []byte("pkg sub"), 0o644); err != nil {
			t.Fatal(err)
		}

		pairs, err := diffDirectories(oldDir, newDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pairs) != 1 {
			t.Fatalf("expected 1 pair, got %d", len(pairs))
		}
		fp := pairs[0]
		if fp.Name != "sub/file.go" {
			t.Errorf("Name = %q, want %q (forward slashes required)", fp.Name, "sub/file.go")
		}
	})
}
