package journal

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestRecoverClaimedStrategyDispatchLeasePairedKRUSRefusesAndReleases(t *testing.T) {
	for _, market := range []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			ctx := context.Background()
			oldOwner, err := j.AcquireStrategyDispatchOwner(ctx, "recovery-old-"+strings.ToLower(market.name))
			if err != nil {
				t.Fatal(err)
			}
			plan := seedBoundStrategyDispatchClaimLease(t, j, oldOwner, "recovery-"+strings.ToLower(market.name),
				"acct-"+strings.ToLower(market.name), market.name, market.symbol)
			claimed, err := j.ClaimStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: 1, OwnerEpoch: oldOwner.Epoch, FencingToken: oldOwner.FencingToken,
			})
			if err != nil || claimed.State != StrategyDispatchLeaseClaimed {
				t.Fatalf("claim=%+v err=%v", claimed, err)
			}
			newOwner, err := j.AcquireStrategyDispatchOwner(ctx, "recovery-new-"+strings.ToLower(market.name))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := j.RecoverClaimedStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: oldOwner.Epoch, FencingToken: oldOwner.FencingToken,
			}); !errors.Is(err, ErrStrategyDispatchFenced) {
				t.Fatalf("stale recovery owner error=%v", err)
			}
			terminal, err := j.RecoverClaimedStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: newOwner.Epoch, FencingToken: newOwner.FencingToken,
			})
			if err != nil {
				t.Fatal(err)
			}
			if terminal.State != StrategyDispatchLeaseRefused || terminal.Disposition != StrategyDispatchReservationReleased ||
				terminal.Revision != claimed.Revision+1 || terminal.RefusalCode != "RECOVERY_CLAIMED_NO_TRANSPORT" ||
				terminal.OwnerEpoch != oldOwner.Epoch || terminal.FencingToken != oldOwner.FencingToken {
				t.Fatalf("terminal recovery lease=%+v", terminal)
			}
			assertStrategyDispatchRealHolds(t, j, terminal, "RELEASED", 0)
			assertNoStrategyDispatchTransport(t, j)
			if _, err := j.RecoverClaimedStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: newOwner.Epoch, FencingToken: newOwner.FencingToken,
			}); !errors.Is(err, ErrStrategyDispatchLeaseConsumed) {
				t.Fatalf("recovery replay error=%v", err)
			}
		})
	}
}

func TestRecoverClaimedStrategyDispatchLeaseClosesPreparedCoreInSameCommitKRUS(t *testing.T) {
	for _, market := range []struct{ name, symbol, currency string }{
		{"KR", "005930", "KRW"}, {"US", "AAPL", "USD"},
	} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			ctx := context.Background()
			_, plan, claimed := seedClaimedStrategyDispatchLease(t, j,
				"recovery-prepared-"+strings.ToLower(market.name), market.name, market.symbol)
			attempt, _ := prepareCoreStrategyAttempt(t, j, plan, market.name, market.symbol, market.currency)
			newOwner, err := j.AcquireStrategyDispatchOwner(ctx, "recovery-prepared-new-"+strings.ToLower(market.name))
			if err != nil {
				t.Fatal(err)
			}
			terminal, err := j.RecoverClaimedStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: newOwner.Epoch, FencingToken: newOwner.FencingToken,
			})
			if err != nil {
				t.Fatal(err)
			}
			record, err := j.LookupAttempt(ctx, attempt.ID())
			if err != nil || record.State != StateNotDispatched ||
				record.ReasonCode != strategyDispatchRecoveryClaimedNoTransport || record.BrokerOrderID != "" {
				t.Fatalf("prepared recovery core=%+v err=%v", record, err)
			}
			if terminal.State != StrategyDispatchLeaseRefused || terminal.Disposition != StrategyDispatchReservationReleased ||
				terminal.RefusalCode != strategyDispatchRecoveryClaimedNoTransport {
				t.Fatalf("prepared recovery terminal=%+v", terminal)
			}
			assertStrategyDispatchRealHolds(t, j, terminal, "RELEASED", 0)
			assertNoStrategyDispatchTransport(t, j)
		})
	}
}

