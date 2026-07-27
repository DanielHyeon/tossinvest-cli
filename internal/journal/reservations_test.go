package journal

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// reservations_test.go pins the entry-side reservation transaction (task 3.1,
// design D5, order-execution "원자적 위험 예약").
//
// The assertions that carry weight are:
//
//   - two decisions cannot both fit into the last slot of an aggregate limit
//     (the requirement's own scenario);
//   - the arithmetic is exact, so a ledger of fractional amounts does not drift
//     under the limit as it grows;
//   - a snapshot that has gone stale, or that describes a superseded ledger, is
//     refused rather than used;
//   - re-collection is bounded, and running out is a refusal.

func openReservationJournal(t *testing.T) (*Journal, *clock.Fake) {
	t.Helper()
	fake := clock.NewFake(time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC))
	j, err := Open(context.Background(), Options{
		Path:     filepath.Join(t.TempDir(), "journal.db"),
		Clock:    fake,
		FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	return j, fake
}

// recordEntryDecision persists an EXPOSURE_RAISING decision the reservation can
// be bound to. Every field it needs is the issuer's, not the reservation's.
func recordEntryDecision(t *testing.T, j *Journal, id, account string) Decision {
	t.Helper()
	issued := j.Now()
	dec, err := j.RecordDecision(context.Background(), DecisionRequest{
		ID:          id,
		AccountRef:  account,
		SafetyClass: SafetyClassExposureRaising,
		Kind:        KindPlace,
		Preimage: RiskIntent{
			AccountRef: account, Market: "kr", Symbol: "005930", Side: "BUY",
			Quantity: "10", EntryPrice: "70000", StopPrice: "68000",
			PolicyVersion: "test-1",
		},
		LimitsJSON: `{"max_notional":"1000000"}`,
		Nonce:      "nonce-" + id,
		IssuedAt:   issued,
		ExpiresAt:  issued.Add(10 * time.Minute),
	})
	if err != nil {
		t.Fatalf("RecordDecision(%s): %v", id, err)
	}
	return dec
}

// exposureReserve builds a one-row OPEN_EXPOSURE request.
func exposureReserve(j *Journal, decisionID, account, amount, usage, limit string, version int64) ReserveRequest {
	return ReserveRequest{
		DecisionID:      decisionID,
		AccountRef:      account,
		SnapshotAsOf:    j.Now(),
		ObservedVersion: version,
		SnapshotUsage: []AggregateAmount{
			{Kind: ReservationKindOpenExposure, Amount: usage, Currency: "KRW"},
		},
		Limits: []AggregateAmount{
			{Kind: ReservationKindOpenExposure, Amount: limit, Currency: "KRW"},
		},
		Reservations: []ReservationRequest{
			{ID: "res-" + decisionID, Kind: ReservationKindOpenExposure, Amount: amount, Currency: "KRW"},
		},
	}
}

func mustVersion(t *testing.T, j *Journal, account string) int64 {
	t.Helper()
	v, err := j.ReservationVersion(context.Background(), account)
	if err != nil {
		t.Fatalf("ReservationVersion: %v", err)
	}
	return v
}

func TestReservationVersionStartsAboveZero(t *testing.T) {
	j, _ := openReservationJournal(t)

	// Zero is never a legitimate version, so a caller that forgot to read one
	// is refused rather than accidentally matching a fresh account.
	if got := mustVersion(t, j, "acct-1"); got != 1 {
		t.Fatalf("a fresh account's ledger version = %d, want 1", got)
	}
	recordEntryDecision(t, j, "d-1", "acct-1")
	if _, err := j.Reserve(context.Background(),
		exposureReserve(j, "d-1", "acct-1", "100", "0", "1000", 0)); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("an unread ledger version must be refused, got %v", err)
	}
}

