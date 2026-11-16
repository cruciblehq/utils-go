package httpx

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/mediatype"
)

type payload struct {
	Name string `codec:"name"`
}

const testType = mediatype.MediaType("application/vnd.test.payload.v1")

func TestDecodeRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"a"}`))
	r.Header.Set(HeaderContentType, testType.With(codec.JSON))

	var p payload
	if err := DecodeRequest(r, testType, codec.JSON, &p); err != nil {
		t.Fatalf("DecodeRequest() = %v, want nil", err)
	}
	if p.Name != "a" {
		t.Errorf("Name = %q, want %q", p.Name, "a")
	}
}

func TestDecodeRequestWrongType(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"a"}`))
	r.Header.Set(HeaderContentType, "application/json")

	var p payload
	if err := DecodeRequest(r, testType, codec.JSON, &p); !errors.Is(err, ErrUnexpectedContentType) {
		t.Errorf("DecodeRequest() = %v, want %v", err, ErrUnexpectedContentType)
	}
}

func TestDecodeRequestInvalidBody(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{bad`))
	r.Header.Set(HeaderContentType, testType.With(codec.JSON))

	var p payload
	if err := DecodeRequest(r, testType, codec.JSON, &p); !errors.Is(err, ErrInvalidBody) {
		t.Errorf("DecodeRequest() = %v, want %v", err, ErrInvalidBody)
	}
}

func TestWrite(t *testing.T) {
	w := httptest.NewRecorder()

	if err := Write(w, testType, codec.JSON, http.StatusCreated, payload{Name: "a"}); err != nil {
		t.Fatalf("Write() = %v, want nil", err)
	}
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
	if ct := w.Header().Get(HeaderContentType); ct != testType.With(codec.JSON) {
		t.Errorf("Content-Type = %q, want %q", ct, testType.With(codec.JSON))
	}
	var p payload
	if err := codec.Decode(w.Body.Bytes(), &p, codec.JSON); err != nil || p.Name != "a" {
		t.Errorf("body = %q, want a payload named a", w.Body.String())
	}
}

func TestWriteUnencodable(t *testing.T) {
	w := httptest.NewRecorder()

	if err := Write(w, testType, codec.JSON, http.StatusOK, func() {}); err == nil {
		t.Fatal("Write() of a function = nil, want an error")
	}
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr
}

func waitForServer(t *testing.T, client *http.Client, url string) {
	t.Helper()
	for range 100 {
		if resp, err := client.Get(url); err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("server at %s did not start", url)
}

func TestServe(t *testing.T) {
	addr := freeAddr(t)
	srv := &http.Server{Addr: addr, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, srv, time.Second)
	}()

	waitForServer(t, http.DefaultClient, fmt.Sprintf("http://%s", addr))
	cancel()
	if err := <-done; err != nil {
		t.Errorf("Serve() = %v, want nil", err)
	}
}

func TestServeTLS(t *testing.T) {
	ts := httptest.NewTLSServer(http.NotFoundHandler())
	cert := ts.TLS.Certificates[0]
	client := ts.Client()
	ts.Close()

	addr := freeAddr(t)
	srv := &http.Server{
		Addr:      addr,
		Handler:   http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, srv, time.Second)
	}()

	waitForServer(t, client, fmt.Sprintf("https://%s", addr))
	cancel()
	if err := <-done; err != nil {
		t.Errorf("Serve() = %v, want nil", err)
	}
}

func TestServeListenError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	srv := &http.Server{Addr: ln.Addr().String()}

	if err := Serve(context.Background(), srv, time.Second); err == nil {
		t.Error("Serve() on a used port = nil, want an error")
	}
}

func TestDecodeRequestYAML(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("name: a\n"))
	r.Header.Set(HeaderContentType, testType.With(codec.YAML))

	var p payload
	if err := DecodeRequest(r, testType, codec.YAML, &p); err != nil {
		t.Fatalf("DecodeRequest() = %v, want nil", err)
	}
	if p.Name != "a" {
		t.Errorf("Name = %q, want %q", p.Name, "a")
	}
}

func TestDecodeRequestFormatMismatch(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"a"}`))
	r.Header.Set(HeaderContentType, testType.With(codec.JSON))

	var p payload
	if err := DecodeRequest(r, testType, codec.YAML, &p); !errors.Is(err, ErrUnexpectedContentType) {
		t.Errorf("DecodeRequest() = %v, want %v", err, ErrUnexpectedContentType)
	}
}
