//go:build !debug

package crex

// Does nothing. Builds with the "debug" tag panic when condition is false.
func Assert(condition bool, message string) {
	// No-op.
}

// Does nothing. Builds with the "debug" tag panic when condition is false.
func Assertf(condition bool, format string, args ...any) {
	// No-op.
}
