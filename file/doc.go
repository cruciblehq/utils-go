// Package file provides file system utilities.
//
// It covers two concerns.
//
// File helpers: atomic writes ([AtomicFile], [WriteFile]) that sync to disk
// before renaming and set the file mode, existence checks, empty-directory
// pruning, subdirectory enumeration. All helpers operate on plain [os.File]
// values and filesystem paths with no domain knowledge.
//
// File locking: exclusive locks via flock(2), with context-aware acquisition
// that honours cancellation and deadlines.
package file
