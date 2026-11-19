package file

import (
	"fmt"
	"os"
	"path/filepath"
)

// Name prefix of the temporary files [CreateAtomic] creates.
//
// A file with this prefix that is still present was left by a process that
// stopped before committing, and can be removed.
const TempPrefix = ".tmp-"

// File written under a temporary name and moved into place on commit.
//
// Readers never see a partial file: the destination either keeps its content
// or holds the complete new content. The data is synced to disk before the
// rename, so a crash cannot leave an empty or truncated file behind it.
type AtomicFile struct {
	f    *os.File    // Temporary file being written.
	mode os.FileMode // Permission mode set before the rename.
	done bool        // Whether Commit or Abort has run.
}

// Creates a temporary file in dir that becomes visible only on
// [AtomicFile.Commit].
//
// dir must be on the same file system as the final path, so the rename is
// atomic. The file gets mode when committed, independent of the umask.
func CreateAtomic(dir string, mode os.FileMode) (*AtomicFile, error) {
	f, err := os.CreateTemp(dir, fmt.Sprintf("%s*", TempPrefix))
	if err != nil {
		return nil, err
	}
	return &AtomicFile{f: f, mode: mode}, nil
}

// Writes p to the temporary file.
func (a *AtomicFile) Write(p []byte) (int, error) {
	return a.f.Write(p)
}

// Syncs the data, applies the mode, and renames the file to dest.
//
// The temporary file is removed when any step fails. Calling Commit after
// Commit or Abort does nothing.
func (a *AtomicFile) Commit(dest string) error {
	if a.done {
		return nil
	}
	a.done = true
	name := a.f.Name()
	if err := a.finish(dest); err != nil {
		a.f.Close()
		os.Remove(name)
		return err
	}
	return nil
}

// Runs the steps of [AtomicFile.Commit].
func (a *AtomicFile) finish(dest string) error {
	if err := a.f.Chmod(a.mode); err != nil {
		return err
	}
	if err := a.f.Sync(); err != nil {
		return err
	}
	if err := a.f.Close(); err != nil {
		return err
	}
	return os.Rename(a.f.Name(), dest)
}

// Closes and removes the temporary file.
//
// Calling Abort after Commit or Abort does nothing, so it is safe to defer.
func (a *AtomicFile) Abort() error {
	if a.done {
		return nil
	}
	a.done = true
	a.f.Close()
	return os.Remove(a.f.Name())
}

// Writes data to path atomically with the given mode.
//
// The temporary file is created next to path. See [AtomicFile].
func WriteFile(path string, data []byte, mode os.FileMode) error {
	a, err := CreateAtomic(filepath.Dir(path), mode)
	if err != nil {
		return err
	}
	defer a.Abort()
	if _, err := a.Write(data); err != nil {
		return err
	}
	return a.Commit(path)
}
