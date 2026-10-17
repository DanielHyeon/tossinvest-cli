//go:build tossos_testseams

package engine

// a137 전략 경로: 일일 손실 한도 거절과 그 뒤의 일일 손실 모드 차단은 계좌 진입 차단 결과로 주기를 조용히 끝냄(레인 결함 · 잠금 아님).
// 다른 모드 차단 · 기동 복원 실패 · 그 밖의 관문 거절은 오늘과 같음. 앞서 건너뛴 범위 거절의 보고는 그대로임.

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func a137Selected() []strategyflow.Result {
	return []strategyflow.Result{
		{Lineage: strategyflow.Lineage{Identity: "a", AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "005930", PositionGeneration: 1}},
		{Lineage: strategyflow.Lineage{Identity: "b", AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "000660", PositionGeneration: 1}},
		{Lineage: strategyflow.Lineage{Identity: "c", AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "035420", PositionGeneration: 1}},
	}
}

func a137Deliver(t *testing.T, answer map[string]error) ([]string, error) {
	t.Helper()
	var seen []string
	err := deliverEachStrategyHandoff(strategyhandoff.AdmitEachOwnerScope(true, a137Selected()), func(delivered strategyhandoff.Delivered) error {
		identity := delivered.Result().Lineage.Identity
		seen = append(seen, identity)
		return answer[identity]
	})
	return seen, err
}

func a137Block() error {
	return &execgw.DailyLossEntryBlock{Cause: &execgw.RejectedError{Reason: execgw.ReasonOperatingModeBlocked, Detail: "ENTRY_BLOCKED (AUTO): DAILY_LOSS_LIMIT_REACHED"}}
}

// freeze 조건 ①: 일일 손실 모드 차단(또는 한도 거절)은 오류 없이 주기를 끝내고 뒤 범위를 전달하지 않음.
func TestA137ADailyLossModeBlockEndsTheCycleQuietly(t *testing.T) {
	for name, err := range map[string]error{
		"bare":                        a137Block(),
		"wrapped like dispatch wraps": fmt.Errorf("engine: first-leg admission DAILY_LOSS_LIMIT: %w", a137Block()),
	} {
		seen, got := a137Deliver(t, map[string]error{"a": err})
		if got != nil || strings.Join(seen, ",") != "a" {
			t.Errorf("%s: delivered=%v err=%v — want the cycle ended after a with no error", name, seen, got)
		}
	}
}

// freeze 조건 ②: 다른 모드 차단 · 기동 복원 실패는 오늘처럼 주기 오류(서명 활성화 레인은 잠김).
func TestA137OtherModeBlocksBehaveAsBefore(t *testing.T) {
	for name, err := range map[string]error{
		"boot restore failure": &execgw.RejectedError{Reason: execgw.ReasonOperatingModeBlocked, Detail: "operating mode could not be restored at startup"},
		"operator halt":        &execgw.RejectedError{Reason: execgw.ReasonOperatingModeBlocked, Detail: "HALT_ALL (OPERATOR): halt"},
		"reconcile latch":      &execgw.RejectedError{Reason: execgw.ReasonReconcileMismatch, Detail: "reconcile"},
	} {
		seen, got := a137Deliver(t, map[string]error{"a": err})
		if !errors.Is(got, err) || strings.Join(seen, ",") != "a" {
			t.Errorf("%s: delivered=%v err=%v — want today's stopping error", name, seen, got)
		}
		var block *execgw.DailyLossEntryBlock
		if errors.As(got, &block) {
			t.Errorf("%s: an ordinary block must not read as a daily-loss account block", name)
		}
	}
}

// freeze 조건 ③: 앞서 건너뛴 범위 거절이 있으면 그 보고는 지금처럼 주기의 오류로 남음 — 조용한 건너뛰기 금지(J4).
func TestA137AMixedCycleStillReportsItsSkippedScopes(t *testing.T) {
	scope := &strategyScopeRefusal{scope: strategyrouter.OwnerKey{AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "005930"},
		detail: "this owner scope's risk bucket is exhausted"}
	seen, err := a137Deliver(t, map[string]error{"a": fmt.Errorf("engine: first-leg admission X: %w", scope), "b": a137Block()})
	if strings.Join(seen, ",") != "a,b" {
		t.Fatalf("delivered=%v — want a skipped, b ending the cycle before c", seen)
	}
	var skipped *strategyScopeRefusal
	if !errors.As(err, &skipped) || skipped != scope {
		t.Fatalf("err=%v — want the skipped scope refusal still reported", err)
	}
}

