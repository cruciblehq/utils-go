package file

import "github.com/cruciblehq/utils-go/crex"

var (
	ErrNotSupported  = crex.New("file locking is not supported on this platform")
	ErrInvalidPath   = crex.New("invalid path")
	ErrUnsafeTempDir = crex.New("unsafe temporary directory")
)
