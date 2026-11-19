//go:build debug

package crex

import (
	"fmt"
	"runtime"
)

// Panics with message if the condition is false.
//
// This function is only active when built with the "debug" build tag. In
// other builds it does nothing.
func Assert(condition bool, message string) {
	assert(condition, message)
}

// Panics with the formatted message if the condition is false.
//
// This function is only active when built with the "debug" build tag. In
// other builds it does nothing.
func Assertf(condition bool, format string, args ...any) {
	assert(condition, fmt.Sprintf(format, args...))
}

// Panics with message and the location of the asserting call if condition is
// false.
func assert(condition bool, message string) {
	if !condition {
		_, file, line, _ := runtime.Caller(2)
		panic(fmt.Sprintf("assertion failed: %s\n  at %s:%d", message, file, line))
	}
}
