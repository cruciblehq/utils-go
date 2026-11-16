package httpx

import "github.com/cruciblehq/utils-go/crex"

var (
	ErrUnexpectedContentType = crex.New("unexpected content type")
	ErrInvalidBody           = crex.New("invalid message body")
	ErrInvalidURL            = crex.New("invalid URL")
)
