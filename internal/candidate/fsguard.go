package candidate

// fsguard.go carries the ledger's durability rule across to the discovery store
// (D2: "ext4 강제·fuseblk 거부 규칙은 원장의 것을 그대로 따른다").
//
// # Why a second copy and not an import
//
// internal/journal owns the original. Importing it here would put journal.Journal
// — RecordIntent, the reservations, the execution contract — inside this
// package's type universe, which is exactly what the isolation requirement
// forbids (D7, spec "발굴은 주문 경로에 도달할 수 없어야 한다"). A filesystem
// allowlist is not worth opening that door.
//
// The obvious cost of a copy is drift, so it is not left to care:
// fsguard_drift_test.go reads internal/journal/fsguard.go as *source* — no import
// — and fails when the two allowlists stop agreeing. Reading a file is not a
// dependency, so the isolation holds and the duplication is watched.
//
// # Why the rule applies here at all
//
// The same partial-write exposure. A store on a filesystem where fsync does not
// mean fsync loses rows on power loss, and the row it loses may be the one
// carrying first_seen_at — which is the one fact this package exists to keep.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Filesystem magic numbers as reported by statfs(2) (linux/magic.h). The values
// and the split between allowed and rejected are the journal's; see the file
// header for why they are spelled again rather than imported.
const (
	// MagicExt covers ext2, ext3 and ext4 — statfs cannot tell them apart.
	MagicExt   int64 = 0xEF53
	MagicXFS   int64 = 0x58465342
	MagicBtrfs int64 = 0x9123683E

	// Recognised and rejected. Named so a refusal can say what it found.
	MagicFUSE    int64 = 0x65735546
	MagicTmpfs   int64 = 0x01021994
	MagicNFS     int64 = 0x6969
	MagicSMB     int64 = 0x517B
	MagicSMB2    int64 = 0xFE534D42
	MagicOverlay int64 = 0x794C7630
	MagicZFS     int64 = 0x2FC12FC1
	MagicExFAT   int64 = 0x2011BAB0
	MagicVFAT    int64 = 0x4D44
	MagicNTFS3   int64 = 0x7366746E
	MagicF2FS    int64 = 0xF2F52010
	MagicCeph    int64 = 0x00C36400
)

var (
	// ErrFilesystemNotAllowed means the store's directory sits outside the
	// allowlist. Opening is refused rather than continued with a durability
	// guarantee we cannot make.
	ErrFilesystemNotAllowed = errors.New("candidate: filesystem is not in the allowlist")
	// ErrFilesystemUnknown means the probe produced a value we do not recognise.
	// Unknown fails closed, exactly as it does for the ledger.
	ErrFilesystemUnknown = errors.New("candidate: filesystem could not be identified")
	// ErrProbeUnsupported means this platform has no filesystem probe.
	ErrProbeUnsupported = errors.New("candidate: filesystem probe is not supported on this platform")
)

// FSInfo is the result of probing the filesystem that hosts a directory.
type FSInfo struct {
	// Name is the human-readable filesystem name, empty when the magic is
	// unrecognised.
	Name string
	// Magic is the raw statfs f_type value, 0 when unavailable.
	Magic int64
	// FreeBytes is the space available to this user on that filesystem, and
	// FreeMeasured is the absent-versus-zero flag for it (D16).
	//
	// The flag is load-bearing and is the reason this is not one field. A prober
	// that cannot answer would otherwise report "0 bytes free", which is the most
	// alarming possible reading of a measurement nobody made — and the opposite
	// mistake, defaulting an unmeasured probe to "plenty", is the one that lets
	// discovery fill the filesystem the ledger writes to. Neither is safe, so the
	// two states are kept apart and the caller decides, exactly as Budget.Reported
	// does for the rate allowance.
	FreeBytes    int64
	FreeMeasured bool
}

// FSProber inspects the filesystem hosting a directory.
type FSProber interface {
	Probe(dir string) (FSInfo, error)
}

// FixedFSProber returns a prober that always answers with info.
//
// TESTS ONLY, for the same reason the journal's is: a developer's TMPDIR is not
// necessarily on an allowlisted filesystem, and the allowlist logic has to be
// testable without mounting anything.
func FixedFSProber(info FSInfo) FSProber { return fixedProber{info: info} }

type fixedProber struct{ info FSInfo }

func (p fixedProber) Probe(string) (FSInfo, error) { return p.info, nil }

