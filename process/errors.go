package process

import "github.com/cruciblehq/utils-go/crex"

var (
	ErrStart = crex.New("command could not start")
	ErrExit  = crex.New("command failed")
)
