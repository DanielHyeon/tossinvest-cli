package verifylive

// reconcile_a164_absent_test.go — a164 tasks 1.2: 편집 **전에** 부재 모드(플래그 없음)의 거절 순서와 쓰기 경계 게이트를
// 행동으로 고정한다. 시험별 coverage 측정(analysis/measure/)에서 두 자리가 행동 시험 0 이었음:
//   - Reconcile B29(쓰기 직전 재검)의 거절 — 구조 핀만 있었음.
//   - 부재 판정 네 가드가 **동시에** 걸릴 때 어느 것이 먼저인지 — 가드마다 하나씩만 비튼 시험뿐이었음.
// a164 는 checkReconcileAbsence 의 A5 판정을 보조 함수로 옮기고 checkReconcileWindow 에 모드 인자를 더하므로, 그 편집이
// 부재 모드 동작을 바꾸지 않음을 이 두 시험이 편집 전후 같은 결과로 보임.

import (
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// TestReconcileAbsenceGuardsRefuseInTheirContractOrder 는 a121 「가드 순서 계약」(OPEN 조건 → 일반 OPEN → CLOSED target →
// 발동 흔적)을 네 조건이 겹친 입력에서 하나씩 벗겨 가며 잰다.
func TestReconcileAbsenceGuardsRefuseInTheirContractOrder(t *testing.T) {
	fired := rcCondRow("CO-OTHER-FIRED", "EXPIRED")
	fired.TriggeredOrderID = "ORD-CHILD-9"
	type layer struct {
		openCond, openPlain, closedTarget, firedTrace bool
	}
	run := func(l layer) *rcHarness {
		h := newRCHarness(t, rcA063Entries())
		if l.openCond {
			h.reader.set(-1, ReconcileGroupConditionalOpen, rcCondPage("", false, rcCondRow("CO-SUCCESSOR", "WATCHING")))
		}
		if l.openPlain {
			h.reader.set(-1, ReconcileGroupPlainOpen, rcPlainPage("", false, official.ReconcileOrderRow{ID: "ORD-REST", Symbol: rcSymbol, Status: "PENDING"}))
		}
		var closed []official.ReconcileConditionalRow
		if l.closedTarget {
			closed = append(closed, rcCondRow(rcTargetID, "EXPIRED"))
		}
		if l.firedTrace {
			closed = append(closed, fired)
		}
		if len(closed) > 0 {
			h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, closed...))
		}
		return h
	}
	cases := []struct {
		l    layer
		want ReconcileRefusalCode
	}{
		{layer{true, true, true, true}, RefuseOpenConditional},
		{layer{false, true, true, true}, RefuseOpenPlainOrder},
		{layer{false, false, true, true}, RefuseClosedTargetExpired},
		{layer{false, false, false, true}, RefuseClosedFiredTrace},
	}
	for _, c := range cases {
		rcRefuse(t, run(c.l), c.want)
	}
}

// TestReconcileRefusesAtTheWriteBoundaryGateItself 는 Reconcile B29 의 행동 쪽이다 — 첫 재검(B16)을 통과한 뒤 쓰기 직전
// 시계가 창을 넘으면 마지막 게이트가 거절하고 기록은 그대로다. 부재 모드의 now() 호출은 정적 Q1(1) → 창 시작(2) →
// 첫 재검(3) → 쓰기 직전(4) 순이므로 넷째 호출에서만 시계를 넘긴다.
func TestReconcileRefusesAtTheWriteBoundaryGateItself(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	p := h.params()
	calls := 0
	p.Now = func() time.Time {
		calls++
		if calls >= 4 {
			return rcNow.Add(rcFreshness + time.Second)
		}
		return rcNow
	}
	before := h.bytes()
	_, err := Reconcile(t.Context(), p)
	rcRequireRefusal(t, err, RefuseFreshnessExceeded)
	rcRequireUnchanged(t, h, before)
	if calls != 4 {
		t.Fatalf("absence mode read the clock %d time(s); the contract points are four (static Q1, window start, two rechecks)", calls)
	}
}

// TestReconcileAbsenceKeepsItsA121Wording 은 design D1 「플래그 없음 = 문구 불변(D8 예외만)」 이다(gstack 리뷰 F1 · Codex P3) —
// 부재 모드의 신선도 거절 문구와 report 의 대사 절 머리는 a121 바이트 그대로다.
func TestReconcileAbsenceKeepsItsA121Wording(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	h.reader.hook = func(c rcCall) {
		if c.Kind == "list" && c.Pass == 1 && c.Group == ReconcileGroupConditionalClosed {
			h.clock.advance(rcFreshness + time.Second)
		}
	}
	r := rcRefuse(t, h, RefuseFreshnessExceeded)
	if !strings.Contains(r.Detail, "elapsed since the first list read — the freshness bound is") {
		t.Fatalf("absence freshness wording changed: %q", r.Detail)
	}
	var report strings.Builder
	writeReconciled(&report, ReconciledArtifacts(rcEntriesWithReconcileLine(t, rcA063Entries())))
	if want := "\n대사로 종결된 객체 1건 (공식 목록 두 번 읽기의 부재 근거 — 성공한 요청·증명이 아니다):\n  conditional-order " +
		rcTargetID + " (" + rcSymbol + ") reconciled absent\n"; report.String() != want {
		t.Fatalf("absence report section changed\n got %q\nwant %q", report.String(), want)
	}
}
