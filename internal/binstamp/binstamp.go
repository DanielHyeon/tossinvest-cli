// Package binstamp fingerprints an executable file so a long-running process can
// notice that the binary it was started from has been replaced
// (openspec change verify-execution-capability, task 1.8 ②).
//
// # Why a fingerprint and not a version string
//
// The two processes this exists for — `tossctl console` and `tossctl soak run` —
// run for days at a time on the operator's own machine while the binary under them
// is rebuilt and reinstalled. A version string cannot see that: a rebuild from the
// same commit produces the same version and a different executable, which is the
// common case during development and exactly the case the operator needs told
// about. So the question asked here is the narrow, honest one: *is the file at this
// path still the file this process was loaded from?*
//
// # Modification time and size, deliberately not a hash
//
// Hashing a 40 MB binary once per soak cycle would be work done to answer a
// question that mtime and size already answer. The failure mode of the cheap check
// is a replacement that keeps both — which for a Go build means the same content
// laid down with a preserved timestamp, and even then the answer "unchanged" is
// correct in every way an operator cares about. The failure direction is therefore
// a missed notice, never a false one: nothing here can claim a binary changed when
// it did not.
//
// # It never reads the file
//
// Stat only. A package that opened the executable would be a package that could be
// asked to read something else, and there is nothing inside the file this needs.
package binstamp

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Stamp identifies one executable file at one moment.
//
// The zero value means "not known", which is a state the callers report rather
// than an error: a console that cannot stat its own binary is still a console, and
// a warning it cannot substantiate is a warning it must not print.
type Stamp struct {
	// Path is where the executable was found.
	Path string `json:"path,omitempty"`
	// Size is the file's size in bytes.
	Size int64 `json:"size,omitempty"`
	// ModTime is the file's modification time, in UTC.
	ModTime time.Time `json:"mod_time,omitempty"`
}

// Known reports that the stamp actually describes a file.
func (s Stamp) Known() bool {
	return s.Path != "" && !s.ModTime.IsZero()
}

// Same reports that two stamps describe the same executable.
//
// Two unknown stamps are the same as each other: nothing was observed, so nothing
// changed. An unknown compared against a known one is *not* different either —
// "we could not look" must not become "it moved on".
func (s Stamp) Same(other Stamp) bool {
	if !s.Known() || !other.Known() {
		return true
	}
	return s.Path == other.Path && s.Size == other.Size && s.ModTime.Equal(other.ModTime)
}

// Describe renders the stamp for an operator.
func (s Stamp) Describe() string {
	if !s.Known() {
		return "(알 수 없음)"
	}
	return fmt.Sprintf("%s · %d바이트 · %s", s.Path, s.Size, s.ModTime.UTC().Format("2006-01-02 15:04:05Z"))
}

// Of fingerprints the file at path.
//
// A missing file is not an error worth propagating — an executable can legitimately
// be replaced by an unlink-and-rename while this runs — so it comes back as the
// zero Stamp with a nil error, and Same then answers "unchanged".
func Of(path string) (Stamp, error) {
	if path == "" {
		return Stamp{}, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Stamp{}, nil
		}
		return Stamp{}, fmt.Errorf("binstamp: stat %s: %w", path, err)
	}
	if info.IsDir() {
		return Stamp{}, fmt.Errorf("binstamp: %s is a directory, not an executable", path)
	}
	return Stamp{Path: path, Size: info.Size(), ModTime: info.ModTime().UTC()}, nil
}

// Self fingerprints the executable at this process's own path.
//
// "At this path", not "the bytes this process is running": once the file has been
// replaced those are different things, and the difference is the whole point. Call
// it once at startup to learn what this process is, and again later to learn what
// is installed.
func Self() (Stamp, error) {
	path, err := SelfPath()
	if err != nil {
		return Stamp{}, err
	}
	return Of(path)
}

// deletedSuffix is what Linux appends to /proc/self/exe when the executable has
// been unlinked, which is what an install by rename leaves behind.
const deletedSuffix = " (deleted)"

// SelfPath is where this process's executable lives.
//
// The suffix is stripped because a reinstall that renames over the binary leaves
// os.Executable reporting a path that does not exist — and the path without it is
// exactly the one the new build now occupies, which is the file the caller is
// asking about.
func SelfPath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("binstamp: locating this executable: %w", err)
	}
	return strings.TrimSuffix(path, deletedSuffix), nil
}
