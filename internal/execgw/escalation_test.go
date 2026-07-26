package execgw_test

// escalation_test.go is the producer half of the automatic-tightening triggers
// (task 3.2's succession: the enumeration and the transition API landed there,
// and nothing called them).
//
// Two of the four producers live in this package:
//
//	BROKER_AUTH_REJECTED       execgw.Retrier, on a 401/403
//	DAILY_LOSS_LIMIT_REACHED   execgw.RiskGuardian, when the chain's daily-loss
//	                           rung refuses
//
// The third (CRITICAL_ALERT_UNDELIVERED) is obs.Notifier's and is tested there;
// the fourth (EXIT_OBSERVATION_OUTAGE) belongs to the exit observation loop,
// task 7.4.
//
// What each test asserts is the same pair, because the pair is the point: the
// in-memory latch stops *this* process, and the persisted transition is what a
// restart still knows. Wiring one without the other leaves a block an operator
// lifts by doing the most natural thing in the world after a credential error.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// authFatal is the broker's rejected-credential error (official.ErrAuth is what
// the client returns for a 401 or a 403).
func authFatal() error { return official.ErrAuth }

// TestARejectedCredentialTightensTheOperatingMode.
func TestARejectedCredentialTightensTheOperatingMode(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	if err := j.SetModeProjector(gate); err != nil {
		t.Fatalf("SetModeProjector: %v", err)
	}
	ctx := context.Background()

	r := &execgw.Retrier{
		Clock:      clk,
		Gate:       gate,
		Escalate:   j,
		AccountRef: "acct-7",
	}
	err := r.Query(ctx, execgw.QueryHoldings, func(context.Context) error { return authFatal() })
	if err == nil {
		t.Fatal("a rejected credential must be returned to the caller")
	}
	// The caller still classifies the query's own error: joining the escalation
	// result onto it must not change what it is.
	if execgw.ClassifyQueryError(err) != execgw.ClassAuthFatal {
		t.Errorf("class = %s, want %s", execgw.ClassifyQueryError(err), execgw.ClassAuthFatal)
	}

	// Half one: this process stops opening positions now.
	if rejected := gate.CheckEntry(); rejected == nil {
		t.Fatal("the entry gate must be latched")
	}
	// Half two: the block is on disk, so it survives the restart an operator
	// performs when they see a credential error.
	snapshot, err := j.CurrentOperatingMode(ctx, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Mode != journal.ModeEntryBlocked {
		t.Errorf("mode = %q, want ENTRY_BLOCKED", snapshot.Mode)
	}
	if snapshot.Cause != journal.ModeTriggerCredentialRejected || snapshot.Actor != journal.ModeActorAuto {
		t.Errorf("row = %+v, want the enumerated trigger recorded as automatic", snapshot)
	}
}

// TestTheRetrierWithoutAnEscalatorStillLatches: the durable half is optional
// wiring and the in-memory half is not. A profile that has not wired the journal
// into the retrier must still stop trading.
func TestTheRetrierWithoutAnEscalatorStillLatches(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})

	r := &execgw.Retrier{Clock: clk, Gate: gate}
	if err := r.Query(context.Background(), execgw.QueryHoldings,
		func(context.Context) error { return authFatal() }); err == nil {
		t.Fatal("the error must still be returned")
	}
	if rejected := gate.CheckEntry(); rejected == nil {
		t.Error("the entry gate must be latched with or without an escalator")
	}
}

