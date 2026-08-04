//go:build linux

package strategyproposal

import (
	"os"
	"syscall"
)

func productionFileUID(info os.FileInfo) (uint32, bool) {
	value, ok := info.Sys().(*syscall.Stat_t)
	return value.Uid, ok
}
