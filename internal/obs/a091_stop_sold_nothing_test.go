package obs_test

// a091 tasks 2.1 · 2.2 · 2.3 — 보호 청산 0주의 새 종류가 critical 로 등록되고, 그 등록이 다른 종류의 등급을 바꾸지 않음을 핀함.
//
// 등급은 종류에만 붙음(SeverityOf 는 criticalEvents 맵 조회 하나). 그래서 새 종류의 등록 누락을 잡는 유일한 장치가 이 파일임 —
// 누락되면 새 종류가 조용히 normal 로 강등되고 durable outbox 행이 생기지 않음(8/2 결함의 재발).

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// 2.1 · 2.3: 새 종류는 critical, 값 문자열은 계약(로그 필터 · outbox event_type · ntfy Tags)이라 핀함.
func TestA091TheStopSoldNothingKindIsCritical(t *testing.T) {
	if got := obs.SeverityOf(obs.EventExitStopSoldNothing); got != obs.SeverityCritical {
		t.Fatalf("SeverityOf(%s) = %s, want critical", obs.EventExitStopSoldNothing, got)
	}
	if obs.EventExitStopSoldNothing != "exit.stop_sold_nothing" {
		t.Errorf("the kind string is a contract (log filters, outbox event_type, ntfy Tags); got %q", obs.EventExitStopSoldNothing)
	}
	if got := obs.EventExitStopSoldNothing.Subject(); got != "exit" {
		t.Errorf("subject = %q, want exit — the class rules over CriticalEvents() key on the subject", got)
	}
	listed := false
	for _, e := range obs.CriticalEvents() {
		if e == obs.EventExitStopSoldNothing {
			listed = true
		}
	}
	if !listed {
		t.Error("the new kind is graded critical but is not in CriticalEvents()")
	}
}

// 2.2: 기존 19 종(base b30318d6 event.go:337-361)의 등급은 그대로이고, 부분 캡 종류는 여전히 normal · 미등록 종류도 normal.
func TestA091NoOtherGradeMoved(t *testing.T) {
	critical := []obs.EventType{
		obs.EventOrderInDoubt, obs.EventOrderUnresolved, obs.EventFillRefused, obs.EventBrokerStateUnknown,
		obs.EventBrokerAuthRejected, obs.EventCredentialExpiring, obs.EventReconcilePermanent, obs.EventFlattenStarted,
		obs.EventFlattenStalled, obs.EventAlertUndelivered, obs.EventOperatingMode, obs.EventEngineLoopFailed,
		obs.EventEngineLoopDegraded, obs.EventExitObservationOutage, obs.EventExitJudgementRefused,
		obs.EventExitSnapshotQuarantined, obs.EventExitProposalRefused, obs.EventExitLiquidationDelayed,
		obs.EventExitPositionAdoptionFailed,
	}
	if len(critical) != 19 {
		t.Fatalf("the base table has 19 critical kinds; this list has %d", len(critical))
	}
	for _, e := range critical {
		if got := obs.SeverityOf(e); got != obs.SeverityCritical {
			t.Errorf("SeverityOf(%s) = %s, want critical (unchanged)", e, got)
		}
	}
	if got := len(obs.CriticalEvents()); got != len(critical)+1 {
		t.Errorf("CriticalEvents() has %d kinds, want %d (19 + exit.stop_sold_nothing)", got, len(critical)+1)
	}
	for _, e := range []obs.EventType{obs.EventExitProposalCapped, obs.EventExitPositionUnmanaged, obs.EventType("exit.not_a_kind")} {
		if got := obs.SeverityOf(e); got != obs.SeverityNormal {
			t.Errorf("SeverityOf(%s) = %s, want normal", e, got)
		}
	}
}
