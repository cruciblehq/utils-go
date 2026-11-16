package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestExtract(t *testing.T) {
	srcDir := t.TempDir()
	createTestFiles(t, srcDir)
	parent := filepath.Join(t.TempDir(), "deep")
	destDir := filepath.Join(parent, "extracted")

	if err := Extract(writeArchive(t, srcDir, Zstd), destDir, Zstd); err != nil {
		t.Fatalf("Extract() = %v, want nil", err)
	}

	assertFileContent(t, filepath.Join(destDir, "file.txt"), "hello")
	assertFileContent(t, filepath.Join(destDir, "subdir", "nested.txt"), "nested")
	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("parent has %d entries, want 1", len(entries))
	}
	info, err := os.Stat(destDir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("dest mode = %o, want %o", info.Mode().Perm(), 0755)
	}
}

func TestExtractInvalidData(t *testing.T) {
	parent := t.TempDir()
	destDir := filepath.Join(parent, "extracted")

	if err := Extract(bytes.NewReader([]byte("not a valid archive")), destDir, Zstd); !errors.Is(err, ErrExtractFailed) {
		t.Fatalf("Extract() = %v, want %v", err, ErrExtractFailed)
	}
	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("parent has %d entries after a failed extraction, want 0", len(entries))
	}
}

func TestExtractDestinationExists(t *testing.T) {
	srcDir := t.TempDir()
	createTestFiles(t, srcDir)
	destDir := t.TempDir()

	err := Extract(writeArchive(t, srcDir, Zstd), destDir, Zstd)
	if !errors.Is(err, os.ErrExist) {
		t.Fatalf("Extract() = %v, want %v", err, os.ErrExist)
	}
}

func TestExtractMaliciousEntryLeavesNothing(t *testing.T) {
	data := tarBytes(t,
		entry{&tar.Header{Name: "good.txt", Mode: 0644, Typeflag: tar.TypeReg}, "hello"},
		entry{&tar.Header{Name: "evil", Typeflag: tar.TypeSymlink, Linkname: "/etc/passwd"}, ""},
	)
	destDir := filepath.Join(t.TempDir(), "extracted")

	if err := Extract(bytes.NewReader(data), destDir, Tar); err == nil {
		t.Fatal("Extract() = nil, want error")
	}
	if _, err := os.Stat(destDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stat dest = %v, want %v", err, os.ErrNotExist)
	}
}

func TestExtractInto(t *testing.T) {
	srcDir := t.TempDir()
	createTestFiles(t, srcDir)
	destDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(destDir, "keep.txt"), []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := ExtractInto(writeArchive(t, srcDir, Gzip), destDir, Gzip); err != nil {
		t.Fatalf("ExtractInto() = %v, want nil", err)
	}

	assertFileContent(t, filepath.Join(destDir, "keep.txt"), "keep")
	assertFileContent(t, filepath.Join(destDir, "file.txt"), "hello")
}

func TestExtractIntoCreatesDirectory(t *testing.T) {
	srcDir := t.TempDir()
	createTestFiles(t, srcDir)
	destDir := filepath.Join(t.TempDir(), "new")

	if err := ExtractInto(writeArchive(t, srcDir, Tar), destDir, Tar); err != nil {
		t.Fatalf("ExtractInto() = %v, want nil", err)
	}

	assertFileContent(t, filepath.Join(destDir, "file.txt"), "hello")
}

func TestExtractIntoKeepsDirectoryOnFailure(t *testing.T) {
	destDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(destDir, "keep.txt"), []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := ExtractInto(bytes.NewReader([]byte("not a valid archive")), destDir, Zstd); !errors.Is(err, ErrExtractFailed) {
		t.Fatalf("ExtractInto() = %v, want %v", err, ErrExtractFailed)
	}
	assertFileContent(t, filepath.Join(destDir, "keep.txt"), "keep")
}

func TestExtractIntoReplacesFile(t *testing.T) {
	destDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(destDir, "file.txt"), []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}
	data := tarBytes(t, entry{&tar.Header{Name: "file.txt", Mode: 0644, Typeflag: tar.TypeReg}, "new"})

	if err := ExtractInto(bytes.NewReader(data), destDir, Tar); err != nil {
		t.Fatalf("ExtractInto() = %v, want nil", err)
	}
	assertFileContent(t, filepath.Join(destDir, "file.txt"), "new")
}

