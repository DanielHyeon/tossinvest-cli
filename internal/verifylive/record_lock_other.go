//go:build !unix

package verifylive

import "errors"

// 기록 파일 잠금이 없는 플랫폼에서는 대사 추가를 거절함(fail-closed).
func flockExclusiveNB(uintptr) error {
	return errors.New("record file locking is unsupported on this platform")
}
