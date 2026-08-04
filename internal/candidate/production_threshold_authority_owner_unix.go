//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package candidate

import (
	"errors"
	"io"
	"os"
	"syscall"
)

func productionThresholdOwnerUID() (uint64, bool) {
	uid := os.Getuid()
	return uint64(uid), uid >= 0
}

func readProductionThresholdFile(path string, owner uint64, mode os.FileMode, maximum int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != mode || info.Size() < 1 || info.Size() > maximum {
		return nil, errors.New("candidate: threshold authority is not an exact owner-only regular file")
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || uint64(stat.Uid) != owner {
		return nil, errors.New("candidate: threshold authority owner mismatch")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return nil, errors.New("candidate: threshold authority changed while opening")
	}
	data := make([]byte, info.Size())
	if _, err := io.ReadFull(file, data); err != nil {
		return nil, err
	}
	var extra [1]byte
	if count, err := file.Read(extra[:]); count != 0 || !errors.Is(err, io.EOF) {
		return nil, errors.New("candidate: threshold authority changed while reading")
	}
	return data, nil
}
