package execgw_test

// riskguardian_test.go covers the Guardian issuer (task 4.1/4.2, design D1).
//
// The assertion the whole file is built around is negative and is the one D1
// exists for: after a refused issuance there is no decision on disk. Everything
// else is there to make that unfakeable — the ledger version has not moved, the
// reservation is absent, and the refused decision's nonce can be used again,
// which is only true if the row was rolled back rather than merely hidden
// (the proof pattern task 0.2 established, applied one layer up).

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// --- the small_live default set ---------------------------------------------

// guardianPolicy is risk-management's conservative default set — the one an
// operator who has chosen no numbers runs on (주문당 notional 1,000,000 KRW·주문당
// 수량 100주·총 노출 10,000,000 KRW·일일 손실 100,000 KRW·자본비 1%·통화 KRW).
func guardianPolicy() risk.Policy { return risk.DefaultPolicy() }

// guardianIntent is an entry that passes every rung of the chain under that
// policy: RR is exactly 2.0 (the boundary, which passes), the size is inside the
// risk budget, and the target clears the real break-even.
func guardianIntent() risk.Intent {
	return risk.Intent{
		Market:      costs.MarketKR,
		Symbol:      "005930",
		Side:        risk.SideBuy,
		Quantity:    "10",
		LimitPrice:  "70000",
		StopPrice:   "69000",
		TargetPrice: "72000",
	}
}

// guardianAccount is an account with room for that entry and nothing latched.
func guardianAccount() risk.AccountState {
	return risk.AccountState{
		Mode:              risk.ModeNormal,
		AllowedSymbols:    []string{"005930"},
		CashAvailable:     riskcalc.Money{Amount: "5000000", Currency: "KRW"},
		OpenExposure:      riskcalc.Money{Amount: "0", Currency: "KRW"},
		DailyRealizedLoss: riskcalc.Money{Amount: "0", Currency: "KRW"},
		AccountEquity:     riskcalc.Money{Amount: "10000000", Currency: "KRW"},
	}
}

// guardianExposure is what one guardianIntent adds: 70000 × 10 plus the
// over-estimated buy cost (0.1% = 700).
const guardianExposure = "700700"

// fixedIDs mints the ids a test needs to name, then falls back to unique ones.
func fixedIDs(ids ...string) func() string {
	var mu sync.Mutex
	var i int
	return func() string {
		mu.Lock()
		defer mu.Unlock()
		i++
		if i <= len(ids) {
			return ids[i-1]
		}
		return fmt.Sprintf("gen-%d", i)
	}
}

type guardianRig struct {
	guardian *execgw.RiskGuardian
	journal  *journal.Journal
	clock    *clock.Fake
	// collections counts the reservation snapshots the loop asked for.
	collections int
}

func newGuardian(t *testing.T, mutate func(*execgw.RiskGuardianOptions)) *guardianRig {
	t.Helper()
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	opts := execgw.RiskGuardianOptions{
		Journal:       j,
		Clock:         clk,
		AccountRef:    "acct-7",
		Policy:        guardianPolicy(),
		Costs:         costs.DefaultModel(),
		PolicyVersion: "add-core-domain/4.1",
	}
	if mutate != nil {
		mutate(&opts)
	}
	g, err := execgw.NewRiskGuardian(opts)
	if err != nil {
		t.Fatalf("NewRiskGuardian: %v", err)
	}
	return &guardianRig{guardian: g, journal: j, clock: clk}
}

// collect is the plain collector: a fresh ledger version, a snapshot dated now,
// and an account with nothing open.
func (r *guardianRig) collect(ctx context.Context, _ int) (execgw.ExposureSnapshot, error) {
	r.collections++
	version, err := r.journal.ReservationVersion(ctx, "acct-7")
	if err != nil {
		return execgw.ExposureSnapshot{}, err
	}
	return execgw.ExposureSnapshot{
		AsOf:         r.clock.Now(),
		Version:      version,
		OpenExposure: riskcalc.Money{Amount: "0", Currency: "KRW"},
	}, nil
}

