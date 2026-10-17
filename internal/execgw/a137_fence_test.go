//go:build tossos_testseams

package execgw_test

// a137 freeze 조건(BTM 220): 강화 전에 승인된 전략 진입이 제출 직전(SUBMITTING 뒤)에 AUTO 일일 손실 모드 변경을 만나면, 그 거절은
// 계좌 진입 차단 원인을 실은 채 돌아오고 브로커 요청은 0 — 다른 원인의 펜스는 오늘과 같음.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

func a137FenceRun(t *testing.T, market, suffix string, change func(*execgw.EntryGate)) (journal.AttemptState, error, int) {
	t.Helper()
	rig, fx, issued, intent := setupQFinalGatewayDecision(t, market, suffix)
	broker := &fakeBroker{}
	gate := execgw.NewEntryGate(rig.clock, map[execgw.RequiredQuery]time.Duration{})
	gateway, err := execgw.New(execgw.Options{Journal: rig.journal, Trading: trading.NewService(openPolicy(), broker),
		Clock: rig.clock, AccountRef: "acct-7", Source: "paired-strategy-test", Entry: gate})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := rig.journal.LookupDecision(context.Background(), issued.Decision.ID)
	if err != nil {
		t.Fatal(err)
	}
	cas := journal.StrategyDispatchLeaseCAS{LeaseID: "lease-" + suffix + "-" + strings.ToLower(market),
		ExpectedRevision: 2, OwnerEpoch: 7, FencingToken: "fence-" + strings.ToLower(market)}
	claimed := claimedGatewayLease(market, intent.Symbol, cas, decision, issued.Reservations[0].ID)
	gateway.SetStrategyDispatchLeaseForTest(func(context.Context, string) (journal.StrategyDispatchLease, error) {
		return claimed, nil
	}, func(context.Context, journal.StrategyDispatchLeaseCAS) (journal.StrategyDispatchLease, error) {
		change(gate)
		started := claimed
		started.State, started.Revision, started.TransportStartedAt = journal.StrategyDispatchLeaseSubmitting, cas.ExpectedRevision+1, fixedNow
		return started, nil
	})
	authority, err := gateway.ObserveStrategyEntryGate(context.Background(), strings.ToLower(market), intent.Symbol)
	if err != nil {
		t.Fatal(err)
	}
	request := execgw.StrategyPlaceRequest{Intent: intent, Decision: issued.Decision, Lease: cas, EntryGateAuthority: authority}
	request.SetAccountBaseFXForTest(fx)
	out, err := gateway.PlaceClaimedStrategy(context.Background(), request)
	places, _, _ := broker.totals()
	return out.State, err, places
}

func TestA137AFenceCausedByTheDailyLossEscalationIsAnAccountBlock(t *testing.T) {
	for _, market := range []string{"KR", "US"} {
		t.Run(market, func(t *testing.T) {
			state, err, places := a137FenceRun(t, market, "a137-daily-loss", func(gate *execgw.EntryGate) {
				gate.ProjectOperatingMode(journal.OperatingModeRecord{AccountRef: "acct-7", Mode: journal.ModeEntryBlocked,
					Actor: journal.ModeActorAuto, Cause: journal.ModeTriggerDailyLossLimit, Seq: 1})
			})
			var block *execgw.DailyLossEntryBlock
			if !errors.As(err, &block) || state != journal.StateNotDispatched || places != 0 {
				t.Fatalf("%s state=%s places=%d err=%v — want a not-dispatched account block", market, state, places, err)
			}
		})
	}
}

func TestA137OtherFencesBehaveAsBefore(t *testing.T) {
	for name, change := range map[string]func(*execgw.EntryGate){
		"reconcile": func(gate *execgw.EntryGate) {
			gate.Block(execgw.ReasonReconcileMismatch, "durable reconcile started after initial entry check")
		},
		"operator entry block": func(gate *execgw.EntryGate) {
			gate.ProjectOperatingMode(journal.OperatingModeRecord{AccountRef: "acct-7", Mode: journal.ModeEntryBlocked,
				Actor: journal.ModeActorOperator, Cause: journal.ModeTriggerDailyLossLimit, Seq: 1})
		},
	} {
		t.Run(name, func(t *testing.T) {
			state, err, places := a137FenceRun(t, "KR", "a137-other-"+strings.ReplaceAll(name, " ", "-"), change)
			var block *execgw.DailyLossEntryBlock
			var rejected *execgw.RejectedError
			if errors.As(err, &block) || !errors.As(err, &rejected) || state != journal.StateNotDispatched || places != 0 {
				t.Fatalf("state=%s places=%d err=%v — want today's ordinary not-dispatched refusal", state, places, err)
			}
		})
	}
}
