// Package httpx holds HTTP plumbing.
//
// Servers decode requests with [DecodeRequest], which checks the media type
// before reading the body, and answer with [Write]. Both take the body format
// as a [codec.Format], so a protocol picks its encoding and the media type
// suffix follows from it. [Serve] runs a server until its context ends and
// then shuts it down gracefully.
//
// Clients build requests with [NewRequest] and decode bodies with [codec.Decode].
// [ErrorBody] turns a failed response into a short message, and [MaxErrorPayload]
// bounds how much of a structured error body to read.
//
// A body that holds a message and a data stream together, such as an archive,
// is a multipart/related body. [MultipartWriter] writes one and
// [ReadMultipart] reads one.
//
// Bodies follow the codec struct tags. [HeaderContentType] and [HeaderAccept]
// name the standard headers.
package httpx
