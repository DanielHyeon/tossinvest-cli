package execgw_test

// riskguardian_race_test.go is task 4.3: what happens when two things race the
// issuance.
//
// Two races, and they are settled in two different places on purpose:
//
//   - Two issuances against one aggregate limit. The chain cannot settle this —
//     both callers hold a snapshot in which the other has not happened — so the
//     authority is the reservation transaction (risk-management: 총계 한도의
//     최종 권위는 예약 트랜잭션).
//   - A mode tightening between issuance and submission. The decision was
//     legitimately issued and is still valid; what changed is the account's
//     mode, and the enforcement point for that is the EntryGate re-check the
//     Gateway performs at submission (design D1/D3).

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// raceSymbols are eight distinct symbols, so nothing in this test is serialised
// by the gateway's one-mutation-per-symbol rule or by a duplicate check. The
// only thing they share is the account's exposure ceiling.
var raceSymbols = []string{"005930", "000660", "035420", "051910", "005380", "068270", "207940", "006400"}

// TestConcurrentIssuancesCannotExceedTheAggregateLimit.
//
// The ceiling admits four holds of 700,700 (2,802,800) and refuses the fifth
// (3,503,500 ≥ 3,000,000 — 도달 시 차단). Eight issuers go at it at once, each
// with a snapshot that says nothing is open, which is exactly the state that
// makes the chain unable to decide: every one of them is individually correct.
//
// What is asserted is the invariant rather than a winner: whatever the
// interleaving, the held total stays under the ceiling and every loser gets a
// stable reason-code rather than a decision.
func TestConcurrentIssuancesCannotExceedTheAggregateLimit(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	ctx := context.Background()

	policy := guardianPolicy()
	policy.MaxOpenExposure = riskcalc.Money{Amount: "3000000", Currency: "KRW"}

	guardian := mustGuardian(t, execgw.RiskGuardianOptions{
		Journal: j, Clock: clk, AccountRef: "acct-7", Policy: policy,
		Costs: costs.DefaultModel(), PolicyVersion: "add-core-domain/4.3",
		// Generous attempts: a version conflict is a race this loop is meant to
		// ride out, and capping it at the default 3 would turn losing a race into
		// the same answer as reaching the limit.
		Recollect: journal.RecollectPolicy{MaxAttempts: 40, Budget: time.Hour},
	})

	var wg sync.WaitGroup
	results := make([]error, len(raceSymbols))
	for i, symbol := range raceSymbols {
		wg.Add(1)
		go func(i int, symbol string) {
			defer wg.Done()
			intent := guardianIntent()
			intent.Symbol = symbol
			account := guardianAccount()
			account.AllowedSymbols = []string{symbol}

			_, err := guardian.IssueEntry(ctx, execgw.EntryIssuance{
				Intent:  intent,
				Account: account,
				Collect: func(ctx context.Context, _ int) (execgw.ExposureSnapshot, error) {
					v, err := j.ReservationVersion(ctx, "acct-7")
					if err != nil {
						return execgw.ExposureSnapshot{}, err
					}
					return execgw.ExposureSnapshot{
						AsOf:         clk.Now(),
						Version:      v,
						OpenExposure: riskcalc.Money{Amount: "0", Currency: "KRW"},
					}, nil
				},
			})
			results[i] = err
		}(i, symbol)
	}
	wg.Wait()

	issued := 0
	for i, err := range results {
		if err == nil {
			issued++
			continue
		}
		var refusal *execgw.IssueRefusal
		if !errors.As(err, &refusal) {
			t.Errorf("%s: err = %v, want an *IssueRefusal", raceSymbols[i], err)
			continue
		}
		switch refusal.Reason {
		case journal.IssueReasonLimitReached, journal.IssueReasonRecollectionExhausted:
		default:
			t.Errorf("%s: reason = %q, want the limit or the exhausted loop", raceSymbols[i], refusal.Reason)
		}
	}
	if issued == 0 || issued > 4 {
		t.Errorf("issued = %d, want between 1 and 4 — the ceiling admits four holds of %s",
			issued, guardianExposure)
	}

	held, err := j.HeldReservations(ctx, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if len(held) != issued {
		t.Errorf("held reservations = %d, issued = %d; every issuance holds exactly one", len(held), issued)
	}
	total := "0"
	for _, r := range held {
		sum, err := riskcalc.AddDecimal(total, r.Amount)
		if err != nil {
			t.Fatal(err)
		}
		total = sum
	}
	// The invariant, stated the way the reservation ledger states it: reaching
	// the ceiling blocks, so the held total must be strictly under it.
	cmp, err := riskcalc.CompareDecimal(total, "3000000")
	if err != nil {
		t.Fatal(err)
	}
	if cmp >= 0 {
		t.Errorf("held total = %s KRW against a ceiling of 3000000; reaching it should have blocked", total)
	}

	// Every winner has a decision on disk and every loser has none: the count of
	// decisions is the count of successes, with no orphans from the losers.
	decisions := 0
	for _, r := range held {
		if _, err := j.LookupDecision(ctx, r.DecisionID); err != nil {
			t.Errorf("reservation %s names decision %s, which is not readable: %v", r.ID, r.DecisionID, err)
			continue
		}
		decisions++
	}
	if decisions != issued {
		t.Errorf("decisions = %d, issued = %d", decisions, issued)
	}
}

// TestATighteningBetweenIssuanceAndSubmissionIsRefusedAtTheGateway is
// risk-management's own scenario (발급과 제출 사이 모드 강화): the decision is
// valid, unexpired, unspent and holds its reservation — and it is still refused,
// because the mode moved and the enforcement point for a mode is the gate the
// sealed submission sequence already consults.
//
// This is also why the re-collection loop does not re-run the chain: the input
// that can change between issuance and submission is covered here, not there.
func TestATighteningBetweenIssuanceAndSubmissionIsRefusedAtTheGateway(t *testing.T) {
	rig := newModeRig(t, filepath.Join(t.TempDir(), "journal.db"),
		domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-1"})
	ctx := context.Background()

	guardian := mustGuardian(t, execgw.RiskGuardianOptions{
		Journal: rig.j, Clock: rig.clk, AccountRef: "acct-7", Policy: guardianPolicy(),
		Costs: costs.DefaultModel(), PolicyVersion: "add-core-domain/4.3",
	})

	issue := func() execgw.Issued {
		t.Helper()
		out, err := guardian.IssueEntry(ctx, execgw.EntryIssuance{
			Intent:  guardianIntent(),
			Account: guardianAccount(),
			Collect: func(ctx context.Context, _ int) (execgw.ExposureSnapshot, error) {
				v, err := rig.j.ReservationVersion(ctx, "acct-7")
				if err != nil {
					return execgw.ExposureSnapshot{}, err
				}
				return execgw.ExposureSnapshot{
					AsOf: rig.clk.Now(), Version: v,
					OpenExposure: riskcalc.Money{Amount: "0", Currency: "KRW"},
				}, nil
			},
		})
		if err != nil {
			t.Fatalf("IssueEntry: %v", err)
		}
		return out
	}

	// Issued under NORMAL, and nothing about the decision changes after this.
	authorised := issue()

	if _, changed, err := rig.j.EscalateOperatingMode(ctx, "acct-7",
		journal.ModeTriggerDailyLossLimit, nil); err != nil || !changed {
		t.Fatalf("escalating: changed=%v err=%v", changed, err)
	}

	out, err := rig.gw.Place(ctx, execgw.PlaceRequest{
		Intent:   guardianPlaceIntent(),
		Decision: authorised.Decision,
	})
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) {
		t.Fatalf("err = %v, want a gateway refusal", err)
	}
	if rejected.Reason != execgw.ReasonOperatingModeBlocked {
		t.Errorf("reason = %q, want %q", rejected.Reason, execgw.ReasonOperatingModeBlocked)
	}
	if out.State != journal.StateNotDispatched {
		t.Errorf("state = %s, want NOT_DISPATCHED", out.State)
	}
	if places, _, _ := rig.broker.totals(); places != 0 {
		t.Errorf("broker place calls = %d; the tightening must be enforced before the call", places)
	}

	// The decision itself is untouched: the refusal is the gate's, not a claim
	// that the authorisation was bad.
	dec, err := rig.j.LookupDecision(ctx, authorised.Decision.ID)
	if err != nil {
		t.Fatalf("the decision must still be readable: %v", err)
	}
	if !dec.ExpiresAt.After(rig.clk.Now()) {
		t.Error("the decision expired; this test is meant to refuse a valid one")
	}
	if !strings.Contains(rejected.Detail, "ENTRY_BLOCKED") {
		t.Errorf("detail %q does not name the mode that refused", rejected.Detail)
	}
}

// guardianPlaceIntent is the order the guardianIntent authorises, in the
// gateway's vocabulary. The two have to agree exactly: an entry is authorised
// for one size at one price.
func guardianPlaceIntent() orderintent.PlaceIntent {
	return orderintent.PlaceIntent{
		Symbol: "005930", Market: "kr", Side: "buy", OrderType: "limit",
		Quantity: 10, Price: 70000, CurrencyMode: "KRW",
	}
}
