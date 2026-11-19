# utils-go

Generic Go utility packages.

* `archive` Writes, extracts, and rebases tar archives, plain or compressed with Gzip or Zstandard.
* buildinfo` Describes a binary from values set at link time.
* `codec` Encodes and decodes Go values as JSON or YAML through struct tags.
* `crex` Structured errors, sentinel wrapping, and a buffering `slog` handler.
* `file` Atomic writes, permission modes, directory helpers, and file locking.
* `httpx` HTTP request decoding, responses, multipart bodies, and graceful serving.
* `mediatype` Media types with structured syntax suffixes.
* `pki` X.509 certificate templates, signing, revocation lists, and key stores.
* `pki/keychain` Key store backed by the OS keychain.
* `process` Runs external commands.
* `retry` Repeats an operation until it succeeds or time runs out.
* `watch` Watches file system paths for changes.
