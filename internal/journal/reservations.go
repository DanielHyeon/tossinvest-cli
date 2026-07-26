package journal

// reservations.go is the entry-side risk reservation (design D5, task 3.1;
// order-execution "원자적 위험 예약").
//
// # What a reservation is for
//
// Per-order limits are checked per order. Account-wide limits — total open
// exposure, the day's loss, cash — cannot be, because two decisions on two
// different symbols can each pass the same snapshot and together exceed the
// aggregate. A reservation is the thing that makes the second one fail: the
// judgement and its consequence are one transaction, so there is no window in
// which a limit is "about to be" consumed.
//
// # The network is outside, the judgement is inside
//
// The broker snapshot is collected by the caller, before this transaction
// starts (SHALL NOT — journal은 단일 커넥션이므로 네트워크 왕복 동안 모든
// mutation 기록이 막힌다). What happens inside is: the snapshot's as-of is
// re-checked against the staleness bound, the ledger version the caller sized
// against is re-checked against the committed one, the aggregate is summed, and
// the rows are inserted. Any of those failing rolls the transaction back and
// the caller re-collects — bounded by ReserveWithRecollection's attempt cap and
// deadline, past which the answer is a refusal and never a guess.
//
// # Why the re-collection loop lives here and not in execgw
//
// It is the other half of this transaction's contract, not a policy on top of
// it: the cap, the deadline and the fail-closed exit exist *because* the
// transaction refuses, and the sentinel errors it retries on (ErrSnapshotStale,
// ErrSnapshotSuperseded) are defined here. Putting the loop in the gateway
// would mean two packages had to agree on which errors are retryable, and the
// gateway would be free to loop forever by accident. The journal still performs
// no I/O of its own: the collection callback is the caller's, and everything it
// returns is data.
//
// # Lock ordering
//
// Exactly one lock is taken here: SQLite's write lock, via BEGIN IMMEDIATE on
// the journal's single connection. Nothing in this file calls an exported
// *Journal method while a transaction is open — every helper takes the *sql.Tx.
// That rule is not stylistic: the pool holds one connection (SetMaxOpenConns(1)),
// so a method that opened its own statement inside a live transaction would
// wait for a connection its own caller is holding, and deadlock until the
// context expired.
//
// The process-wide order, for the two locks that exist at all, is
//
//	Attempt.mu  →  SQLite write lock
//
// which is the order durability.go's transition() already takes them in. No
// path in this file takes Attempt.mu, so the order cannot be inverted here.
//
// # Arithmetic
//
// Decimal strings, added exactly (internal/riskcalc): 예약 산술은 decimal 문자열
// 연산이며 float 누적을 사용하지 않는다(SHALL NOT). Ten reservations of "0.1"
// sum to "1" and not to 0.9999999999999999, which is the difference between a
// limit that holds and one that admits an eleventh position.
//
// # Broker-behaviour claims
//
// None. Every value here is TossOS-internal; the broker's contribution is the
// snapshot the caller collected, and this file only judges its age.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

var (
	// ErrSnapshotStale means the collected snapshot was older than the
	// staleness bound at the moment the transaction judged it. The caller
	// re-collects.
	ErrSnapshotStale = errors.New("journal: the broker snapshot is older than its staleness bound")

	// ErrSnapshotSuperseded means the reservation ledger moved between the
	// caller reading its version and this transaction running: somebody else's
	// decision was reserved, or lapsed reservations were released, so the
	// aggregate the caller sized against is no longer the aggregate. The caller
	// re-collects.
	ErrSnapshotSuperseded = errors.New("journal: the reservation ledger moved since the snapshot was collected")

	// ErrReservationLimitExceeded means the aggregate would reach or pass a
	// configured limit. Reaching it counts: risk-management refuses when a
	// limit is *reached* (한도 중 하나라도 도달한 상태 → 거부), unlike the
	// per-order maxima, which are inclusive ceilings.
	ErrReservationLimitExceeded = errors.New("journal: the reservation would reach an aggregate limit")

	// ErrRecollectionExhausted means the attempt cap or the deadline was spent
	// without a snapshot that held. It is a refusal, never a fallback to the
	// last snapshot seen.
	ErrRecollectionExhausted = errors.New("journal: snapshot re-collection was exhausted without a usable snapshot")

	// ErrReservationNotFound means no risk_reservations row carries that id.
	ErrReservationNotFound = errors.New("journal: risk reservation not found")
)

