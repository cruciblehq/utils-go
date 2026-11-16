//go:build !windows

package process

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	var out bytes.Buffer
	code, err := Run(context.Background(), Command{Path: "sh", Args: []string{"-c", "echo hi; exit 3"}, Stdout: &out})
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if code != 3 {
		t.Errorf("Run() code = %d, want 3", code)
	}
	if out.String() != "hi\n" {
		t.Errorf("stdout = %q, want %q", out.String(), "hi\n")
	}
}

func TestRunStdin(t *testing.T) {
	var out bytes.Buffer
	if _, err := Run(context.Background(), Command{Path: "cat", Stdin: strings.NewReader("data"), Stdout: &out}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "data" {
		t.Errorf("stdout = %q, want %q", out.String(), "data")
	}
}

func TestRunEnv(t *testing.T) {
	var out bytes.Buffer
	if _, err := Run(context.Background(), Command{Path: "sh", Args: []string{"-c", "echo $FOO"}, Env: []string{"FOO=bar"}, Stdout: &out}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "bar\n" {
		t.Errorf("stdout = %q, want %q", out.String(), "bar\n")
	}
}

func TestRunNotFound(t *testing.T) {
	code, err := Run(context.Background(), Command{Path: "/nonexistent/program"})
	if !errors.Is(err, ErrStart) {
		t.Errorf("Run() error = %v, want %v", err, ErrStart)
	}
	if code != noExitCode {
		t.Errorf("Run() code = %d, want %d", code, noExitCode)
	}
}

func TestRunCancelled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	code, err := Run(ctx, Command{Path: "sleep", Args: []string{"5"}})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Run() error = %v, want %v", err, context.DeadlineExceeded)
	}
	if code != noExitCode {
		t.Errorf("Run() code = %d, want %d", code, noExitCode)
	}
}

func TestCheck(t *testing.T) {
	if err := Check(context.Background(), Command{Path: "true"}); err != nil {
		t.Errorf("Check(true) = %v, want nil", err)
	}
}

func TestCheckFailure(t *testing.T) {
	var stderr bytes.Buffer
	err := Check(context.Background(), Command{Path: "sh", Args: []string{"-c", "echo boom >&2; exit 2"}, Stderr: &stderr})
	if !errors.Is(err, ErrExit) {
		t.Fatalf("Check() = %v, want %v", err, ErrExit)
	}
	if !strings.Contains(err.Error(), "boom") || !strings.Contains(err.Error(), "code 2") {
		t.Errorf("Check() message = %q, want it to name the code and stderr", err.Error())
	}
	if stderr.String() != "boom\n" {
		t.Errorf("stderr = %q, want %q", stderr.String(), "boom\n")
	}
}

func TestCheckFailureSilent(t *testing.T) {
	err := Check(context.Background(), Command{Path: "false"})
	if !errors.Is(err, ErrExit) {
		t.Fatalf("Check() = %v, want %v", err, ErrExit)
	}
}

func TestOutput(t *testing.T) {
	out, err := Output(context.Background(), Command{Path: "sh", Args: []string{"-c", "echo '  value  '"}})
	if err != nil {
		t.Fatalf("Output() = %v, want nil", err)
	}
	if out != "value" {
		t.Errorf("Output() = %q, want %q", out, "value")
	}
}

func TestOutputFailure(t *testing.T) {
	if _, err := Output(context.Background(), Command{Path: "false"}); !errors.Is(err, ErrExit) {
		t.Errorf("Output() = %v, want %v", err, ErrExit)
	}
}

func TestEnviron(t *testing.T) {
	t.Setenv("PROCESS_TEST_SET", "yes")
	t.Setenv("PROCESS_TEST_EMPTY", "")

	env := Environ("PROCESS_TEST_SET", "PROCESS_TEST_EMPTY", "PROCESS_TEST_UNSET_XYZ")
	if len(env) != 1 || env[0] != "PROCESS_TEST_SET=yes" {
		t.Errorf("Environ() = %v, want [PROCESS_TEST_SET=yes]", env)
	}
}