// 첫 다리 승인: 일일 손실 한도 체인 거절(StageChain)이면 강화를 부르고 DAILY_LOSS_LIMIT + 계좌 차단 원인.
func TestA137ADailyLossRefusalEscalatesAndIsNotALaneFault(t *testing.T) {
	for _, market := range pairedStrategyFirstLegResults(t) {
		loader := &firstLegAuthorityLoaderStub{issuance: firstLegBridgeIssuanceFixture(market.result)}
		refusal := &execgw.IssueRefusal{Stage: execgw.StageChain, Reason: string(risk.ReasonDailyLossLimitReached), Detail: "limit"}
		issuer := &firstLegGuardianSpy{precheckErr: refusal}
		outcome := newStrategyFirstLegAdmissionBridge(issuer, loader).admit(context.Background(), market.result)
		if outcome.Code != StrategyFirstLegDailyLossLimit || issuer.escalations != 1 || issuer.issueCalls != 0 {
			t.Fatalf("outcome=%+v issuer=%+v", outcome, issuer)
		}
		var block *execgw.DailyLossEntryBlock
		var scope *strategyScopeRefusal
		if !errors.As(outcome.cause, &block) || errors.As(outcome.cause, &scope) || isCentralStrategyIntegrity(outcome.cause) {
			t.Fatalf("cause=%v — want a daily-loss account block, not a scope refusal or central fault", outcome.cause)
		}
		if !strings.HasPrefix(issuer.events, "measure,precheck") {
			t.Fatalf("events=%q — the measurement must be taken right before the precheck", issuer.events)
		}
	}
}

// 강화가 실패하면 계좌 차단이 아니라 거절과 강화 오류를 함께 든 오류(조용히 삼켜지지 않음).
func TestA137AFailedEscalationIsJoinedNotLost(t *testing.T) {
	market := pairedStrategyFirstLegResults(t)[0]
	loader := &firstLegAuthorityLoaderStub{issuance: firstLegBridgeIssuanceFixture(market.result)}
	refusal := &execgw.IssueRefusal{Stage: execgw.StageChain, Reason: string(risk.ReasonDailyLossLimitReached)}
	escalationErr := errors.New("operating mode did not persist")
	issuer := &firstLegGuardianSpy{precheckErr: refusal, escalateErr: escalationErr}
	outcome := newStrategyFirstLegAdmissionBridge(issuer, loader).admit(context.Background(), market.result)
	var block *execgw.DailyLossEntryBlock
	if errors.As(outcome.cause, &block) || !errors.Is(outcome.cause, escalationErr) || !errors.Is(outcome.cause, refusal) {
		t.Fatalf("cause=%v — want refusal + escalation error joined, not a quiet block", outcome.cause)
	}
}

// 모든 precheck 거절은 원인을 실음(F8) — 문구는 그대로.
func TestA137APrecheckRefusalKeepsItsCause(t *testing.T) {
	market := pairedStrategyFirstLegResults(t)[0]
	loader := &firstLegAuthorityLoaderStub{issuance: firstLegBridgeIssuanceFixture(market.result)}
	refusal := &execgw.IssueRefusal{Stage: execgw.StageChain, Reason: string(risk.ReasonOpenExposureExceeded), Detail: "exposure"}
	outcome := newStrategyFirstLegAdmissionBridge(&firstLegGuardianSpy{precheckErr: refusal}, loader).admit(context.Background(), market.result)
	if outcome.Code != StrategyFirstLegAuthorityMismatch || !errors.Is(outcome.cause, refusal) || outcome.Detail != refusal.Error() {
		t.Fatalf("outcome=%+v", outcome)
	}
}

// 분류 불변 표(P2-1): 일일 손실 한도 + StageChain 만 강화. 그 밖은 모두 강화 없이 AUTHORITY_MISMATCH, 중앙 무결성도 범위 거절도 아님
// (버킷 고갈만 범위 거절 — 오늘 그대로).
func TestA137ClassificationIsUnchangedForEveryPrecheckRefusal(t *testing.T) {
	market := pairedStrategyFirstLegResults(t)[0]
	cases := map[string]struct {
		err      error
		code     StrategyFirstLegAdmissionCode
		escalate bool
	}{
		"daily loss at the chain":          {&execgw.IssueRefusal{Stage: execgw.StageChain, Reason: string(risk.ReasonDailyLossLimitReached)}, StrategyFirstLegDailyLossLimit, true},
		"daily loss reason at issuance":    {&execgw.IssueRefusal{Stage: execgw.StageIssuance, Reason: string(risk.ReasonDailyLossLimitReached)}, StrategyFirstLegAuthorityMismatch, false},
		"daily loss input unavailable":     {&execgw.IssueRefusal{Stage: execgw.StageChain, Reason: string(risk.ReasonInputUnavailable)}, StrategyFirstLegAuthorityMismatch, false},
		"kill switch":                      {&execgw.IssueRefusal{Stage: execgw.StageChain, Reason: string(risk.ReasonKillSwitchActive)}, StrategyFirstLegAuthorityMismatch, false},
		"plain error with the reason text": {errors.New("DAILY_LOSS_LIMIT_REACHED"), StrategyFirstLegAuthorityMismatch, false},
		"q_final refusal":                  {&execgw.QFinalRefusal{Code: "OWNER_CONFLICT"}, StrategyFirstLegAuthorityMismatch, false},
	}
	for name, c := range cases {
		loader := &firstLegAuthorityLoaderStub{issuance: firstLegBridgeIssuanceFixture(market.result)}
		issuer := &firstLegGuardianSpy{precheckErr: c.err}
		outcome := newStrategyFirstLegAdmissionBridge(issuer, loader).admit(context.Background(), market.result)
		var scope *strategyScopeRefusal
		if outcome.Code != c.code || (issuer.escalations == 1) != c.escalate || errors.As(outcome.cause, &scope) || isCentralStrategyIntegrity(outcome.cause) {
			t.Errorf("%s: outcome=%+v escalations=%d", name, outcome, issuer.escalations)
		}
	}
}

