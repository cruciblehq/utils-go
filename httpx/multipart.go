package httpx

import (
	"bufio"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"

	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/mediatype"
)

// Media type of a body made of a message part followed by a data part.
const multipartRelated = "multipart/related"

// Parameters of the multipart media type.
const (
	boundaryParam = "boundary" // Separator between the parts.
	typeParam     = "type"     // Media type of the message part.
)

// Writes a body made of a message part followed by a data part.
//
// The body is multipart/related (RFC 2387). The message part comes first and
// holds a value encoded in a codec format. Its media type is also the type
// parameter of the body, so the Content-Type alone names the message. The data
// part holds a stream, such as an archive.
type MultipartWriter struct {
	w  *multipart.Writer   // Writer of the parts, holding the boundary.
	mt mediatype.MediaType // Media type of the message part.
	f  codec.Format        // Format the message part is encoded in.
}

// Returns a writer to w of a multipart body whose message has media type mt
// and is encoded in f.
func NewMultipartWriter(w io.Writer, mt mediatype.MediaType, f codec.Format) *MultipartWriter {
	return &MultipartWriter{w: multipart.NewWriter(w), mt: mt, f: f}
}

// Returns the Content-Type of the body, including its message type and boundary.
func (m *MultipartWriter) ContentType() string {
	return mime.FormatMediaType(multipartRelated, map[string]string{
		typeParam:     m.mt.With(m.f),
		boundaryParam: m.w.Boundary(),
	})
}

// Writes v as the message part, then data as the data part, and ends the body.
//
// The data part has media type dataType. Nothing is written until data yields
// its first byte or ends, so when encoding v or reading data fails at once,
// the underlying writer is left untouched.
func (m *MultipartWriter) Write(v any, dataType mediatype.MediaType, data io.Reader) error {
	msg, err := codec.Encode(v, m.f)
	if err != nil {
		return err
	}
	br := bufio.NewReader(data)
	if _, err := br.Peek(1); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	part, err := m.w.CreatePart(textproto.MIMEHeader{HeaderContentType: {m.mt.With(m.f)}})
	if err != nil {
		return err
	}
	if _, err := part.Write(msg); err != nil {
		return err
	}
	part, err = m.w.CreatePart(textproto.MIMEHeader{HeaderContentType: {dataType.String()}})
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, br); err != nil {
		return err
	}
	return m.w.Close()
}

// Decodes the message part of a multipart body into v and returns the data part.
//
// contentType is the Content-Type of the body. Its type parameter and the
// message part must both have media type mt with the suffix of f, and the data
// part must have media type dataType. Returns [ErrUnexpectedContentType] when
// a media type does not match, and [ErrInvalidBody] when the body is not a
// valid multipart body. The returned reader reads the data part from body, so
// body must stay open until the data part is read.
func ReadMultipart(contentType string, body io.Reader, mt mediatype.MediaType, f codec.Format, v any, dataType mediatype.MediaType) (io.Reader, error) {
	kind, params, err := mime.ParseMediaType(contentType)
	if err != nil || kind != multipartRelated || params[boundaryParam] == "" {
		return nil, crex.Newf(ErrUnexpectedContentType, "%q is not a multipart body", contentType)
	}
	expected := mt.With(f)
	if !mediatype.Matches(params[typeParam], expected) {
		return nil, crex.Newf(ErrUnexpectedContentType, "multipart body declares message type %q where %q is expected", params[typeParam], expected)
	}
	r := multipart.NewReader(body, params[boundaryParam])

	part, err := r.NextRawPart()
	if err != nil {
		return nil, crex.Wrap(ErrInvalidBody, err)
	}
	if ct := part.Header.Get(HeaderContentType); !mediatype.Matches(ct, expected) {
		return nil, crex.Newf(ErrUnexpectedContentType, "message part has media type %q where %q is expected", ct, expected)
	}
	msg, err := io.ReadAll(part)
	if err != nil {
		return nil, crex.Wrap(ErrInvalidBody, err)
	}
	if err := codec.Decode(msg, v, f); err != nil {
		return nil, crex.Wrap(ErrInvalidBody, err)
	}

	part, err = r.NextRawPart()
	if err != nil {
		return nil, crex.Wrap(ErrInvalidBody, err)
	}
	if ct := part.Header.Get(HeaderContentType); !mediatype.Matches(ct, dataType.String()) {
		return nil, crex.Newf(ErrUnexpectedContentType, "data part has media type %q where %q is expected", ct, dataType)
	}
	return part, nil
}
