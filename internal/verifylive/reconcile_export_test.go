package verifylive

// reconcile_export_test.go — 외부 시험 패키지(verifylive_test)의 롤백 시험이 쓰는 픽스처 노출(export_test 관례).

import (
	"os"
	"path/filepath"
	"testing"
)

// RCTargetIDForTest 는 a063 형 기록의 대사 대상 id 다.
const RCTargetIDForTest = rcTargetID

// RCA063RecordForTest 는 a063 형 기록 파일의 원문 바이트다.
func RCA063RecordForTest(t *testing.T) []byte {
	t.Helper()
	path := filepath.Join(t.TempDir(), RecordFileName(MarketKR))
	rcWriteEntries(t, path, rcA063Entries())
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// RCExpiredRecordAndLineForTest 는 a063 형 기록 원문과, 그 위에서 만료 대사(a164)가 실제로 쓴 줄(개행 포함)이다.
func RCExpiredRecordAndLineForTest(t *testing.T) ([]byte, string) {
	t.Helper()
	h := newRCExpiredHarness(t, true)
	before := h.bytes()
	_, _, raw := rcAccept(t, h)
	return before, raw + "\n"
}

// RCReconcileLineForTest 는 설계 모양 대사 줄(개행 포함)이다.
func RCReconcileLineForTest() string {
	return rcReconcileLineJSON(rcTarget(), rcNow, ReconcileBasisDomain+":sha256:00")
}
