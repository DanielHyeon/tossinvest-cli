//go:build unix

package strategyrouter

import (
	"io"
	"os"
)

func productionRouteOwnerUID() (uint32, bool) { return uint32(os.Geteuid()), true }

func readProductionRouteFile(path string, owner uint32, mode os.FileMode, maximum int64) ([]byte, error) {
	before, err := os.Lstat(path)
	if err != nil || before.Mode()&os.ModeSymlink != 0 || !before.Mode().IsRegular() || before.Mode().Perm() != mode {
		return nil, ErrProductionRouteUnavailable
	}
	uid, ok := productionRouteFileUID(before)
	if !ok || uid != owner || before.Size() <= 0 || before.Size() > maximum {
		return nil, ErrProductionRouteUnavailable
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(before, opened) || opened.Mode().Perm() != mode {
		return nil, ErrProductionRouteUnavailable
	}
	data, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil || int64(len(data)) > maximum {
		return nil, ErrProductionRouteUnavailable
	}
	after, err := os.Lstat(path)
	if err != nil || !os.SameFile(before, after) {
		return nil, ErrProductionRouteUnavailable
	}
	return data, nil
}

func validateProductionRouteJournalFile(path string, owner uint32) error {
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		return ErrProductionRouteUnavailable
	}
	uid, ok := productionRouteFileUID(info)
	if !ok || uid != owner {
		return ErrProductionRouteUnavailable
	}
	return nil
}
