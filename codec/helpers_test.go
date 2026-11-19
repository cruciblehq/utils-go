package codec

type sample struct {
	Name    string `codec:"name"`    // Name string.
	Version int    `codec:"version"` // Version number.
}

type nested struct {
	Inner sample `codec:"inner"` // Nested sample value.
}

type squashed struct {
	sample        // Embedded name and version fields.
	Extra  string `codec:"extra"` // Additional field alongside the squashed fields.
}

type custom struct {
	Value string // Encoded or decoded string value.
}

func (c *custom) Encode(_ *Codec) (any, error) {
	return map[string]any{"custom": c.Value}, nil
}

func (c *custom) Decode(_ *Codec, raw any) error {
	m, _ := raw.(map[string]any)
	c.Value, _ = m["custom"].(string)
	return nil
}