func (r *guardianRig) issue(ctx context.Context) (execgw.Issued, error) {
	return r.guardian.IssueEntry(ctx, execgw.EntryIssuance{
		Intent:  guardianIntent(),
		Account: guardianAccount(),
		Collect: r.collect,
	})
}

func decisionOnDisk(t *testing.T, j *journal.Journal, id string) bool {
	t.Helper()
	_, err := j.LookupDecision(context.Background(), id)
	switch {
	case err == nil:
		return true
	case errors.Is(err, journal.ErrDecisionNotFound):
		return false
	default:
		t.Fatalf("LookupDecision(%s): %v", id, err)
		return false
	}
}

func version(t *testing.T, j *journal.Journal) int64 {
	t.Helper()
	v, err := j.ReservationVersion(context.Background(), "acct-7")
	if err != nil {
		t.Fatalf("ReservationVersion: %v", err)
	}
	return v
}

// --- the happy path ---------------------------------------------------------

// TestTheGuardianIssuesTheDecisionAndItsReservationTogether is the procedure
// risk-management fixes: 체인 ALLOW → 결정 영속과 예약 삽입을 하나의 트랜잭션 →
// Gateway 제출. What is asserted is that the second step produced both rows and
// that the reference names a decision the gateway can actually read.
func TestTheGuardianIssuesTheDecisionAndItsReservationTogether(t *testing.T) {
	rig := newGuardian(t, nil)
	ctx := context.Background()

	issued, err := rig.issue(ctx)
	if err != nil {
		t.Fatalf("IssueEntry: %v", err)
	}
	if issued.Decision.IsZero() {
		t.Fatal("the issuance returned no reference")
	}
	if issued.ExpiresAt != rig.clock.Now().UTC().Add(execgw.DefaultDecisionTTL) {
		t.Errorf("expiry = %s, want issue time + %s (the implemented TTL)",
			issued.ExpiresAt, execgw.DefaultDecisionTTL)
	}

	dec, err := rig.journal.LookupDecision(ctx, issued.Decision.ID)
	if err != nil {
		t.Fatalf("the reference must name a persisted decision: %v", err)
	}
	if dec.SafetyClass != journal.SafetyClassExposureRaising {
		t.Errorf("safety class = %s, want EXPOSURE_RAISING", dec.SafetyClass)
	}
	preimage, err := journal.ParsePreimage(dec.PreimageKind, dec.RiskPreimage)
	if err != nil {
		t.Fatalf("ParsePreimage: %v", err)
	}
	intent, ok := preimage.(journal.RiskIntent)
	if !ok {
		t.Fatalf("preimage = %T, want a RiskIntent", preimage)
	}
	// The stop is on the row, not merely checked and discarded: it becomes the
	// exit policy's t0 baseline (risk-management → exit-policy).
	if intent.StopPrice != "69000" || intent.TargetPrice != "72000" || intent.Quantity != "10" {
		t.Errorf("preimage = %+v, want the intent's own numbers", intent)
	}
	if intent.PolicyVersion != "add-core-domain/4.1" {
		t.Errorf("policy version = %q; a decision that cannot name its rules cannot be audited",
			intent.PolicyVersion)
	}

	if len(issued.Reservations) != 1 {
		t.Fatalf("reservations = %d, want the one open-exposure hold", len(issued.Reservations))
	}
	res := issued.Reservations[0]
	if res.Kind != journal.ReservationKindOpenExposure || res.Amount != guardianExposure {
		t.Errorf("reservation = %+v, want %s of OPEN_EXPOSURE — the amount the chain found room for",
			res, guardianExposure)
	}
	if res.DecisionID != issued.Decision.ID || !res.Held() {
		t.Errorf("reservation %+v is not a HELD hold bound to the decision", res)
	}
	held, err := rig.journal.HeldReservations(ctx, "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	if len(held) != 1 {
		t.Errorf("held reservations = %d, want 1", len(held))
	}
}

