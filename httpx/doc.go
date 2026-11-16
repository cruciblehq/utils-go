// Package httpx holds HTTP plumbing.
//
// Servers decode requests with [DecodeRequest], which checks the media type
// before reading the body, and answer with [Write]. Both take the body format
// as a [codec.Format], so a protocol picks its encoding and the media type
// suffix follows from it. [Serve] runs a server until its context ends and then
// shuts it down gracefully. Clients build requests with [NewRequest] and turn a
// failed response into a short message with [ErrorBody], and decode bodies with
// [codec.Decode]. Bodies follow the codec struct tags.
package httpx
