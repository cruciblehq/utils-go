// Package archive writes, extracts, and rewrites tar archives as streams.
//
// Archives are tar or tar compressed with Gzip or Zstandard (see [Format]).
// Every function reads from an [io.Reader] or writes to an [io.Writer], so
// the caller decides where the bytes live. [FormatOfName] finds the format
// from a file name and [DetectFormat] finds it from the leading bytes of a
// stream.
//
// [Write] archives a file or directory tree. Only regular files and directories
// are archivable. Symlinks and special files (devices, sockets, named pipes)
// are rejected with [ErrUnsupportedFileType]. Permissions are kept with the
// setuid, setgid, and sticky bits removed.
//
// [Extract] unpacks an archive into a new directory, which appears only once
// extraction has fully succeeded. [ExtractInto] unpacks into a directory that
// may already exist. Both accept regular files, directories, symlinks, and hard
// links, and reject any entry that would land outside the destination with
// [ErrInvalidPath].
//
// [Rebase] copies a tar stream while moving its entries to another directory,
// which is useful to place an archive at a different path before applying it.
//
// Writing a directory to a file:
//
//	f, err := os.Create("site.tar.zst")
//	if err != nil {
//		return err
//	}
//	defer f.Close()
//	if err := archive.Write(f, "site", "", archive.Zstd); err != nil {
//		return err
//	}
//
// Extracting a downloaded archive:
//
//	resp, err := http.Get(url)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//	if err := archive.Extract(resp.Body, "tools", archive.Gzip); err != nil {
//		return err
//	}
package archive