// TestConcurrentDecisionsCannotBothTakeTheLastSlot is the requirement's own
// scenario: 총 개방 노출 한도의 잔여분이 1건분만 남은 상태에서 서로 다른 두
// 심볼의 결정이 동시에 요청되면 하나만 예약에 성공한다.
func TestConcurrentDecisionsCannotBothTakeTheLastSlot(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	recordEntryDecision(t, j, "d-a", "acct-1")
	recordEntryDecision(t, j, "d-b", "acct-1")

	// Room for one 600 reservation under a 1000 limit; a second would reach it.
	version := mustVersion(t, j, "acct-1")

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		okCount int
		errs    []error
	)
	for _, id := range []string{"d-a", "d-b"} {
		wg.Add(1)
		go func(decision string) {
			defer wg.Done()
			_, err := j.Reserve(ctx, exposureReserve(j, decision, "acct-1", "600", "0", "1000", version))
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				okCount++
				return
			}
			errs = append(errs, err)
		}(id)
	}
	wg.Wait()

	if okCount != 1 {
		t.Fatalf("exactly one decision may reserve the last slot, %d succeeded (errors: %v)", okCount, errs)
	}
	if len(errs) != 1 {
		t.Fatalf("want one refusal, got %v", errs)
	}
	// Either refusal is correct: the loser either sized against a superseded
	// ledger or was measured against the winner's hold. Both are the aggregate
	// doing its job; neither is a silent success.
	if !errors.Is(errs[0], ErrSnapshotSuperseded) && !errors.Is(errs[0], ErrReservationLimitExceeded) {
		t.Fatalf("the losing decision must be refused for a superseded snapshot or the limit, got %v", errs[0])
	}

	held, err := j.HeldReservations(ctx, "acct-1")
	if err != nil {
		t.Fatalf("HeldReservations: %v", err)
	}
	if len(held) != 1 {
		t.Fatalf("want exactly one held reservation, got %d", len(held))
	}
}

// TestReachingTheAggregateLimitIsRefused pins the tie semantics the Manager
// fixed: 총계 한도는 도달 시 차단 (unlike the per-order maxima, which are
// inclusive ceilings).
func TestReachingTheAggregateLimitIsRefused(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	recordEntryDecision(t, j, "d-tie", "acct-1")

	_, err := j.Reserve(ctx, exposureReserve(j, "d-tie", "acct-1", "400", "600", "1000",
		mustVersion(t, j, "acct-1")))
	if !errors.Is(err, ErrReservationLimitExceeded) {
		t.Fatalf("600 + 400 against a limit of 1000 must be refused, got %v", err)
	}
	held, _ := j.HeldReservations(ctx, "acct-1")
	if len(held) != 0 {
		t.Fatalf("a refused reservation must write nothing, found %d rows", len(held))
	}
}

// TestReservationArithmeticIsExact is the fractional case the task names.
//
// Nine held reservations of "0.1" plus a tenth reach a limit of "1" exactly and
// are refused. A float64 accumulation of the same values totals
// 0.9999999999999999, which is under the limit and would admit the tenth — so
// this assertion fails on any implementation that accumulates in binary.
func TestReservationArithmeticIsExact(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()

	for i := 0; i < 9; i++ {
		id := "d-frac-" + string(rune('a'+i))
		recordEntryDecision(t, j, id, "acct-1")
		if _, err := j.Reserve(ctx, exposureReserve(j, id, "acct-1", "0.1", "0", "1",
			mustVersion(t, j, "acct-1"))); err != nil {
			t.Fatalf("reservation %d of 0.1 must fit under a limit of 1: %v", i+1, err)
		}
	}

	recordEntryDecision(t, j, "d-frac-tenth", "acct-1")
	_, err := j.Reserve(ctx, exposureReserve(j, "d-frac-tenth", "acct-1", "0.1", "0", "1",
		mustVersion(t, j, "acct-1")))
	if !errors.Is(err, ErrReservationLimitExceeded) {
		t.Fatalf("the tenth 0.1 reaches the limit of 1 exactly and must be refused, got %v", err)
	}

	// The same tenth fits under a limit that is genuinely larger, which proves
	// the refusal above is the tie and not an off-by-one.
	if _, err := j.Reserve(ctx, exposureReserve(j, "d-frac-tenth", "acct-1", "0.1", "0", "1.1",
		mustVersion(t, j, "acct-1"))); err != nil {
		t.Fatalf("the tenth 0.1 must fit under a limit of 1.1: %v", err)
	}
}

