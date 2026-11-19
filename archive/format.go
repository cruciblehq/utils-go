package archive

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"strings"

	"github.com/klauspost/compress/zstd"

	"github.com/cruciblehq/utils-go/crex"
)

// Compression applied to a tar archive.
//
// The zero value is [Tar], an uncompressed archive.
type Format int

const (
	Tar  Format = iota // Plain tar, no compression (.tar).
	Gzip               // Gzip compression (.tar.gz, .tgz).
	Zstd               // Zstandard compression (.tar.zst).
)

const (
	extTar  = ".tar"     // File extension for plain tar archives.
	extGzip = ".tar.gz"  // File extension for tar archives compressed with Gzip.
	extTgz  = ".tgz"     // Short file extension for tar archives compressed with Gzip.
	extZstd = ".tar.zst" // File extension for tar archives compressed with Zstandard.
)

// Leading bytes of compressed streams.
var (
	gzipMagic = []byte{0x1f, 0x8b}             // Gzip member header.
	zstdMagic = []byte{0x28, 0xb5, 0x2f, 0xfd} // Zstandard frame header.
)

// Returns the canonical file extension for the format.
func (f Format) String() string {
	switch f {
	case Tar:
		return extTar
	case Gzip:
		return extGzip
	case Zstd:
		return extZstd
	default:
		return ""
	}
}

// Returns the format named by the extension of a file name.
//
// The check ignores letter case. Returns [ErrUnsupportedFormat] if the
// extension is not one of the extensions listed in [Format].
func FormatOfName(name string) (Format, error) {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, extZstd):
		return Zstd, nil
	case strings.HasSuffix(lower, extGzip), strings.HasSuffix(lower, extTgz):
		return Gzip, nil
	case strings.HasSuffix(lower, extTar):
		return Tar, nil
	default:
		return Tar, crex.Newf(ErrUnsupportedFormat, "unknown archive extension in %q", name)
	}
}

// Finds the format of a stream from its leading bytes.
//
// Returns the format and a reader that yields the whole stream, including the
// bytes inspected. A stream that is not Gzip or Zstandard is reported as [Tar].
func DetectFormat(r io.Reader) (Format, io.Reader, error) {
	br := bufio.NewReader(r)
	head, err := br.Peek(len(zstdMagic))
	if err != nil && !errors.Is(err, io.EOF) {
		return Tar, br, crex.Wrap(ErrReadFailed, err)
	}
	switch {
	case bytes.HasPrefix(head, zstdMagic):
		return Zstd, br, nil
	case bytes.HasPrefix(head, gzipMagic):
		return Gzip, br, nil
	default:
		return Tar, br, nil
	}
}

// Returns a writer that compresses what is written to it into w in format f.
//
// The caller must close the returned writer to flush the compressed data. It
// does not close w.
func Compress(w io.Writer, f Format) (io.WriteCloser, error) {
	switch f {
	case Tar:
		return nopWriteCloser{w}, nil
	case Gzip:
		return gzip.NewWriter(w), nil
	case Zstd:
		return zstd.NewWriter(w)
	default:
		return nil, crex.ProgrammingErrorf("cannot compress archive", "unsupported archive format %d", f).
			Cause(ErrUnsupportedFormat).
			Err()
	}
}

// Returns a reader that decompresses r from format f.
//
// The caller must close the returned reader to release its resources. It does
// not close r.
func Decompress(r io.Reader, f Format) (io.ReadCloser, error) {
	switch f {
	case Tar:
		return io.NopCloser(r), nil
	case Gzip:
		return gzip.NewReader(r)
	case Zstd:
		zr, err := zstd.NewReader(r)
		if err != nil {
			return nil, err
		}
		return zr.IOReadCloser(), nil
	default:
		return nil, crex.ProgrammingErrorf("cannot decompress archive", "unsupported archive format %d", f).
			Cause(ErrUnsupportedFormat).
			Err()
	}
}

// Writer with a Close method that does nothing.
type nopWriteCloser struct {
	io.Writer
}

// Does nothing and returns nil.
func (nopWriteCloser) Close() error {
	return nil
}
