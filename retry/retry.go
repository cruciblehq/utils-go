package retry

import (
	"context"
	"time"

	"github.com/cruciblehq/utils-go/crex"
)

// Calls fn until it returns nil, waiting interval between attempts.
//
// Gives up when waiting another interval would pass timeout, returning
// [ErrTimeout] that wraps the last error fn returned. Returns the context's
// error when ctx ends first. fn is always called at least once.
func Poll(ctx context.Context, timeout, interval time.Duration, fn func(context.Context) error) error {
	deadline := time.Now().Add(timeout)
	for {
		err := fn(ctx)
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if time.Now().Add(interval).After(deadline) {
			return crex.Wrapf(ErrTimeout, err, "no success within %s", timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}
