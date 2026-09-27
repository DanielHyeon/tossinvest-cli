//go:build tossos_testseams

package execgw_test

// a066 6.1 (B)(2) — Gateway.submit B41: 전략 plan 의 마지막 순간 q_final 장벽(`checkReservation` 을 broker 호출 직전,
// scheduler 최종 권위 재확인 뒤에 다시 부름)은 a066 커밋 a37d97f5 가 넣은 분기인데 어떤 시험도 실행하지 않았음
// (review.md 2.x 측정: submit 16/57 미실행 중 하나). 비전략 경로의 같은 장벽은
// TestGatewayLastMomentQFinalBarrierRefusesHoldReleaseAfterInitialAdmissionCheck 가 잼.
// 이 시험은 초기 검사를 통과한 뒤 최종 권위 재확인 시점에 진입 손실 잠금을 켜서, 전략 경로도 broker 전에 거절함을 고정함.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

func TestA066StrategyLastMomentQFinalBarrierRefusesALockTakenAfterTheInitialCheck(t *testing.T) {
	for _, market := range []string{"KR", "US"} {
		t.Run(market, func(t *testing.T) {
			rig, fx, issued, intent := setupQFinalGatewayDecision(t, market, "last-moment-lock")
			broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "order-" + strings.ToLower(market)}}
			gateway, err := execgw.New(execgw.Options{Journal: rig.journal, Trading: trading.NewService(openPolicy(), broker),
				Clock: rig.clock, AccountRef: "acct-7", Source: "a066-strategy-last-moment"})
			if err != nil {
				t.Fatal(err)
			}
			decision, err := rig.journal.LookupDecision(context.Background(), issued.Decision.ID)
			if err != nil {
				t.Fatal(err)
			}
			cas := journal.StrategyDispatchLeaseCAS{LeaseID: "lease-last-moment-" + strings.ToLower(market),
				ExpectedRevision: 2, OwnerEpoch: 7, FencingToken: "fence-" + strings.ToLower(market)}
			claimed := claimedGatewayLease(market, intent.Symbol, cas, decision, issued.Reservations[0].ID)
			gateway.SetStrategyDispatchLeaseForTest(func(context.Context, string) (journal.StrategyDispatchLease, error) {
				return claimed, nil
			}, func(context.Context, journal.StrategyDispatchLeaseCAS) (journal.StrategyDispatchLease, error) {
				started := claimed
				started.State, started.Revision, started.TransportStartedAt = journal.StrategyDispatchLeaseSubmitting, cas.ExpectedRevision+1, fixedNow
				return started, nil
			})
			finalChecks := 0
			request := execgw.StrategyPlaceRequest{Intent: intent, Decision: issued.Decision, Lease: cas,
				// 최종 권위 재확인은 통과시키되, 그 순간 이 시장의 두 horizon 을 잠금 — 초기 checkReservation 은 이미 지났음.
				FinalAuthorityCheck: func(ctx context.Context) error {
					finalChecks++
					for _, horizon := range []riskbucket.Horizon{riskbucket.HorizonShort, riskbucket.HorizonMedium} {
						if _, _, err := rig.journal.ActivateEntryLossLock(ctx, journal.EntryLossLock{AccountRef: "acct-7",
							Market: riskbucket.Market(market), Horizon: horizon, Cause: "a066 last-moment probe", ActivatedAt: fixedNow}); err != nil {
							return err
						}
					}
					return nil
				}}
			request.SetAccountBaseFXForTest(fx)
			out, err := gateway.PlaceClaimedStrategy(context.Background(), request)
			var rejected *execgw.RejectedError
			if finalChecks != 1 {
				t.Fatalf("%s: final authority check ran %d times — the lock was not taken after the initial check", market, finalChecks)
			}
			// 전략 경로의 call() 오류는 전부 ReasonStrategyDispatchFenced 로 눕혀짐(gateway.go "strategy final authority changed
			// before transport", 소유 8022f578) — 잠금 사유는 Detail 에만 남음. 그래서 사유 코드는 눕혀진 값으로, 원인은 Detail 의
			// 거절 문장으로 단언함(타입 보존은 review.md 6.1 잔여).
			if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonStrategyDispatchFenced ||
				!strings.Contains(rejected.Detail, "rejected (entry_loss_lock_active)") ||
				out.State != journal.StateNotDispatched {
				t.Fatalf("%s: outcome=%+v err=%v, want NOT_DISPATCHED fenced by the last-moment entry_loss_lock_active refusal", market, out, err)
			}
			if places, _, _ := broker.totals(); places != 0 {
				t.Fatalf("%s: broker places=%d after a last-moment lock, want 0", market, places)
			}
		})
	}
}
