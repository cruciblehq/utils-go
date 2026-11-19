package archive

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Extracts the archive r in format f into the new directory dest.
//
// The archive is extracted into a temporary directory next to dest, which is
// renamed to dest only once extraction succeeds, so dest never holds a partial
// tree. dest must not exist; if it does, the error wraps [os.ErrExist]. Parent
// directories are created as needed. dest gets mode [file.DefaultDirMode].
func Extract(r io.Reader, dest string, f Format) error {
	if _, err := os.Lstat(dest); err == nil {
		return crex.Wrap(ErrExtractFailed, os.ErrExist)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return crex.Wrap(ErrExtractFailed, err)
	}

	parent := filepath.Dir(dest)
	if err := os.MkdirAll(parent, os.ModePerm); err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}
	tmp, err := os.MkdirTemp(parent, fmt.Sprintf("%s*", file.TempPrefix))
	if err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}
	defer os.RemoveAll(tmp)

	if err := extract(r, tmp, f); err != nil {
		return err
	}
	if err := os.Chmod(tmp, file.DefaultDirMode); err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}
	return nil
}

// Extracts the archive r in format f into dir, adding to what it holds.
//
// dir is created if it does not exist. Entries replace files of the same name.
// Nothing is removed when extraction fails, so dir may then hold part of the
// archive.
func ExtractInto(r io.Reader, dir string, f Format) error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}
	return extract(r, dir, f)
}

// Decompresses r in format f and extracts its entries into the directory dest.
func extract(r io.Reader, dest string, f Format) error {
	dr, err := Decompress(r, f)
	if err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}
	defer dr.Close()

	if err := readTar(tar.NewReader(dr), dest); err != nil {
		return crex.Wrap(ErrExtractFailed, err)
	}
	return nil
}

// Reads tar entries and extracts them to dest.
//
// Validates each entry path for security before extraction. Returns the first
// error encountered or nil on successful completion. The tar reader is read
// sequentially until the end of the archive or an error occurs. The caller is
// responsible for closing the underlying decompression reader.
func readTar(tr *tar.Reader, dest string) error {
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		target, err := validateAndJoinPath(dest, header.Name)
		if err != nil {
			return err
		}

		if target == "" {
			continue
		}

		if err := validatePathComponents(dest, filepath.Dir(target)); err != nil {
			return err
		}

		if err := extractEntry(header, tr, dest, target); err != nil {
			return err
		}
	}
}

// Validates and joins an archive path with the destination directory.
//
// Strips leading "./" prefixes common in tar archives, then uses
// [filepath.Localize] to convert slash-separated paths to OS format and
// [filepath.IsLocal] to ensure the path is local (not absolute, no ".."
// traversal, no reserved names on Windows). Returns the validated path
// joined with dest, or ("", nil) for root entries like "./" that resolve
// to the destination itself.
func validateAndJoinPath(dest, name string) (string, error) {
	name = strings.TrimPrefix(name, "./")
	name = strings.TrimRight(name, "/")

	if name == "." || name == "" {
		return "", nil
	}

	localName, err := filepath.Localize(name)
	if err != nil {
		return "", ErrInvalidPath
	}

	// Not empty, not absolute path, no ".." traversal, no reserved names on Windows
	if !filepath.IsLocal(localName) {
		return "", ErrInvalidPath
	}

	return filepath.Join(dest, localName), nil
}