// TestOnlyTheCredentialFailureEscalates: the other error classes are not
// triggers, and the enumeration is what says so. A rate limit and a transient
// failure are conditions that recover on their own; escalating on them would
// turn a slow morning into a blocked account.
func TestOnlyTheCredentialFailureEscalates(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"rate limited", official.ErrRateLimited},
		{"transient", official.ErrServer},
		{"permanent", &official.APIError{Code: 400}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clk := clock.NewFake(fixedNow)
			j := openJournal(t, clk)
			ctx := context.Background()

			r := &execgw.Retrier{
				Clock:      clk,
				Policy:     execgw.RetryPolicy{MaxAttempts: 1, Budget: time.Second},
				Escalate:   j,
				AccountRef: "acct-7",
			}
			_ = r.Query(ctx, execgw.QueryHoldings, func(context.Context) error { return tc.err })

			snapshot, err := j.CurrentOperatingMode(ctx, "acct-7")
			if err != nil {
				t.Fatal(err)
			}
			if snapshot.Recorded {
				t.Errorf("a %s failure wrote an operating-mode row (%+v); it is not an enumerated trigger",
					tc.name, snapshot)
			}
		})
	}
}

// TestAFailedEscalationIsNotSwallowed. A tightening that did not persist is a
// block a restart lifts, so it has to reach somebody. It cannot replace the
// query's error — the caller classifies that one — so it is joined onto it.
func TestAFailedEscalationIsNotSwallowed(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})

	r := &execgw.Retrier{
		Clock:      clk,
		Gate:       gate,
		Escalate:   refusingEscalator{},
		AccountRef: "acct-7",
	}
	err := r.Query(context.Background(), execgw.QueryHoldings,
		func(context.Context) error { return authFatal() })
	if err == nil {
		t.Fatal("the query error must still be returned")
	}
	if execgw.ClassifyQueryError(err) != execgw.ClassAuthFatal {
		t.Errorf("the joined error must still classify as the original: %s",
			execgw.ClassifyQueryError(err))
	}
	if !strings.Contains(err.Error(), "restart would lift the block") {
		t.Errorf("err = %v, want the escalation failure visible in it", err)
	}
	if rejected := gate.CheckEntry(); rejected == nil {
		t.Error("the latch is unconditional; only the durable half failed")
	}
}

type refusingEscalator struct{}

func (refusingEscalator) EscalateOperatingMode(context.Context, string, string,
	journal.ModeAnnouncer) (journal.OperatingModeRecord, bool, error) {
	return journal.OperatingModeRecord{}, false, errors.New("the journal is unwritable")
}

// --- the daily-loss producer ------------------------------------------------