// TestFractionalShareAmountsSurviveTheLedger covers the US fractional-share
// case: quantities four decimal places out must not be rounded into the
// aggregate.
func TestFractionalShareAmountsSurviveTheLedger(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	recordEntryDecision(t, j, "d-us", "acct-us")

	req := exposureReserve(j, "d-us", "acct-us", "0.00010", "0", "1", mustVersion(t, j, "acct-us"))
	req.Reservations[0].Currency = "USD"
	req.SnapshotUsage[0].Currency = "USD"
	req.Limits[0].Currency = "USD"

	out, err := j.Reserve(ctx, req)
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if got := out.Reservations[0].Amount; got != "0.0001" {
		t.Fatalf("stored amount = %q, want the canonical 0.0001", got)
	}
	stored, err := j.LookupReservation(ctx, out.Reservations[0].ID)
	if err != nil {
		t.Fatalf("LookupReservation: %v", err)
	}
	if stored.Amount != "0.0001" || stored.Currency != "USD" || !stored.Held() {
		t.Fatalf("stored reservation = %+v", stored)
	}
}

func TestStaleSnapshotIsRefused(t *testing.T) {
	j, fake := openReservationJournal(t)
	ctx := context.Background()
	recordEntryDecision(t, j, "d-stale", "acct-1")

	asOf := j.Now()
	// riskcalc.AccountSnapshotStaleness is 10s; 11 puts the snapshot past it.
	fake.Advance(11 * time.Second)

	req := exposureReserve(j, "d-stale", "acct-1", "100", "0", "1000", mustVersion(t, j, "acct-1"))
	req.SnapshotAsOf = asOf
	if _, err := j.Reserve(ctx, req); !errors.Is(err, ErrSnapshotStale) {
		t.Fatalf("an 11s-old snapshot must be refused, got %v", err)
	}
}

func TestCallerMayNarrowTheStalenessBoundButNotWidenIt(t *testing.T) {
	j, fake := openReservationJournal(t)
	ctx := context.Background()
	recordEntryDecision(t, j, "d-narrow", "acct-1")
	recordEntryDecision(t, j, "d-wide", "acct-1")

	asOf := j.Now()
	fake.Advance(5 * time.Second)

	narrow := exposureReserve(j, "d-narrow", "acct-1", "1", "0", "1000", mustVersion(t, j, "acct-1"))
	narrow.SnapshotAsOf = asOf
	narrow.Staleness = 2 * time.Second
	if _, err := j.Reserve(ctx, narrow); !errors.Is(err, ErrSnapshotStale) {
		t.Fatalf("a caller-narrowed 2s bound must refuse a 5s-old snapshot, got %v", err)
	}

	// Asking for a wider window than the transcribed default does not widen it.
	fake.Advance(20 * time.Second)
	wide := exposureReserve(j, "d-wide", "acct-1", "1", "0", "1000", mustVersion(t, j, "acct-1"))
	wide.SnapshotAsOf = asOf
	wide.Staleness = time.Hour
	if _, err := j.Reserve(ctx, wide); !errors.Is(err, ErrSnapshotStale) {
		t.Fatalf("a caller must not be able to widen the staleness bound, got %v", err)
	}
}

func TestFutureSnapshotIsRefused(t *testing.T) {
	j, _ := openReservationJournal(t)
	recordEntryDecision(t, j, "d-future", "acct-1")

	req := exposureReserve(j, "d-future", "acct-1", "1", "0", "1000", mustVersion(t, j, "acct-1"))
	req.SnapshotAsOf = j.Now().Add(time.Minute)
	if _, err := j.Reserve(context.Background(), req); !errors.Is(err, ErrSnapshotStale) {
		t.Fatalf("a snapshot dated in the future must be refused, got %v", err)
	}
}

func TestUnknownAggregateIsRefusedRatherThanTreatedAsZero(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	recordEntryDecision(t, j, "d-unknown", "acct-1")

	missingUsage := exposureReserve(j, "d-unknown", "acct-1", "1", "0", "1000", mustVersion(t, j, "acct-1"))
	missingUsage.SnapshotUsage = nil
	if _, err := j.Reserve(ctx, missingUsage); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("an absent usage must be refused as unknown, got %v", err)
	}

	missingLimit := exposureReserve(j, "d-unknown", "acct-1", "1", "0", "1000", mustVersion(t, j, "acct-1"))
	missingLimit.Limits = nil
	if _, err := j.Reserve(ctx, missingLimit); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("an absent limit must be refused, got %v", err)
	}

	zeroLimit := exposureReserve(j, "d-unknown", "acct-1", "1", "0", "0", mustVersion(t, j, "acct-1"))
	if _, err := j.Reserve(ctx, zeroLimit); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("a zero limit is not an unlimited one, got %v", err)
	}
}

