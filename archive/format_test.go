package archive

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestFormatOf(t *testing.T) {
	tests := []struct {
		name    string
		want    Format
		wantErr bool
	}{
		{"archive.tar.zst", Zstd, false},
		{"archive.tar.gz", Gzip, false},
		{"archive.tgz", Gzip, false},
		{"ARCHIVE.TAR.ZST", Zstd, false},
		{"ARCHIVE.TAR.GZ", Gzip, false},
		{"archive.tar", Tar, false},
		{"archive.zip", Tar, true},
		{"archive.rar", Tar, true},
		{"noextension", Tar, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatOfName(tt.name)
			if (err != nil) != tt.wantErr {
				t.Fatalf("FormatOfName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, ErrUnsupportedFormat) {
				t.Fatalf("FormatOfName(%q) = %v, want %v", tt.name, err, ErrUnsupportedFormat)
			}
			if got != tt.want {
				t.Fatalf("FormatOfName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestFormatString(t *testing.T) {
	tests := []struct {
		f    Format
		want string
	}{
		{Tar, ".tar"},
		{Gzip, ".tar.gz"},
		{Zstd, ".tar.zst"},
		{Format(99), ""},
	}

	for _, tt := range tests {
		if got := tt.f.String(); got != tt.want {
			t.Errorf("Format(%d).String() = %q, want %q", tt.f, got, tt.want)
		}
	}
}

func TestFormatZeroValue(t *testing.T) {
	var f Format
	if f != Tar {
		t.Fatalf("zero Format = %v, want %v", f, Tar)
	}
}

func TestDetect(t *testing.T) {
	data := tarBytes(t, entry{&tar.Header{Name: "file.txt", Mode: 0644, Typeflag: tar.TypeReg}, "hello"})

	for _, f := range []Format{Tar, Gzip, Zstd} {
		t.Run(f.String(), func(t *testing.T) {
			got, r, err := DetectFormat(compressBytes(t, data, f))
			if err != nil {
				t.Fatalf("DetectFormat() error = %v, want nil", err)
			}
			if got != f {
				t.Fatalf("DetectFormat() = %v, want %v", got, f)
			}
			if !bytes.Equal(decompressAll(t, r, got), data) {
				t.Fatal("DetectFormat() reader does not yield the whole stream")
			}
		})
	}
}

func compressBytes(t *testing.T, data []byte, f Format) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	cw, err := Compress(&buf, f)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cw.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := cw.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf
}

func decompressAll(t *testing.T, r io.Reader, f Format) []byte {
	t.Helper()

	dr, err := Decompress(r, f)
	if err != nil {
		t.Fatal(err)
	}
	defer dr.Close()
	out, err := io.ReadAll(dr)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestDetectEmptyStream(t *testing.T) {
	got, _, err := DetectFormat(bytes.NewReader(nil))
	if err != nil {
		t.Fatalf("DetectFormat() error = %v, want nil", err)
	}
	if got != Tar {
		t.Fatalf("DetectFormat() = %v, want %v", got, Tar)
	}
}

func TestDetectReadError(t *testing.T) {
	_, _, err := DetectFormat(errReader{})
	if !errors.Is(err, ErrReadFailed) {
		t.Fatalf("DetectFormat() = %v, want %v", err, ErrReadFailed)
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestCompressUnsupportedFormat(t *testing.T) {
	if _, err := Compress(io.Discard, Format(99)); !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("Compress() = %v, want %v", err, ErrUnsupportedFormat)
	}
}

func TestDecompressUnsupportedFormat(t *testing.T) {
	if _, err := Decompress(bytes.NewReader(nil), Format(99)); !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("Decompress() = %v, want %v", err, ErrUnsupportedFormat)
	}
}
