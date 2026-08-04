//go:build unix

package strategyevidence

import "os"

func evidenceOwnerUID() (uint32, bool) { return uint32(os.Geteuid()), true }

func validateEvidenceReadOnlyFile(path string, owner uint32) error {
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		return ErrSnapshotUnavailable
	}
	uid, ok := evidenceFileUID(info)
	if !ok || uid != owner {
		return ErrSnapshotUnavailable
	}
	return nil
}
