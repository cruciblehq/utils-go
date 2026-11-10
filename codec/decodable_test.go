package codec

import (
	"errors"
	"testing"
)

func TestUnmarshal_JSON(t *testing.T) {
	var s sample
	err := Unmarshal([]byte(`{"name":"bar","version":2}`), &s, JSON)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "bar" || s.Version != 2 {
		t.Errorf("Unmarshal(JSON) = %+v", s)
	}
}

func TestUnmarshal_YAML(t *testing.T) {
	var s sample
	err := Unmarshal([]byte("name: baz\nversion: 3\n"), &s, YAML)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "baz" || s.Version != 3 {
		t.Errorf("Unmarshal(YAML) = %+v", s)
	}
}

func TestUnmarshal_UnsupportedFormat(t *testing.T) {
	var s sample
	err := Unmarshal([]byte("{}"), &s, Format(99))
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("Unmarshal(Format(99)) error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestUnmarshal_InvalidData(t *testing.T) {
	var s sample
	err := Unmarshal([]byte("not json"), &s, JSON)
	if err == nil {
		t.Error("Unmarshal(invalid JSON) should fail")
	}
}

func TestUnmarshal_CustomDecodable(t *testing.T) {
	c := &custom{}
	err := Unmarshal([]byte(`{"custom":"raw-input"}`), c, JSON)
	if err != nil {
		t.Fatal(err)
	}
	if c.Value != "raw-input" {
		t.Errorf("Unmarshal(custom) = %q", c.Value)
	}
}

func TestDecode(t *testing.T) {
	m := map[string]any{"name": "y", "version": 10}
	var s sample
	if err := Decode(m, &s); err != nil {
		t.Fatal(err)
	}
	if s.Name != "y" || s.Version != 10 {
		t.Errorf("Decode = %+v", s)
	}
}

func TestDecode_WeaklyTyped(t *testing.T) {
	m := map[string]any{"name": "w", "version": "8"}
	var s sample
	if err := Decode(m, &s); err != nil {
		t.Fatal(err)
	}
	if s.Version != 8 {
		t.Errorf("Decode weakly typed version = %d, want 8", s.Version)
	}
}

func TestDecode_Squash(t *testing.T) {
	m := map[string]any{"name": "sq", "version": 7, "extra": "e"}
	var s squashed
	if err := Decode(m, &s); err != nil {
		t.Fatal(err)
	}
	if s.Name != "sq" {
		t.Errorf("Decode squash name = %q, want %q", s.Name, "sq")
	}
	if s.Version != 7 {
		t.Errorf("Decode squash version = %d, want 7", s.Version)
	}
	if s.Extra != "e" {
		t.Errorf("Decode squash extra = %q, want %q", s.Extra, "e")
	}
}

func TestField(t *testing.T) {
	src := map[string]any{"name": "alice", "version": 7}
	var s sample
	if err := Field(src, &s, "Name"); err != nil {
		t.Fatal(err)
	}
	if err := Field(src, &s, "Version"); err != nil {
		t.Fatal(err)
	}
	if s.Name != "alice" {
		t.Errorf("Field Name = %q, want %q", s.Name, "alice")
	}
	if s.Version != 7 {
		t.Errorf("Field Version = %d, want 7", s.Version)
	}
}

func TestField_WeaklyTyped(t *testing.T) {
	src := map[string]any{"version": "42"}
	var s sample
	if err := Field(src, &s, "Version"); err != nil {
		t.Fatal(err)
	}
	if s.Version != 42 {
		t.Errorf("Field Version = %d, want 42", s.Version)
	}
}

func TestField_Default(t *testing.T) {
	src := map[string]any{}
	var s withDefaults
	if err := Field(src, &s, "Weight"); err != nil {
		t.Fatal(err)
	}
	if s.Weight != 100 {
		t.Errorf("Field Weight = %d, want 100", s.Weight)
	}
}

func TestField_Nested(t *testing.T) {
	src := map[string]any{"inner": map[string]any{"name": "n", "version": 3}}
	var s nested
	if err := Field(src, &s, "Inner"); err != nil {
		t.Fatal(err)
	}
	if s.Inner.Name != "n" || s.Inner.Version != 3 {
		t.Errorf("Field Inner = %+v", s.Inner)
	}
}

func TestField_Absent(t *testing.T) {
	src := map[string]any{}
	var s sample
	if err := Field(src, &s, "Name"); err != nil {
		t.Fatal(err)
	}
	if s.Name != "" {
		t.Errorf("Field absent Name = %q, want empty", s.Name)
	}
}

func TestField_NotPointer(t *testing.T) {
	src := map[string]any{"name": "x"}
	var s sample
	err := Field(src, s, "Name")
	if err == nil {
		t.Fatal("Field(non-pointer) should fail")
	}
}

func TestField_BadFieldName(t *testing.T) {
	src := map[string]any{}
	var s sample
	err := Field(src, &s, "DoesNotExist")
	if err == nil {
		t.Fatal("Field(bad name) should fail")
	}
}

func TestField_Decodable(t *testing.T) {
	type wrapper struct {
		C custom `codec:"c"`
	}
	src := map[string]any{"c": map[string]any{"custom": "hooked"}}
	var w wrapper
	if err := Field(src, &w, "C"); err != nil {
		t.Fatal(err)
	}
	if w.C.Value != "hooked" {
		t.Errorf("Field Decodable = %q, want %q", w.C.Value, "hooked")
	}
}

type payload struct {
	Data []byte `codec:"data"`
}

type payloadHolder struct {
	Inner payload
}

func (p *payloadHolder) Decode(c *Codec, raw any) error {
	return c.Decode(raw.(map[string]any), &p.Inner)
}

func TestUnmarshal_JSONBytes(t *testing.T) {
	want := []byte{0xff, 0x00, 'h', '\n'}
	data, err := Encode(payload{Data: want}, JSON)
	if err != nil {
		t.Fatal(err)
	}
	var got payload
	if err := Unmarshal(data, &got, JSON); err != nil {
		t.Fatal(err)
	}
	if string(got.Data) != string(want) {
		t.Errorf("Unmarshal(JSON) data = %q, want %q", got.Data, want)
	}
}

func TestUnmarshal_JSONBytesInvalidBase64(t *testing.T) {
	var got payload
	err := Unmarshal([]byte(`{"data":"not base64!"}`), &got, JSON)
	if !errors.Is(err, ErrInvalidBase64) {
		t.Errorf("Unmarshal(JSON) error = %v, want ErrInvalidBase64", err)
	}
}

func TestUnmarshal_JSONBytesInDecodable(t *testing.T) {
	want := []byte("nested")
	data, err := Encode(payload{Data: want}, JSON)
	if err != nil {
		t.Fatal(err)
	}
	var got payloadHolder
	if err := Unmarshal(data, &got, JSON); err != nil {
		t.Fatal(err)
	}
	if string(got.Inner.Data) != string(want) {
		t.Errorf("Unmarshal(JSON) nested data = %q, want %q", got.Inner.Data, want)
	}
}

func TestUnmarshal_YAMLBytes(t *testing.T) {
	want := []byte{0xff, 0x00, 'h'}
	data, err := Encode(payload{Data: want}, YAML)
	if err != nil {
		t.Fatal(err)
	}
	var got payload
	if err := Unmarshal(data, &got, YAML); err != nil {
		t.Fatal(err)
	}
	if string(got.Data) != string(want) {
		t.Errorf("Unmarshal(YAML) data = %q, want %q", got.Data, want)
	}
}

func TestDecode_StringIntoBytesUnchanged(t *testing.T) {
	var got payload
	if err := Decode(map[string]any{"data": "raw"}, &got); err != nil {
		t.Fatal(err)
	}
	if string(got.Data) != "raw" {
		t.Errorf("Decode data = %q, want %q", got.Data, "raw")
	}
}
