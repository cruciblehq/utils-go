// Package file provides file system utilities.
//
// It covers three concerns.
//
// Path validation: ValidateAbsPath validates slash-separated absolute paths
// with POSIX semantics, independent of the host operating system.
//
// File helpers: atomic writes with SHA-256 digest tracking, existence checks,
// empty-directory pruning, subdirectory enumeration. All helpers operate on
// plain [os.File] values and filesystem paths with no domain knowledge.
//
// File locking: exclusive locks via flock(2), with context-aware acquisition
// that honours cancellation and deadlines.
package file