// TestTheDecisionCarriesThePolicysOwnLimits: the snapshot stamped on the
// decision is the one derived from the policy the chain sized against. The
// gateway re-measures the order against it, so a Guardian that stamped anything
// else would be authorising against numbers nobody audited.
func TestTheDecisionCarriesThePolicysOwnLimits(t *testing.T) {
	rig := newGuardian(t, nil)
	issued, err := rig.issue(context.Background())
	if err != nil {
		t.Fatalf("IssueEntry: %v", err)
	}
	dec, err := rig.journal.LookupDecision(context.Background(), issued.Decision.ID)
	if err != nil {
		t.Fatal(err)
	}
	stamped, err := execgw.DecodeLimits(dec.LimitsJSON)
	if err != nil {
		t.Fatalf("DecodeLimits: %v", err)
	}
	if !reflect.DeepEqual(stamped, rig.guardian.ExposureLimits()) {
		t.Errorf("stamped %+v, declared %+v — one source or the audit is fiction",
			stamped, rig.guardian.ExposureLimits())
	}
}

// TestTheDefaultPolicyStatesTheSmallLiveSet pins the five fields plus the
// currency the interlock asks for, Set bits included (risk-management "정책
// 수치의 provenance": 사용자 미확정 시 보수 기본값 전체 집합).
func TestTheDefaultPolicyStatesTheSmallLiveSet(t *testing.T) {
	limits, err := execgw.ExposureLimitsFor(risk.DefaultPolicy())
	if err != nil {
		t.Fatalf("ExposureLimitsFor: %v", err)
	}
	want := execgw.Limits{
		MaxQuantity:        execgw.Bound(100),
		MaxNotional:        execgw.Bound(1_000_000),
		MaxTotalExposure:   execgw.Bound(10_000_000),
		MaxDailyLossAmount: execgw.Bound(100_000),
		MaxDailyLossRatio:  execgw.Bound(0.01),
		Currency:           "KRW",
	}
	if !reflect.DeepEqual(limits, want) {
		t.Errorf("limits = %+v, want the small_live set %+v", limits, want)
	}
}

// --- refusals ---------------------------------------------------------------

// TestAChainRefusalIssuesNothing: the chain refuses before anything is written,
// and the refusal carries the chain's own stable code.
func TestAChainRefusalIssuesNothing(t *testing.T) {
	rig := newGuardian(t, nil)
	ctx := context.Background()

	account := guardianAccount()
	account.AllowedSymbols = []string{"000660"}

	before := version(t, rig.journal)
	_, err := rig.guardian.IssueEntry(ctx, execgw.EntryIssuance{
		Intent: guardianIntent(), Account: account, Collect: rig.collect,
	})
	var refusal *execgw.IssueRefusal
	if !errors.As(err, &refusal) {
		t.Fatalf("err = %v, want an *IssueRefusal", err)
	}
	if refusal.Stage != execgw.StageChain || refusal.Reason != string(risk.ReasonSymbolNotAllowed) {
		t.Errorf("refusal = %+v, want the chain's SYMBOL_NOT_ALLOWED", refusal)
	}
	if rig.collections != 0 {
		t.Errorf("the collector ran %d times for an intent the chain refused; "+
			"a refused intent costs no broker round trip", rig.collections)
	}
	if after := version(t, rig.journal); after != before {
		t.Errorf("the ledger version moved from %d to %d on a chain refusal", before, after)
	}
}

