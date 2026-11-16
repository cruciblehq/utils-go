package pki

import "github.com/cruciblehq/utils-go/crex"

var (
	ErrKeyNotFound    = crex.New("key not found")
	ErrInvalidKey     = crex.New("invalid key")
	ErrInvalidKeyName = crex.New("invalid key name")
)
