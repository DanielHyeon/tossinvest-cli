//go:build tossos_testseams

package engine

// a157 엔진 경로 — 실제 a112 거래 fixture(생산 적재기 · 실제 Guardian · 실제 원장)의 첫 다리 승인 `admit` 이 결속 결함과 진입 거절이
// 겹친 요청을 어떻게 분류하는지 잼.
//
// 엔진 경로에서 뒤집을 수 있는 결속 결함은 B2(정책 버전 — 거래 ID) · B8(주간 예약 결속)뿐임: B3~B7 꼴은 validateStrategyFirstLegAuthority
// 가 precheck 앞에서 먼저 막음(design F8 · D3). 그래서 엔진 행은 이 둘로만 만듦. 적재기가 모은 권한을 precheck 직전에 바꾸는 감싸개로 심음.

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// a157MutatingLoader 는 생산 적재기가 모은 권한에 결속 결함 하나를 심음(validateStrategyFirstLegAuthority 를 통과하는 꼴만).
type a157MutatingLoader struct {
	inner  strategyFirstLegAuthorityLoader
	mutate func(*execgw.QFinalCampaignFirstLegIssuance)
}

func (l a157MutatingLoader) collectStrategyFirstLegAuthority(ctx context.Context, accepted strategyFirstLegAccepted) (execgw.QFinalCampaignFirstLegIssuance, error) {
	authority, err := l.inner.collectStrategyFirstLegAuthority(ctx, accepted)
	if err == nil && l.mutate != nil {
		l.mutate(&authority)
	}
	return authority, err
}

// a157EscalationCounter 는 실제 Guardian 을 그대로 쓰되 일일 손실 강화 · 발급 호출을 셈. escalationFailure 가 있으면 강화가 실패한 것처럼
// 실제 Guardian 의 실패 갈래와 같은 모양(errors.Join(거절, 강화 오류) — a137_daily_loss.go EscalateDailyLossLimit)을 돌려줌(원장 쓰기 없음).
type a157EscalationCounter struct {
	*execgw.RiskGuardian
	escalationFailure error
	escalations       atomic.Int32
	issues            atomic.Int32
}

func (g *a157EscalationCounter) EscalateDailyLossLimit(ctx context.Context, refusal error) error {
	g.escalations.Add(1)
	if g.escalationFailure != nil {
		return errors.Join(refusal, g.escalationFailure)
	}
	return g.RiskGuardian.EscalateDailyLossLimit(ctx, refusal)
}

func (g *a157EscalationCounter) IssuePrecheckedQFinalCampaignFirstLeg(ctx context.Context, precheck execgw.QFinalCampaignFirstLegPrecheck) (journal.QFinalCampaignFirstLegReceipt, error) {
	g.issues.Add(1)
	return g.RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg(ctx, precheck)
}

// 엔진 경로의 두 결속 결함 — 결함 이름과 심는 변형.
var a157EngineBindings = map[string]func(*execgw.QFinalCampaignFirstLegIssuance){
	"B2-policy-version": func(r *execgw.QFinalCampaignFirstLegIssuance) { r.Entry.Admission.TransactionID = "tx|bad" },
	// 주간이 아닌 레인(KR 단기)에 예약 결속이 붙어 옴 — guardianWeeklyFirstLegBinding 이 충돌로 거절.
	"B8-weekly": func(r *execgw.QFinalCampaignFirstLegIssuance) {
		r.Weekly = &journal.WeeklyFirstLegReservationBinding{ReservationID: "weekly-a157", StableWeek: "KR-XKRX-2026-W14", PlannedOrdinal: 1,
			ScopeVersion: 1, RequestDigest: "request", RecordDigest: "record", CalendarGeneration: "generation", CalendarDigest: "calendar"}
	},
}

// 각 결속 결함 거절의 원인 sentinel — B2 는 journal.QFinalPolicyVersion 의 요청 무효, B8 은 주간 예약 충돌.
var a157EngineBindingCause = map[string]error{
	"B2-policy-version": journal.ErrInvalidRequest,
	"B8-weekly":         journal.ErrWeeklyReservationConflict,
}

// a157DailyLoss 는 일일 손실 원장 상황 — 없음 · 한도 도달(큰 동결 손실) · 측정 불가(성과 행 없는 청산).
type a157DailyLoss string

const (
	a157NoLoss       a157DailyLoss = ""
	a157LossLimit    a157DailyLoss = "limit"
	a157Unmeasurable a157DailyLoss = "unmeasurable"
)