// TestARefusedReservationLeavesNoSubmittableDecision is D1's whole point, one
// layer above the journal's own test: the chain allowed, the ledger refused, and
// what is left behind is nothing.
//
// The nonce is the unfakeable half. It is UNIQUE across the decisions table, so
// a second issuance that reuses it can only succeed if the first row is gone
// rather than merely unreadable.
func TestARefusedReservationLeavesNoSubmittableDecision(t *testing.T) {
	ctx := context.Background()
	tight := guardianPolicy()
	tight.MaxOpenExposure = riskcalc.Money{Amount: "1000000", Currency: "KRW"}

	rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) {
		o.Policy = tight
		o.NewID = fixedIDs("d-first", "nonce-first", "d-refused", "nonce-refused")
	})

	// The first entry holds 700,700 of the 1,000,000 ceiling.
	if _, err := rig.issue(ctx); err != nil {
		t.Fatalf("the first issuance must succeed: %v", err)
	}
	before := version(t, rig.journal)

	// The second reaches the ceiling. The chain still allows it — the caller's
	// snapshot says nothing is open, which is exactly the race the reservation
	// exists to settle (risk-management: 동시 결정이 한도 잔여를 소진해).
	_, err := rig.issue(ctx)
	var refusal *execgw.IssueRefusal
	if !errors.As(err, &refusal) {
		t.Fatalf("err = %v, want an *IssueRefusal", err)
	}
	if refusal.Stage != execgw.StageIssuance || refusal.Reason != journal.IssueReasonLimitReached {
		t.Errorf("refusal = %+v, want the issuance's LIMIT_REACHED", refusal)
	}
	if !errors.Is(err, journal.ErrReservationLimitExceeded) {
		t.Errorf("err = %v, want the limit sentinel to survive the wrapping", err)
	}

	if decisionOnDisk(t, rig.journal, "d-refused") {
		t.Fatal("a refused issuance left a submittable decision on disk")
	}
	reservations, err := rig.journal.ReservationsForDecision(ctx, "d-refused")
	if err != nil {
		t.Fatal(err)
	}
	if len(reservations) != 0 {
		t.Errorf("the refused issuance left %d reservations", len(reservations))
	}
	if after := version(t, rig.journal); after != before {
		t.Errorf("the ledger version moved from %d to %d on a refusal", before, after)
	}

	// The nonce is free again. Raise the ceiling so the retry has room, and mint
	// the refused decision's nonce a second time.
	roomy := guardianPolicy()
	rig.guardian = mustGuardian(t, execgw.RiskGuardianOptions{
		Journal: rig.journal, Clock: rig.clock, AccountRef: "acct-7",
		Policy: roomy, Costs: costs.DefaultModel(), PolicyVersion: "add-core-domain/4.1",
		NewID: fixedIDs("d-retry", "nonce-refused"),
	})
	if _, err := rig.issue(ctx); err != nil {
		t.Fatalf("the refused decision's nonce must be free again: %v", err)
	}
}

func mustGuardian(t *testing.T, opts execgw.RiskGuardianOptions) *execgw.RiskGuardian {
	t.Helper()
	g, err := execgw.NewRiskGuardian(opts)
	if err != nil {
		t.Fatalf("NewRiskGuardian: %v", err)
	}
	return g
}

