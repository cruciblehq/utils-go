// Package keychain stores private keys in the OS keychain.
//
// It implements [pki.KeyStore] on the macOS Keychain, the Secret Service on
// Linux, and the Windows Credential Manager. Keys are never written to disk
// by this package and are unlocked by the OS login. It lives apart from pki
// so only programs that use it pull in its dependencies.
package keychain
