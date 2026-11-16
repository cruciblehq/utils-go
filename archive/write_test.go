package archive

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndExtract(t *testing.T) {
	for _, f := range []Format{Tar, Gzip, Zstd} {
		t.Run(f.String(), func(t *testing.T) {
			srcDir := t.TempDir()
			createTestFiles(t, srcDir)

			destDir := filepath.Join(t.TempDir(), "extracted")
			if err := Extract(writeArchive(t, srcDir, f), destDir, f); err != nil {
				t.Fatalf("Extract() = %v, want nil", err)
			}

			assertFileContent(t, filepath.Join(destDir, "file.txt"), "hello")
			assertFileContent(t, filepath.Join(destDir, "subdir", "nested.txt"), "nested")
			assertDirExists(t, filepath.Join(destDir, "emptydir"))
		})
	}
}

func TestWriteNames(t *testing.T) {
	srcDir := t.TempDir()
	createTestFiles(t, srcDir)

	var buf bytes.Buffer
	if err := Write(&buf, srcDir, "", Tar); err != nil {
		t.Fatalf("Write() = %v, want nil", err)
	}

	headers := readHeaders(t, buf.Bytes())
	for _, name := range []string{"file.txt", "subdir/", "subdir/nested.txt", "emptydir/"} {
		if headers[name] == nil {
			t.Errorf("entry %q missing, have %v", name, headers)
		}
	}
	if len(headers) != 4 {
		t.Errorf("len(entries) = %d, want 4", len(headers))
	}
}

func TestWritePrefix(t *testing.T) {
	srcDir := t.TempDir()
	createTestFiles(t, srcDir)

	var buf bytes.Buffer
	if err := Write(&buf, srcDir, "/opt/app/", Tar); err != nil {
		t.Fatalf("Write() = %v, want nil", err)
	}

	headers := readHeaders(t, buf.Bytes())
	for _, name := range []string{"opt/app/", "opt/app/file.txt", "opt/app/subdir/nested.txt"} {
		if headers[name] == nil {
			t.Errorf("entry %q missing, have %v", name, headers)
		}
	}
}

func TestWriteSingleFile(t *testing.T) {
	src := filepath.Join(t.TempDir(), "app.bin")
	if err := os.WriteFile(src, []byte("bin"), 0755); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := Write(&buf, src, "usr/bin", Tar); err != nil {
		t.Fatalf("Write() = %v, want nil", err)
	}

	headers := readHeaders(t, buf.Bytes())
	hdr := headers["usr/bin/app.bin"]
	if hdr == nil {
		t.Fatalf("entry usr/bin/app.bin missing, have %v", headers)
	}
	if hdr.Mode != 0755 {
		t.Errorf("mode = %o, want %o", hdr.Mode, 0755)
	}
	if len(headers) != 1 {
		t.Errorf("len(entries) = %d, want 1", len(headers))
	}
}

func TestWriteStripsSpecialBits(t *testing.T) {
	src := filepath.Join(t.TempDir(), "tool")
	if err := os.WriteFile(src, []byte("x"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(src, 0755|os.ModeSetuid); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := Write(&buf, src, "", Tar); err != nil {
		t.Fatalf("Write() = %v, want nil", err)
	}

	if mode := readHeaders(t, buf.Bytes())["tool"].Mode; mode != 0755 {
		t.Errorf("mode = %o, want %o", mode, 0755)
	}
}

func TestWriteSymlinkError(t *testing.T) {
	srcDir := t.TempDir()
	target := filepath.Join(srcDir, "target.txt")
	if err := os.WriteFile(target, []byte("target"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(srcDir, "link")); err != nil {
		t.Fatal(err)
	}

	err := Write(io.Discard, srcDir, "", Zstd)
	if !errors.Is(err, ErrUnsupportedFileType) {
		t.Fatalf("Write() = %v, want %v", err, ErrUnsupportedFileType)
	}
	if !errors.Is(err, ErrWriteFailed) {
		t.Fatalf("Write() = %v, want %v", err, ErrWriteFailed)
	}
}

func TestWriteEmptyDirectory(t *testing.T) {
	destDir := filepath.Join(t.TempDir(), "extracted")
	if err := Extract(writeArchive(t, t.TempDir(), Zstd), destDir, Zstd); err != nil {
		t.Fatalf("Extract() = %v, want nil", err)
	}

	assertDirExists(t, destDir)
}

func TestWriteNonExistentSource(t *testing.T) {
	err := Write(io.Discard, filepath.Join(t.TempDir(), "missing"), "", Zstd)
	if !errors.Is(err, ErrWriteFailed) {
		t.Fatalf("Write() = %v, want %v", err, ErrWriteFailed)
	}
}

func TestWriteUnsupportedFormat(t *testing.T) {
	err := Write(io.Discard, t.TempDir(), "", Format(99))
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("Write() = %v, want %v", err, ErrUnsupportedFormat)
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func TestWriteReportsFlushError(t *testing.T) {
	srcDir := t.TempDir()
	createTestFiles(t, srcDir)

	if err := Write(failWriter{}, srcDir, "", Zstd); !errors.Is(err, ErrWriteFailed) {
		t.Fatalf("Write() = %v, want %v", err, ErrWriteFailed)
	}
}

func TestCleanName(t *testing.T) {
	tests := map[string]string{
		"":          "",
		"/":         "",
		".":         "",
		"./a/b":     "a/b",
		"/opt/app/": "opt/app",
		"a/../b":    "b",
		"../a":      "a",
	}
	for in, want := range tests {
		if got := cleanName(in); got != want {
			t.Errorf("cleanName(%q) = %q, want %q", in, got, want)
		}
	}
}