// TestTheIssuanceReasonsAreDistinguishable walks the reservation failures a
// Guardian can produce and pins the stable code each one surfaces as
// (risk-management: 예약 실패는 원인별 안정 reason-code로 기록된다).
//
// Three of the four are reachable through this type. VERSION_CONFLICT is not,
// and that is structural rather than missing: the Guardian always runs the
// bounded re-collection loop, and the loop converges every stale-or-superseded
// refusal into SNAPSHOT_RECOLLECTION_EXHAUSTED at its end (issues.md, task 0.2).
// The single-shot journal API still reports it, so the code stays; what this
// test can honestly assert about it is the mapping, which the journal's own
// IssueRefusalReason test pins.
func TestTheIssuanceReasonsAreDistinguishable(t *testing.T) {
	ctx := context.Background()

	t.Run("LIMIT_REACHED", func(t *testing.T) {
		tight := guardianPolicy()
		tight.MaxOpenExposure = riskcalc.Money{Amount: "1000000", Currency: "KRW"}
		rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) { o.Policy = tight })
		if _, err := rig.issue(ctx); err != nil {
			t.Fatalf("seeding the hold: %v", err)
		}
		_, err := rig.issue(ctx)
		assertIssueReason(t, mustFail(t, err), journal.IssueReasonLimitReached)
	})

	t.Run("SNAPSHOT_RECOLLECTION_EXHAUSTED", func(t *testing.T) {
		rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) {
			o.Recollect = journal.RecollectPolicy{MaxAttempts: 3, Budget: time.Minute}
		})
		// A collector that never reports the version the ledger is actually at.
		// Every attempt is superseded, and running out of attempts is a refusal
		// and never a fallback to the last snapshot seen.
		_, err := rig.guardian.IssueEntry(ctx, execgw.EntryIssuance{
			Intent: guardianIntent(), Account: guardianAccount(),
			Collect: func(ctx context.Context, attempt int) (execgw.ExposureSnapshot, error) {
				rig.collections++
				return execgw.ExposureSnapshot{
					AsOf:         rig.clock.Now(),
					Version:      version(t, rig.journal) + 99,
					OpenExposure: riskcalc.Money{Amount: "0", Currency: "KRW"},
				}, nil
			},
		})
		assertIssueReason(t, err, journal.IssueReasonRecollectionExhausted)
		if rig.collections != 3 {
			t.Errorf("collections = %d, want the 3 the policy allows", rig.collections)
		}
	})

	t.Run("DECISION_EXPIRED", func(t *testing.T) {
		rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) { o.TTL = 5 * time.Second })
		// The decision's TTL is what bounds the freshness of everything the chain
		// read, so a collection that takes longer than the TTL must not produce an
		// issuance (design D1).
		_, err := rig.guardian.IssueEntry(ctx, execgw.EntryIssuance{
			Intent: guardianIntent(), Account: guardianAccount(),
			Collect: func(ctx context.Context, attempt int) (execgw.ExposureSnapshot, error) {
				rig.collections++
				rig.clock.Advance(10 * time.Second)
				return rig.collect(ctx, attempt)
			},
		})
		assertIssueReason(t, err, journal.IssueReasonDecisionExpired)
		if !errors.Is(err, journal.ErrDecisionExpired) {
			t.Errorf("err = %v, want the expiry sentinel", err)
		}
		held, herr := rig.journal.HeldReservations(ctx, "acct-7")
		if herr != nil {
			t.Fatal(herr)
		}
		if len(held) != 0 {
			t.Errorf("an expired issuance left %d holds", len(held))
		}
	})
}

func mustFail(t *testing.T, err error) error {
	t.Helper()
	if err == nil {
		t.Fatal("the issuance must be refused")
	}
	return err
}

func assertIssueReason(t *testing.T, err error, want string) {
	t.Helper()
	var refusal *execgw.IssueRefusal
	if !errors.As(err, &refusal) {
		t.Fatalf("err = %v, want an *IssueRefusal", err)
	}
	if refusal.Stage != execgw.StageIssuance {
		t.Errorf("stage = %s, want %s", refusal.Stage, execgw.StageIssuance)
	}
	if refusal.Reason != want {
		t.Errorf("reason = %q, want %q (%s)", refusal.Reason, want, refusal.Detail)
	}
}

// --- the chain runs once ----------------------------------------------------