// 구현 리뷰 P1-3(Manager 결정 ①): 측정 불가(타입 표지)는 강화 없이 DAILY_LOSS_UNMEASURABLE — 레인 결함이 아님. 표지 없는 입력 불가는 오늘처럼.
func TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault(t *testing.T) {
	market := pairedStrategyFirstLegResults(t)[0]
	refusal := &execgw.IssueRefusal{Stage: execgw.StageChain, Reason: string(risk.ReasonInputUnavailable), Detail: "cannot be measured"}
	marked := errors.Join(refusal, &execgw.DailyLossUnmeasurable{Cause: errors.New("closed today without a frozen outcome row")})
	issuer := &firstLegGuardianSpy{precheckErr: marked}
	loader := &firstLegAuthorityLoaderStub{issuance: firstLegBridgeIssuanceFixture(market.result)}
	outcome := newStrategyFirstLegAdmissionBridge(issuer, loader).admit(context.Background(), market.result)
	var unmeasurable *execgw.DailyLossUnmeasurable
	if outcome.Code != StrategyFirstLegDailyLossUnmeasurable || issuer.escalations != 0 || !errors.As(outcome.cause, &unmeasurable) {
		t.Fatalf("outcome=%+v escalations=%d", outcome, issuer.escalations)
	}
	seen, err := a137Deliver(t, map[string]error{"a": fmt.Errorf("engine: first-leg admission %s: %w", outcome.Code, outcome.cause)})
	if err != nil || strings.Join(seen, ",") != "a" {
		t.Fatalf("delivered=%v err=%v — want the cycle ended quietly", seen, err)
	}

	plain := &execgw.IssueRefusal{Stage: execgw.StageChain, Reason: string(risk.ReasonInputUnavailable), Detail: "equity missing"}
	loader = &firstLegAuthorityLoaderStub{issuance: firstLegBridgeIssuanceFixture(market.result)}
	outcome = newStrategyFirstLegAdmissionBridge(&firstLegGuardianSpy{precheckErr: plain}, loader).admit(context.Background(), market.result)
	if outcome.Code != StrategyFirstLegAuthorityMismatch {
		t.Fatalf("an unmarked INPUT_UNAVAILABLE must stay AUTHORITY_MISMATCH, got %+v", outcome)
	}
	seen, err = a137Deliver(t, map[string]error{"a": fmt.Errorf("engine: first-leg admission %s: %w", outcome.Code, outcome.cause)})
	if err == nil || strings.Join(seen, ",") != "a" {
		t.Fatalf("delivered=%v err=%v — want today's stopping error", seen, err)
	}
}

// 구현 리뷰 P3: 승인은 Guardian 이 잰 토큰을 그대로 precheck 요청에 실음.
func TestA137AdmissionCarriesTheMeasuredTokenIntoThePrecheck(t *testing.T) {
	market := pairedStrategyFirstLegResults(t)[0]
	token := a137RealToken(t)
	issuer := &firstLegGuardianSpy{token: token, precheckErr: errors.New("stop after precheck")}
	loader := &firstLegAuthorityLoaderStub{issuance: firstLegBridgeIssuanceFixture(market.result)}
	newStrategyFirstLegAdmissionBridge(issuer, loader).admit(context.Background(), market.result)
	if len(issuer.requests) != 1 || !reflect.DeepEqual(issuer.requests[0].Entry.DailyLoss, token) {
		t.Fatalf("the precheck request did not carry the measured token")
	}
}

// a137RealToken 은 실제 Guardian 이 빈 원장에서 잰 토큰 — 엔진 패키지는 토큰을 만들 수 없음(불투명).
func a137RealToken(t *testing.T) execgw.DailyLossMeasurement {
	t.Helper()
	clk := clock.NewFake(time.Date(2026, 10, 14, 1, 30, 0, 0, time.UTC))
	j, err := journal.Open(context.Background(), journal.Options{Path: filepath.Join(t.TempDir(), journal.DBFileName), Clock: clk,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = j.Close() })
	g, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{Journal: j, Clock: clk, AccountRef: "acct", Policy: risk.DefaultPolicy(),
		Costs: costs.DefaultModel(), PolicyVersion: "a137/test", ManualBuys: a136ZeroManualBuyBoardInternal(clk)})
	if err != nil {
		t.Fatal(err)
	}
	return g.MeasureDailyLoss(context.Background())
}
