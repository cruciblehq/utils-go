package pki

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/cruciblehq/utils-go/file"
)

func TestDirKeyStore(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "keys")
	s := NewDirKeyStore(dir)
	_, key, _ := ed25519.GenerateKey(rand.Reader)

	if err := s.Store("machine", key); err != nil {
		t.Fatalf("Store() = %v, want nil", err)
	}
	got, err := s.Load("machine")
	if err != nil {
		t.Fatalf("Load() = %v, want nil", err)
	}
	if !key.Equal(got) {
		t.Error("Load() returned a different key")
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Join(dir, "machine.key"))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != file.OwnerFileMode {
			t.Errorf("key mode = %v, want %v", info.Mode().Perm(), file.OwnerFileMode)
		}
		dirInfo, err := os.Stat(dir)
		if err != nil {
			t.Fatal(err)
		}
		if dirInfo.Mode().Perm() != file.OwnerDirMode {
			t.Errorf("dir mode = %v, want %v", dirInfo.Mode().Perm(), file.OwnerDirMode)
		}
	}
}

func TestDirKeyStoreReplace(t *testing.T) {
	s := NewDirKeyStore(t.TempDir())
	_, first, _ := ed25519.GenerateKey(rand.Reader)
	_, second, _ := ed25519.GenerateKey(rand.Reader)
	if err := s.Store("k", first); err != nil {
		t.Fatal(err)
	}

	if err := s.Store("k", second); err != nil {
		t.Fatalf("Store() = %v, want nil", err)
	}
	got, err := s.Load("k")
	if err != nil {
		t.Fatal(err)
	}
	if !second.Equal(got) {
		t.Error("Load() after replace returned the old key")
	}
}

func TestDirKeyStoreNotFound(t *testing.T) {
	if _, err := NewDirKeyStore(t.TempDir()).Load("missing"); !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("Load() = %v, want %v", err, ErrKeyNotFound)
	}
}

func TestDirKeyStoreInvalidName(t *testing.T) {
	s := NewDirKeyStore(t.TempDir())
	_, key, _ := ed25519.GenerateKey(rand.Reader)

	if err := s.Store("../escape", key); !errors.Is(err, ErrInvalidKeyName) {
		t.Errorf("Store(../escape) = %v, want %v", err, ErrInvalidKeyName)
	}
	if _, err := s.Load(""); !errors.Is(err, ErrInvalidKeyName) {
		t.Errorf("Load(empty) = %v, want %v", err, ErrInvalidKeyName)
	}
}

func TestDirKeyStoreCorruptKey(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.key"), []byte("garbage"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := NewDirKeyStore(dir).Load("bad"); !errors.Is(err, ErrInvalidKey) {
		t.Errorf("Load() = %v, want %v", err, ErrInvalidKey)
	}
}

func TestMarshalParseKey(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	der, err := MarshalKey(key)
	if err != nil {
		t.Fatalf("MarshalKey() = %v, want nil", err)
	}
	got, err := ParseKey(der)
	if err != nil {
		t.Fatalf("ParseKey() = %v, want nil", err)
	}
	if !key.Equal(got) {
		t.Error("ParseKey() returned a different key")
	}
}

func TestParseKeyInvalid(t *testing.T) {
	if _, err := ParseKey([]byte("garbage")); !errors.Is(err, ErrInvalidKey) {
		t.Errorf("ParseKey() = %v, want %v", err, ErrInvalidKey)
	}
}
