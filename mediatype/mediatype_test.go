package mediatype

import (
	"testing"

	"github.com/cruciblehq/utils-go/codec"
)

func TestString(t *testing.T) {
	if got := MediaType("application/vnd.test.v1").String(); got != "application/vnd.test.v1" {
		t.Errorf("String() = %q, want %q", got, "application/vnd.test.v1")
	}
}

func TestWith(t *testing.T) {
	m := MediaType("application/vnd.test.v1")
	if got := m.With(codec.JSON); got != "application/vnd.test.v1+json" {
		t.Errorf("With(JSON) = %q, want %q", got, "application/vnd.test.v1+json")
	}
	if got := m.With(codec.YAML); got != "application/vnd.test.v1+yaml" {
		t.Errorf("With(YAML) = %q, want %q", got, "application/vnd.test.v1+yaml")
	}
}

func TestMatches(t *testing.T) {
	tests := []struct {
		header   string
		expected string
		want     bool
	}{
		{"application/vnd.test.v1+json", "application/vnd.test.v1+json", true},
		{"application/vnd.test.v1+json; charset=utf-8", "application/vnd.test.v1+json", true},
		{"Application/VND.Test.v1+JSON", "application/vnd.test.v1+json", true},
		{"application/vnd.test.v1", "application/vnd.test.v1+json", false},
		{"application/json", "application/vnd.test.v1+json", false},
		{"", "application/vnd.test.v1+json", false},
		{"not a media type;;", "application/vnd.test.v1+json", false},
	}
	for _, tt := range tests {
		if got := Matches(tt.header, tt.expected); got != tt.want {
			t.Errorf("Matches(%q, %q) = %v, want %v", tt.header, tt.expected, got, tt.want)
		}
	}
}
