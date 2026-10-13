//go:build unix

package main

import (
	"os"
	"syscall"
)

// fileOwner 는 경로 소유자 uid 를 돌려줌(a131 정리 명령의 uid 확인).
func fileOwner(path string) (int, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, false
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, false
	}
	return int(st.Uid), true
}