// Extracts a single tar entry to target.
//
// Handles directories, regular files, symlinks, and hard links. PAX extended
// headers and GNU long name headers are skipped. Returns [ErrUnsupportedFileType]
// for all other entry types (devices, FIFOs, etc.).
func extractEntry(header *tar.Header, tr *tar.Reader, dest, target string) error {
	switch header.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, os.FileMode(header.Mode)&os.ModePerm)

	case tar.TypeReg:
		if err := removeExisting(target); err != nil {
			return err
		}
		return extractFile(tr, target, os.FileMode(header.Mode)&os.ModePerm)

	case tar.TypeSymlink:
		if err := validateSymlink(dest, target, header.Linkname); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
			return err
		}
		if err := removeExisting(target); err != nil {
			return err
		}
		return os.Symlink(header.Linkname, target)

	case tar.TypeLink:
		linkTarget, err := validateHardlink(dest, header.Linkname)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
			return err
		}
		if err := removeExisting(target); err != nil {
			return err
		}
		return os.Link(linkTarget, target)

	case tar.TypeXHeader, tar.TypeXGlobalHeader, tar.TypeGNULongName, tar.TypeGNULongLink:
		return nil

	default:
		return ErrUnsupportedFileType
	}
}

// Rejects symlinks whose resolved target escapes the destination tree.
//
// A relative linkname is resolved from the directory that holds the symlink.
// An absolute linkname is used as is. For example, a symlink at "foo/bar/link"
// with linkname "../x" resolves to "foo/x". It is valid when dest is "foo",
// but invalid when dest is "foo/bar". The resolved link target must be equal
// to or inside the cleaned destination path.
func validateSymlink(dest, target, linkname string) error {
	resolved := linkname
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(filepath.Dir(target), resolved)
	}
	resolved = filepath.Clean(resolved)
	cleanDest := filepath.Clean(dest)
	if resolved != cleanDest && !file.IsInside(resolved, cleanDest) {
		return ErrInvalidPath
	}
	return nil
}

// Validates each component of the path for directory escape.
//
// Each component of the path is checked for existence and if it is a symlink,
// the resolved target is validated to ensure it does not escape the destination
// tree. This prevents directory traversal attacks via symlinks in intermediate
// directories.
func validatePathComponents(dest, dir string) error {
	cleanDest := filepath.Clean(dest)
	rel, err := filepath.Rel(cleanDest, dir)
	if err != nil {
		return ErrInvalidPath
	}

	current := cleanDest
	for part := range strings.SplitSeq(rel, string(filepath.Separator)) {
		if part == "." || part == "" {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return ErrInvalidPath
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if err := checkSymlinkEscape(current, cleanDest); err != nil {
				return err
			}
		}
	}
	return nil
}

// Rejects symlinks that escape the destination tree.
//
// Unlike [validateSymlink], which validates the declared linkname string, this
// function follows the full OS symlink chain. It is used to detect cases where
// previously extracted or pre-existing symlinks in intermediate path components
// chain together to escape the destination tree.
func checkSymlinkEscape(current, cleanDest string) error {
	resolved, err := filepath.EvalSymlinks(current)
	if err != nil {
		return ErrInvalidPath
	}
	resolved = filepath.Clean(resolved)
	if resolved != cleanDest && !file.IsInside(resolved, cleanDest) {
		return ErrInvalidPath
	}
	return nil
}

// Rejects hard links whose target would escape the destination tree.
//
// Hard link Linkname in a tar archive is an archive-relative path to a
// previously extracted entry. It is validated and joined with dest the
// same way regular entry names are, ensuring the resolved target stays
// within the destination tree.
func validateHardlink(dest, linkname string) (string, error) {
	target, err := validateAndJoinPath(dest, linkname)
	if err != nil {
		return "", err
	}
	if target == "" {
		return "", ErrInvalidPath
	}
	return target, nil
}

// Removes what is at target unless it is a directory.
//
// A new entry then replaces an existing file or symlink instead of writing
// through it, so a symlink cannot redirect the write outside the destination.
func removeExisting(target string) error {
	info, err := os.Lstat(target)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	return os.Remove(target)
}

// Extracts a regular file from r to target with the given mode.
//
// Creates parent directories as needed with [os.ModePerm], subject to the
// process umask.
func extractFile(r io.Reader, target string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = io.Copy(f, r); err != nil {
		return err
	}

	return nil
}
