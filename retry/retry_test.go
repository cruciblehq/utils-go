package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

var errNotYet = errors.New("not yet")

func TestPollSucceeds(t *testing.T) {
	calls := 0
	err := Poll(context.Background(), time.Second, time.Millisecond, func(context.Context) error {
		calls++
		if calls < 3 {
			return errNotYet
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Poll() = %v, want nil", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestPollTimeout(t *testing.T) {
	err := Poll(context.Background(), 20*time.Millisecond, 5*time.Millisecond, func(context.Context) error {
		return errNotYet
	})
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("Poll() = %v, want %v", err, ErrTimeout)
	}
	if !errors.Is(err, errNotYet) {
		t.Errorf("Poll() = %v, want it to wrap the last error", err)
	}
}

func TestPollCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	err := Poll(ctx, time.Second, time.Millisecond, func(context.Context) error {
		calls++
		return errNotYet
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Poll() = %v, want %v", err, context.Canceled)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestPollCancelledWhileWaiting(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := Poll(ctx, time.Minute, time.Second, func(context.Context) error {
		return errNotYet
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Poll() = %v, want %v", err, context.DeadlineExceeded)
	}
}