// TestReCollectionDoesNotReRunTheChain is design D1 as a call count.
//
// Re-collection refreshes the *reservation snapshot* — the ledger version and
// the account's open exposure — and nothing else. Re-running the chain between
// attempts would silently re-judge the intent against inputs the decision's
// expiry never bounded, and the mode and latch it would be re-reading are
// already re-checked by the Gateway at submission.
func TestReCollectionDoesNotReRunTheChain(t *testing.T) {
	var chainCalls int
	restore := execgw.SetChainForTest(func(in risk.Input) risk.Decision {
		chainCalls++
		return risk.Evaluate(in)
	})
	defer restore()

	rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) {
		o.Recollect = journal.RecollectPolicy{MaxAttempts: 5, Budget: time.Minute}
	})
	ctx := context.Background()

	// Two superseded collections, then a good one.
	issued, err := rig.guardian.IssueEntry(ctx, execgw.EntryIssuance{
		Intent: guardianIntent(), Account: guardianAccount(),
		Collect: func(ctx context.Context, attempt int) (execgw.ExposureSnapshot, error) {
			rig.collections++
			snapshot, err := rig.collect(ctx, attempt)
			if err != nil {
				return execgw.ExposureSnapshot{}, err
			}
			if attempt < 3 {
				snapshot.Version += 7
			}
			return snapshot, nil
		},
	})
	if err != nil {
		t.Fatalf("the third attempt must succeed: %v", err)
	}
	if issued.Decision.IsZero() {
		t.Fatal("no reference was returned")
	}
	// collect() counts once per call and the wrapper counts once more, so six
	// counts is three collections.
	if rig.collections != 6 {
		t.Errorf("collections = %d, want 3 (counted twice each)", rig.collections)
	}
	if chainCalls != 1 {
		t.Errorf("the chain ran %d times for one issuance; re-collection must not re-run it", chainCalls)
	}
}

// --- reductions -------------------------------------------------------------

// TestAReductionIsIssuedWithNoLimitsAndNoReservation: an exit carries no limit
// snapshot (a limit that could refuse a liquidation is a trap, §0.3) and takes
// no reservation (it lowers the aggregate it would be reserving against).
func TestAReductionIsIssuedWithNoLimitsAndNoReservation(t *testing.T) {
	rig := newGuardian(t, nil)
	ctx := context.Background()

	account := guardianAccount()
	account.HeldQuantity = "10"
	// Everything an entry would be refused for, and none of it applies to an exit.
	account.KillSwitchActive = true
	account.Mode = risk.ModeHaltAll
	account.EntryBlockedLatch = true
	account.AllowedSymbols = nil

	intent := guardianIntent()
	intent.Side = risk.SideSell
	intent.Quantity = "4"

	issued, err := rig.guardian.IssueReduction(ctx, execgw.ReductionIssuance{
		Intent: intent, Account: account, Reason: "ratchet stop",
	})
	if err != nil {
		t.Fatalf("IssueReduction: %v", err)
	}
	if len(issued.Reservations) != 0 {
		t.Errorf("a reduction took %d reservations", len(issued.Reservations))
	}
	dec, err := rig.journal.LookupDecision(ctx, issued.Decision.ID)
	if err != nil {
		t.Fatal(err)
	}
	if dec.SafetyClass != journal.SafetyClassRiskReducing {
		t.Errorf("safety class = %s, want RISK_REDUCING", dec.SafetyClass)
	}
	if dec.LimitsJSON != "" {
		t.Errorf("the reduction carries a limit snapshot: %q", dec.LimitsJSON)
	}
	preimage, err := journal.ParsePreimage(dec.PreimageKind, dec.RiskPreimage)
	if err != nil {
		t.Fatal(err)
	}
	reduction, ok := preimage.(journal.ReductionIntent)
	if !ok {
		t.Fatalf("preimage = %T, want a ReductionIntent", preimage)
	}
	if reduction.MaxQuantity != "4" || reduction.Reason != "ratchet stop" {
		t.Errorf("preimage = %+v, want the ceiling and the reason", reduction)
	}
	if held, _ := rig.journal.HeldReservations(ctx, "acct-7"); len(held) != 0 {
		t.Errorf("held reservations = %d after a reduction", len(held))
	}
}

