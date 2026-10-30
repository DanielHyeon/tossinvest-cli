//go:build linux

package clock

import (
	"errors"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

type fakeBootProbe struct {
	fsErr, errorRead, errorNS, errorTime error
	boot                                 []string
	namespaces                           []BootNamespace
	offsets                              string
	sample                               unix.Timespec
	bootIndex, nsIndex                   int
}

func (p *fakeBootProbe) procFS() error { return p.fsErr }
func (p *fakeBootProbe) read(path string, max int) ([]byte, error) {
	if p.errorRead != nil {
		return nil, p.errorRead
	}
	if path == offsetsPath {
		return []byte(p.offsets), nil
	}
	v := p.boot[p.bootIndex]
	p.bootIndex++
	return []byte(v + "\n"), nil
}
func (p *fakeBootProbe) namespace(string) (BootNamespace, error) {
	if p.errorNS != nil {
		return BootNamespace{}, p.errorNS
	}
	v := p.namespaces[p.nsIndex]
	p.nsIndex++
	return v, nil
}
func (p *fakeBootProbe) bootTime() (unix.Timespec, error) { return p.sample, p.errorTime }
func goodProbe() *fakeBootProbe {
	n := BootNamespace{Device: 1, Inode: 2}
	return &fakeBootProbe{boot: []string{testBootA, testBootA}, namespaces: []BootNamespace{n, n, n, n}, offsets: "monotonic 0 0\nboottime 0 0\n", sample: unix.Timespec{Sec: 12, Nsec: 3}}
}
func TestBootSnapshotValid(t *testing.T) {
	p := goodProbe()
	s, e := (systemBootSource{p}).Snapshot()
	if e != nil || s.BootID != testBootA || s.Namespace != (BootNamespace{Device: 1, Inode: 2}) || s.Uptime != 12*time.Second+3 {
		t.Fatalf("invalid snapshot: %v", e)
	}
}
func TestBootSnapshotFailureCases(t *testing.T) {
	errInjected := errors.New("fixture failure")
	cases := []struct {
		name   string
		mutate func(*fakeBootProbe)
	}{
		{"procfs", func(p *fakeBootProbe) { p.fsErr = errInjected }}, {"read", func(p *fakeBootProbe) { p.errorRead = errInjected }}, {"namespace", func(p *fakeBootProbe) { p.errorNS = errInjected }}, {"syscall", func(p *fakeBootProbe) { p.errorTime = errInjected }},
		{"boot-changed", func(p *fakeBootProbe) { p.boot[1] = testBootB }}, {"boot-invalid", func(p *fakeBootProbe) { p.boot[0] = "BAD" }},
		{"children-differ", func(p *fakeBootProbe) { p.namespaces[1].Inode++ }},
		{"children-persistently-differ", func(p *fakeBootProbe) { p.namespaces[1].Inode++; p.namespaces[3].Inode++ }}, {"current-changed", func(p *fakeBootProbe) { p.namespaces[2].Inode++ }}, {"children-changed", func(p *fakeBootProbe) { p.namespaces[3].Inode++ }},
		{"nonzero-offset", func(p *fakeBootProbe) { p.offsets = "monotonic 0 0\nboottime 0 1\n" }}, {"negative-clock", func(p *fakeBootProbe) { p.sample.Sec = -1 }}, {"overflow-clock", func(p *fakeBootProbe) { p.sample.Sec = math.MaxInt64 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := goodProbe()
			tc.mutate(p)
			s, e := (systemBootSource{p}).Snapshot()
			if !errors.Is(e, ErrBootStamp) || s != (BootStamp{}) {
				t.Fatalf("failure accepted: %v", e)
			}
		})
	}
}
func TestBootOffsetParserStrict(t *testing.T) {
	valid := []string{"monotonic 0 0\nboottime 0 0\n", "boottime 0 0\nmonotonic 0 0\n"}
	for _, s := range valid {
		if e := parseZeroOffsets([]byte(s)); e != nil {
			t.Fatal(e)
		}
	}
	invalid := []string{"", "monotonic 0 0\n", "monotonic 0 0\nmonotonic 0 0\n", "monotonic 0 0\nboottime 0 0\nextra 0 0\n", "monotonic -1 0\nboottime 0 0\n", "monotonic 0 0\nboottime 0 1\n", "monotonic 0 0\nboottime 0 999999999999999999999999\n", "monotonic 0 0\nother 0 0\n", "monotonic 0 0 0\nboottime 0 0\n"}
	for _, s := range invalid {
		if e := parseZeroOffsets([]byte(s)); e == nil {
			t.Fatalf("invalid offset accepted %q", s)
		}
	}
}
func TestBootIDStrict(t *testing.T) {
	valid := testBootA + "\n"
	if _, e := parseBootID([]byte(valid)); e != nil {
		t.Fatal(e)
	}
	for _, s := range []string{"", strings.ToUpper(testBootA), "00000000-0000-0000-0000-000000000000", testBootA + "\n\n", " " + testBootA, testBootA[:35]} {
		if _, e := parseBootID([]byte(s)); e == nil {
			t.Fatal("invalid boot id accepted")
		}
	}
}
func TestBootTimespecBounds(t *testing.T) {
	for _, s := range []unix.Timespec{{Sec: -1}, {Nsec: -1}, {Nsec: 1000000000}, {Sec: math.MaxInt64}, {Sec: math.MaxInt64 / 1000000000, Nsec: 999999999}} {
		if _, e := timespecDuration(s); e == nil {
			t.Fatal("invalid timespec accepted")
		}
	}
}
func TestBootLocalSmoke(t *testing.T) {
	s, e := SystemBootSource().Snapshot()
	if e != nil {
		if !errors.Is(e, ErrBootStamp) {
			t.Fatal("unsupported result lacks typed error")
		}
		t.Skip("local boot source unsupported")
	}
	if !validStamp(s) {
		t.Fatal("invalid local snapshot")
	}
}
func TestBootProcReadBound(t *testing.T) {
	f, e := os.CreateTemp(t.TempDir(), "bounded")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f.WriteString(strings.Repeat("x", 65)); e != nil {
		t.Fatal(e)
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
	if _, e = (systemBootProbe{}).read(f.Name(), 64); e == nil {
		t.Fatal("oversized proc input accepted")
	}
}

func TestBootIDRejectsInvalidUUIDVersionAndVariant(t *testing.T) {
	for _, v := range []string{"12345678-1234-0234-8234-123456789abc", "12345678-1234-4234-1234-123456789abc"} {
		if _, e := parseBootID([]byte(v)); e == nil {
			t.Fatal("invalid UUID metadata accepted")
		}
	}
}
