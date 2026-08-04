//go:build unix

package strategyproposal

import (
	"io"
	"os"
)

func productionOwnerUID() (uint32, bool) { return uint32(os.Geteuid()), true }

func readProductionFile(path string, owner uint32, mode os.FileMode, maximum int64) ([]byte, error) {
	before, err := os.Lstat(path)
	if err != nil || before.Mode()&os.ModeSymlink != 0 || !before.Mode().IsRegular() || before.Mode().Perm() != mode {
		return nil, ErrProductionProposalUnavailable
	}
	uid, ok := productionFileUID(before)
	if !ok || uid != owner || before.Size() <= 0 || before.Size() > maximum {
		return nil, ErrProductionProposalUnavailable
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(before, opened) || opened.Mode().Perm() != mode {
		return nil, ErrProductionProposalUnavailable
	}
	data, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil || int64(len(data)) > maximum {
		return nil, ErrProductionProposalUnavailable
	}
	after, err := os.Lstat(path)
	if err != nil || !os.SameFile(before, after) {
		return nil, ErrProductionProposalUnavailable
	}
	return data, nil
}
