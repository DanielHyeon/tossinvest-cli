package journal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Filesystem magic numbers as reported by statfs(2) (linux/magic.h). Exported so
// tests can build probe results without a syscall, and so the refusal message can
// name what it found.
const (
	// MagicExt covers ext2, ext3 and ext4 — statfs cannot distinguish them, they
	// share EXT2_SUPER_MAGIC. ext2 has no journal, but it is a local block
	// filesystem and vanishingly rare on a modern host; treating the family as
	// allowed is the pragmatic reading of the spec's "local journaling
	// filesystem" allowlist.
	MagicExt   int64 = 0xEF53
	MagicXFS   int64 = 0x58465342
	MagicBtrfs int64 = 0x9123683E

	// Below: recognised but rejected. Named only so the refusal tells the
	// operator what the journal directory is actually sitting on.
	MagicFUSE    int64 = 0x65735546 // fuse and fuseblk (ntfs-3g, sshfs, …)
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
	// ErrFilesystemNotAllowed means the journal directory sits on a filesystem
	// outside the allowlist. Startup is refused rather than continuing with
	// durability guarantees we cannot make.
	ErrFilesystemNotAllowed = errors.New("journal: filesystem is not in the allowlist")

	// ErrFilesystemUnknown means the probe produced a value we do not recognise.
	// Unknown is treated exactly like not-allowed: fail closed.
	ErrFilesystemUnknown = errors.New("journal: filesystem could not be identified")

	// ErrProbeUnsupported means this platform has no filesystem probe. The engine
	// journal is Linux-targeted; on other platforms the guard cannot be
	// satisfied and startup is refused.
	ErrProbeUnsupported = errors.New("journal: filesystem probe is not supported on this platform")
)

// FSInfo is the result of probing the filesystem that hosts a directory.
type FSInfo struct {
	// Name is the human-readable filesystem name, e.g. "ext2/ext3/ext4" or
	// "fuse/fuseblk". Empty when the magic is unrecognised.
	Name string
	// Magic is the raw statfs f_type value, 0 when unavailable.
	Magic int64
}

// FSProber inspects the filesystem hosting a directory. It is an interface so the
// allowlist logic can be tested on any host without mounting anything, while the
// production implementation goes through statfs(2).
type FSProber interface {
	Probe(dir string) (FSInfo, error)
}

// FixedFSProber returns a prober that always answers with info.
//
// TESTS ONLY. Wiring this into production defeats the durability guard the
// order-execution spec requires; the journal's own tests use it because they must
// run on whatever filesystem the developer's TMPDIR happens to be.
func FixedFSProber(info FSInfo) FSProber { return fixedProber{info: info} }

type fixedProber struct{ info FSInfo }

func (p fixedProber) Probe(string) (FSInfo, error) { return p.info, nil }

// fsNames maps every magic we recognise to a stable diagnostic name.
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

// allowedMagics is the allowlist by statfs magic.
var allowedMagics = map[int64]bool{
	MagicExt:   true,
	MagicXFS:   true,
	MagicBtrfs: true,
}

// allowedNames is the allowlist by name, for probers that report a name instead
// of a magic (non-Linux implementations, tests).
var allowedNames = map[string]bool{
	"ext2":           true,
	"ext3":           true,
	"ext4":           true,
	"ext2/ext3/ext4": true,
	"xfs":            true,
	"btrfs":          true,
}

// AllowedFilesystems returns the allowlist as it appears in operator-facing
// messages.
func AllowedFilesystems() []string { return []string{"ext4", "xfs", "btrfs"} }

// FilesystemName returns the diagnostic name for a statfs magic, or "" when the
// magic is unrecognised.
func FilesystemName(magic int64) string { return fsNames[magic] }

// isAllowedFS reports whether a probe result satisfies the allowlist. Magic wins
// when present; the name is the fallback.
func isAllowedFS(info FSInfo) bool {
	if info.Magic != 0 {
		return allowedMagics[info.Magic]
	}
	return allowedNames[strings.ToLower(strings.TrimSpace(info.Name))]
}

// CheckFilesystem verifies that dir (or, when it does not exist yet, its nearest
// existing ancestor) sits on an allowlisted filesystem.
//
// Failure modes all end the same way — refuse to start:
//   - probe error (unsupported platform, permission denied): propagated
//   - unrecognised filesystem: ErrFilesystemUnknown
//   - recognised but not allowlisted: ErrFilesystemNotAllowed
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
		return FSInfo{}, fmt.Errorf("journal: probing the filesystem of %s: %w", target, err)
	}
	if info.Name == "" && info.Magic != 0 {
		info.Name = FilesystemName(info.Magic)
	}

	switch {
	case isAllowedFS(info):
		if info.Name == "" {
			info.Name = FilesystemName(info.Magic)
		}
		return info, nil
	case info.Name == "" && info.Magic == 0:
		return info, fmt.Errorf("%w: %s (probe returned nothing); allowed: %s",
			ErrFilesystemUnknown, target, strings.Join(AllowedFilesystems(), ", "))
	case info.Name == "":
		return info, fmt.Errorf("%w: %s reports filesystem magic %#x; allowed: %s",
			ErrFilesystemUnknown, target, info.Magic, strings.Join(AllowedFilesystems(), ", "))
	default:
		return info, fmt.Errorf("%w: %s is on %s; allowed: %s — move the journal with %s",
			ErrFilesystemNotAllowed, target, info.Name,
			strings.Join(AllowedFilesystems(), ", "), EnvDataDir)
	}
}

// nearestExistingDir walks up from dir to the first path that exists, so a
// first-run data directory can be validated before it is created.
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
			// A file where a directory was expected: let the caller's mkdir
			// report it, but probe its parent so the guard still runs.
			return filepath.Dir(current), nil
		case !errors.Is(err, os.ErrNotExist):
			return "", fmt.Errorf("journal: inspecting %s: %w", current, err)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("%w: no existing ancestor of %s", ErrFilesystemUnknown, dir)
		}
		current = parent
	}
}
