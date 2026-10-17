//go:build tossos_testseams

package execgw_test

// a137 구현 리뷰 P1-1 · P1-2 · P2-3: 살아 있는 진입 문(PrecheckQFinalEntry)이 실제로 측정 바닥을 쓰는지, 그리고 계좌 일 키 창
// (KST 08:00 롤오버)이 실제 원장 읽기에서 지켜지는지 — 변이 M2c(바닥 미적용) · M1b(창의 롤오버 삭제)가 살아남지 않게.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

func a137Rows(rows ...journal.DailyLossRow) func(context.Context, string, time.Time, time.Time) ([]journal.DailyLossRow, error) {
	return func(context.Context, string, time.Time, time.Time) ([]journal.DailyLossRow, error) { return rows, nil }
}

func a137Row(id, market, closedAt, pnl string) journal.DailyLossRow {
	return journal.DailyLossRow{PositionID: id, Market: market, Symbol: "005930", ClosedAt: closedAt, Outcome: &pnl}
}

// fixedNow = KST 2026-03-30 10:30 — 일 키 2026-03-30, 창 [2026-03-29T23:00Z, 2026-03-30T23:00Z).
func a137Precheck(t *testing.T, rows func(context.Context, string, time.Time, time.Time) ([]journal.DailyLossRow, error),
	mutate func(*execgw.QFinalEntryIssuance)) error {
	t.Helper()
	rig := newGuardian(t, nil)
	if rows != nil {
		rig.guardian.SetDailyLossRowsForTest(rows)
	}
	request := qFinalKRRequest(t, rig, "a137", 20)
	if mutate != nil {
		mutate(&request)
	}
	request.DailyLoss = rig.guardian.MeasureDailyLoss(context.Background())
	_, err := rig.guardian.PrecheckQFinalEntry(request)
	return err
}

func a137ChainReason(t *testing.T, err error) string {
	t.Helper()
	var refusal *execgw.IssueRefusal
	if !errors.As(err, &refusal) || refusal.Stage != execgw.StageChain {
		t.Fatalf("err = %v, want a chain refusal", err)
	}
	return refusal.Reason
}

func TestA137ThePrecheckDoorRefusesAMeasuredLossAtTheLimit(t *testing.T) {
	err := a137Precheck(t, a137Rows(a137Row("p-1", "kr", "2026-03-30T00:30:00Z", "-100000")), nil)
	if reason := a137ChainReason(t, err); reason != string(risk.ReasonDailyLossLimitReached) {
		t.Fatalf("reason = %s, want DAILY_LOSS_LIMIT_REACHED (the signed caller value is 0)", reason)
	}
}

func TestA137ThePrecheckDoorRefusesAnUnmeasurableDayAsTyped(t *testing.T) {
	noOutcome := journal.DailyLossRow{PositionID: "p-1", Market: "kr", Symbol: "005930", ClosedAt: "2026-03-30T00:30:00Z"}
	err := a137Precheck(t, a137Rows(noOutcome), nil)
	if reason := a137ChainReason(t, err); reason != string(risk.ReasonInputUnavailable) {
		t.Fatalf("reason = %s, want INPUT_UNAVAILABLE", reason)
	}
	var unmeasurable *execgw.DailyLossUnmeasurable
	if !errors.As(err, &unmeasurable) {
		t.Fatalf("err = %v, want the typed unmeasurable marker", err)
	}
	// 측정과 무관한 입력 불가(호출자 값이 음수)는 표지가 없음 — 오늘처럼 결함.
	err = a137Precheck(t, a137Rows(), func(r *execgw.QFinalEntryIssuance) { r.Account.DailyRealizedLoss.Amount = "-1" })
	if reason := a137ChainReason(t, err); reason != string(risk.ReasonInputUnavailable) || errors.As(err, &unmeasurable) {
		t.Fatalf("err = %v, want an unmarked INPUT_UNAVAILABLE", err)
	}
}

// KR 요청의 US 손실: 같은 파도 짝 FX 가 없으면 거절, 있으면 올림 환산되어 한도 아래면 통과.
func TestA137ThePrecheckDoorConvertsAUSLossOnlyWithThePairedAuthority(t *testing.T) {
	us := a137Row("p-us", "us", "2026-03-30T00:30:00Z", "-10")
	if reason := a137ChainReason(t, a137Precheck(t, a137Rows(us), nil)); reason != string(risk.ReasonInputUnavailable) {
		t.Fatalf("reason = %s, want INPUT_UNAVAILABLE without the paired US authority", reason)
	}
	paired, err := officialfx.EvidenceForAuthorityTest("USD", "KRW", "1400", "1.01", fixedNow.Add(-time.Minute), fixedNow.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if err := a137Precheck(t, a137Rows(us), func(r *execgw.QFinalEntryIssuance) { r.PairedFXAuthority = paired }); err != nil {
		t.Fatalf("a 14140 KRW loss under the 100000 limit must pass with the paired authority: %v", err)
	}
}

// P2-3: 자본 0 이하는 체인이 DAILY_LOSS_LIMIT_REACHED 로 거절함(기존 B3) — 살아 있는 문에서도 같은 사유라 첫 다리 승인이 강화함(보수).
func TestA137ThePrecheckDoorReportsNoCapitalAsTheDailyLossLimit(t *testing.T) {
	err := a137Precheck(t, a137Rows(), func(r *execgw.QFinalEntryIssuance) { r.Account.AccountEquity.Amount = "0" })
	if reason := a137ChainReason(t, err); reason != string(risk.ReasonDailyLossLimitReached) {
		t.Fatalf("reason = %s, want DAILY_LOSS_LIMIT_REACHED for an account with no capital", reason)
	}
}

// P1-2: 실제 원장 읽기(시험 자리 없음)로 창의 네 경계를 잼 — 한도와 같은 손실 하나가 창 안이면 거절, 밖이면 통과.
func TestA137TheAccountDayWindowHoldsThroughTheRealJournalRead(t *testing.T) {
	for _, c := range []struct {
		closed string
		inside bool
	}{
		{"2026-03-29T22:59:59Z", false}, // KST 03-30 07:59:59 — 전 일 키
		{"2026-03-29T23:00:00Z", true},  // KST 03-30 08:00:00 — 일 키 시작
		{"2026-03-30T22:59:59Z", true},  // KST 03-31 07:59:59 — 아직 일 키 03-30
		{"2026-03-30T23:00:00Z", false}, // KST 03-31 08:00:00 — 다음 일 키
	} {
		t.Run(c.closed, func(t *testing.T) {
			rig := newGuardian(t, nil)
			if err := rig.journal.InsertClosedManagedPositionForTest(context.Background(), "p-edge", "acct-7", "kr", "005930",
				"2026-03-29T20:00:00Z", c.closed, "-100000"); err != nil {
				t.Fatal(err)
			}
			request := qFinalKRRequest(t, rig, "a137-window", 20)
			_, err := rig.guardian.PrecheckQFinalEntry(request)
			if c.inside {
				if reason := a137ChainReason(t, err); reason != string(risk.ReasonDailyLossLimitReached) {
					t.Fatalf("reason = %s, want DAILY_LOSS_LIMIT_REACHED for a close inside the account day", reason)
				}
			} else if err != nil {
				t.Fatalf("a close outside the account day must not count: %v", err)
			}
		})
	}
}