var fsNames = map[int64]string{
	MagicExt:     "ext2/ext3/ext4",
	MagicXFS:     "xfs",
	MagicBtrfs:   "btrfs",
	MagicFUSE:    "fuse/fuseblk",
	MagicTmpfs:   "tmpfs",
	MagicNFS:     "nfs",
	MagicSMB:     "smb",
	MagicSMB2:    "smb2",
	MagicOverlay: "overlayfs",
	MagicZFS:     "zfs",
	MagicExFAT:   "exfat",
	MagicVFAT:    "vfat",
	MagicNTFS3:   "ntfs3",
	MagicF2FS:    "f2fs",
	MagicCeph:    "ceph",
}

var allowedMagics = map[int64]bool{
	MagicExt:   true,
	MagicXFS:   true,
	MagicBtrfs: true,
}

var allowedNames = map[string]bool{
	"ext2":           true,
	"ext3":           true,
	"ext4":           true,
	"ext2/ext3/ext4": true,
	"xfs":            true,
	"btrfs":          true,
}

// AllowedFilesystems returns the allowlist as it appears in operator messages.
func AllowedFilesystems() []string { return []string{"ext4", "xfs", "btrfs"} }

// FilesystemName returns the diagnostic name for a statfs magic, "" when
// unrecognised.
func FilesystemName(magic int64) string { return fsNames[magic] }

func isAllowedFS(info FSInfo) bool {
	if info.Magic != 0 {
		return allowedMagics[info.Magic]
	}
	return allowedNames[strings.ToLower(strings.TrimSpace(info.Name))]
}

// CheckFilesystem verifies that dir — or, when it does not exist yet, its nearest
// existing ancestor — sits on an allowlisted filesystem.
func CheckFilesystem(prober FSProber, dir string) (FSInfo, error) {
	if prober == nil {
		return FSInfo{}, fmt.Errorf("%w: no prober configured", ErrProbeUnsupported)
	}
	target, err := nearestExistingDir(dir)
	if err != nil {
		return FSInfo{}, err
	}
	info, err := prober.Probe(target)
	if err != nil {
		return FSInfo{}, fmt.Errorf("candidate: probing the filesystem of %s: %w", target, err)
	}
	if info.Name == "" && info.Magic != 0 {
		info.Name = FilesystemName(info.Magic)
	}

	switch {
	case isAllowedFS(info):
		return info, nil
	case info.Name == "" && info.Magic == 0:
		return info, fmt.Errorf("%w: %s (probe returned nothing); allowed: %s",
			ErrFilesystemUnknown, target, strings.Join(AllowedFilesystems(), ", "))
	case info.Name == "":
		return info, fmt.Errorf("%w: %s reports filesystem magic %#x; allowed: %s",
			ErrFilesystemUnknown, target, info.Magic, strings.Join(AllowedFilesystems(), ", "))
	default:
		return info, fmt.Errorf("%w: %s is on %s; allowed: %s — move the data directory with %s",
			ErrFilesystemNotAllowed, target, info.Name,
			strings.Join(AllowedFilesystems(), ", "), EnvDataDir)
	}
}

// nearestExistingDir walks up until it finds a directory that exists, so the
// guard can run before the data directory is created.
//
// The empty-string case is refused rather than cleaned. filepath.Clean("") is
// ".", so without this the guard would probe the process's current working
// directory and hand back a durability approval for a directory that has nothing
// to do with the store — a pass, on the wrong filesystem, for a path nobody asked
// about. The ledger's copy refuses it and so does this one.
func nearestExistingDir(dir string) (string, error) {
	if strings.TrimSpace(dir) == "" {
		return "", fmt.Errorf("%w: empty directory", ErrFilesystemUnknown)
	}
	current := filepath.Clean(dir)
	for {
		info, err := os.Stat(current)
		switch {
		case err == nil && info.IsDir():
			return current, nil
		case err == nil:
			// A file where a directory belongs. The caller's MkdirAll will report
			// it far better than this can, so probe the parent and let the guard
			// still run rather than turning a naming mistake into a filesystem
			// verdict.
			return filepath.Dir(current), nil
		case !os.IsNotExist(err):
			return "", fmt.Errorf("candidate: inspecting %s: %w", current, err)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("%w: no existing ancestor of %s", ErrFilesystemUnknown, dir)
		}
		current = parent
	}
}
