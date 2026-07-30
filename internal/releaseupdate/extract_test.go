package releaseupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"strings"
	"testing"
)

type tarEntry struct {
	name     string
	kind     byte
	body     []byte
	linkname string
	pax      map[string]string
	format   tar.Format
}

func makeArchive(t *testing.T, entries ...tarEntry) []byte {
	t.Helper()
	var out bytes.Buffer
	gz := gzip.NewWriter(&out)
	tw := tar.NewWriter(gz)
	for _, entry := range entries {
		kind := entry.kind
		if kind == 0 {
			kind = tar.TypeReg
		}
		h := &tar.Header{
			Name: entry.name, Typeflag: kind, Mode: 0o755,
			Size: int64(len(entry.body)), Linkname: entry.linkname,
			PAXRecords: entry.pax, Format: entry.format,
		}
		if kind == tar.TypeDir {
			h.Size = 0
		}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(entry.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func TestExtractTossctlAcceptsOfficialAuthHelperLayout(t *testing.T) {
	archive := makeArchive(t,
		tarEntry{name: "tossctl", body: []byte("binary")},
		tarEntry{name: "auth-helper/", kind: tar.TypeDir},
		tarEntry{name: "auth-helper/tests/", kind: tar.TypeDir},
		tarEntry{name: "auth-helper/tests/fixture.json", body: []byte("{}")},
	)
	got, err := extractTossctl(archive, archiveLimits{
		maxArchive: 1 << 20, maxBinary: 1 << 20, maxExpanded: 2 << 20, maxEntries: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "binary" {
		t.Fatalf("binary = %q", got)
	}
}

func TestExtractTossctlRejectsPathAndSpecialEntryAttacks(t *testing.T) {
	for _, tc := range []struct {
		name    string
		entries []tarEntry
	}{
		{"duplicate", []tarEntry{{name: "tossctl", body: []byte("a")}, {name: "tossctl", body: []byte("b")}}},
		{"traversal", []tarEntry{{name: "tossctl", body: []byte("a")}, {name: "../escape", body: []byte("x")}}},
		{"absolute", []tarEntry{{name: "tossctl", body: []byte("a")}, {name: "/escape", body: []byte("x")}}},
		{"backslash", []tarEntry{{name: "tossctl", body: []byte("a")}, {name: `auth-helper\\escape`, body: []byte("x")}}},
		{"symlink", []tarEntry{{name: "tossctl", kind: tar.TypeSymlink, linkname: "other"}}},
		{"hardlink", []tarEntry{{name: "tossctl", kind: tar.TypeLink, linkname: "other"}}},
		{"fifo", []tarEntry{{name: "tossctl", kind: tar.TypeFifo}}},
		{"pax metadata", []tarEntry{{name: "tossctl", body: []byte("a"), pax: map[string]string{"comment": "forbidden"}}}},
		{"gnu metadata", []tarEntry{{name: "tossctl", body: []byte("a"), format: tar.FormatGNU}}},
		{"gnu sparse", []tarEntry{{name: "tossctl", body: []byte("a"), kind: tar.TypeGNUSparse, format: tar.FormatGNU}}},
		{"unexpected regular", []tarEntry{{name: "tossctl", body: []byte("a")}, {name: "README", body: []byte("x")}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := extractTossctl(makeArchive(t, tc.entries...), archiveLimits{
				maxArchive: 1 << 20, maxBinary: 1 << 20, maxExpanded: 2 << 20, maxEntries: 20,
			}); err == nil {
				t.Fatal("archive accepted")
			}
		})
	}
}

func TestExtractTossctlRejectsBoundsAndGzipMultistream(t *testing.T) {
	t.Run("binary", func(t *testing.T) {
		archive := makeArchive(t, tarEntry{name: "tossctl", body: []byte(strings.Repeat("x", 32))})
		if _, err := extractTossctl(archive, archiveLimits{
			maxArchive: 1 << 20, maxBinary: 8, maxExpanded: 1 << 20, maxEntries: 10,
		}); err == nil {
			t.Fatal("oversized binary accepted")
		}
	})
	t.Run("entries", func(t *testing.T) {
		archive := makeArchive(t,
			tarEntry{name: "auth-helper/", kind: tar.TypeDir},
			tarEntry{name: "tossctl", body: []byte("binary")},
		)
		if _, err := extractTossctl(archive, archiveLimits{
			maxArchive: 1 << 20, maxBinary: 1 << 20, maxExpanded: 1 << 20, maxEntries: 1,
		}); err == nil {
			t.Fatal("entry overflow accepted")
		}
	})
	t.Run("multistream", func(t *testing.T) {
		one := makeArchive(t, tarEntry{name: "tossctl", body: []byte("one")})
		two := makeArchive(t, tarEntry{name: "tossctl", body: []byte("two")})
		if _, err := extractTossctl(append(one, two...), archiveLimits{
			maxArchive: 1 << 20, maxBinary: 1 << 20, maxExpanded: 1 << 20, maxEntries: 10,
		}); err == nil {
			t.Fatal("gzip multistream accepted")
		}
	})
	t.Run("trailing tar payload in one gzip member", func(t *testing.T) {
		archive := makeArchive(t, tarEntry{name: "tossctl", body: []byte("one")})
		reader, err := gzip.NewReader(bytes.NewReader(archive))
		if err != nil {
			t.Fatal(err)
		}
		raw, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		if err := reader.Close(); err != nil {
			t.Fatal(err)
		}
		raw = append(raw, []byte("trailing")...)
		var repacked bytes.Buffer
		writer := gzip.NewWriter(&repacked)
		if _, err := writer.Write(raw); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := extractTossctl(repacked.Bytes(), archiveLimits{
			maxArchive: 1 << 20, maxBinary: 1 << 20, maxExpanded: 1 << 20, maxEntries: 10,
		}); err == nil {
			t.Fatal("trailing uncompressed payload accepted")
		}
	})
}
