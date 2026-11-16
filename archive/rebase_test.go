package archive

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestRebaseDirectory(t *testing.T) {
	data := tarBytes(t,
		entry{&tar.Header{Name: "old/", Typeflag: tar.TypeDir, Mode: 0755}, ""},
		entry{&tar.Header{Name: "old/a.txt", Mode: 0644, Typeflag: tar.TypeReg}, "a"},
		entry{&tar.Header{Name: "older/b.txt", Mode: 0644, Typeflag: tar.TypeReg}, "b"},
		entry{&tar.Header{Name: "old/link", Typeflag: tar.TypeLink, Linkname: "old/a.txt"}, ""},
	)

	var out bytes.Buffer
	if err := Rebase(&out, bytes.NewReader(data), "/old", "new/dir"); err != nil {
		t.Fatalf("Rebase() = %v, want nil", err)
	}

	headers := readHeaders(t, out.Bytes())
	for _, name := range []string{"new/dir/", "new/dir/a.txt", "older/b.txt", "new/dir/link"} {
		if headers[name] == nil {
			t.Errorf("entry %q missing, have %v", name, headers)
		}
	}
	if link := headers["new/dir/link"]; link != nil && link.Linkname != "new/dir/a.txt" {
		t.Errorf("link target = %q, want %q", link.Linkname, "new/dir/a.txt")
	}
}

func TestRebaseSingleFile(t *testing.T) {
	data := tarBytes(t, entry{&tar.Header{Name: "build.img", Mode: 0644, Typeflag: tar.TypeReg}, "img"})

	var out bytes.Buffer
	if err := Rebase(&out, bytes.NewReader(data), "build.img", "disk/root.img"); err != nil {
		t.Fatalf("Rebase() = %v, want nil", err)
	}

	if headers := readHeaders(t, out.Bytes()); headers["disk/root.img"] == nil {
		t.Fatalf("entry disk/root.img missing, have %v", headers)
	}
}

func TestRebaseEverything(t *testing.T) {
	data := tarBytes(t,
		entry{&tar.Header{Name: "./", Typeflag: tar.TypeDir, Mode: 0755}, ""},
		entry{&tar.Header{Name: "./bin/sh", Mode: 0755, Typeflag: tar.TypeReg, Uid: 7}, "sh"},
		entry{&tar.Header{Name: "bin/ls", Typeflag: tar.TypeSymlink, Linkname: "sh"}, ""},
	)

	var out bytes.Buffer
	if err := Rebase(&out, bytes.NewReader(data), "", "/opt/app"); err != nil {
		t.Fatalf("Rebase() = %v, want nil", err)
	}

	headers := readHeaders(t, out.Bytes())
	if len(headers) != 2 {
		t.Fatalf("len(entries) = %d, want 2; have %v", len(headers), headers)
	}
	if sh := headers["opt/app/bin/sh"]; sh == nil || sh.Uid != 7 || sh.Mode != 0755 {
		t.Errorf("opt/app/bin/sh = %+v, want uid 7 and mode 0755", sh)
	}
	if ls := headers["opt/app/bin/ls"]; ls == nil || ls.Linkname != "sh" {
		t.Errorf("opt/app/bin/ls = %+v, want symlink to sh", ls)
	}
}

func TestRebaseRootKeepsNames(t *testing.T) {
	data := tarBytes(t,
		entry{&tar.Header{Name: "./etc/", Typeflag: tar.TypeDir, Mode: 0755}, ""},
		entry{&tar.Header{Name: "./etc/hosts", Mode: 0644, Typeflag: tar.TypeReg}, "hosts"},
	)

	var out bytes.Buffer
	if err := Rebase(&out, bytes.NewReader(data), "", "/"); err != nil {
		t.Fatalf("Rebase() = %v, want nil", err)
	}

	headers := readHeaders(t, out.Bytes())
	if headers["etc/"] == nil || headers["etc/hosts"] == nil {
		t.Fatalf("entries = %v, want etc/ and etc/hosts", headers)
	}
}

func TestRebaseInvalidStream(t *testing.T) {
	err := Rebase(io.Discard, bytes.NewReader([]byte("not a tar stream at all, just some bytes")), "", "x")
	if !errors.Is(err, ErrReadFailed) {
		t.Fatalf("Rebase() = %v, want %v", err, ErrReadFailed)
	}
}
