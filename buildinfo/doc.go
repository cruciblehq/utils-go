// Package buildinfo describes a binary from values set at link time.
//
// Programs declare string variables and set them with -ldflags "-X ...", then
// pass them to [New]. A version or commit that was not set is taken from what
// the Go toolchain recorded in the binary, when available. Values that are
// still unknown are reported as "(undefined)", and a binary without a version,
// stage, and commit is treated as a local build.
package buildinfo
