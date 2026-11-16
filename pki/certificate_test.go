package pki

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newAuthority(t *testing.T) (*x509.Certificate, ed25519.PrivateKey) {
	t.Helper()
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := Sign(AuthorityTemplate("root", ValidFor(time.Hour), false), nil, pub, key)
	if err != nil {
		t.Fatal(err)
	}
	return cert, key
}

func TestValidFor(t *testing.T) {
	before := time.Now()
	v := ValidFor(time.Hour)

	if !v.NotBefore.Before(before) {
		t.Errorf("NotBefore = %v, want before %v", v.NotBefore, before)
	}
	if v.NotAfter.Before(before.Add(time.Hour)) {
		t.Errorf("NotAfter = %v, want at least %v", v.NotAfter, before.Add(time.Hour))
	}
}

func TestValidUntil(t *testing.T) {
	end := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	v := ValidUntil(end)

	if !v.NotAfter.Equal(end) {
		t.Errorf("NotAfter = %v, want %v", v.NotAfter, end)
	}
	if !v.NotBefore.Before(time.Now()) {
		t.Errorf("NotBefore = %v, want before now", v.NotBefore)
	}
}

func TestSerialNumber(t *testing.T) {
	a, err := SerialNumber()
	if err != nil {
		t.Fatal(err)
	}
	b, err := SerialNumber()
	if err != nil {
		t.Fatal(err)
	}
	if a.Cmp(b) == 0 {
		t.Error("SerialNumber() returned the same value twice, want different")
	}
	if a.BitLen() > serialBits {
		t.Errorf("BitLen() = %d, want at most %d", a.BitLen(), serialBits)
	}
}

func TestSignSelfSigned(t *testing.T) {
	root, _ := newAuthority(t)

	if !root.IsCA {
		t.Error("IsCA = false, want true")
	}
	if err := root.CheckSignatureFrom(root); err != nil {
		t.Errorf("CheckSignatureFrom(itself) = %v, want nil", err)
	}
	if root.SerialNumber == nil || root.SerialNumber.Sign() <= 0 {
		t.Errorf("SerialNumber = %v, want a positive number", root.SerialNumber)
	}
}

func TestSignChain(t *testing.T) {
	root, rootKey := newAuthority(t)
	issuerPub, issuerKey, _ := ed25519.GenerateKey(rand.Reader)
	issuer, err := Sign(AuthorityTemplate("issuer", ValidFor(time.Hour), true), root, issuerPub, rootKey)
	if err != nil {
		t.Fatal(err)
	}
	leafPub, _, _ := ed25519.GenerateKey(rand.Reader)
	leaf, err := Sign(LeafTemplate("leaf", ValidFor(time.Hour), x509.ExtKeyUsageServerAuth, "example.test"), issuer, leafPub, issuerKey)
	if err != nil {
		t.Fatal(err)
	}

	roots := x509.NewCertPool()
	roots.AddCert(root)
	intermediates := x509.NewCertPool()
	intermediates.AddCert(issuer)
	if _, err := leaf.Verify(x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		DNSName:       "example.test",
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}); err != nil {
		t.Errorf("Verify() = %v, want nil", err)
	}
	if !issuer.MaxPathLenZero {
		t.Error("issuer MaxPathLenZero = false, want true")
	}
}

func TestSignKeepsSerial(t *testing.T) {
	pub, key, _ := ed25519.GenerateKey(rand.Reader)
	tmpl := AuthorityTemplate("root", ValidFor(time.Hour), false)
	serial, err := SerialNumber()
	if err != nil {
		t.Fatal(err)
	}
	tmpl.SerialNumber = serial

	cert, err := Sign(tmpl, nil, pub, key)
	if err != nil {
		t.Fatal(err)
	}
	if cert.SerialNumber.Cmp(serial) != 0 {
		t.Errorf("SerialNumber = %v, want %v", cert.SerialNumber, serial)
	}
}

func TestSignRevocationList(t *testing.T) {
	root, rootKey := newAuthority(t)
	revoked := []x509.RevocationListEntry{{SerialNumber: root.SerialNumber, RevocationTime: time.Now()}}

	crl, err := SignRevocationList(7, ValidFor(time.Hour), revoked, root, rootKey)
	if err != nil {
		t.Fatalf("SignRevocationList() = %v, want nil", err)
	}
	if crl.Number.Int64() != 7 {
		t.Errorf("Number = %s, want 7", crl.Number)
	}
	if err := crl.CheckSignatureFrom(root); err != nil {
		t.Errorf("CheckSignatureFrom() = %v, want nil", err)
	}
	if len(crl.RevokedCertificateEntries) != 1 {
		t.Errorf("len(RevokedCertificateEntries) = %d, want 1", len(crl.RevokedCertificateEntries))
	}
}

func TestReadCertificate(t *testing.T) {
	root, _ := newAuthority(t)
	path := filepath.Join(t.TempDir(), "root.crt")
	if err := os.WriteFile(path, root.Raw, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := ReadCertificate(path)
	if err != nil {
		t.Fatalf("ReadCertificate() = %v, want nil", err)
	}
	if !got.Equal(root) {
		t.Error("ReadCertificate() differs from the written certificate")
	}
}

func TestReadCertificateMissing(t *testing.T) {
	if _, err := ReadCertificate(filepath.Join(t.TempDir(), "missing.crt")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("ReadCertificate() = %v, want %v", err, os.ErrNotExist)
	}
}

func TestReadCertificateInvalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.crt")
	if err := os.WriteFile(path, []byte("not a certificate"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := ReadCertificate(path); err == nil {
		t.Error("ReadCertificate() of garbage = nil, want an error")
	}
}
