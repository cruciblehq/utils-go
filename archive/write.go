package archive

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/cruciblehq/utils-go/crex"
)

// Writes src to w as an archive in format f.
//
// When src is a directory, its contents are archived with names relative to
// it. When src is a file, it is archived under its base name. Each name is
// placed under prefix, a slash-separated directory inside the archive; an
// empty prefix places entries at the top level. A non-empty prefix also gets
// its own directory entry when src is a directory. Only regular files and
// directories are archivable; anything else returns [ErrUnsupportedFileType].
// Permissions are kept with the setuid, setgid, and sticky bits removed.
func Write(w io.Writer, src, prefix string, f Format) error {
	cw, err := Compress(w, f)
	if err != nil {
		return crex.Wrap(ErrWriteFailed, err)
	}
	tw := tar.NewWriter(cw)
	err = writeTree(tw, src, cleanName(prefix))
	if err == nil {
		err = tw.Close()
	}
	if closeErr := cw.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return crex.Wrap(ErrWriteFailed, err)
	}
	return nil
}

// Writes the file or directory tree at src to tw under prefix.
func writeTree(tw *tar.Writer, src, prefix string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return writeEntry(tw, src, path.Join(prefix, filepath.Base(src)), info)
	}
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		name := path.Join(prefix, filepath.ToSlash(rel))
		if name == "." {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		return writeEntry(tw, p, name, info)
	})
}

// Writes the file or directory at p to tw as the entry name.
func writeEntry(tw *tar.Writer, p, name string, info fs.FileInfo) error {
	mode := info.Mode()
	if !mode.IsRegular() && !mode.IsDir() {
		return crex.Newf(ErrUnsupportedFileType, "%q is not a regular file or directory", p)
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = name
	if mode.IsDir() {
		header.Name = fmt.Sprintf("%s/", name)
	}
	header.Mode = int64(mode.Perm())

	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	if mode.IsRegular() {
		return copyFile(tw, p)
	}
	return nil
}

// Copies the contents of the file at p to w.
func copyFile(w io.Writer, p string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}

// Returns a slash-separated archive name without leading slashes or dot
// segments. The root of the archive is the empty string.
func cleanName(name string) string {
	return path.Clean(fmt.Sprintf("/%s", filepath.ToSlash(name)))[1:]
}
