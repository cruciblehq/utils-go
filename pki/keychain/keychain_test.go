package keychain

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/cruciblehq/utils-go/pki"
	"github.com/zalando/go-keyring"
)

func TestStore(t *testing.T) {
	keyring.MockInit()
	s := New("test")
	_, key, _ := ed25519.GenerateKey(rand.Reader)

	if err := s.Store("root", key); err != nil {
		t.Fatalf("Store() = %v, want nil", err)
	}
	got, err := s.Load("root")
	if err != nil {
		t.Fatalf("Load() = %v, want nil", err)
	}
	if !key.Equal(got) {
		t.Error("Load() returned a different key")
	}
}

func TestStoreSeparatesServices(t *testing.T) {
	keyring.MockInit()
	_, key, _ := ed25519.GenerateKey(rand.Reader)
	if err := New("a").Store("root", key); err != nil {
		t.Fatal(err)
	}

	if _, err := New("b").Load("root"); !errors.Is(err, pki.ErrKeyNotFound) {
		t.Errorf("Load() from another service = %v, want %v", err, pki.ErrKeyNotFound)
	}
}

func TestLoadNotFound(t *testing.T) {
	keyring.MockInit()

	if _, err := New("test").Load("missing"); !errors.Is(err, pki.ErrKeyNotFound) {
		t.Errorf("Load() = %v, want %v", err, pki.ErrKeyNotFound)
	}
}

func TestLoadCorrupt(t *testing.T) {
	keyring.MockInit()
	if err := keyring.Set("test", "bad", "not base64!"); err != nil {
		t.Fatal(err)
	}

	if _, err := New("test").Load("bad"); !errors.Is(err, pki.ErrInvalidKey) {
		t.Errorf("Load() = %v, want %v", err, pki.ErrInvalidKey)
	}
}

func TestKeychainUnavailable(t *testing.T) {
	keyring.MockInitWithError(errors.New("locked"))
	_, key, _ := ed25519.GenerateKey(rand.Reader)

	if err := New("test").Store("root", key); !errors.Is(err, ErrKeychain) {
		t.Errorf("Store() = %v, want %v", err, ErrKeychain)
	}
	if _, err := New("test").Load("root"); !errors.Is(err, ErrKeychain) {
		t.Errorf("Load() = %v, want %v", err, ErrKeychain)
	}
}
