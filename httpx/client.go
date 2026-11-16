package httpx

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/cruciblehq/utils-go/crex"
)

// Largest part of a failed response body kept for an error message.
const maxErrorBody = 4096

// Largest structured error body a client should read.
//
// Wrap a failed response's body in [io.LimitReader] with this size before
// decoding it, so a misbehaving server cannot make the client read without end.
const MaxErrorPayload = 1 << 20

// Builds a request for path under base with the given query and body.
//
// base is the scheme and host, optionally with a path that path is joined
// to. query may be nil.
func NewRequest(ctx context.Context, method, base, path string, query url.Values, body io.Reader) (*http.Request, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, crex.Wrap(ErrInvalidURL, err)
	}
	u = u.JoinPath(path)
	u.RawQuery = query.Encode()
	return http.NewRequestWithContext(ctx, method, u.String(), body)
}

// Returns the start of a failed response's body, trimmed, for an error message.
//
// Reads at most 4 KiB, so a large body cannot exhaust memory.
func ErrorBody(resp *http.Response) string {
	data, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
	return strings.TrimSpace(string(data))
}
