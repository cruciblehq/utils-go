// Package process runs external commands.
//
// [Run] reports the exit code and treats a non-zero exit as a normal result.
// [Check] and [Output] treat it as an error that includes what the command
// wrote to stderr. Every command stops when its context ends, and its output
// pipes get a bounded time to drain after it is killed. [Environ] builds an
// environment from selected variables of the current one.
package process
