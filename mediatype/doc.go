// Package mediatype names the content of HTTP messages.
//
// A [MediaType] is the base type (e.g., "application/vnd.example.thing.v1").
// [MediaType.With] adds the structured syntax suffix (RFC 6839) of the format
// a body is encoded in, such as +json. [Matches] compares a Content-Type header
// against an expected type, ignoring parameters and letter case.
package mediatype
