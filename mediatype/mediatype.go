package mediatype

import (
	"fmt"
	"mime"
	"strings"

	"github.com/cruciblehq/utils-go/codec"
)

// Media type without a structured syntax suffix.
type MediaType string

// Returns the media type as a string.
func (m MediaType) String() string {
	return string(m)
}

// Returns the media type with the structured syntax suffix of format f.
//
// The suffix is the name of the format, so JSON gives "+json" and YAML gives
// "+yaml".
func (m MediaType) With(f codec.Format) string {
	return fmt.Sprintf("%s+%s", m, f)
}

// Whether header names the media type expected.
//
// header is a Content-Type value. Parameters such as charset are ignored, and
// the comparison ignores letter case. A malformed header never matches.
func Matches(header, expected string) bool {
	mt, _, err := mime.ParseMediaType(header)
	return err == nil && strings.EqualFold(mt, expected)
}
