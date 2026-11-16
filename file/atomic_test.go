package file

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func assertNoTemp(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), TempPrefix) {
			t.Errorf("temporary file %s left behind", e.Name())
		}
	}
}

func assertMode(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if runtime.GOOS == "windows" {
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != mode {
		t.Errorf("mode = %v, want %v", info.Mode().Perm(), mode)
	}
}

func TestAtomicFileCommit(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out")

	a, err := CreateAtomic(dir, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Write([]byte("data")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Errorf("destination exists before Commit, stat = %v", err)
	}
	if err := a.Commit(dest); err != nil {
		t.Fatalf("Commit() = %v, want nil", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "data" {
		t.Errorf("content = %q, want %q", data, "data")
	}
	assertMode(t, dest, 0o600)
	assertNoTemp(t, dir)
	if err := a.Abort(); err != nil {
		t.Errorf("Abort() after Commit = %v, want nil", err)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("Abort() after Commit removed the file, stat = %v", err)
	}
}

func TestAtomicFileAbort(t *testing.T) {
	dir := t.TempDir()

	a, err := CreateAtomic(dir, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Write([]byte("data")); err != nil {
		t.Fatal(err)
	}
	if err := a.Abort(); err != nil {
		t.Fatalf("Abort() = %v, want nil", err)
	}
	assertNoTemp(t, dir)
	if err := a.Commit(filepath.Join(dir, "out")); err != nil {
		t.Errorf("Commit() after Abort = %v, want nil", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "out")); !os.IsNotExist(err) {
		t.Errorf("Commit() after Abort created the file, stat = %v", err)
	}
}

func TestAtomicFileCommitFails(t *testing.T) {
	dir := t.TempDir()

	a, err := CreateAtomic(dir, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Commit(filepath.Join(dir, "missing", "out")); err == nil {
		t.Fatal("Commit() into a missing directory = nil, want an error")
	}
	assertNoTemp(t, dir)
}

func TestCreateAtomicMissingDir(t *testing.T) {
	if _, err := CreateAtomic(filepath.Join(t.TempDir(), "missing"), 0o600); err == nil {
		t.Error("CreateAtomic() in a missing directory = nil, want an error")
	}
}

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out")
	if err := os.WriteFile(dest, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := WriteFile(dest, []byte("new"), 0o600); err != nil {
		t.Fatalf("WriteFile() = %v, want nil", err)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Errorf("content = %q, want %q", data, "new")
	}
	assertMode(t, dest, 0o600)
	assertNoTemp(t, dir)
}

func TestWriteFileMissingDir(t *testing.T) {
	if err := WriteFile(filepath.Join(t.TempDir(), "missing", "out"), []byte("x"), 0o600); err == nil {
		t.Error("WriteFile() into a missing directory = nil, want an error")
	}
}

func TestWriteAtomic(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.bin")
	content := "hello world"

	digest, size, err := WriteAtomic(strings.NewReader(content), dest, DefaultFileMode)
	if err != nil {
		t.Fatal(err)
	}
	if size != int64(len(content)) {
		t.Errorf("size = %d, want %d", size, len(content))
	}
	h := sha256.Sum256([]byte(content))
	want := fmt.Sprintf("sha256:%s", hex.EncodeToString(h[:]))
	if digest != want {
		t.Errorf("digest = %s, want %s", digest, want)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("content = %q, want %q", data, content)
	}
	assertMode(t, dest, DefaultFileMode)
	assertNoTemp(t, dir)
}

func TestWriteAtomicReadError(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.bin")

	if _, _, err := WriteAtomic(errReader{}, dest, DefaultFileMode); err == nil {
		t.Fatal("WriteAtomic() with a failing reader = nil, want an error")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Errorf("destination exists after a failed write, stat = %v", err)
	}
	assertNoTemp(t, dir)
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, os.ErrClosed
}
