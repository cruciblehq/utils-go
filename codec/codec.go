package codec

// Encodes and decodes values using a named struct tag.
//
// The tag selects which struct fields map to serialized keys. [Default] uses
// the "codec" tag; [New] binds a different one.
type Codec struct {
	tag         string // Struct tag used to map fields.
	base64Bytes bool   // Whether strings decode into []byte fields as base64.
}

// Returns a Codec bound to the default "codec" struct tag.
func Default() *Codec {
	return &Codec{tag: defaultCodecTag}
}

// Returns a Codec bound to the given struct tag.
//
// An empty tag falls back to the default "codec" tag.
func New(tag string) *Codec {
	if tag == "" {
		tag = defaultCodecTag
	}
	return &Codec{tag: tag}
}

// Converts v to bytes in the given format using the default codec tag.
func Encode(v any, f Format) ([]byte, error) {
	return Default().Encode(v, f)
}

// Populates v from data in the given format using the default codec tag.
func Decode(data []byte, v any, f Format) error {
	return Default().Decode(data, v, f)
}
