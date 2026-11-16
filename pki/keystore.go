package pki

import (
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/file"
)

// Extension of key files in a [DirKeyStore].
const keyExtension = ".key"

// Place where private keys are kept by name.
type KeyStore interface {
	Load(name string) (crypto.Signer, error)    // Returns the key stored under name, or [ErrKeyNotFound].
	Store(name string, key crypto.Signer) error // Stores key under name, replacing any previous key.
}

// Key store that keeps each key as a file in a directory.
//
// The directory and the files are readable only by their owner. Keys are
// written atomically in PKCS #8 DER form.
type DirKeyStore struct {
	dir string // Directory holding the key files.
}

// Returns a key store backed by dir.
//
// dir is created on the first [DirKeyStore.Store].
func NewDirKeyStore(dir string) *DirKeyStore {
	return &DirKeyStore{dir: dir}
}

// Returns the key stored under name, or [ErrKeyNotFound].
func (s *DirKeyStore) Load(name string) (crypto.Signer, error) {
	path, err := s.path(name)
	if err != nil {
		return nil, err
	}
	der, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, crex.Newf(ErrKeyNotFound, "no key named %q in %q", name, s.dir)
	}
	if err != nil {
		return nil, err
	}
	return ParseKey(der)
}

// Stores key under name, replacing any previous key.
func (s *DirKeyStore) Store(name string, key crypto.Signer) error {
	path, err := s.path(name)
	if err != nil {
		return err
	}
	der, err := MarshalKey(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.dir, file.OwnerDirMode); err != nil {
		return err
	}
	return file.WriteFile(path, der, file.OwnerFileMode)
}

// Returns the file holding the key named name.
func (s *DirKeyStore) path(name string) (string, error) {
	if name == "" || !filepath.IsLocal(name) || filepath.Base(name) != name {
		return "", crex.Newf(ErrInvalidKeyName, "%q is not a plain file name", name)
	}
	return filepath.Join(s.dir, fmt.Sprintf("%s%s", name, keyExtension)), nil
}

// Encodes key in PKCS #8 DER form.
func MarshalKey(key crypto.Signer) ([]byte, error) {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, crex.Wrap(ErrInvalidKey, err)
	}
	return der, nil
}

// Decodes a PKCS #8 DER private key that can sign.
func ParseKey(der []byte) (crypto.Signer, error) {
	parsed, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, crex.Wrap(ErrInvalidKey, err)
	}
	key, ok := parsed.(crypto.Signer)
	if !ok {
		return nil, crex.Newf(ErrInvalidKey, "unexpected key type %T", parsed)
	}
	return key, nil
}
