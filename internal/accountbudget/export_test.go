//go:build linux

package accountbudget

import "errors"

// InstallFirstWriteForTest는 요청 시작 전에만 시험 관측자를 설정함.
func InstallFirstWriteForTest(t *Transport, observe func()) error {
	if t == nil || observe == nil {
		return errors.New("invalid test transport or observer")
	}
	t.firstWrite = observe
	return nil
}
