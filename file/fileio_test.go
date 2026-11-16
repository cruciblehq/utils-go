package file

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathExists(t *testing.T) {
	dir := t.TempDir()

	exists, err := PathExists(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("expected existing directory to return true")
	}

	exists, err = PathExists(filepath.Join(dir, "nope"))
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("expected non-existent path to return false")
	}
}

func TestRemoveDirIfEmpty(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "empty")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := RemoveDirIfEmpty(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("expected empty directory to be removed")
	}
}

func TestRemoveDirIfEmptyNonEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveDirIfEmpty(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(dir); err != nil {
		t.Fatal("expected non-empty directory to remain")
	}
}

func TestRemoveDirIfEmptyMissing(t *testing.T) {
	err := RemoveDirIfEmpty(filepath.Join(t.TempDir(), "gone"))
	if err != nil {
		t.Fatalf("expected nil for missing directory, got %v", err)
	}
}

func TestRemoveEmptyParents(t *testing.T) {
	root := t.TempDir()
	deep := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(deep, 0755); err != nil {
		t.Fatal(err)
	}

	if err := RemoveEmptyParents(deep, root); err != nil {
		t.Fatalf("RemoveEmptyParents() = %v, want nil", err)
	}
	if _, err := os.Stat(filepath.Join(root, "a")); !os.IsNotExist(err) {
		t.Errorf("empty parent still exists, stat = %v", err)
	}
	if _, err := os.Stat(root); err != nil {
		t.Errorf("root was removed, stat = %v", err)
	}
}

func TestRemoveEmptyParentsStopsAtNonEmpty(t *testing.T) {
	root := t.TempDir()
	deep := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(deep, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a", "keep"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveEmptyParents(deep, root); err != nil {
		t.Fatalf("RemoveEmptyParents() = %v, want nil", err)
	}
	if _, err := os.Stat(deep); !os.IsNotExist(err) {
		t.Errorf("empty directory still exists, stat = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "a")); err != nil {
		t.Errorf("non-empty parent was removed, stat = %v", err)
	}
}

func TestRemoveEmptyParentsMissing(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "a"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := RemoveEmptyParents(filepath.Join(root, "a", "gone"), root); err != nil {
		t.Fatalf("RemoveEmptyParents() = %v, want nil", err)
	}
	if _, err := os.Stat(filepath.Join(root, "a")); !os.IsNotExist(err) {
		t.Errorf("empty parent of a missing directory still exists, stat = %v", err)
	}
}

func TestRemoveEmptyParentsOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	if err := RemoveEmptyParents(outside, root); err != nil {
		t.Fatalf("RemoveEmptyParents() = %v, want nil", err)
	}
	if _, err := os.Stat(outside); err != nil {
		t.Errorf("directory outside root was removed, stat = %v", err)
	}
}

func TestIsInside(t *testing.T) {
	root := filepath.Join("/", "data", "root")
	tests := []struct {
		path string
		want bool
	}{
		{filepath.Join(root, "a"), true},
		{filepath.Join(root, "a", "b"), true},
		{filepath.Join(root, ".hidden"), true},
		{root, false},
		{filepath.Join("/", "data"), false},
		{filepath.Join("/", "data", "rootx"), false},
		{filepath.Join(root, "..", "other"), false},
	}
	for _, tt := range tests {
		if got := IsInside(tt.path, root); got != tt.want {
			t.Errorf("IsInside(%q, %q) = %v, want %v", tt.path, root, got, tt.want)
		}
	}
}

func TestListSubdirs(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"alpha", "beta"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	// Create a regular file that should be excluded.
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	names, err := ListSubdirs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 subdirs, got %d: %v", len(names), names)
	}
}

func TestListSubdirsEmpty(t *testing.T) {
	names, err := ListSubdirs(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Fatalf("expected 0 subdirs, got %d", len(names))
	}
}

func TestListSubdirsMissing(t *testing.T) {
	_, err := ListSubdirs(filepath.Join(t.TempDir(), "nope"))
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}