func TestMixedCurrenciesAreRefused(t *testing.T) {
	j, _ := openReservationJournal(t)
	recordEntryDecision(t, j, "d-fx", "acct-1")

	req := exposureReserve(j, "d-fx", "acct-1", "1", "0", "1000", mustVersion(t, j, "acct-1"))
	req.Reservations[0].Currency = "USD"
	if _, err := j.Reserve(context.Background(), req); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("the journal must not convert currencies, got %v", err)
	}
}

func TestReservationRequiresARecordedUnexpiredDecision(t *testing.T) {
	j, fake := openReservationJournal(t)
	ctx := context.Background()

	if _, err := j.Reserve(ctx, exposureReserve(j, "d-ghost", "acct-1", "1", "0", "1000", 1)); !errors.Is(err, ErrDecisionNotFound) {
		t.Fatalf("a reservation for an unrecorded decision must be refused, got %v", err)
	}

	recordEntryDecision(t, j, "d-expired", "acct-1")
	fake.Advance(11 * time.Minute) // past the decision's 10-minute validity
	req := exposureReserve(j, "d-expired", "acct-1", "1", "0", "1000", mustVersion(t, j, "acct-1"))
	req.SnapshotAsOf = j.Now()
	if _, err := j.Reserve(ctx, req); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("an expired decision reserves nothing, got %v", err)
	}
}

func TestADecisionReservesOnce(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	recordEntryDecision(t, j, "d-once", "acct-1")

	if _, err := j.Reserve(ctx, exposureReserve(j, "d-once", "acct-1", "1", "0", "1000",
		mustVersion(t, j, "acct-1"))); err != nil {
		t.Fatalf("first reservation: %v", err)
	}
	second := exposureReserve(j, "d-once", "acct-1", "1", "0", "1000", mustVersion(t, j, "acct-1"))
	second.Reservations[0].ID = "res-second"
	if _, err := j.Reserve(ctx, second); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("a decision must not reserve twice, got %v", err)
	}
}

func TestDailyLossReservationNeedsItsTradingDay(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	recordEntryDecision(t, j, "d-day", "acct-1")

	req := exposureReserve(j, "d-day", "acct-1", "1", "0", "1000", mustVersion(t, j, "acct-1"))
	req.Reservations[0].Kind = ReservationKindDailyLoss
	req.SnapshotUsage[0].Kind = ReservationKindDailyLoss
	req.Limits[0].Kind = ReservationKindDailyLoss
	if _, err := j.Reserve(ctx, req); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("a DAILY_LOSS reservation with no trading day must be refused, got %v", err)
	}

	req.Reservations[0].TradingDay = "2026-03-30"
	if _, err := j.Reserve(ctx, req); err != nil {
		t.Fatalf("a DAILY_LOSS reservation with its day must be accepted: %v", err)
	}

	// The converse: only DAILY_LOSS lapses with the day, so nothing else may
	// carry one.
	recordEntryDecision(t, j, "d-day2", "acct-1")
	other := exposureReserve(j, "d-day2", "acct-1", "1", "0", "1000", mustVersion(t, j, "acct-1"))
	other.Reservations[0].TradingDay = "2026-03-30"
	if _, err := j.Reserve(ctx, other); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("only DAILY_LOSS carries a trading day, got %v", err)
	}
}

// TestPrepareBindsTheDecisionsReservationsToTheAttempt pins the backfill design
// D9 requires: the reservation exists before the attempt, and this is the join
// the terminal record releases it through.
func TestPrepareBindsTheDecisionsReservationsToTheAttempt(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	dec := recordEntryDecision(t, j, "d-bind", "acct-1")

	out, err := j.Reserve(ctx, exposureReserve(j, dec.ID, "acct-1", "100", "0", "1000",
		mustVersion(t, j, "acct-1")))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if out.Reservations[0].AttemptID != "" {
		t.Fatal("a reservation is taken before its attempt exists; attempt_id must start empty")
	}

	if _, err := j.Prepare(ctx, reservationPrepare(dec, "attempt-1")); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	bound, err := j.LookupReservation(ctx, out.Reservations[0].ID)
	if err != nil {
		t.Fatalf("LookupReservation: %v", err)
	}
	if bound.AttemptID != "attempt-1" {
		t.Fatalf("attempt_id = %q, want attempt-1", bound.AttemptID)
	}
}

