package main

import (
	"os"
	"path/filepath"
	"testing"
)

// --- diffDirectories gitignore filtering tests ---

func TestDiffDirectories_gitignore_skipsIgnoredInOld(t *testing.T) {
	// oldDir is a real git repo with .gitignore containing "dist/".
	// keep.go is committed; dist/app is ignored.
	// Only keep.go should appear in pairs (as removed, since newDir is empty).
	oldDir := makeTestRepo(t, map[string]string{
		"keep.go":    "content",
		".gitignore": "dist/\n",
	})
	newDir := t.TempDir()

	// Write dist/app to oldDir disk (not committed — ignored).
	distDir := filepath.Join(oldDir, "dist")
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(distDir, "app"), []byte("artifact"), 0o644); err != nil {
		t.Fatal(err)
	}

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
	// newDir is a real git repo with .gitignore containing "dist/".
	// keep.go is committed; dist/app is ignored.
	// oldDir is an empty temp dir (no repo).
	oldDir := t.TempDir()
	newDir := makeTestRepo(t, map[string]string{
		"keep.go":    "content",
		".gitignore": "dist/\n",
	})

	// Write dist/app to newDir disk (not committed — ignored).
	distDir := filepath.Join(newDir, "dist")
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(distDir, "app"), []byte("artifact"), 0o644); err != nil {
		t.Fatal(err)
	}

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
	// Neither dir is a git repo → all files included (fail-open).
	oldDir := t.TempDir()
	newDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(oldDir, "keep.go"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "keep.go"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	pairs, err := diffDirectories(oldDir, newDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 1 || pairs[0].Name != "keep.go" {
		t.Errorf("expected keep.go pair; got: %+v", pairs)
	}
}

func TestDiffDirectories_gitignore_gitNotFound_walksAll(t *testing.T) {
	// Plain temp dirs (no git repo) → fail-open, all files included.
	oldDir := t.TempDir()
	newDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(oldDir, "keep.go"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "keep.go"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

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