// TestReachingTheDailyLossLimitTightensTheOperatingMode.
//
// The chain's daily-loss rung is where "한도 도달" is judged, so the issuer is
// where the transition belongs (issues.md, task 3.2's producer table). The
// realised loss itself is an input the caller supplies from riskcalc.DailyLoss;
// this seam does not compute a P&L and must not.
func TestReachingTheDailyLossLimitTightensTheOperatingMode(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	ctx := context.Background()

	guardian := mustGuardian(t, execgw.RiskGuardianOptions{
		Journal: j, Clock: clk, AccountRef: "acct-7", Policy: guardianPolicy(),
		Costs: costs.DefaultModel(), PolicyVersion: "add-core-domain/3.2",
	})

	account := guardianAccount()
	// Exactly at the ceiling: 도달 시 차단, and the boundary is the only place
	// this matters.
	account.DailyRealizedLoss = riskcalc.Money{Amount: "100000", Currency: "KRW"}

	_, err := guardian.IssueEntry(ctx, execgw.EntryIssuance{
		Intent: guardianIntent(), Account: account,
		Collect: func(ctx context.Context, _ int) (execgw.ExposureSnapshot, error) {
			return execgw.ExposureSnapshot{}, errors.New("the collector must not be reached")
		},
	})
	var refusal *execgw.IssueRefusal
	if !errors.As(err, &refusal) || refusal.Reason != "DAILY_LOSS_LIMIT_REACHED" {
		t.Fatalf("err = %v, want the chain's DAILY_LOSS_LIMIT_REACHED", err)
	}

	snapshot, err := j.CurrentOperatingMode(ctx, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Mode != journal.ModeEntryBlocked {
		t.Fatalf("mode = %q, want ENTRY_BLOCKED — the limit is reached and it must outlive the process",
			snapshot.Mode)
	}
	if snapshot.Cause != journal.ModeTriggerDailyLossLimit || snapshot.Actor != journal.ModeActorAuto {
		t.Errorf("row = %+v, want the enumerated trigger recorded as automatic", snapshot)
	}
}

// TestAnAccountWithNoCapitalTightensTheSameWay: the equity-at-or-below-zero
// branch reports the same reason (계좌자본 0 이하이면 즉시 차단) and means the
// same thing — there is no loss budget left to measure against.
func TestAnAccountWithNoCapitalTightensTheSameWay(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	ctx := context.Background()

	guardian := mustGuardian(t, execgw.RiskGuardianOptions{
		Journal: j, Clock: clk, AccountRef: "acct-7", Policy: guardianPolicy(),
		Costs: costs.DefaultModel(), PolicyVersion: "add-core-domain/3.2",
	})

	account := guardianAccount()
	account.AccountEquity = riskcalc.Money{Amount: "0", Currency: "KRW"}

	if _, err := guardian.IssueEntry(ctx, execgw.EntryIssuance{
		Intent: guardianIntent(), Account: account,
		Collect: func(ctx context.Context, _ int) (execgw.ExposureSnapshot, error) {
			return execgw.ExposureSnapshot{}, errors.New("the collector must not be reached")
		},
	}); err == nil {
		t.Fatal("an account with no capital must be refused")
	}
	snapshot, err := j.CurrentOperatingMode(ctx, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Mode != journal.ModeEntryBlocked {
		t.Errorf("mode = %q, want ENTRY_BLOCKED", snapshot.Mode)
	}
}

// TestOtherChainRefusalsAreNotTriggers is the closed-enumeration property from
// the producer's side: an intent refused for its own shape says nothing about
// the account, and blocking the account for it would be a mode transition caused
// by a bad signal.
func TestOtherChainRefusalsAreNotTriggers(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	ctx := context.Background()

	guardian := mustGuardian(t, execgw.RiskGuardianOptions{
		Journal: j, Clock: clk, AccountRef: "acct-7", Policy: guardianPolicy(),
		Costs: costs.DefaultModel(), PolicyVersion: "add-core-domain/3.2",
	})

	for _, tc := range []struct {
		name  string
		build func() (execgw.EntryIssuance, string)
	}{
		{"a symbol nobody allowed", func() (execgw.EntryIssuance, string) {
			account := guardianAccount()
			account.AllowedSymbols = []string{"000660"}
			return execgw.EntryIssuance{Intent: guardianIntent(), Account: account}, "SYMBOL_NOT_ALLOWED"
		}},
		{"a reward:risk under the floor", func() (execgw.EntryIssuance, string) {
			intent := guardianIntent()
			intent.TargetPrice = "70500"
			return execgw.EntryIssuance{Intent: intent, Account: guardianAccount()}, "MIN_RR_NOT_MET"
		}},
		{"not enough cash", func() (execgw.EntryIssuance, string) {
			account := guardianAccount()
			account.CashAvailable = riskcalc.Money{Amount: "1000", Currency: "KRW"}
			return execgw.EntryIssuance{Intent: guardianIntent(), Account: account}, "INSUFFICIENT_CASH"
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			issuance, wantReason := tc.build()
			issuance.Collect = func(ctx context.Context, _ int) (execgw.ExposureSnapshot, error) {
				return execgw.ExposureSnapshot{}, errors.New("the collector must not be reached")
			}
			_, err := guardian.IssueEntry(ctx, issuance)
			var refusal *execgw.IssueRefusal
			if !errors.As(err, &refusal) || refusal.Reason != wantReason {
				t.Fatalf("err = %v, want %s", err, wantReason)
			}
			snapshot, serr := j.CurrentOperatingMode(ctx, "acct-7")
			if serr != nil {
				t.Fatal(serr)
			}
			if snapshot.Recorded {
				t.Errorf("%s wrote an operating-mode row (%+v); only the enumerated triggers may",
					wantReason, snapshot)
			}
		})
	}
}
