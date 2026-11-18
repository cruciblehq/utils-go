package httpx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/mediatype"
)

const testDataType = mediatype.MediaType("application/vnd.test.data.v1")

type failingReader struct {
	err error
}

func (f failingReader) Read([]byte) (int, error) {
	return 0, f.err
}

func TestMultipartRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	mw := NewMultipartWriter(&buf, testType, codec.JSON)
	if err := mw.Write(payload{Name: "a"}, testDataType, strings.NewReader("archive")); err != nil {
		t.Fatalf("Write() = %v, want nil", err)
	}

	var p payload
	data, err := ReadMultipart(mw.ContentType(), &buf, testType, codec.JSON, &p, testDataType)
	if err != nil {
		t.Fatalf("ReadMultipart() = %v, want nil", err)
	}
	if p.Name != "a" {
		t.Errorf("Name = %q, want %q", p.Name, "a")
	}
	got, err := io.ReadAll(data)
	if err != nil {
		t.Fatalf("read data: %v", err)
	}
	if string(got) != "archive" {
		t.Errorf("data = %q, want %q", got, "archive")
	}
}

func TestMultipartEmptyData(t *testing.T) {
	var buf bytes.Buffer
	mw := NewMultipartWriter(&buf, testType, codec.JSON)
	if err := mw.Write(payload{}, testDataType, strings.NewReader("")); err != nil {
		t.Fatalf("Write() = %v, want nil", err)
	}

	var p payload
	data, err := ReadMultipart(mw.ContentType(), &buf, testType, codec.JSON, &p, testDataType)
	if err != nil {
		t.Fatalf("ReadMultipart() = %v, want nil", err)
	}
	got, _ := io.ReadAll(data)
	if len(got) != 0 {
		t.Errorf("data = %q, want empty", got)
	}
}

func TestMultipartWriteDataFailsFirst(t *testing.T) {
	var buf bytes.Buffer
	want := errors.New("no such image")

	err := NewMultipartWriter(&buf, testType, codec.JSON).Write(payload{}, testDataType, failingReader{err: want})
	if !errors.Is(err, want) {
		t.Fatalf("Write() = %v, want %v", err, want)
	}
	if buf.Len() != 0 {
		t.Errorf("wrote %d bytes, want 0", buf.Len())
	}
}

func TestReadMultipartNotMultipart(t *testing.T) {
	var p payload
	_, err := ReadMultipart(testType.With(codec.JSON), strings.NewReader(`{}`), testType, codec.JSON, &p, testDataType)
	if !errors.Is(err, ErrUnexpectedContentType) {
		t.Errorf("ReadMultipart() = %v, want %v", err, ErrUnexpectedContentType)
	}
}

func TestReadMultipartWrongMessageType(t *testing.T) {
	var buf bytes.Buffer
	mw := NewMultipartWriter(&buf, "application/vnd.test.other.v1", codec.JSON)
	if err := mw.Write(payload{}, testDataType, strings.NewReader("x")); err != nil {
		t.Fatalf("Write() = %v, want nil", err)
	}

	var p payload
	_, err := ReadMultipart(mw.ContentType(), &buf, testType, codec.JSON, &p, testDataType)
	if !errors.Is(err, ErrUnexpectedContentType) {
		t.Errorf("ReadMultipart() = %v, want %v", err, ErrUnexpectedContentType)
	}
}

func TestReadMultipartMessagePartMismatch(t *testing.T) {
	body := "--b\r\nContent-Type: application/json\r\n\r\n{}\r\n--b--\r\n"
	ct := fmt.Sprintf("multipart/related; type=%q; boundary=b", testType.With(codec.JSON))

	var p payload
	_, err := ReadMultipart(ct, strings.NewReader(body), testType, codec.JSON, &p, testDataType)
	if !errors.Is(err, ErrUnexpectedContentType) {
		t.Errorf("ReadMultipart() = %v, want %v", err, ErrUnexpectedContentType)
	}
}

func TestReadMultipartWrongDataType(t *testing.T) {
	var buf bytes.Buffer
	mw := NewMultipartWriter(&buf, testType, codec.JSON)
	if err := mw.Write(payload{}, "application/vnd.test.other.v1", strings.NewReader("x")); err != nil {
		t.Fatalf("Write() = %v, want nil", err)
	}

	var p payload
	_, err := ReadMultipart(mw.ContentType(), &buf, testType, codec.JSON, &p, testDataType)
	if !errors.Is(err, ErrUnexpectedContentType) {
		t.Errorf("ReadMultipart() = %v, want %v", err, ErrUnexpectedContentType)
	}
}

func TestReadMultipartMissingDataPart(t *testing.T) {
	body := fmt.Sprintf("--b\r\nContent-Type: %s\r\n\r\n{}\r\n--b--\r\n", testType.With(codec.JSON))

	var p payload
	_, err := ReadMultipart(fmt.Sprintf("multipart/related; type=%q; boundary=b", testType.With(codec.JSON)), strings.NewReader(body), testType, codec.JSON, &p, testDataType)
	if !errors.Is(err, ErrInvalidBody) {
		t.Errorf("ReadMultipart() = %v, want %v", err, ErrInvalidBody)
	}
}
