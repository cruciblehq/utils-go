//go:build windows

package file

import (
	"context"
	"os"

	"github.com/cruciblehq/utils-go/crex"
)

// System error returned by every file locking function.
var errNotSupported = crex.SystemError("file locking is unavailable", "file locking is not supported on Windows").
	Cause(ErrNotSupported).
	Err()

// Acquires an exclusive file lock on f.
//
// File locking is not supported on Windows; always returns an error that wraps
// [ErrNotSupported].
func LockWithContext(_ context.Context, _ *os.File) error {
	return errNotSupported
}

// Acquires an exclusive file lock on f.
//
// File locking is not supported on Windows; always returns an error that wraps
// [ErrNotSupported].
func Lock(_ *os.File) error {
	return errNotSupported
}

// Releases the file lock on f.
//
// File locking is not supported on Windows; always returns an error that wraps
// [ErrNotSupported].
func Unlock(_ *os.File) error {
	return errNotSupported
}
