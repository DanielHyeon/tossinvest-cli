package binstamp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o700); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

// TestAReplacedBinaryIsNoticed is the whole point of the package: a soak that has
// been running for two days has to be able to tell that the file under it moved on.
func TestAReplacedBinaryIsNoticed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tossctl")
	write(t, path, "old")

	before, err := Of(path)
	if err != nil {
		t.Fatalf("Of: %v", err)
	}
	if !before.Known() {
		t.Fatal("the stamp of a file that exists reports itself unknown")
	}

	write(t, path, "a different build entirely")
	// Size alone would carry this, so move the timestamp too and make sure the
	// comparison is not relying on one of the two by accident.
	future := before.ModTime.Add(time.Hour)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	after, err := Of(path)
	if err != nil {
		t.Fatalf("Of after the replacement: %v", err)
	}
	if before.Same(after) {
		t.Errorf("the replacement was not noticed: %s vs %s", before.Describe(), after.Describe())
	}
}

// TestOnlyTheTimestampMovingIsEnough covers the reinstall that happens to produce
// the same number of bytes.
func TestOnlyTheTimestampMovingIsEnough(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tossctl")
	write(t, path, "same size")
	before, _ := Of(path)

	future := before.ModTime.Add(time.Minute)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}
	after, _ := Of(path)

	if before.Same(after) {
		t.Error("a rebuild that kept the size was reported as unchanged")
	}
}

// TestNothingObservedIsNeverAChange is the direction that matters.
//
// A warning the tool cannot substantiate is worse than no warning: it teaches the
// operator to ignore the banner, and the banner is the only place a stale console
// announces itself.
func TestNothingObservedIsNeverAChange(t *testing.T) {
	known := Stamp{Path: "/x/tossctl", Size: 10, ModTime: time.Unix(1, 0).UTC()}
	cases := map[string][2]Stamp{
		"both unknown":  {{}, {}},
		"lost the file": {known, {}},
		"gained a file": {{}, known},
	}
	for name, pair := range cases {
		if !pair[0].Same(pair[1]) {
			t.Errorf("%s: reported as a change; it is an absence of observation", name)
		}
	}
}

// TestAMissingBinaryIsNotAnError: an executable can be replaced by an unlink and a
// rename while this runs, and that must not take the process down.
func TestAMissingBinaryIsNotAnError(t *testing.T) {
	stamp, err := Of(filepath.Join(t.TempDir(), "never-existed"))
	if err != nil {
		t.Fatalf("Of on a missing file returned %v; it should return the zero stamp", err)
	}
	if stamp.Known() {
		t.Errorf("Of on a missing file returned %+v, want the zero stamp", stamp)
	}
	if stamp.Describe() == "" {
		t.Error("an unknown stamp describes itself as an empty string; it has to say something")
	}
}

// TestADirectoryIsAnError, because a path that resolves to one means the caller is
// wrong about where the binary is, and answering "unchanged" forever would hide it.
func TestADirectoryIsAnError(t *testing.T) {
	if _, err := Of(t.TempDir()); err == nil {
		t.Error("Of accepted a directory")
	}
}

// TestSelfFindsTheTestBinary. Self is what the console and the soak actually call.
func TestSelfFindsTheTestBinary(t *testing.T) {
	stamp, err := Self()
	if err != nil {
		t.Fatalf("Self: %v", err)
	}
	if !stamp.Known() {
		t.Fatalf("Self returned an unknown stamp: %+v", stamp)
	}
	if !stamp.Same(stamp) {
		t.Error("a stamp differs from itself")
	}
}

// TestSelfPathSurvivesAReinstallByRename.
//
// Linux reports /proc/self/exe as "/path (deleted)" once the file has been
// unlinked, which is what installing over a running binary looks like. The path
// without the suffix is where the new build now is, and that is the file the
// caller wants stat'ed.
func TestSelfPathSurvivesAReinstallByRename(t *testing.T) {
	path, err := SelfPath()
	if err != nil {
		t.Fatalf("SelfPath: %v", err)
	}
	if strings.HasSuffix(path, deletedSuffix) {
		t.Errorf("SelfPath returned %q with the deleted marker still on it", path)
	}
	if path == "" {
		t.Error("SelfPath returned nothing")
	}
}
