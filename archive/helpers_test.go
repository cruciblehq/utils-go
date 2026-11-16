package archive

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func createTestFiles(t *testing.T, dir string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	subdir := filepath.Join(dir, "subdir")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(subdir, "nested.txt"), []byte("nested"), 0644); err != nil {
		t.Fatal(err)
	}

	emptydir := filepath.Join(dir, "emptydir")
	if err := os.MkdirAll(emptydir, 0755); err != nil {
		t.Fatal(err)
	}
}

func writeArchive(t *testing.T, src string, f Format) *bytes.Reader {
	t.Helper()

	var buf bytes.Buffer
	if err := Write(&buf, src, "", f); err != nil {
		t.Fatalf("Write() = %v, want nil", err)
	}
	return bytes.NewReader(buf.Bytes())
}

type entry struct {
	header *tar.Header
	body   string
}

func tarBytes(t *testing.T, entries ...entry) []byte {
	t.Helper()

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, e := range entries {
		h := *e.header
		if h.Typeflag == tar.TypeReg {
			h.Size = int64(len(e.body))
		}
		if err := tw.WriteHeader(&h); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(e.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func zstdBytes(t *testing.T, data []byte) *bytes.Reader {
	t.Helper()

	var buf bytes.Buffer
	zw, err := zstd.NewWriter(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := zw.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(buf.Bytes())
}

func readHeaders(t *testing.T, data []byte) map[string]*tar.Header {
	t.Helper()

	headers := make(map[string]*tar.Header)
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return headers
		}
		if err != nil {
			t.Fatal(err)
		}
		headers[hdr.Name] = hdr
	}
}

func assertFileContent(t *testing.T, path, expected string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}

	if string(content) != expected {
		t.Fatalf("%s = %q, want %q", path, string(content), expected)
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("directory %s does not exist: %v", path, err)
	}

	if !info.IsDir() {
		t.Fatalf("%s is not a directory", path)
	}
}
