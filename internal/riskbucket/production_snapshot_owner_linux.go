//go:build linux

package riskbucket

import (
	"os"
	"syscall"
)

func productionRiskFileUID(info os.FileInfo) (uint32, bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, false
	}
	return stat.Uid, true
}
