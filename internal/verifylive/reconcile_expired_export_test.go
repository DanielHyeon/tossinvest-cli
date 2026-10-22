package verifylive

// reconcile_expired_export_test.go — 외부 시험 패키지(verifylive_test)의 a164 롤백 시험이 쓰는 노출(export_test 관례).

import "testing"

// RCExpiredRecordAndLineForTest 는 a063 형 기록 원문과, 그 위에서 만료 대사(a164)가 실제로 쓴 줄(개행 포함)이다.
func RCExpiredRecordAndLineForTest(t *testing.T) ([]byte, string) {
	t.Helper()
	h := newRCExpiredHarness(t, true)
	before := h.bytes()
	_, _, raw := rcAccept(t, h)
	return before, raw + "\n"
}
