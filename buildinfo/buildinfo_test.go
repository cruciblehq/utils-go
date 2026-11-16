package buildinfo

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"testing"
)

func TestUndefined(t *testing.T) {
	i := New(" ", "", "")
	if i.Version() != undefined || i.Stage() != undefined || i.Commit() != undefined {
		t.Errorf("Version, Stage, Commit = %q, %q, %q, want %q", i.Version(), i.Stage(), i.Commit(), undefined)
	}
	if !i.IsLocal() {
		t.Error("IsLocal() = false, want true")
	}
	if i.String() != localBuild {
		t.Errorf("String() = %q, want %q", i.String(), localBuild)
	}
}

func TestVersionNormalized(t *testing.T) {
	if got := New(" V1.2.3 ", "", "").Version(); got != "1.2.3" {
		t.Errorf("Version() = %q, want %q", got, "1.2.3")
	}
}

func TestString(t *testing.T) {
	want := fmt.Sprintf("1.2.3+beta abc [%s]", runtime.GOARCH)
	if got := New("v1.2.3", "Beta", "abc").String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestStringMainStage(t *testing.T) {
	want := fmt.Sprintf("1.2.3 abc [%s]", runtime.GOARCH)
	if got := New("1.2.3", "main", "abc").String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestWithRecorded(t *testing.T) {
	bi := &debug.BuildInfo{
		Main:     debug.Module{Version: "v1.4.0"},
		Settings: []debug.BuildSetting{{Key: vcsRevision, Value: "0123456789abcdef0123"}},
	}

	i := Info{}.withRecorded(bi)
	if i.Version() != "1.4.0" {
		t.Errorf("Version() = %q, want %q", i.Version(), "1.4.0")
	}
	if i.Commit() != "0123456789ab" {
		t.Errorf("Commit() = %q, want %q", i.Commit(), "0123456789ab")
	}
}

func TestWithRecordedKeepsLinkTimeValues(t *testing.T) {
	bi := &debug.BuildInfo{
		Main:     debug.Module{Version: "v1.4.0"},
		Settings: []debug.BuildSetting{{Key: vcsRevision, Value: "0123456789abcdef"}},
	}

	i := Info{version: "2.0.0", commit: "abc"}.withRecorded(bi)
	if i.Version() != "2.0.0" || i.Commit() != "abc" {
		t.Errorf("Version, Commit = %q, %q, want %q, %q", i.Version(), i.Commit(), "2.0.0", "abc")
	}
}

func TestWithRecordedDevelVersion(t *testing.T) {
	i := Info{}.withRecorded(&debug.BuildInfo{Main: debug.Module{Version: develVersion}})
	if i.Version() != undefined || i.Commit() != undefined {
		t.Errorf("Version, Commit = %q, %q, want %q", i.Version(), i.Commit(), undefined)
	}
}
