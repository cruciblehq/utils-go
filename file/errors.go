package file

import "github.com/cruciblehq/utils-go/crex"

var (
	ErrNotSupported = crex.New("file locking is not supported on this platform")
)