// a157Admit 은 거래 fixture 를 세우고(예산 · 일일 손실 원장 행), 결속 결함(빈 이름이면 없음)을 심어 KR 첫 범위(005930)를 승인에 넘김.
func a157Admit(t *testing.T, riskBudget string, loss a157DailyLoss, binding string) (StrategyFirstLegAdmissionResult, *a157EscalationCounter) {
	t.Helper()
	return a157AdmitWith(t, riskBudget, loss, binding, nil)
}

// a157AdmitWith 는 a157Admit 에 강화 실패(escalationFailure, nil 이면 실제 강화)를 더한 것.
func a157AdmitWith(t *testing.T, riskBudget string, loss a157DailyLoss, binding string, escalationFailure error) (StrategyFirstLegAdmissionResult, *a157EscalationCounter) {
	t.Helper()
	fixture := newA112TradingFixture(t, a112TradingOptions{riskBudget: riskBudget})
	if loss != a157NoLoss {
		// 손실 행은 다른 종목(999999)에 둠 — 일일 손실은 계좌 단위이고, 같은 종목이면 캠페인 세대 판독이 달라져 승인 앞에서 다른 거절이 남.
		closed := fixture.now.Add(-time.Minute)
		day, err := clock.AccountTradingDay(closed)
		today, todayErr := clock.AccountTradingDay(fixture.now)
		if err != nil || todayErr != nil || day != today {
			t.Fatalf("arrangement: the loss row must close inside today's account day (closed=%s now=%s)", closed, fixture.now)
		}
		pnl := "-100000000"
		if loss == a157Unmeasurable {
			pnl = ""
		}
		if err := fixture.journal.InsertClosedManagedPositionForTest(context.Background(), "p-a157", "acct-risk-loader", "kr", "999999",
			closed.Add(-time.Hour).UTC().Format(time.RFC3339), closed.UTC().Format(time.RFC3339), pnl); err != nil {
			t.Fatal(err)
		}
	}
	counter := &a157EscalationCounter{RiskGuardian: fixture.guardian, escalationFailure: escalationFailure}
	var loader strategyFirstLegAuthorityLoader = fixture.loader
	if binding != "" {
		mutate, ok := a157EngineBindings[binding]
		if !ok {
			t.Fatalf("arrangement: unknown binding defect %q", binding)
		}
		loader = a157MutatingLoader{inner: fixture.loader, mutate: mutate}
	}
	result := fixture.winner.Proposal()
	return newStrategyFirstLegAdmissionBridge(counter, loader).admit(context.Background(), result), counter
}

// 결속 결함 단독(B2 · B8) — AUTHORITY_MISMATCH(편집 전과 같음).
func TestA157EngineABindingDefectAloneIsAnAuthorityMismatch(t *testing.T) {
	for binding := range a157EngineBindings {
		outcome, counter := a157Admit(t, "", a157NoLoss, binding)
		// 원인이 그 결속 판정의 것인지까지 봄 — 결함이 precheck 의 결속 판정에 실제로 닿았다는 증거(다른 거절의 AUTHORITY_MISMATCH 가 아님).
		if outcome.Code != StrategyFirstLegAuthorityMismatch || !errors.Is(outcome.cause, a157EngineBindingCause[binding]) ||
			counter.escalations.Load() != 0 || counter.issues.Load() != 0 {
			t.Errorf("%s: outcome=%+v cause=%v escalations=%d issues=%d — want AUTHORITY_MISMATCH from the binding check, no escalation, no issue",
				binding, outcome, outcome.cause, counter.escalations.Load(), counter.issues.Load())
		}
	}
}

// 일일 손실 한도 + 결속 결함(없음 · B2 · B8) — DAILY_LOSS_LIMIT, 강화 한 번(실제 원장의 운영 모드 강화 성공 → 계좌 진입 차단 원인), 발급 0.
// 사람 결정 1-1: 결속 결함이 겹쳐도 편집 전과 같아야 함.
func TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect(t *testing.T) {
	for _, binding := range []string{"", "B2-policy-version", "B8-weekly"} {
		outcome, counter := a157Admit(t, "", a157LossLimit, binding)
		var block *execgw.DailyLossEntryBlock
		if outcome.Code != StrategyFirstLegDailyLossLimit || counter.escalations.Load() != 1 || counter.issues.Load() != 0 ||
			!errors.As(outcome.cause, &block) {
			t.Errorf("binding=%q: outcome=%+v cause=%v escalations=%d issues=%d — want DAILY_LOSS_LIMIT, one escalation, an account block",
				binding, outcome, outcome.cause, counter.escalations.Load(), counter.issues.Load())
		}
	}
}

