//go:build unix

package verifylive

// record_lock_unix.go — a121 대사 추가의 기록 파일 flock(codex CG-2). 커널이 유지하는 배타 잠금이라 프로세스가 죽으면
// 함께 풀림. 하드링크 별칭은 같은 inode 라 같은 잠금을 다툼.

import (
	"errors"
	"syscall"
)

func flockExclusiveNB(fd uintptr) error {
	if err := syscall.Flock(int(fd), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return errRecordLocked
		}
		return err
	}
	return nil
}