func TestRecoverClaimedStrategyDispatchLeaseRejectsSubmittingWithoutReleaseKRUS(t *testing.T) {
	for _, market := range []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			ctx := context.Background()
			oldOwner, err := j.AcquireStrategyDispatchOwner(ctx, "submitting-old-"+strings.ToLower(market.name))
			if err != nil {
				t.Fatal(err)
			}
			plan := seedBoundStrategyDispatchClaimLease(t, j, oldOwner, "submitting-recovery-"+strings.ToLower(market.name),
				"acct-"+strings.ToLower(market.name), market.name, market.symbol)
			claimed, err := j.ClaimStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: 1, OwnerEpoch: oldOwner.Epoch, FencingToken: oldOwner.FencingToken,
			})
			if err != nil {
				t.Fatal(err)
			}
			submitting, err := j.BeginStrategyDispatchSubmitting(ctx, StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: oldOwner.Epoch, FencingToken: oldOwner.FencingToken,
			})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := j.AcquireStrategyDispatchOwner(ctx, "submitting-racing-"+strings.ToLower(market.name)); !errors.Is(err, ErrStrategyDispatchOwnerBusy) {
				t.Fatalf("active SUBMITTING takeover error=%v", err)
			}
			j.clk.(*clock.Fake).Set(submitting.ExpiresAt.Add(strategyDispatchOwnerTakeoverGrace))
			newOwner, err := j.AcquireStrategyDispatchOwner(ctx, "submitting-new-"+strings.ToLower(market.name))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := j.RecoverClaimedStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: submitting.Revision,
				OwnerEpoch: newOwner.Epoch, FencingToken: newOwner.FencingToken,
			}); !errors.Is(err, ErrStrategyDispatchLeaseConsumed) {
				t.Fatalf("SUBMITTING recovery error=%v", err)
			}
			after, err := j.LookupStrategyDispatchLease(ctx, plan.LeaseID)
			if err != nil || after.State != StrategyDispatchLeaseSubmitting || after.Disposition != StrategyDispatchReservationReserved {
				t.Fatalf("SUBMITTING recovery mutated lease=%+v err=%v", after, err)
			}
			assertStrategyDispatchRealHolds(t, j, after, "HELD", 5)
		})
	}
}

func TestRecoverClaimedStrategyDispatchLeaseRollbackIsAtomicKRUS(t *testing.T) {
	for _, market := range []struct{ name, symbol, currency string }{
		{"KR", "005930", "KRW"}, {"US", "AAPL", "USD"},
	} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			ctx := context.Background()
			_, plan, claimed := seedClaimedStrategyDispatchLease(t, j,
				"recovery-rollback-"+strings.ToLower(market.name), market.name, market.symbol)
			attempt, _ := prepareCoreStrategyAttempt(t, j, plan, market.name, market.symbol, market.currency)
			newOwner, err := j.AcquireStrategyDispatchOwner(ctx, "rollback-new-"+strings.ToLower(market.name))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.Exec(`CREATE TRIGGER fail_claimed_recovery_outcome BEFORE INSERT ON strategy_dispatch_outcomes
				WHEN NEW.to_state='REFUSED' AND NEW.transition_code='RECOVERY_CLAIMED_NO_TRANSPORT'
				BEGIN SELECT RAISE(ABORT,'synthetic claimed recovery failure'); END`); err != nil {
				t.Fatal(err)
			}
			_, err = j.RecoverClaimedStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: newOwner.Epoch, FencingToken: newOwner.FencingToken,
			})
			if err == nil || !strings.Contains(err.Error(), "synthetic claimed recovery failure") {
				t.Fatalf("recovery rollback error=%v", err)
			}
			after, err := j.LookupStrategyDispatchLease(ctx, plan.LeaseID)
			if err != nil || after.State != StrategyDispatchLeaseClaimed || after.Disposition != StrategyDispatchReservationReserved ||
				after.Revision != claimed.Revision {
				t.Fatalf("recovery rollback lease=%+v err=%v", after, err)
			}
			record, err := j.LookupAttempt(ctx, attempt.ID())
			if err != nil || record.State != StateRecorded {
				t.Fatalf("recovery rollback core=%+v err=%v", record, err)
			}
			assertStrategyDispatchRealHolds(t, j, after, "HELD", 5)
			assertNoStrategyDispatchTransport(t, j)
		})
	}
}
