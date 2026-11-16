package buildinfo

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// Reported for a value that was not set at link time.
const undefined = "(undefined)"

// Reported instead of a version string for a local build.
const localBuild = "(local)"

// Stage that is left out of version strings.
const mainStage = "main"

// Module version the Go toolchain records for a build outside a module download.
const develVersion = "(devel)"

// Build setting that holds the version control revision.
const vcsRevision = "vcs.revision"

// Length a revision recorded by the Go toolchain is shortened to.
const shortRevisionLength = 12

// Build identity of a binary.
type Info struct {
	version string // Version as set at link time.
	stage   string // Stage or branch as set at link time.
	commit  string // Git commit as set at link time.
}

// Returns the build identity for the given link time values.
//
// A version or commit that was not set at link time is taken from what the Go
// toolchain recorded in the binary, when it recorded one: the module version
// for a binary installed with "go install module@version", and the revision of
// the version control checkout the binary was built from.
func New(version, stage, commit string) Info {
	i := Info{
		version: strings.TrimSpace(version),
		stage:   strings.TrimSpace(stage),
		commit:  strings.TrimSpace(commit),
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		i = i.withRecorded(bi)
	}
	return i
}

// Returns i with an empty version or commit filled from bi.
func (i Info) withRecorded(bi *debug.BuildInfo) Info {
	if i.version == "" && bi.Main.Version != develVersion {
		i.version = bi.Main.Version
	}
	if i.commit == "" {
		for _, s := range bi.Settings {
			if s.Key == vcsRevision {
				i.commit = s.Value[:min(len(s.Value), shortRevisionLength)]
			}
		}
	}
	return i
}

// Returns the version without a leading "v", or "(undefined)".
func (i Info) Version() string {
	if i.version == "" {
		return undefined
	}
	return strings.TrimPrefix(strings.ToLower(i.version), "v")
}

// Returns the development stage in lowercase, or "(undefined)".
//
// The stage is usually the git branch the binary was built from.
func (i Info) Stage() string {
	if i.stage == "" {
		return undefined
	}
	return strings.ToLower(i.stage)
}

// Returns the git commit, or "(undefined)".
func (i Info) Commit() string {
	if i.commit == "" {
		return undefined
	}
	return i.commit
}

// Returns the architecture the binary was built for.
func (i Info) Arch() string {
	return runtime.GOARCH
}

// Whether the binary was built without version, stage, or commit.
func (i Info) IsLocal() bool {
	return i.version == "" || i.stage == "" || i.commit == ""
}

// Returns a version string such as "1.2.3+beta a1b2c3d [arm64]".
//
// The stage is left out for the main stage. A local build reports "(local)".
func (i Info) String() string {
	if i.IsLocal() {
		return localBuild
	}
	stage := ""
	if s := i.Stage(); s != mainStage {
		stage = fmt.Sprintf("+%s", s)
	}
	return fmt.Sprintf("%s%s %s [%s]", i.Version(), stage, i.Commit(), i.Arch())
}