func TestExtractIntoDoesNotWriteThroughSymlink(t *testing.T) {
	parent := t.TempDir()
	outside := filepath.Join(parent, "outside.txt")
	if err := os.WriteFile(outside, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	destDir := filepath.Join(parent, "dest")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(destDir, "file.txt")); err != nil {
		t.Fatal(err)
	}
	data := tarBytes(t, entry{&tar.Header{Name: "file.txt", Mode: 0644, Typeflag: tar.TypeReg}, "pwned"})

	if err := ExtractInto(bytes.NewReader(data), destDir, Tar); err != nil {
		t.Fatalf("ExtractInto() = %v, want nil", err)
	}
	assertFileContent(t, outside, "original")
	assertFileContent(t, filepath.Join(destDir, "file.txt"), "pwned")
	info, err := os.Lstat(filepath.Join(destDir, "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error("file.txt is still a symlink")
	}
}

func TestExtractIntoSymlinkChain(t *testing.T) {
	parent := t.TempDir()
	outside := filepath.Join(parent, "outside")
	if err := os.MkdirAll(outside, 0755); err != nil {
		t.Fatal(err)
	}
	destDir := filepath.Join(parent, "dest")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../outside", filepath.Join(destDir, "evil")); err != nil {
		t.Fatal(err)
	}
	data := tarBytes(t, entry{&tar.Header{Name: "evil/file.txt", Mode: 0644, Typeflag: tar.TypeReg}, "pwned"})

	err := ExtractInto(bytes.NewReader(data), destDir, Tar)
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("ExtractInto() = %v, want %v", err, ErrInvalidPath)
	}
	if _, err := os.Stat(filepath.Join(outside, "file.txt")); err == nil {
		t.Fatal("file was written outside destination")
	}
}

func TestExtractPathTraversal(t *testing.T) {
	data := tarBytes(t, entry{&tar.Header{Name: "../etc/passwd", Mode: 0644, Typeflag: tar.TypeReg}, "pwned"})

	err := Extract(zstdBytes(t, data), filepath.Join(t.TempDir(), "extracted"), Zstd)
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("Extract() = %v, want %v", err, ErrInvalidPath)
	}
}

func TestExtractAbsolutePath(t *testing.T) {
	data := tarBytes(t, entry{&tar.Header{Name: "/etc/passwd", Mode: 0644, Typeflag: tar.TypeReg}, "pwned"})

	err := Extract(zstdBytes(t, data), filepath.Join(t.TempDir(), "extracted"), Zstd)
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("Extract() = %v, want %v", err, ErrInvalidPath)
	}
}

func TestExtractGzipPathTraversal(t *testing.T) {
	data := tarBytes(t, entry{&tar.Header{Name: "../etc/passwd", Mode: 0644, Typeflag: tar.TypeReg}, "pwned"})
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}

	err := Extract(&buf, filepath.Join(t.TempDir(), "extracted"), Gzip)
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("Extract() = %v, want %v", err, ErrInvalidPath)
	}
}

