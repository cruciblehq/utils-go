// Package codec provides format-neutral encoding and decoding.
//
// Struct fields are mapped by a struct tag, "codec" by default ([New] binds
// another one). The codec package converts between Go values and byte
// representations in JSON or YAML using mapstructure as the struct-to-map
// bridge and standard library encoders for the final byte format.
//
// [Encode] serializes a Go value to bytes. If the value implements [Encodable],
// its Encode method produces the intermediate representation; otherwise
// [Codec.ToMap] converts the struct to a map using struct tags. The result is
// then written in the requested format.
//
//	data, err := codec.Encode(v, codec.YAML)
//
// [Decode] parses bytes (JSON or YAML) into a map, then populates a Go struct
// with [Codec.FromMap], which walks the type tree, applying weak type coercion
// (for example, a string into an int) and tag-declared defaults for absent
// fields. JSON carries []byte values as base64 strings, so decoding JSON turns
// them back into bytes.
//
//	err := codec.Decode(data, &v, codec.YAML)
//
// Types that need custom logic implement [Encodable] or [Decodable]. When
// decoding meets a [Decodable] type at any depth, it calls Decode with the
// active codec and the raw value instead of mapping fields automatically. The
// raw value is typically a map[string]any but may be a string or other scalar.
// Use the passed codec's [Codec.Field] to decode individual fields one at a
// time, retaining coercion and defaults, and [Codec.FromMap] on nested structs
// of different types. Encode hooks use [Codec.ToMap] the same way.
//
//	func (m *MyType) Decode(c *codec.Codec, raw any) error {
//	    src := raw.(map[string]any)
//	    c.Field(src, m, "Name")
//	    c.Field(src, m, "Version")
//	    // custom dispatch or iteration for remaining fields
//	    return nil
//	}
//
// Struct fields may declare a default value in their tag:
//
//	Weight uint16 `codec:"weight" default:"100"`
//
// Defaults are applied only when the field's key is absent from the source
// map. Explicitly provided values, including zero, are never overwritten.
//
// Types implementing [Validatable] check their own invariants. Validation is
// always a separate step from decoding, allowing partial or invalid data to be
// decoded first and validated later.
package codec
