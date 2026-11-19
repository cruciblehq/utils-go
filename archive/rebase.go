package archive

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/cruciblehq/utils-go/crex"
)

// Copies the uncompressed tar stream r to w, moving the entries under from to
// to.
//
// from and to are slash-separated directories inside the archive; leading
// slashes are ignored. An entry named from, or inside from, is moved under to.
// When from is empty, every entry is moved under to. Other entries are copied
// unchanged. Hard link targets move the same way as names, so links stay
// valid. The entry for the archive root is dropped. All other header fields
// are kept, except the tar format, which the writer picks again.
func Rebase(w io.Writer, r io.Reader, from, to string) error {
	from, to = cleanName(from), cleanName(to)

	tr := tar.NewReader(r)
	tw := tar.NewWriter(w)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return crex.Wrap(ErrReadFailed, err)
		}

		name := rebaseName(cleanName(hdr.Name), from, to)
		if name == "" {
			continue
		}
		hdr.Name = name
		if hdr.Typeflag == tar.TypeDir {
			hdr.Name = fmt.Sprintf("%s/", name)
		}
		if hdr.Typeflag == tar.TypeLink {
			hdr.Linkname = rebaseName(cleanName(hdr.Linkname), from, to)
		}
		hdr.Format = tar.FormatUnknown

		if err := tw.WriteHeader(hdr); err != nil {
			return crex.Wrap(ErrWriteFailed, err)
		}
		if _, err := io.Copy(tw, tr); err != nil {
			return crex.Wrap(ErrWriteFailed, err)
		}
	}
	if err := tw.Close(); err != nil {
		return crex.Wrap(ErrWriteFailed, err)
	}
	return nil
}

// Returns name moved from the directory from to the directory to.
//
// Names outside from are returned unchanged. The archive root stays empty.
func rebaseName(name, from, to string) string {
	if name == "" {
		return ""
	}
	if from != "" {
		if name == from {
			return to
		}
		rest, ok := strings.CutPrefix(name, fmt.Sprintf("%s/", from))
		if !ok {
			return name
		}
		name = rest
	}
	return path.Join(to, name)
}
