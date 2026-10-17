package execgw_test

// a066 6.5 발견 1(Manager 판정 2026-09-28): 이미 발급된 q_final 결정이라도 제출 시점에 그 owner 범위가 대사 중이거나
// latch 되어 있으면 broker 전에 거절됨. 사유 코드는 기존 guardian_risk_bucket_mismatch 를 재사용하고, 원인(latch·대사)은
// 같은 규칙 함수가 만든 detail 이 이름으로 말함.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

func TestA066IssuedEntryIsRefusedAtSubmitWhenItsScopeLatchedAfterIssuance(t *testing.T) {
	for _, reconcile := range []bool{false, true} {
		name := "control: no latch submits"
		if reconcile {
			name = "active reconcile after issuance refuses the submit"
		}
		t.Run(name, func(t *testing.T) {
			rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) {
				options.NewID = fixedIDs("latch-submit-decision", "latch-submit-nonce")
			})
			ctx := context.Background()
			issued, err := rig.guardian.IssueQFinalEntry(ctx, lossLockQFinalRequest(t, rig, "latch-submit", riskbucket.HorizonShort))
			if err != nil {
				t.Fatalf("issuance: %v", err)
			}
			if reconcile {
				if _, entered, err := rig.journal.EnterReconcile(ctx, journal.EnterReconcileRequest{AccountRef: "acct-7", Symbol: "005930",
					Cause: journal.ReconcileCauseQuantityMismatch, Evidence: "a066 6.5 submit latch probe"}); err != nil || !entered {
					t.Fatalf("enter reconcile: entered=%v err=%v", entered, err)
				}
			}
			broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-latch-submit"}}
			opts := execgw.Options{Journal: rig.journal, Trading: trading.NewService(openPolicy(), broker), Clock: rig.clock,
				AccountRef: "acct-7", Source: "a066-submit-latch-test"}
			opts.SetMarketProtectionForTest(func(market string, _ int) (bool, string) { return true, market + ":stable-protection" })
			gw, err := execgw.New(opts)
			if err != nil {
				t.Fatal(err)
			}
			intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{Symbol: "005930", Market: "kr", Side: "buy", OrderType: "limit", Quantity: 10, Price: 70000, CurrencyMode: "KRW"})
			if err != nil {
				t.Fatal(err)
			}
			out, err := gw.Place(ctx, execgw.PlaceRequest{Intent: intent, Decision: issued.Decision})
			places, _, _ := broker.totals()
			if !reconcile {
				if err != nil || places != 1 {
					t.Fatalf("control: state=%s places=%d err=%v", out.State, places, err)
				}
				return
			}
			var rejected *execgw.RejectedError
			if !errors.As(err, &rejected) || places != 0 || out.State != journal.StateNotDispatched ||
				rejected.Reason != execgw.ReasonGuardianRiskBucketMismatch || !strings.Contains(rejected.Detail, "RECONCILE") {
				t.Fatalf("latched scope submitted or refusal unnamed: rejected=%+v state=%s places=%d err=%v", rejected, out.State, places, err)
			}
		})
	}
}
