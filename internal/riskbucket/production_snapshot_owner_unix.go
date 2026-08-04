//go:build unix

package riskbucket

import (
	"io"
	"os"
)

func productionRiskOwnerUID() (uint32, bool) { return uint32(os.Geteuid()), true }

func readProductionRiskFile(path string, owner uint32, mode os.FileMode, maximum int64) ([]byte, error) {
	before, err := os.Lstat(path)
	if err != nil || before.Mode()&os.ModeSymlink != 0 || !before.Mode().IsRegular() || before.Mode().Perm() != mode {
		return nil, ErrProductionRiskSnapshotUnavailable
	}
	uid, ok := productionRiskFileUID(before)
	if !ok || uid != owner || before.Size() <= 0 || before.Size() > maximum {
		return nil, ErrProductionRiskSnapshotUnavailable
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(before, opened) || opened.Mode().Perm() != mode {
		return nil, ErrProductionRiskSnapshotUnavailable
	}
	data, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil || int64(len(data)) > maximum {
		return nil, ErrProductionRiskSnapshotUnavailable
	}
	after, err := os.Lstat(path)
	if err != nil || !os.SameFile(before, after) {
		return nil, ErrProductionRiskSnapshotUnavailable
	}
	return data, nil
}

func validateProductionRiskJournalFile(path string, owner uint32) error {
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		return ErrProductionRiskSnapshotUnavailable
	}
	uid, ok := productionRiskFileUID(info)
	if !ok || uid != owner {
		return ErrProductionRiskSnapshotUnavailable
	}
	return nil
}