func TestExtractFilePermissions(t *testing.T) {
	srcDir := t.TempDir()
	createTestFiles(t, srcDir)
	destDir := filepath.Join(t.TempDir(), "extracted")

	if err := Extract(writeArchive(t, srcDir, Zstd), destDir, Zstd); err != nil {
		t.Fatalf("Extract() = %v, want nil", err)
	}

	info, err := os.Stat(filepath.Join(destDir, "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("file mode = %o, want %o", info.Mode().Perm(), 0644)
	}
	info, err = os.Stat(filepath.Join(destDir, "subdir"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("dir mode = %o, want %o", info.Mode().Perm(), 0755)
	}
}

func TestExtractPreservesFileMode(t *testing.T) {
	data := tarBytes(t,
		entry{&tar.Header{Name: "script.sh", Mode: 0755, Typeflag: tar.TypeReg}, "#!/bin/sh\necho hello"},
		entry{&tar.Header{Name: "data.txt", Mode: 0644, Typeflag: tar.TypeReg}, "plain data"},
	)
	dest := filepath.Join(t.TempDir(), "preserved")

	if err := Extract(bytes.NewReader(data), dest, Tar); err != nil {
		t.Fatalf("Extract() = %v, want nil", err)
	}

	info, err := os.Stat(filepath.Join(dest, "script.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("script.sh mode = %o, want %o", info.Mode().Perm(), 0755)
	}
	info, err = os.Stat(filepath.Join(dest, "data.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("data.txt mode = %o, want %o", info.Mode().Perm(), 0644)
	}
}

func TestExtractUnsupportedFormat(t *testing.T) {
	err := Extract(bytes.NewReader(nil), filepath.Join(t.TempDir(), "extracted"), Format(99))
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("Extract() = %v, want %v", err, ErrUnsupportedFormat)
	}
}

func TestExtractSkipsPAXHeaders(t *testing.T) {
	data := tarBytes(t,
		entry{&tar.Header{Typeflag: tar.TypeXGlobalHeader, Name: "GlobalHead.0", PAXRecords: map[string]string{"comment": "global"}}, ""},
		entry{&tar.Header{Name: "file.txt", Mode: 0644, Typeflag: tar.TypeReg, Format: tar.FormatPAX, PAXRecords: map[string]string{"comment": "local"}}, "hello"},
	)
	dest := filepath.Join(t.TempDir(), "extracted")

	if err := Extract(bytes.NewReader(data), dest, Tar); err != nil {
		t.Fatalf("Extract() = %v, want nil", err)
	}
	assertFileContent(t, filepath.Join(dest, "file.txt"), "hello")
}

func TestExtractHardLink(t *testing.T) {
	data := tarBytes(t,
		entry{&tar.Header{Name: "original.txt", Mode: 0644, Typeflag: tar.TypeReg}, "linked content"},
		entry{&tar.Header{Name: "subdir/link.txt", Typeflag: tar.TypeLink, Linkname: "original.txt"}, ""},
	)
	dest := filepath.Join(t.TempDir(), "extracted")

	if err := Extract(bytes.NewReader(data), dest, Tar); err != nil {
		t.Fatalf("Extract() = %v, want nil", err)
	}
	assertFileContent(t, filepath.Join(dest, "original.txt"), "linked content")
	assertFileContent(t, filepath.Join(dest, "subdir", "link.txt"), "linked content")
}

func TestExtractHardLinkWithDotSlashPrefix(t *testing.T) {
	data := tarBytes(t,
		entry{&tar.Header{Name: "./original.txt", Mode: 0644, Typeflag: tar.TypeReg}, "gnu-style content"},
		entry{&tar.Header{Name: "./link.txt", Typeflag: tar.TypeLink, Linkname: "./original.txt"}, ""},
	)
	dest := filepath.Join(t.TempDir(), "extracted")

	if err := Extract(bytes.NewReader(data), dest, Tar); err != nil {
		t.Fatalf("Extract() = %v, want nil", err)
	}
	assertFileContent(t, filepath.Join(dest, "original.txt"), "gnu-style content")
	assertFileContent(t, filepath.Join(dest, "link.txt"), "gnu-style content")
}

func TestExtractHardLinkEscape(t *testing.T) {
	data := tarBytes(t, entry{&tar.Header{Name: "evil.txt", Typeflag: tar.TypeLink, Linkname: "../etc/passwd"}, ""})

	err := Extract(bytes.NewReader(data), filepath.Join(t.TempDir(), "extracted"), Tar)
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("Extract() = %v, want %v", err, ErrInvalidPath)
	}
}

func TestExtractHardLinkAbsolutePath(t *testing.T) {
	data := tarBytes(t, entry{&tar.Header{Name: "evil.txt", Typeflag: tar.TypeLink, Linkname: "/etc/passwd"}, ""})

	err := Extract(bytes.NewReader(data), filepath.Join(t.TempDir(), "extracted"), Tar)
	if !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("Extract() = %v, want %v", err, ErrInvalidPath)
	}
}
