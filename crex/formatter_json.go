package crex

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

// Formats log records as JSON objects, one per line.
//
// Each object contains at minimum "level" and "message" fields, plus "time"
// when the record has one. When verbose mode is disabled (default), only those
// are included. When enabled, additional attributes are included as top-level
// fields, and group values are recursively resolved into nested objects.
type JSONFormatter struct {
	Verbose bool // Whether to include all attributes in the output.
}

// Creates a new JSON formatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// Writes a log record as a JSON line.
func (f *JSONFormatter) Write(w io.Writer, rctx *RecordContext) error {
	entry := map[string]any{
		"level":   strings.ToLower(rctx.Record.Level.String()),
		"message": rctx.Record.Message,
	}
	if !rctx.Record.Time.IsZero() {
		entry["time"] = rctx.Record.Time.Format(time.RFC3339Nano)
	}

	if f.Verbose {
		if len(rctx.Groups) > 0 {
			entry["groups"] = rctx.Groups
		}

		rctx.Record.Attrs(func(attr slog.Attr) bool {
			entry[attr.Key] = resolveValue(attr.Value)
			return true
		})
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(w, string(data))
	return err
}

// Recursively resolves a slog.Value to a Go value. A [slog.LogValuer] is
// resolved first. Groups are converted to maps; other values use their native
// representation.
func resolveValue(v slog.Value) any {
	v = v.Resolve()
	switch v.Kind() {
	case slog.KindGroup:
		m := make(map[string]any)
		for _, attr := range v.Group() {
			m[attr.Key] = resolveValue(attr.Value)
		}
		return m
	default:
		return v.Any()
	}
}
