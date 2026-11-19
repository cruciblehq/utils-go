// Package file provides file system utilities.
//
// It covers two concerns.
//
// File helpers: atomic writes ([AtomicFile], [WriteFile]) that sync to disk
// before renaming and set the file mode, common permission modes, existence
// checks, removal of empty directories, listing of subdirectories, and a check
// that a path is inside a root ([IsInside]). All helpers operate on plain
// [os.File] values and file system paths with no domain knowledge.
//
// File locking: exclusive locks via flock(2), with acquisition that honours
// context cancellation and deadlines. Locking is not supported on Windows.
package file