// reservationPrepare builds the PrepareRequest a decision's attempt uses.
func reservationPrepare(dec Decision, attemptID string) PrepareRequest {
	return PrepareRequest{
		Intent: Intent{
			ID: "intent-" + attemptID, Market: "kr", TradingDay: "2026-03-30",
			AccountRef: dec.AccountRef, Symbol: "005930", Side: "BUY", OrderType: "LIMIT",
			Quantity: "10", Price: "70000", Currency: "KRW", Source: "test",
			Fingerprint: "fp-" + attemptID,
		},
		Kind:          KindPlace,
		AttemptID:     attemptID,
		AccountRef:    dec.AccountRef,
		DecisionID:    dec.ID,
		SafetyClass:   dec.SafetyClass,
		ClientOrderID: dec.ClientOrderID,
	}
}

// --- the bounded re-collection loop -----------------------------------------

func TestRecollectionStopsAtTheAttemptCap(t *testing.T) {
	j, _ := openReservationJournal(t)
	recordEntryDecision(t, j, "d-cap", "acct-1")

	attempts := 0
	_, err := j.ReserveWithRecollection(context.Background(),
		func(_ context.Context, attempt int) (ReserveRequest, error) {
			attempts = attempt
			// A version nothing will ever match: every attempt is superseded.
			return exposureReserve(j, "d-cap", "acct-1", "1", "0", "1000", 99), nil
		}, RecollectPolicy{})
	if !errors.Is(err, ErrRecollectionExhausted) {
		t.Fatalf("exhausting the cap must be a refusal, got %v", err)
	}
	if !errors.Is(err, ErrSnapshotSuperseded) {
		t.Fatalf("the refusal must carry the last reason, got %v", err)
	}
	if attempts != DefaultRecollectAttempts {
		t.Fatalf("collected %d times, want the documented cap of %d", attempts, DefaultRecollectAttempts)
	}
}

func TestRecollectionStopsAtTheDeadline(t *testing.T) {
	j, fake := openReservationJournal(t)
	recordEntryDecision(t, j, "d-deadline", "acct-1")

	attempts := 0
	_, err := j.ReserveWithRecollection(context.Background(),
		func(_ context.Context, attempt int) (ReserveRequest, error) {
			attempts = attempt
			// Collecting takes longer than the whole budget allows.
			fake.Advance(4 * time.Second)
			return exposureReserve(j, "d-deadline", "acct-1", "1", "0", "1000", 99), nil
		}, RecollectPolicy{MaxAttempts: 10, Budget: 5 * time.Second})
	if !errors.Is(err, ErrRecollectionExhausted) {
		t.Fatalf("spending the budget must be a refusal, got %v", err)
	}
	if attempts >= 10 {
		t.Fatalf("the deadline must stop the loop before the cap, ran %d attempts", attempts)
	}
}

func TestRecollectionDoesNotRetryALimitRefusal(t *testing.T) {
	j, _ := openReservationJournal(t)
	recordEntryDecision(t, j, "d-limit", "acct-1")

	attempts := 0
	_, err := j.ReserveWithRecollection(context.Background(),
		func(_ context.Context, attempt int) (ReserveRequest, error) {
			attempts = attempt
			return exposureReserve(j, "d-limit", "acct-1", "900", "200", "1000",
				mustVersion(t, j, "acct-1")), nil
		}, RecollectPolicy{})
	if !errors.Is(err, ErrReservationLimitExceeded) {
		t.Fatalf("a limit refusal must be returned as-is, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("a limit refusal must not be retried; collected %d times", attempts)
	}
}

func TestRecollectionSucceedsOnASecondSnapshot(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	recordEntryDecision(t, j, "d-retry", "acct-1")

	attempts := 0
	out, err := j.ReserveWithRecollection(ctx,
		func(_ context.Context, attempt int) (ReserveRequest, error) {
			attempts = attempt
			version := int64(99) // superseded on the first pass
			if attempt > 1 {
				version = mustVersion(t, j, "acct-1")
			}
			return exposureReserve(j, "d-retry", "acct-1", "100", "0", "1000", version), nil
		}, RecollectPolicy{})
	if err != nil {
		t.Fatalf("ReserveWithRecollection: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("want two collections, got %d", attempts)
	}
	if len(out.Reservations) != 1 {
		t.Fatalf("want one reservation, got %d", len(out.Reservations))
	}
}
