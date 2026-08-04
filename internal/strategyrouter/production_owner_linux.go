//go:build linux

package strategyrouter

import (
	"os"
	"syscall"
)

func productionRouteFileUID(info os.FileInfo) (uint32, bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, false
	}
	return stat.Uid, true
}
