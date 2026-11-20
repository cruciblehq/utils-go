package pki

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
)

// Allowance for clocks that run behind the issuer's.
//
// Certificates become valid this long before they are issued, so a peer whose
// clock is slightly behind still accepts them.
const clockSkew = time.Hour

// Number of random bits in a serial number.
const serialBits = 128

// Period during which a certificate or revocation list is valid.
type Validity struct {
	NotBefore time.Time // Start of the period.
	NotAfter  time.Time // End of the period.
}

// Returns a validity that starts now and lasts d.
//
// The start is moved back to allow for clock skew between peers.
func ValidFor(d time.Duration) Validity {
	now := time.Now()
	return Validity{NotBefore: now.Add(-clockSkew), NotAfter: now.Add(d)}
}

// Returns a validity that starts now and ends at notAfter.
//
// The start is moved back to allow for clock skew between peers.
func ValidUntil(notAfter time.Time) Validity {
	return Validity{NotBefore: time.Now().Add(-clockSkew), NotAfter: notAfter}
}

// Returns a random serial number.
func SerialNumber() (*big.Int, error) {
	return rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), serialBits))
}

// Returns a template for a certificate authority named name.
//
// The authority can sign certificates and revocation lists. When pathLenZero is
// set, it can sign only leaves, not other authorities.
func AuthorityTemplate(name string, v Validity, pathLenZero bool) *x509.Certificate {
	return &x509.Certificate{
		Subject:               pkix.Name{CommonName: name},
		NotBefore:             v.NotBefore,
		NotAfter:              v.NotAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        pathLenZero,
	}
}

// Returns a template for a leaf certificate named name.
//
// usage says what the certificate may authenticate, and dnsNames are the
// names it is valid for.
func LeafTemplate(name string, v Validity, usage x509.ExtKeyUsage, dnsNames ...string) *x509.Certificate {
	return &x509.Certificate{
		Subject:     pkix.Name{CommonName: name},
		DNSNames:    dnsNames,
		NotBefore:   v.NotBefore,
		NotAfter:    v.NotAfter,
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{usage},
	}
}

// Signs tmpl for pub with signer and returns the parsed certificate.
//
// parent is the issuing certificate. When parent is nil, the certificate is
// signed by itself and signer must hold the key for pub. tmpl gets a random
// serial number when it has none.
func Sign(tmpl, parent *x509.Certificate, pub crypto.PublicKey, signer crypto.Signer) (*x509.Certificate, error) {
	if tmpl.SerialNumber == nil {
		serial, err := SerialNumber()
		if err != nil {
			return nil, err
		}
		tmpl.SerialNumber = serial
	}
	if parent == nil {
		parent = tmpl
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, parent, pub, signer)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(der)
}

// Signs a revocation list numbered number for issuer and returns it parsed.
//
// revoked lists the withdrawn certificates. issuer must be allowed to sign
// revocation lists, and signer must hold its key.
func SignRevocationList(number int64, v Validity, revoked []x509.RevocationListEntry, issuer *x509.Certificate, signer crypto.Signer) (*x509.RevocationList, error) {
	der, err := x509.CreateRevocationList(rand.Reader, &x509.RevocationList{
		Number:                    big.NewInt(number),
		ThisUpdate:                v.NotBefore,
		NextUpdate:                v.NotAfter,
		RevokedCertificateEntries: revoked,
	}, issuer, signer)
	if err != nil {
		return nil, err
	}
	return x509.ParseRevocationList(der)
}
