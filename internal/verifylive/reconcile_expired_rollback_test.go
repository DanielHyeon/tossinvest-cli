package verifylive_test

// reconcile_expired_rollback_test.go — a164 롤백 방향: reconcile_rollback_test.go 의 구 바이너리 사본(Outstanding ·
// PendingCleanup · oldDecode)으로 만료 대사 줄을 읽는다.

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// TestAnOlderBinaryStillSeesTheExpiredArtifactAsOutstanding 는 a164 D10 이다 — reconciled_expired 를 모르는 구 바이너리(a121
// 이전 사본; a121 바이너리도 이 필드를 모르는 점은 같다)는 만료 대사 줄을 비-종결로 읽어 artifact 를 다시 outstanding·보유로
// 본다. 실제 Reconcile 이 쓴 줄로 잰다.
func TestAnOlderBinaryStillSeesTheExpiredArtifactAsOutstanding(t *testing.T) {
	record, line := verifylive.RCExpiredRecordAndLineForTest(t)
	without := oldDecode(t, record)
	if !oldHas(Outstanding(without), verifylive.RCTargetIDForTest) || !oldHas(PendingCleanup(without), verifylive.RCTargetIDForTest) {
		t.Fatal("control: the old binary must plan the released artifact for cleanup")
	}
	with := oldDecode(t, append(record, []byte(line)...))
	if !oldHas(Outstanding(with), verifylive.RCTargetIDForTest) {
		t.Fatal("an older binary hides a reconciled-expired artifact — rollback must fail in the 'still visible' direction")
	}
	if oldHas(PendingCleanup(with), verifylive.RCTargetIDForTest) {
		t.Fatal("an older binary plans a DELETE for the reconciled-expired artifact; it must be held again")
	}
}
