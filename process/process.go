package process

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cruciblehq/utils-go/crex"
)

// Time output pipes get to drain after a command is killed.
const defaultWaitDelay = 5 * time.Second

// Exit code reported when a command did not run to completion.
const noExitCode = -1

// External command to run.
type Command struct {
	Path      string        // Program to run, found through PATH when it has no separator.
	Args      []string      // Arguments after the program name.
	Env       []string      // Environment as KEY=value pairs; nil inherits the current one.
	Stdin     io.Reader     // Standard input; nil reads nothing.
	Stdout    io.Writer     // Standard output; nil discards it.
	Stderr    io.Writer     // Standard error; nil discards it.
	WaitDelay time.Duration // Time output pipes get to drain after a kill; zero means 5 seconds.
}

// Runs c and returns its exit code.
//
// A non-zero exit code is a normal result, not an error. The error is set
// only when the command could not start ([ErrStart]) or ctx ended before it
// finished, in which case the exit code is -1.
func Run(ctx context.Context, c Command) (int, error) {
	err := c.cmd(ctx).Run()
	if ctx.Err() != nil {
		return noExitCode, ctx.Err()
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), nil
	}
	if err != nil {
		return noExitCode, crex.Wrap(ErrStart, err)
	}
	return 0, nil
}

// Runs c and fails with [ErrExit] when it exits with a non-zero code.
//
// The error message includes what the command wrote to stderr. c.Stderr, when
// set, still receives the output.
func Check(ctx context.Context, c Command) error {
	var stderr bytes.Buffer
	if c.Stderr != nil {
		c.Stderr = io.MultiWriter(c.Stderr, &stderr)
	} else {
		c.Stderr = &stderr
	}
	code, err := Run(ctx, c)
	if err != nil {
		return err
	}
	if code != 0 {
		return exitError(c, code, stderr.String())
	}
	return nil
}

// Runs c and returns its standard output with surrounding space trimmed.
//
// Fails like [Check] when the command exits with a non-zero code. c.Stdout is
// ignored.
func Output(ctx context.Context, c Command) (string, error) {
	var stdout strings.Builder
	c.Stdout = &stdout
	if err := Check(ctx, c); err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout.String()), nil
}

// Returns KEY=value pairs for the named variables that have a value.
//
// Variables that are unset or empty in the current environment are left out.
func Environ(keys ...string) []string {
	var env []string
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			env = append(env, fmt.Sprintf("%s=%s", key, val))
		}
	}
	return env
}

// Returns the [exec.Cmd] for c.
func (c Command) cmd(ctx context.Context) *exec.Cmd {
	cmd := exec.CommandContext(ctx, c.Path, c.Args...)
	cmd.Env = c.Env
	cmd.Stdin = c.Stdin
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stderr
	cmd.WaitDelay = c.WaitDelay
	if cmd.WaitDelay == 0 {
		cmd.WaitDelay = defaultWaitDelay
	}
	return cmd
}

// Returns the [ErrExit] error for c exiting with code.
func exitError(c Command, code int, stderr string) error {
	line := strings.Join(append([]string{c.Path}, c.Args...), " ")
	if msg := strings.TrimSpace(stderr); msg != "" {
		return crex.Newf(ErrExit, "%q exited with code %d and reported %q", line, code, msg)
	}
	return crex.Newf(ErrExit, "%q exited with code %d", line, code)
}
