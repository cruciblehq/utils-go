package httpx

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/mediatype"
)

// Decodes the body of r into v after checking its media type.
//
// The Content-Type header must name expected with the suffix of format f (see
// [mediatype.MediaType.With]). Returns [ErrUnexpectedContentType] otherwise,
// without reading the body. Returns [ErrInvalidBody] when the body cannot be
// read or decoded as f.
func DecodeRequest(r *http.Request, expected mediatype.MediaType, f codec.Format, v any) error {
	ct := r.Header.Get(HeaderContentType)
	if mt := expected.With(f); !mediatype.Matches(ct, mt) {
		return crex.Newf(ErrUnexpectedContentType, "request body has media type %q where %q is expected", ct, mt)
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return crex.Wrap(ErrInvalidBody, err)
	}
	if err := codec.Decode(data, v, f); err != nil {
		return crex.Wrap(ErrInvalidBody, err)
	}
	return nil
}

// Writes v as a response encoded in format f, with media type mt and status.
//
// The Content-Type header is mt with the suffix of f. When v cannot be encoded,
// a 500 is written instead and the error returned. Returns the error when the
// body cannot be written.
func Write(w http.ResponseWriter, mt mediatype.MediaType, f codec.Format, status int, v any) error {
	data, err := codec.Encode(v, f)
	if err != nil {
		http.Error(w, "could not encode response", http.StatusInternalServerError)
		return err
	}
	w.Header().Set(HeaderContentType, mt.With(f))
	w.WriteHeader(status)
	_, err = w.Write(data)
	return err
}

// Runs srv until ctx ends, then shuts it down gracefully.
//
// The server uses TLS when srv.TLSConfig is set, with the certificates it
// provides. In-flight requests get up to timeout to finish once ctx ends.
// Returns the error that stopped the server, or nil after a clean shutdown.
func Serve(ctx context.Context, srv *http.Server, timeout time.Duration) error {
	errs := make(chan error, 1)
	go func() {
		if srv.TLSConfig != nil {
			errs <- srv.ListenAndServeTLS("", "")
		} else {
			errs <- srv.ListenAndServe()
		}
	}()

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-errs; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