// Reservation is a stored risk_reservations row.
type Reservation struct {
	ID         string
	DecisionID string
	// AttemptID is empty until Prepare binds the decision's attempt to it.
	AttemptID  string
	AccountRef string
	Kind       string
	Amount     string
	Currency   string
	// TradingDay is set for DAILY_LOSS only, market-local YYYY-MM-DD.
	TradingDay   string
	SnapshotAsOf time.Time
	State        string
	ReleasedAt   time.Time
	// ReleaseReason is one of the ReleaseReason* constants, empty while HELD.
	ReleaseReason string
}

// Held reports whether the reservation still consumes limit headroom.
func (r Reservation) Held() bool { return r.State == ReservationHeld }

// ReservationRequest is one row to insert.
type ReservationRequest struct {
	// ID is caller-supplied and stable, like every other primary key in the
	// journal: a retry after an ambiguous failure must be recognisable.
	ID string
	// Kind is OPEN_EXPOSURE, DAILY_LOSS or CASH.
	Kind string
	// Amount is a non-negative decimal string in Currency.
	Amount   string
	Currency string
	// TradingDay is required for DAILY_LOSS and refused otherwise: a daily-loss
	// reservation lapses with the market's trading day, and one that does not
	// say which day cannot lapse at all.
	TradingDay string
}

// AggregateAmount is one kind's amount in one currency. It carries both the
// caller's snapshot-derived usage and the configured limit, because the two are
// compared and a comparison across currencies is not one.
type AggregateAmount struct {
	Kind     string
	Amount   string
	Currency string
}

// ReserveRequest is one reservation transaction.
//
// Note what is *not* here: no clock, and no limit lookup. The instant comes
// from the journal's injected clock, so a caller cannot make a stale snapshot
// fresh by passing an optimistic "now"; the limits come from the caller,
// because the journal has no business reading configuration and a storage layer
// that decides risk limits is a storage layer nobody can audit.
type ReserveRequest struct {
	// DecisionID is the decision these reservations belong to. It must already
	// be persisted — the schema's foreign key makes an unrecorded decision
	// structurally unable to reserve anything.
	DecisionID string
	AccountRef string
	// SnapshotAsOf is when the broker data behind SnapshotUsage was current.
	SnapshotAsOf time.Time
	// ObservedVersion is the value ReservationVersion returned *before* the
	// snapshot was collected. Zero is refused: a version nobody read is not
	// evidence that nothing moved.
	ObservedVersion int64
	// Staleness may narrow riskcalc's transcribed bound; it can never widen it.
	Staleness time.Duration
	// SnapshotUsage is what the broker snapshot already accounts for, per kind.
	// Held reservations are *not* included — this transaction sums those itself,
	// which is the whole point of holding them in the journal.
	//
	// Every kind being reserved needs an entry. An absent one is an unknown
	// quantity, not a zero: 정의되지 않은 양에 예약을 걸어서는 안 된다.
	SnapshotUsage []AggregateAmount
	// Limits is the configured limit per kind, in the same currency as the
	// usage and the reservation.
	Limits []AggregateAmount
	// Reservations are the rows to insert.
	Reservations []ReservationRequest
}

// ReserveResult is what one successful reservation transaction did.
type ReserveResult struct {
	// Version is the ledger version after the commit. A caller that reserves
	// twice in a row uses it as the next ObservedVersion without re-reading.
	Version      int64
	Reservations []Reservation
}

// ReservationVersion returns the account's reservation-ledger version.
//
// It counts state changes rather than rows — an insert and a release each move
// it — so it is monotone: reservations are never deleted, and a release never
// un-releases. A caller reads it before collecting a broker snapshot and hands
// it back to Reserve, which refuses if it has moved. That is what makes the
// as-of check a real re-validation rather than a formality: under a single
// connection the two transactions cannot interleave, so the only way to detect
// that a snapshot describes a superseded ledger is to compare something the
// ledger itself advances.
//
// The version starts at 1, so the zero value is never a legitimate one and a
// caller that forgot to read it is refused rather than silently accepted.
func (j *Journal) ReservationVersion(ctx context.Context, accountRef string) (int64, error) {
	return reservationVersion(ctx, j.db, accountRef)
}

type querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func reservationVersion(ctx context.Context, q querier, accountRef string) (int64, error) {
	account := strings.TrimSpace(accountRef)
	if account == "" {
		return 0, fmt.Errorf("%w: a reservation ledger version is per account", ErrInvalidRequest)
	}
	var version int64
	if err := q.QueryRowContext(ctx,
		`SELECT 1 + count(*) + coalesce(sum(CASE WHEN state = 'RELEASED' THEN 1 ELSE 0 END), 0)
		 FROM risk_reservations WHERE account_ref = ?`, account).Scan(&version); err != nil {
		return 0, fmt.Errorf("journal: reading the reservation ledger version for %s: %w", account, err)
	}
	return version, nil
}

// Reserve validates a pre-collected snapshot and inserts the reservations in
// one BEGIN IMMEDIATE transaction.
//
// It returns ErrSnapshotStale or ErrSnapshotSuperseded when the caller should
// re-collect, ErrReservationLimitExceeded when it should not, and writes
// nothing in either case.
func (j *Journal) Reserve(ctx context.Context, req ReserveRequest) (ReserveResult, error) {
	plan, err := req.build()
	if err != nil {
		return ReserveResult{}, err
	}

	now := j.clk.Now().UTC()

	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE (DSN _txlock=immediate)
	if err != nil {
		return ReserveResult{}, fmt.Errorf("journal: starting the reservation transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. The snapshot's age, judged against the injected clock rather than
	//    against anything the caller said about the time.
	if err := checkSnapshotFresh(req.SnapshotAsOf, now, req.Staleness); err != nil {
		return ReserveResult{}, err
	}

	// 2. The ledger version, before anything this transaction does moves it.
	current, err := reservationVersion(ctx, tx, plan.accountRef)
	if err != nil {
		return ReserveResult{}, err
	}
	if current != req.ObservedVersion {
		return ReserveResult{}, fmt.Errorf(
			"%w: the caller sized against version %d, the ledger is at %d",
			ErrSnapshotSuperseded, req.ObservedVersion, current)
	}

	// 3. Lapsed holds, computed here rather than on a timer (task 3.2): an
	//    expiry whose nonce was never spent, and a daily-loss hold whose market
	//    has moved to another trading day. It runs *after* the version check so
	//    that this transaction's own housekeeping does not invalidate the
	//    caller's snapshot — a release only frees headroom, and the caller's
	//    snapshot never counted our holds in the first place.
	if _, err := sweepLapsedReservations(ctx, tx, plan.accountRef, now); err != nil {
		return ReserveResult{}, err
	}

	// 4. The decision, read inside the transaction. A decision that expired
	//    between issue and reservation authorises nothing.
	if err := checkDecisionReservable(ctx, tx, plan, now); err != nil {
		return ReserveResult{}, err
	}

	// 5. The aggregate, summed exactly.
	held, err := heldTotals(ctx, tx, plan.accountRef)
	if err != nil {
		return ReserveResult{}, err
	}
	if err := plan.checkLimits(held); err != nil {
		return ReserveResult{}, err
	}

	// 6. Insert.
	out := ReserveResult{Reservations: make([]Reservation, 0, len(plan.rows))}
	for _, row := range plan.rows {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO risk_reservations
			   (id, decision_id, attempt_id, account_ref, kind, amount, currency,
			    trading_day, snapshot_as_of, state, released_at, release_reason)
			 VALUES (?,?,NULL,?,?,?,?,?,?,?,NULL,NULL)`,
			row.ID, plan.decisionID, plan.accountRef, row.Kind, row.Amount, row.Currency,
			nullableString(row.TradingDay), formatJournalTime(req.SnapshotAsOf),
			ReservationHeld); err != nil {
			return ReserveResult{}, fmt.Errorf("journal: recording reservation %s: %w", row.ID, err)
		}
		out.Reservations = append(out.Reservations, Reservation{
			ID:           row.ID,
			DecisionID:   plan.decisionID,
			AccountRef:   plan.accountRef,
			Kind:         row.Kind,
			Amount:       row.Amount,
			Currency:     row.Currency,
			TradingDay:   row.TradingDay,
			SnapshotAsOf: req.SnapshotAsOf.UTC(),
			State:        ReservationHeld,
		})
	}

	version, err := reservationVersion(ctx, tx, plan.accountRef)
	if err != nil {
		return ReserveResult{}, err
	}
	out.Version = version

	if err := tx.Commit(); err != nil {
		return ReserveResult{}, fmt.Errorf("journal: committing the reservation for decision %s: %w",
			plan.decisionID, err)
	}
	return out, nil
}

// CollectSnapshot produces one ReserveRequest from a freshly collected broker
// snapshot. Attempt counts from 1.
//
// It is the caller's function and it is where the network round trip happens —
// outside the transaction, by construction, because this signature is the only
// way a snapshot reaches Reserve.
type CollectSnapshot func(ctx context.Context, attempt int) (ReserveRequest, error)

// RecollectPolicy bounds the re-collection loop.
type RecollectPolicy struct {
	// MaxAttempts counts the first collection: 3 means one collection and two
	// re-collections. Zero takes DefaultRecollectAttempts.
	MaxAttempts int
	// Budget is the total wall time the loop may spend, measured on the
	// journal's clock. Zero takes DefaultRecollectBudget. It is a hard stop:
	// the alternative to a deadline is an entry decision that takes as long as
	// the broker's worst minute.
	Budget time.Duration
}

const (
	// DefaultRecollectAttempts is the cap the spec names (기본 3회).
	DefaultRecollectAttempts = 3
	// DefaultRecollectBudget is the total deadline. Ten seconds is the account
	// snapshot's own staleness bound (riskcalc.AccountSnapshotStaleness): a loop
	// allowed to run longer than the data stays fresh cannot succeed anyway, it
	// can only keep an entry decision open.
	DefaultRecollectBudget = riskcalc.AccountSnapshotStaleness
)

func (p RecollectPolicy) resolve() RecollectPolicy {
	if p.MaxAttempts <= 0 {
		p.MaxAttempts = DefaultRecollectAttempts
	}
	if p.Budget <= 0 {
		p.Budget = DefaultRecollectBudget
	}
	return p
}

// ReserveWithRecollection runs the collect-validate-insert loop.
//
// It re-collects only on the two errors that mean "the snapshot is no longer a
// description of the account" — a stale as-of and a superseded ledger version.
// A limit refusal is returned immediately: collecting the same account again
// will not make it fit, and retrying a limit is how a cap becomes a suggestion.
//
// Exhausting the attempts or the budget is a refusal (ErrRecollectionExhausted),
// wrapping the last reason. Nothing is reserved and nothing is submitted.
func (j *Journal) ReserveWithRecollection(ctx context.Context, collect CollectSnapshot, policy RecollectPolicy) (ReserveResult, error) {
	if collect == nil {
		return ReserveResult{}, fmt.Errorf("%w: reserving needs a snapshot collector", ErrInvalidRequest)
	}
	policy = policy.resolve()
	deadline := j.clk.Now().Add(policy.Budget)

	var last error
	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		if now := j.clk.Now(); now.After(deadline) {
			return ReserveResult{}, fmt.Errorf(
				"%w: the %s budget expired after %d attempt(s): %w",
				ErrRecollectionExhausted, policy.Budget, attempt-1, orNoAttempt(last))
		}
		req, err := collect(ctx, attempt)
		if err != nil {
			return ReserveResult{}, fmt.Errorf("journal: collecting the broker snapshot: %w", err)
		}
		result, err := j.Reserve(ctx, req)
		switch {
		case err == nil:
			return result, nil
		case errors.Is(err, ErrSnapshotStale), errors.Is(err, ErrSnapshotSuperseded):
			last = err
		default:
			return ReserveResult{}, err
		}
	}
	return ReserveResult{}, fmt.Errorf("%w: %d attempts spent: %w",
		ErrRecollectionExhausted, policy.MaxAttempts, orNoAttempt(last))
}

func orNoAttempt(err error) error {
	if err == nil {
		return errors.New("no attempt was made")
	}
	return err
}

// HeldReservations lists the account's reservations that still consume limit
// headroom, oldest first.
func (j *Journal) HeldReservations(ctx context.Context, accountRef string) ([]Reservation, error) {
	return j.queryReservations(ctx,
		reservationSelect+" WHERE account_ref = ? AND state = ? ORDER BY rowid",
		strings.TrimSpace(accountRef), ReservationHeld)
}

// ReservationsForDecision lists every reservation a decision took, held or not.
func (j *Journal) ReservationsForDecision(ctx context.Context, decisionID string) ([]Reservation, error) {
	return j.queryReservations(ctx,
		reservationSelect+" WHERE decision_id = ? ORDER BY rowid", strings.TrimSpace(decisionID))
}

// LookupReservation reads one reservation back.
func (j *Journal) LookupReservation(ctx context.Context, id string) (Reservation, error) {
	return scanReservation(j.db.QueryRowContext(ctx, reservationSelect+" WHERE id = ?", strings.TrimSpace(id)))
}

func (j *Journal) queryReservations(ctx context.Context, query string, args ...any) ([]Reservation, error) {
	rows, err := j.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("journal: listing reservations: %w", err)
	}
	defer rows.Close()

	var out []Reservation
	for rows.Next() {
		res, err := scanReservation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, res)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing reservations: %w", err)
	}
	return out, nil
}

const reservationSelect = `SELECT id, decision_id, attempt_id, account_ref, kind, amount,
	currency, trading_day, snapshot_as_of, state, released_at, release_reason
	FROM risk_reservations`

func scanReservation(row rowScanner) (Reservation, error) {
	var (
		res                       Reservation
		attemptID, tradingDay     sql.NullString
		releasedAt, releaseReason sql.NullString
		asOf                      string
	)
	err := row.Scan(&res.ID, &res.DecisionID, &attemptID, &res.AccountRef, &res.Kind,
		&res.Amount, &res.Currency, &tradingDay, &asOf, &res.State, &releasedAt, &releaseReason)
	if errors.Is(err, sql.ErrNoRows) {
		return Reservation{}, ErrReservationNotFound
	}
	if err != nil {
		return Reservation{}, fmt.Errorf("journal: reading a reservation: %w", err)
	}
	res.AttemptID = attemptID.String
	res.TradingDay = tradingDay.String
	res.ReleaseReason = releaseReason.String
	if res.SnapshotAsOf, err = parseJournalTime(asOf); err != nil {
		return Reservation{}, fmt.Errorf("journal: reservation %s snapshot_as_of: %w", res.ID, err)
	}
	if releasedAt.Valid && releasedAt.String != "" {
		if res.ReleasedAt, err = parseJournalTime(releasedAt.String); err != nil {
			return Reservation{}, fmt.Errorf("journal: reservation %s released_at: %w", res.ID, err)
		}
	}
	return res, nil
}

// --- the validated plan -----------------------------------------------------

// reservePlan is the request after validation: everything below this point
// works on canonical decimals and known-good kinds, so the transaction body has
// no input handling in it.
type reservePlan struct {
	decisionID string
	accountRef string
	rows       []ReservationRequest
	// usage and limits are keyed by kind, with the currency each was given in.
	usage  map[string]AggregateAmount
	limits map[string]AggregateAmount
	// newTotals is the sum of this request's own rows, per kind.
	newTotals map[string]string
}

func (r ReserveRequest) build() (reservePlan, error) {
	invalid := func(format string, args ...any) (reservePlan, error) {
		return reservePlan{}, fmt.Errorf("%w: %s", ErrInvalidRequest, fmt.Sprintf(format, args...))
	}
	plan := reservePlan{
		decisionID: strings.TrimSpace(r.DecisionID),
		accountRef: strings.TrimSpace(r.AccountRef),
		usage:      map[string]AggregateAmount{},
		limits:     map[string]AggregateAmount{},
		newTotals:  map[string]string{},
	}
	switch {
	case plan.decisionID == "":
		return invalid("a reservation belongs to a decision; none was named")
	case plan.accountRef == "":
		return invalid("a reservation is scoped to an account; none was named")
	case r.SnapshotAsOf.IsZero():
		return invalid("the snapshot carries no as-of, so its age cannot be established")
	case r.ObservedVersion <= 0:
		return invalid("no reservation ledger version was observed before the snapshot was collected; " +
			"a version nobody read is not evidence that nothing moved")
	case len(r.Reservations) == 0:
		return invalid("a reservation transaction with no reservations reserves nothing")
	}

	for _, row := range r.Reservations {
		id := strings.TrimSpace(row.ID)
		kind := strings.TrimSpace(row.Kind)
		currency := strings.ToUpper(strings.TrimSpace(row.Currency))
		day := strings.TrimSpace(row.TradingDay)
		switch {
		case id == "":
			return invalid("every reservation needs an id")
		case !validReservationKind(kind):
			return invalid("reservation kind %q is not one this build reserves", row.Kind)
		case currency == "":
			return invalid("reservation %s carries no currency, and there is no default one", id)
		case kind == ReservationKindDailyLoss && day == "":
			return invalid("reservation %s is a DAILY_LOSS reservation with no trading day; "+
				"one that cannot name its day cannot lapse with it", id)
		case kind != ReservationKindDailyLoss && day != "":
			return invalid("reservation %s is a %s reservation carrying a trading day; "+
				"only DAILY_LOSS lapses at the day boundary", id, kind)
		}
		amount, err := nonNegativeDecimal("reservation "+id+" amount", row.Amount)
		if err != nil {
			return reservePlan{}, err
		}
		plan.rows = append(plan.rows, ReservationRequest{
			ID: id, Kind: kind, Amount: amount, Currency: currency, TradingDay: day,
		})

		running, seen := plan.newTotals[kind]
		if !seen {
			running = "0"
		}
		sum, err := riskcalc.AddDecimal(running, amount)
		if err != nil {
			return reservePlan{}, fmt.Errorf("%w: summing the %s reservations: %v",
				ErrInvalidRequest, kind, err)
		}
		plan.newTotals[kind] = sum
	}

	for _, entry := range r.SnapshotUsage {
		normalised, err := normaliseAggregate("snapshot usage", entry)
		if err != nil {
			return reservePlan{}, err
		}
		plan.usage[normalised.Kind] = normalised
	}
	for _, entry := range r.Limits {
		normalised, err := normaliseAggregate("limit", entry)
		if err != nil {
			return reservePlan{}, err
		}
		plan.limits[normalised.Kind] = normalised
	}

	// Every kind being reserved needs both a usage and a limit, in the same
	// currency as the reservation. Missing either is an unknown quantity, and
	// a limit over an unknown quantity is not a control.
	for _, row := range plan.rows {
		usage, ok := plan.usage[row.Kind]
		if !ok {
			return invalid("no %s usage was supplied for the snapshot; an absent aggregate is unknown, not zero",
				row.Kind)
		}
		limit, ok := plan.limits[row.Kind]
		if !ok {
			return invalid("no %s limit was supplied; an absent limit is not an unlimited one", row.Kind)
		}
		if usage.Currency != row.Currency || limit.Currency != row.Currency {
			return invalid("reservation %s is in %s, the %s usage in %s and the limit in %s; "+
				"the journal does not convert currencies",
				row.ID, row.Currency, row.Kind, usage.Currency, limit.Currency)
		}
		positive, err := riskcalc.CompareDecimal(limit.Amount, "0")
		if err != nil {
			return reservePlan{}, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
		}
		if positive <= 0 {
			return invalid("the %s limit is %s; an absent or non-positive limit is not an unlimited one",
				row.Kind, limit.Amount)
		}
	}
	return plan, nil
}

// checkLimits applies usage + held + new >= limit → refuse, per kind.
func (p reservePlan) checkLimits(held map[string]AggregateAmount) error {
	for kind, added := range p.newTotals {
		usage := p.usage[kind]
		limit := p.limits[kind]

		total := usage.Amount
		if existing, ok := held[kind]; ok {
			if existing.Currency != limit.Currency {
				return fmt.Errorf(
					"%w: %s reservations are already held in %s but the limit is in %s; "+
						"the journal does not convert currencies",
					ErrInvalidRequest, kind, existing.Currency, limit.Currency)
			}
			sum, err := riskcalc.AddDecimal(total, existing.Amount)
			if err != nil {
				return fmt.Errorf("%w: summing held %s reservations: %v", ErrInvalidRequest, kind, err)
			}
			total = sum
		}
		sum, err := riskcalc.AddDecimal(total, added)
		if err != nil {
			return fmt.Errorf("%w: summing the new %s reservations: %v", ErrInvalidRequest, kind, err)
		}
		total = sum

		cmp, err := riskcalc.CompareDecimal(total, limit.Amount)
		if err != nil {
			return fmt.Errorf("%w: comparing the %s aggregate: %v", ErrInvalidRequest, kind, err)
		}
		if cmp >= 0 {
			return fmt.Errorf("%w: %s would reach %s %s against a limit of %s (snapshot %s, already held %s, new %s)",
				ErrReservationLimitExceeded, kind, total, limit.Currency, limit.Amount,
				usage.Amount, heldAmount(held, kind), added)
		}
	}
	return nil
}

func heldAmount(held map[string]AggregateAmount, kind string) string {
	if entry, ok := held[kind]; ok {
		return entry.Amount
	}
	return "0"
}

// --- transaction helpers (every one of these takes the tx) -------------------

// checkSnapshotFresh judges the snapshot's age. A future as-of is refused for
// the same reason riskcalc refuses one: a negative age means the clocks
// disagree, and a clock nobody can trust is not a freshness proof.
func checkSnapshotFresh(asOf, now time.Time, want time.Duration) error {
	// Narrower wins, exactly as riskcalc.Staleness does: a caller may tighten
	// the transcribed bound and can never loosen it (§0.9 as behaviour rather
	// than as a comment).
	bound := riskcalc.AccountSnapshotStaleness
	if want > 0 && want < bound {
		bound = want
	}

	age := now.Sub(asOf.UTC())
	if age < 0 {
		return fmt.Errorf("%w: the snapshot is dated %s in the future; the clocks disagree",
			ErrSnapshotStale, -age)
	}
	if age > bound {
		return fmt.Errorf("%w: the snapshot is %s old, past the %s bound",
			ErrSnapshotStale, age, bound)
	}
	return nil
}

// checkDecisionReservable reads the decision row inside the transaction and
// refuses anything that cannot authorise a reservation.
func checkDecisionReservable(ctx context.Context, tx *sql.Tx, plan reservePlan, now time.Time) error {
	var accountRef, expiresAt string
	err := tx.QueryRowContext(ctx,
		"SELECT account_ref, expires_at FROM decisions WHERE id = ?", plan.decisionID).
		Scan(&accountRef, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: reservation names decision %s, which is not recorded",
			ErrDecisionNotFound, plan.decisionID)
	}
	if err != nil {
		return fmt.Errorf("journal: reading decision %s for its reservation: %w", plan.decisionID, err)
	}
	if accountRef != plan.accountRef {
		return fmt.Errorf("%w: decision %s is for account %q, the reservation is on %q",
			ErrInvalidRequest, plan.decisionID, accountRef, plan.accountRef)
	}
	expires, err := parseJournalTime(expiresAt)
	if err != nil {
		return fmt.Errorf("journal: decision %s expires_at: %w", plan.decisionID, err)
	}
	if !expires.After(now) {
		return fmt.Errorf("%w: decision %s expired at %s; an expired decision reserves nothing",
			ErrInvalidRequest, plan.decisionID, expiresAt)
	}

	// A decision reserves once. Re-reserving would double-count the same
	// entitlement against the aggregate, and the release path would then have
	// to guess which of the two rows the attempt's terminal state closes.
	var existing int
	if err := tx.QueryRowContext(ctx,
		"SELECT count(*) FROM risk_reservations WHERE decision_id = ?", plan.decisionID).
		Scan(&existing); err != nil {
		return fmt.Errorf("journal: checking the reservations of decision %s: %w", plan.decisionID, err)
	}
	if existing > 0 {
		return fmt.Errorf("%w: decision %s has already reserved; a decision reserves once",
			ErrInvalidRequest, plan.decisionID)
	}
	return nil
}

// heldTotals sums the account's held reservations per kind, exactly.
func heldTotals(ctx context.Context, tx *sql.Tx, accountRef string) (map[string]AggregateAmount, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT kind, currency, amount FROM risk_reservations
		 WHERE account_ref = ? AND state = ?`, accountRef, ReservationHeld)
	if err != nil {
		return nil, fmt.Errorf("journal: summing held reservations for %s: %w", accountRef, err)
	}
	defer rows.Close()

	out := map[string]AggregateAmount{}
	for rows.Next() {
		var kind, currency, amount string
		if err := rows.Scan(&kind, &currency, &amount); err != nil {
			return nil, fmt.Errorf("journal: summing held reservations for %s: %w", accountRef, err)
		}
		entry, seen := out[kind]
		if !seen {
			out[kind] = AggregateAmount{Kind: kind, Amount: amount, Currency: currency}
			continue
		}
		if entry.Currency != currency {
			return nil, fmt.Errorf(
				"%w: %s reservations are held in both %s and %s on account %s; "+
					"the journal does not convert currencies",
				ErrInvalidRequest, kind, entry.Currency, currency, accountRef)
		}
		sum, err := riskcalc.AddDecimal(entry.Amount, amount)
		if err != nil {
			return nil, fmt.Errorf("%w: summing held %s reservations: %v", ErrInvalidRequest, kind, err)
		}
		entry.Amount = sum
		out[kind] = entry
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: summing held reservations for %s: %w", accountRef, err)
	}
	return out, nil
}

// bindReservationsToAttempt backfills risk_reservations.attempt_id inside the
// Prepare transaction (design D9: "attempt_id … Prepare 시 backfill").
//
// The reservation is taken before the attempt exists — that is the point of a
// reservation — so this is the join that later lets the attempt's terminal
// record release the hold in the same transaction as the record itself.
func bindReservationsToAttempt(ctx context.Context, tx *sql.Tx, decisionID, attemptID string) error {
	decision := strings.TrimSpace(decisionID)
	if decision == "" {
		return nil
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE risk_reservations SET attempt_id = ?
		 WHERE decision_id = ? AND attempt_id IS NULL`, attemptID, decision); err != nil {
		return fmt.Errorf("journal: binding the reservations of decision %s to attempt %s: %w",
			decision, attemptID, err)
	}
	return nil
}

// --- small validators -------------------------------------------------------

func validReservationKind(kind string) bool {
	switch kind {
	case ReservationKindOpenExposure, ReservationKindDailyLoss, ReservationKindCash:
		return true
	default:
		return false
	}
}

func nonNegativeDecimal(field, value string) (string, error) {
	canonical, err := riskcalc.CanonicalDecimal(value)
	if err != nil {
		return "", fmt.Errorf("%w: %s: %v", ErrInvalidRequest, field, err)
	}
	negative, err := riskcalc.IsNegativeDecimal(canonical)
	if err != nil {
		return "", fmt.Errorf("%w: %s: %v", ErrInvalidRequest, field, err)
	}
	if negative {
		return "", fmt.Errorf("%w: %s is %s; a negative amount would enlarge the headroom it consumes",
			ErrInvalidRequest, field, canonical)
	}
	return canonical, nil
}

func normaliseAggregate(label string, entry AggregateAmount) (AggregateAmount, error) {
	kind := strings.TrimSpace(entry.Kind)
	currency := strings.ToUpper(strings.TrimSpace(entry.Currency))
	if !validReservationKind(kind) {
		return AggregateAmount{}, fmt.Errorf("%w: %s names kind %q, which is not one this build reserves",
			ErrInvalidRequest, label, entry.Kind)
	}
	if currency == "" {
		return AggregateAmount{}, fmt.Errorf("%w: the %s %s carries no currency, and there is no default one",
			ErrInvalidRequest, kind, label)
	}
	amount, err := nonNegativeDecimal(kind+" "+label, entry.Amount)
	if err != nil {
		return AggregateAmount{}, err
	}
	return AggregateAmount{Kind: kind, Amount: amount, Currency: currency}, nil
}