// 일일 손실 한도 + 결속 결함 + **강화 실패** — 계좌 차단이 아니라 거절과 강화 오류를 함께 든 결함이고, 그 모양 · 문구가 결속 결함이 없을 때와
// 같음(design D2 (c) 「성공/실패 갈래 모두 같음」, 구현 리뷰 Opus P3-3). 발급 0.
func TestA157EngineAFailedEscalationIsTheSameJoinedFaultWithABindingDefect(t *testing.T) {
	failure := errors.New("operating mode did not persist")
	var baseline StrategyFirstLegAdmissionResult
	for _, binding := range []string{"", "B2-policy-version", "B8-weekly"} {
		outcome, counter := a157AdmitWith(t, "", a157LossLimit, binding, failure)
		var block *execgw.DailyLossEntryBlock
		var issue *execgw.IssueRefusal
		if outcome.Code != StrategyFirstLegDailyLossLimit || counter.escalations.Load() != 1 || counter.issues.Load() != 0 ||
			errors.As(outcome.cause, &block) || !errors.Is(outcome.cause, failure) || !errors.As(outcome.cause, &issue) ||
			issue.Stage != execgw.StageChain || issue.Reason != "DAILY_LOSS_LIMIT_REACHED" {
			t.Errorf("binding=%q: outcome=%+v cause=%v — want DAILY_LOSS_LIMIT with the chain refusal and the escalation error joined, no account block",
				binding, outcome, outcome.cause)
			continue
		}
		if binding == "" {
			baseline = outcome
			continue
		}
		if outcome.Detail != baseline.Detail || outcome.cause.Error() != baseline.cause.Error() {
			t.Errorf("binding=%q: detail=%q cause=%q — want exactly the no-defect fault detail=%q cause=%q",
				binding, outcome.Detail, outcome.cause, baseline.Detail, baseline.cause)
		}
	}
}

// 일일 손실 측정 불가 + 결속 결함(없음 · B2 · B8) — DAILY_LOSS_UNMEASURABLE(표지 원인), 강화 0, 발급 0. 편집 전과 같아야 함.
func TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect(t *testing.T) {
	for _, binding := range []string{"", "B2-policy-version", "B8-weekly"} {
		outcome, counter := a157Admit(t, "", a157Unmeasurable, binding)
		var unmeasurable *execgw.DailyLossUnmeasurable
		if outcome.Code != StrategyFirstLegDailyLossUnmeasurable || counter.escalations.Load() != 0 || counter.issues.Load() != 0 ||
			!errors.As(outcome.cause, &unmeasurable) {
			t.Errorf("binding=%q: outcome=%+v cause=%v escalations=%d issues=%d — want DAILY_LOSS_UNMEASURABLE with its marker",
				binding, outcome, outcome.cause, counter.escalations.Load(), counter.issues.Load())
		}
	}
}

// Guardian 상한 0(예산 1 KRW) 단독 — a146 스킵 그대로.
func TestA157EngineACapZeroAloneIsStillASkip(t *testing.T) {
	outcome, counter := a157Admit(t, "1", a157NoLoss, "")
	var skip *strategyEntrySkip
	if outcome.Code != StrategyFirstLegGuardianCapZero || !errors.As(outcome.cause, &skip) || counter.issues.Load() != 0 {
		t.Fatalf("outcome=%+v cause=%v — want the a146 skip", outcome, outcome.cause)
	}
}

// Guardian 상한 0(예산 1 KRW) + 결속 결함(B2 · B8) — a157 로 뒤집힌 행(이 change 의 목적, design D2 (a)). 편집 전에는 상한 0 이 결함을 가려
// 조용한 스킵이었음(`analysis/red/characterization-head.txt`); 이제는 결속 결함의 AUTHORITY_MISMATCH 이고 스킵 기록 · 발급은 없음.
func TestA157EngineACapZeroNoLongerHidesABindingDefect(t *testing.T) {
	for binding := range a157EngineBindings {
		outcome, counter := a157Admit(t, "1", a157NoLoss, binding)
		var skip *strategyEntrySkip
		if outcome.Code != StrategyFirstLegAuthorityMismatch || !errors.Is(outcome.cause, a157EngineBindingCause[binding]) ||
			errors.As(outcome.cause, &skip) || counter.escalations.Load() != 0 || counter.issues.Load() != 0 {
			t.Errorf("%s: outcome=%+v cause=%v — want AUTHORITY_MISMATCH from the binding check, not the cap-zero skip", binding, outcome, outcome.cause)
		}
	}
}