// TestAReductionBeyondTheHoldingIsRefused: the one thing an exit can get wrong.
func TestAReductionBeyondTheHoldingIsRefused(t *testing.T) {
	rig := newGuardian(t, nil)
	account := guardianAccount()
	account.HeldQuantity = "3"

	intent := guardianIntent()
	intent.Side = risk.SideSell
	intent.Quantity = "4"

	_, err := rig.guardian.IssueReduction(context.Background(), execgw.ReductionIssuance{
		Intent: intent, Account: account, Reason: "oversell",
	})
	var refusal *execgw.IssueRefusal
	if !errors.As(err, &refusal) {
		t.Fatalf("err = %v, want an *IssueRefusal", err)
	}
	if refusal.Reason != string(risk.ReasonSellExceedsHoldings) {
		t.Errorf("reason = %q, want SELL_EXCEEDS_HOLDINGS", refusal.Reason)
	}
}

// --- construction -----------------------------------------------------------

// TestAGuardianRefusesToExistWithoutWhatItNeeds. Each of these would otherwise
// surface as a refusal of every entry, reported at the least useful moment.
func TestAGuardianRefusesToExistWithoutWhatItNeeds(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	base := execgw.RiskGuardianOptions{
		Journal: j, Clock: clk, AccountRef: "acct-7",
		Policy: guardianPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "v1",
	}
	for _, tc := range []struct {
		name   string
		mutate func(*execgw.RiskGuardianOptions)
	}{
		{"no journal", func(o *execgw.RiskGuardianOptions) { o.Journal = nil }},
		{"no account", func(o *execgw.RiskGuardianOptions) { o.AccountRef = " " }},
		{"no policy version", func(o *execgw.RiskGuardianOptions) { o.PolicyVersion = "" }},
		{"no cost model", func(o *execgw.RiskGuardianOptions) { o.Costs = costs.Model{} }},
		{"an unusable policy", func(o *execgw.RiskGuardianOptions) {
			p := guardianPolicy()
			p.MaxOpenExposure = riskcalc.Money{}
			o.Policy = p
		}},
		{"a policy with no currency", func(o *execgw.RiskGuardianOptions) {
			p := guardianPolicy()
			p.MaxOrderNotional.Currency = ""
			o.Policy = p
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts := base
			tc.mutate(&opts)
			if _, err := execgw.NewRiskGuardian(opts); err == nil {
				t.Fatal("the Guardian must refuse to exist")
			}
		})
	}
	if _, err := execgw.NewRiskGuardian(base); err != nil {
		t.Errorf("a complete configuration must construct: %v", err)
	}
}

// TestTheDraftAuthorizeRefusesBothClasses. AuthorizationRequest predates the
// issuer and carries neither a stop nor a holding; issuing on a guessed default
// is what a checkpoint exists not to do.
func TestTheDraftAuthorizeRefusesBothClasses(t *testing.T) {
	rig := newGuardian(t, nil)
	for _, side := range []string{"BUY", "SELL"} {
		if _, err := rig.guardian.Authorize(context.Background(), execgw.AuthorizationRequest{
			Kind: journal.KindPlace, AccountRef: "acct-7", Market: "kr", Symbol: "005930",
			Side: side, Quantity: 1, Price: 70000, Currency: "KRW",
		}); err == nil {
			t.Errorf("Authorize(%s) returned a decision", side)
		}
	}
}

// TestTheGuardianRefusesAnIntentForAnotherAccount: a mis-scoped intent is a
// mistake worth stopping on, not a field worth correcting.
func TestTheGuardianRefusesAnIntentForAnotherAccount(t *testing.T) {
	rig := newGuardian(t, nil)
	intent := guardianIntent()
	intent.AccountRef = "acct-9"
	if _, err := rig.guardian.IssueEntry(context.Background(), execgw.EntryIssuance{
		Intent: intent, Account: guardianAccount(), Collect: rig.collect,
	}); err == nil {
		t.Fatal("an intent naming another account must be refused")
	}
}
