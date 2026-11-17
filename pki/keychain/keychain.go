package keychain

import (
	"crypto"
	"encoding/base64"
	"errors"

	"github.com/cruciblehq/utils-go/crex"
	"github.com/cruciblehq/utils-go/pki"
	"github.com/zalando/go-keyring"
)

// Key store backed by the OS keychain.
//
// Each key is an entry under service, with the key name as its account. Keys
// are stored in PKCS #8 DER form, base64 encoded.
type Store struct {
	service string // Keychain service name shared by all entries.
}

// Returns a key store that keeps its entries under service.
func New(service string) *Store {
	return &Store{service: service}
}

// Returns the key stored under name, or [pki.ErrKeyNotFound].
func (s *Store) Load(name string) (crypto.Signer, error) {
	encoded, err := keyring.Get(s.service, name)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, crex.Newf(pki.ErrKeyNotFound, "no key named %q in keychain service %q", name, s.service)
	}
	if err != nil {
		return nil, crex.Wrap(ErrKeychain, err)
	}
	der, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, crex.Wrap(pki.ErrInvalidKey, err)
	}
	return pki.DecodeKey(der)
}

// Stores key under name, replacing any previous key.
func (s *Store) Store(name string, key crypto.Signer) error {
	der, err := pki.EncodeKey(key)
	if err != nil {
		return err
	}
	if err := keyring.Set(s.service, name, base64.StdEncoding.EncodeToString(der)); err != nil {
		return crex.Wrap(ErrKeychain, err)
	}
	return nil
}
