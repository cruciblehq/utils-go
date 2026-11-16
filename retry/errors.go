package retry

import "github.com/cruciblehq/utils-go/crex"

var ErrTimeout = crex.New("operation did not succeed in time")
