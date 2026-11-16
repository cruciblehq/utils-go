package file

import (
	"os"
	"path/filepath"
)

// Default permission modes.
const (

	// Default permission mode used when creating directories.
	DefaultDirMode os.FileMode = 0755

	// Default permission mode used when creating files.
	DefaultFileMode os.FileMode = 0644

	// Default permission mode for executable files.
	DefaultExecMode os.FileMode = 0755

	// Permission mode for directories that only their owner can read, write,
	// or enter.
	OwnerDirMode os.FileMode = 0o700

	// Permission mode for files that only their owner can read or write.
	OwnerFileMode os.FileMode = 0o600
)

// Removes dir if it contains no entries.
//
// If dir does not exist, returns nil. If dir is not empty, returns nil and does
// not remove it. Returns an error if dir exists but cannot be read or removed.
func RemoveDirIfEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(entries) > 0 {
		return nil
	}
	err = os.Remove(dir)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Removes dir and each parent while it is empty, stopping before root.
//
// dir must be inside root. Stops at the first directory that is not empty.
// Missing directories are skipped. root itself is never removed.
func RemoveEmptyParents(dir, root string) error {
	root = filepath.Clean(root)
	for current := filepath.Clean(dir); IsInside(current, root); current = filepath.Dir(current) {
		entries, err := os.ReadDir(current)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return nil
		}
		if err := os.Remove(current); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// Whether path is strictly inside root.
func IsInside(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != "." && filepath.IsLocal(rel)
}

// Lists the names of immediate subdirectories of dir.
//
// Regular files are ignored. The returned slice is nil when dir contains no
// subdirectories, and the names are in directory-read order (not sorted).
func ListSubdirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// Whether path exists.
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
