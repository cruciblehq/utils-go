// Package pki issues and stores X.509 certificates and private keys.
//
// It covers the mechanics every certificate authority needs, without any
// policy: random serial numbers, certificate templates for authorities and
// leaves, signing, revocation lists, reading certificates, encoding private
// keys ([EncodeKey], [DecodeKey]), and storing them behind a [KeyStore].
// Which names, lifetimes, and hierarchy to use is left to the caller.
//
// [DirKeyStore] keeps keys as files readable only by their owner. The
// keychain subpackage keeps them in the OS keychain.
package pki
