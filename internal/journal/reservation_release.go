package journal

// reservation_release.go is the other half of the reservation contract: what
// lets a hold go (design D5, task 3.2; order-execution "원자적 위험 예약").
//
// # There are exactly four exits, and none of them is a guess
//
//	BROKER_TERMINAL     the attempt reached a *derived* terminal state
//	EXPIRED_UNCONSUMED  the decision expired with its nonce never spent
//	OPERATOR            a human, with a recorded reason and an audit line
//	DAY_BOUNDARY        a daily-loss hold lapsing with the market's trading day
//
// The one that is deliberately absent is "the order is probably gone by now".
// How an expired order appears in the broker's status vocabulary is
// [미측정 — 2b 2.1], and the derivation keeps the ambiguous combination
// (CLOSED, nothing filled, no cancellation) at UNKNOWN_BROKER_STATE rather than
// calling it cancelled. A hold released on that guess would free limit headroom
// for an order that is still live at the broker, which is the one direction
// that adds exposure nobody decided to take. So the hold stays, and an operator
// is told about it — ReservationsAwaitingOperator and the alerts RecordFill
// reports are what make a fail-closed ratchet something a human can see (D5:
// "운영자가 볼 수 없는 래칫은 아니다").
//
// # Release rides with the record that caused it
//
// Every automatic release happens inside the transaction that writes the
// triggering record — the attempt's terminal transition, or the fill snapshot
// that made the order terminal. Not afterwards, and not in a sweeper: a crash
// between "the order is finished" and "the hold is freed" would leave the
// account permanently smaller than its limits allow, and no producer (the fill
// detector, the resolution procedure, an operator) has to remember to do it.
//
// # Lazy, not timed
//
// The expiry and trading-day exits are computed when the ledger is next read or
// written, and once at startup — never on a timer. A background goroutine that
// releases risk holds is a second writer to the same rows with no transaction
// discipline, and it would keep running after the engine has stopped trading.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// AuditActionReservationRelease is the audit log's action string for an
// operator releasing a held reservation. The journal owns the string because
// the journal owns the event; internal/audit owns where it is written.
const AuditActionReservationRelease = "risk_reservation.release"

// ReservationAuditor records an operator release where an operator can find it.
//
// It is a one-method interface satisfied by *audit.Log rather than a direct
// dependency, so the journal does not import internal/audit: a storage layer
// that knows where the operator log lives is a storage layer with an opinion
// about deployment.
type ReservationAuditor interface {
	RecordAction(action, setting, value, detail string) error
}

// ReservationRelease is one hold that was let go, and why.
type ReservationRelease struct {
	ReservationID string
	DecisionID    string
	AccountRef    string
	Kind          string
	Amount        string
	Currency      string
	// Reason is one of the ReleaseReason* constants.
	Reason     string
	ReleasedAt time.Time
	// Detail is the operator-facing explanation.
	Detail string
}

// ReservationAlert is a hold that only an operator can release, with the
// evidence for why. It is what the spec's "운영자 해제 경로가 근거 기록과 함께
// 제공되고 알림이 발송된다" needs to be actionable.
type ReservationAlert struct {
	Reservation  Reservation
	AttemptID    string
	AttemptState AttemptState
	// Cause names the condition holding the reservation:
	// UNRESOLVED_IN_DOUBT, UNKNOWN_BROKER_STATE, NONCE_SPENT_EXPIRED or
	// UNKNOWN_MARKET.
	Cause  string
	Detail string
}

// The causes a ReservationAlert reports. Stable strings: alerting reads them.
const (
	// AlertCauseUnresolved — the attempt is parked as UNRESOLVED_IN_DOUBT.
	AlertCauseUnresolved = "UNRESOLVED_IN_DOUBT"
	// AlertCauseBrokerStateUnknown — the last observation of the order failed
	// closed, so nothing derived it as terminal.
	AlertCauseBrokerStateUnknown = "UNKNOWN_BROKER_STATE"
	// AlertCauseNonceSpentExpired — the decision expired, but its nonce was
	// spent: something was sent, so expiry proves nothing.
	AlertCauseNonceSpentExpired = "NONCE_SPENT_EXPIRED"
	// AlertCauseUnknownMarket — the reservation's market cannot be resolved, so
	// its trading day cannot be computed and the day boundary cannot lapse it.
	AlertCauseUnknownMarket = "UNKNOWN_MARKET"
)

// --- (a) derived terminal states --------------------------------------------

// releasesReservations reports whether reaching this attempt state frees the
// decision's holds.
//
// CONFIRMED does not: the order exists, so the exposure is real and the hold
// stands until the *order* reaches a derived terminal state (which arrives
// through RecordFill). UNRESOLVED_IN_DOUBT does not either — it is precisely
// the state whose only exit is an operator.
func releasesReservations(state AttemptState) bool {
	switch state {
	case StateNotDispatched, StateFailedConfirmed:
		return true
	default:
		return false
	}
}

// releaseReservationsForAttempt frees every hold bound to an attempt, inside
// the caller's transaction.
func releaseReservationsForAttempt(ctx context.Context, tx *sql.Tx, attemptID, reason, detail, now string) ([]ReservationRelease, error) {
	if strings.TrimSpace(attemptID) == "" {
		return nil, nil
	}
	return releaseWhere(ctx, tx, reason, detail, now,
		"attempt_id = ? AND state = ?", attemptID, ReservationHeld)
}

// releaseReservationsForOrder frees the holds of whichever attempt named this
// broker order.
//
// The identifier is compared byte-for-byte: `orderId` is an opaque token
// (openapi contracts no shape for it), so the join is on what the broker sent
// and nothing is trimmed or folded on the way in.
func releaseReservationsForOrder(ctx context.Context, tx *sql.Tx, orderID, intentID, reason, detail,
	now string,
) ([]ReservationRelease, error) {
	if orderID == "" || strings.TrimSpace(intentID) == "" {
		return nil, nil
	}
	return releaseWhere(ctx, tx, reason, detail, now,
		`state = ? AND attempt_id IN (
			SELECT id FROM mutation_attempts WHERE broker_order_id = ? AND intent_id = ?
		)`,
		ReservationHeld, orderID, strings.TrimSpace(intentID))
}

// releaseWhere is the single UPDATE every automatic release goes through. It
// reads the rows first so the caller can report exactly what it freed, and both
// statements are in the caller's transaction.
func releaseWhere(ctx context.Context, tx *sql.Tx, reason, detail, now, where string, args ...any) ([]ReservationRelease, error) {
	if !validReleaseReason(reason) {
		return nil, fmt.Errorf("%w: release reason %q is not one this build records", ErrInvalidRequest, reason)
	}
	rows, err := tx.QueryContext(ctx,
		`SELECT id, decision_id, account_ref, kind, amount, currency FROM risk_reservations WHERE `+where, args...)
	if err != nil {
		return nil, fmt.Errorf("journal: finding the reservations to release: %w", err)
	}
	var released []ReservationRelease
	for rows.Next() {
		var rel ReservationRelease
		if err := rows.Scan(&rel.ReservationID, &rel.DecisionID, &rel.AccountRef,
			&rel.Kind, &rel.Amount, &rel.Currency); err != nil {
			rows.Close()
			return nil, fmt.Errorf("journal: finding the reservations to release: %w", err)
		}
		rel.Reason = reason
		rel.Detail = detail
		released = append(released, rel)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("journal: finding the reservations to release: %w", err)
	}
	rows.Close()
	if len(released) == 0 {
		return nil, nil
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE risk_reservations SET state = ?, released_at = ?, release_reason = ? WHERE `+where,
		append([]any{ReservationReleased, now, reason}, args...)...); err != nil {
		return nil, fmt.Errorf("journal: releasing reservations: %w", err)
	}

	releasedAt, err := parseJournalTime(now)
	if err != nil {
		return nil, err
	}
	for i := range released {
		released[i].ReleasedAt = releasedAt
	}
	return released, nil
}

func validReleaseReason(reason string) bool {
	switch reason {
	case ReleaseReasonBrokerTerminal, ReleaseReasonExpiredUnconsumed,
		ReleaseReasonOperator, ReleaseReasonDayBoundary:
		return true
	default:
		return false
	}
}

// alertsForOrder reports the holds an order's fail-closed observation left
// standing, so the caller can raise the operator alert at the moment the
// ambiguity was observed rather than on the next sweep.
func alertsForOrder(ctx context.Context, tx *sql.Tx, scope FillSnapshotScope,
	evidenceAt, cause, detail string) ([]ReservationAlert, error) {
	scope = canonicalFillSnapshotScope(scope)
	if !scope.complete() {
		return nil, nil
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT r.id, r.decision_id, r.attempt_id, r.account_ref, r.kind, r.amount,
		       r.currency, r.trading_day, r.snapshot_as_of, r.state, r.released_at, r.release_reason
		  FROM risk_reservations r
		  JOIN mutation_attempts a ON a.id = r.attempt_id
		  JOIN intents i ON i.id = a.intent_id
		 WHERE r.state = ? AND r.decision_id = a.decision_id
		   AND a.state = ? AND a.kind IN ('PLACE','AMEND') AND a.settled_at < ?
		   AND a.broker_order_id = ?
		   AND TRIM(i.account_ref) = ? AND LOWER(TRIM(i.market)) = ?
		   AND TRIM(i.trading_day) = ? AND UPPER(TRIM(i.symbol)) = ? AND UPPER(TRIM(i.side)) = ?
		   AND 1 = (SELECT COUNT(DISTINCT owner.intent_id)
			  FROM mutation_attempts owner JOIN intents owned ON owned.id = owner.intent_id
			 WHERE owner.state = ? AND owner.kind IN ('PLACE','AMEND')
			   AND owner.settled_at < ? AND owner.broker_order_id = ?
			   AND TRIM(owned.account_ref) = ? AND LOWER(TRIM(owned.market)) = ?
			   AND TRIM(owned.trading_day) = ? AND UPPER(TRIM(owned.symbol)) = ?
			   AND UPPER(TRIM(owned.side)) = ?)`,
		ReservationHeld, string(StateConfirmed), evidenceAt,
		scope.OrderID, scope.AccountRef, scope.Market,
		scope.TradingDay, scope.Symbol, scope.Side,
		string(StateConfirmed), evidenceAt, scope.OrderID, scope.AccountRef, scope.Market,
		scope.TradingDay, scope.Symbol, scope.Side)
	if err != nil {
		return nil, fmt.Errorf("journal: finding the reservations held by order %s: %w", scope.OrderID, err)
	}
	defer rows.Close()

	var out []ReservationAlert
	for rows.Next() {
		res, err := scanReservation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, ReservationAlert{
			Reservation: res,
			AttemptID:   res.AttemptID,
			Cause:       cause,
			Detail:      detail,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: finding the reservations held by order %s: %w", scope.OrderID, err)
	}
	return out, nil
}

// --- (b) + (d) the lapse sweep ----------------------------------------------

// ReservationSweep is what one lapse pass did.
type ReservationSweep struct {
	Released []ReservationRelease
	// Preserved are holds the pass deliberately left alone, with the reason an
	// operator needs.
	Preserved []ReservationAlert
}

// sweepLapsedReservations applies the two time-driven exits inside the caller's
// transaction: an expiry whose nonce was never spent, and a daily-loss hold
// whose market has moved on to another trading day.
//
// accountRef narrows the pass to one account; empty sweeps every account, which
// is what startup does.
func sweepLapsedReservations(ctx context.Context, tx *sql.Tx, accountRef string, now time.Time) (ReservationSweep, error) {
	var sweep ReservationSweep
	nowText := formatJournalTime(now)

	expired, err := sweepExpiredUnconsumed(ctx, tx, accountRef, nowText)
	if err != nil {
		return ReservationSweep{}, err
	}
	sweep.Released = append(sweep.Released, expired.Released...)
	sweep.Preserved = append(sweep.Preserved, expired.Preserved...)

	lapsed, err := sweepTradingDayBoundary(ctx, tx, accountRef, now, nowText)
	if err != nil {
		return ReservationSweep{}, err
	}
	sweep.Released = append(sweep.Released, lapsed.Released...)
	sweep.Preserved = append(sweep.Preserved, lapsed.Preserved...)
	return sweep, nil
}

// sweepExpiredUnconsumed releases the holds of decisions that expired without
// ever being sent.
//
// The nonce is the evidence. An unspent nonce means MarkDispatchStarted never
// ran, so nothing left the process and the hold is dead weight. A *spent* nonce
// means a request was sent and the response may have been lost — the order
// could be live — so the expiry releases nothing and the hold waits for the
// resolution procedure or an operator (spec: 소비 후 만료는 해제하지 않는다).
func sweepExpiredUnconsumed(ctx context.Context, tx *sql.Tx, accountRef, nowText string) (ReservationSweep, error) {
	query := `SELECT r.id, d.nonce FROM risk_reservations r
	          JOIN decisions d ON d.id = r.decision_id
	          WHERE r.state = ? AND d.expires_at <= ?`
	args := []any{ReservationHeld, nowText}
	if accountRef != "" {
		query += " AND r.account_ref = ?"
		args = append(args, accountRef)
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return ReservationSweep{}, fmt.Errorf("journal: finding expired reservations: %w", err)
	}
	type candidate struct{ id, nonce string }
	var candidates []candidate
	for rows.Next() {
		var c candidate
		if err := rows.Scan(&c.id, &c.nonce); err != nil {
			rows.Close()
			return ReservationSweep{}, fmt.Errorf("journal: finding expired reservations: %w", err)
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return ReservationSweep{}, fmt.Errorf("journal: finding expired reservations: %w", err)
	}
	rows.Close()

	var sweep ReservationSweep
	for _, c := range candidates {
		var consumedAt string
		err := tx.QueryRowContext(ctx,
			"SELECT consumed_at FROM spent_nonces WHERE nonce = ?", c.nonce).Scan(&consumedAt)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			released, err := releaseWhere(ctx, tx, ReleaseReasonExpiredUnconsumed,
				"the decision expired with its nonce never spent; nothing was sent",
				nowText, "id = ? AND state = ?", c.id, ReservationHeld)
			if err != nil {
				return ReservationSweep{}, err
			}
			sweep.Released = append(sweep.Released, released...)
		case err != nil:
			return ReservationSweep{}, fmt.Errorf("journal: reading the spent nonce of reservation %s: %w", c.id, err)
		default:
			res, err := scanReservation(tx.QueryRowContext(ctx, reservationSelect+" WHERE id = ?", c.id))
			if err != nil {
				return ReservationSweep{}, err
			}
			sweep.Preserved = append(sweep.Preserved, ReservationAlert{
				Reservation: res,
				AttemptID:   res.AttemptID,
				Cause:       AlertCauseNonceSpentExpired,
				Detail: fmt.Sprintf(
					"the decision expired but its nonce was spent at %s; a request was sent and may have been accepted",
					consumedAt),
			})
		}
	}
	return sweep, nil
}

// sweepTradingDayBoundary lapses daily-loss holds whose trading day has passed.
//
// The day is per market (internal/clock): one UTC instant is two different
// trading days in Seoul and New York, so a single "today" would release a
// Korean hold while a US one is still inside its day. The market comes from the
// decision's preimage, which is the only place it is recorded at reservation
// time — the attempt, and therefore the intent, may not exist yet.
//
// A market this build cannot parse preserves the hold and reports it. Refusing
// to lapse is the conservative direction: the limit stays tighter than it needs
// to be, rather than looser than it should be.
func sweepTradingDayBoundary(ctx context.Context, tx *sql.Tx, accountRef string, now time.Time, nowText string) (ReservationSweep, error) {
	query := `SELECT r.id, r.trading_day, d.preimage_kind, d.risk_preimage
	          FROM risk_reservations r JOIN decisions d ON d.id = r.decision_id
	          WHERE r.state = ? AND r.kind = ?`
	args := []any{ReservationHeld, ReservationKindDailyLoss}
	if accountRef != "" {
		query += " AND r.account_ref = ?"
		args = append(args, accountRef)
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return ReservationSweep{}, fmt.Errorf("journal: finding daily-loss reservations: %w", err)
	}
	type candidate struct{ id, day, kind, preimage string }
	var candidates []candidate
	for rows.Next() {
		var (
			c   candidate
			day sql.NullString
		)
		if err := rows.Scan(&c.id, &day, &c.kind, &c.preimage); err != nil {
			rows.Close()
			return ReservationSweep{}, fmt.Errorf("journal: finding daily-loss reservations: %w", err)
		}
		c.day = day.String
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return ReservationSweep{}, fmt.Errorf("journal: finding daily-loss reservations: %w", err)
	}
	rows.Close()

	var sweep ReservationSweep
	for _, c := range candidates {
		today, err := tradingDayForPreimage(c.kind, c.preimage, now)
		if err != nil {
			res, scanErr := scanReservation(tx.QueryRowContext(ctx, reservationSelect+" WHERE id = ?", c.id))
			if scanErr != nil {
				return ReservationSweep{}, scanErr
			}
			sweep.Preserved = append(sweep.Preserved, ReservationAlert{
				Reservation: res,
				AttemptID:   res.AttemptID,
				Cause:       AlertCauseUnknownMarket,
				Detail: fmt.Sprintf(
					"the trading day of reservation %s cannot be computed (%v); it is held rather than lapsed",
					c.id, err),
			})
			continue
		}
		if today == c.day {
			continue
		}
		released, err := releaseWhere(ctx, tx, ReleaseReasonDayBoundary,
			fmt.Sprintf("the trading day moved from %s to %s", c.day, today),
			nowText, "id = ? AND state = ?", c.id, ReservationHeld)
		if err != nil {
			return ReservationSweep{}, err
		}
		sweep.Released = append(sweep.Released, released...)
	}
	return sweep, nil
}

// tradingDayForPreimage resolves the market-local trading day a decision's
// reservation belongs to.
func tradingDayForPreimage(kind, preimage string, now time.Time) (string, error) {
	parsed, err := ParsePreimage(kind, preimage)
	if err != nil {
		return "", err
	}
	_, market, _, _ := PreimageVenue(parsed)
	parsedMarket, err := clock.ParseMarket(market)
	if err != nil {
		return "", err
	}
	return parsedMarket.TradingDay(now)
}

// --- the startup sweep (task 3.3) -------------------------------------------

// SweepReservations recovers orphaned holds and applies the time-driven exits,
// across every account. It is what a restart runs once, before the engine takes
// any decision.
//
// Three things happen, and the split between them is the whole point:
//
//   - a hold whose attempt or order already reached a releasing terminal state
//     is released. Going forward that cannot happen — the release rides in the
//     same transaction as the record — but a row written by an older build, or
//     a database restored from the pre-migration backup, can be in that state,
//     and a hold nothing will ever release shrinks the account's limits forever.
//   - an expired decision whose nonce was never spent is released, and a
//     daily-loss hold from a previous trading day lapses.
//   - everything else is *preserved* and reported. An UNRESOLVED_IN_DOUBT
//     attempt, an order whose last observation failed closed, an expiry whose
//     nonce was spent: in each of those the order may exist, so the hold is
//     correct and only an operator may remove it.
//
// It is idempotent: running it twice releases nothing the second time.
func (j *Journal) SweepReservations(ctx context.Context) (ReservationSweep, error) {
	now := j.clk.Now().UTC()

	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return ReservationSweep{}, fmt.Errorf("journal: starting the reservation sweep: %w", err)
	}
	defer tx.Rollback()

	sweep, err := sweepLapsedReservations(ctx, tx, "", now)
	if err != nil {
		return ReservationSweep{}, err
	}
	orphans, err := sweepOrphanedTerminals(ctx, tx, formatJournalTime(now))
	if err != nil {
		return ReservationSweep{}, err
	}
	sweep.Released = append(sweep.Released, orphans...)

	if err := tx.Commit(); err != nil {
		return ReservationSweep{}, fmt.Errorf("journal: committing the reservation sweep: %w", err)
	}

	// The operator-facing half is read after the commit, so it reports what is
	// still held *after* the sweep rather than what was held before it.
	awaiting, err := j.ReservationsAwaitingOperator(ctx)
	if err != nil {
		return ReservationSweep{}, err
	}
	sweep.Preserved = mergeAlerts(sweep.Preserved, awaiting)
	return sweep, nil
}

// sweepOrphanedTerminals releases holds whose attempt or order is already over.
//
// "Already over" is read from the same two sources the live path uses and from
// nothing else: an attempt state that releases (NOT_DISPATCHED,
// FAILED_CONFIRMED), or a fill snapshot the caller derived as terminal without
// failing closed. A terminal snapshot releases only the reservation whose
// confirmed attempt and intent match its complete canonical scope, and only
// when that scope has one intent owner. Missing scope or ambiguous ownership
// stays held; ambiguity also enters the existing IDENTIFIER_CONFLICT contract.
// A snapshot that failed closed is not terminal, which is what keeps the
// assumed-expiry case out of this sweep as well.
func sweepOrphanedTerminals(ctx context.Context, tx *sql.Tx, nowText string) ([]ReservationRelease, error) {
	byAttemptState, err := releaseWhere(ctx, tx, ReleaseReasonBrokerTerminal,
		"recovered at startup: the attempt was already in a terminal state that releases", nowText,
		`state = ? AND attempt_id IN (SELECT id FROM mutation_attempts WHERE state IN (?,?))`,
		ReservationHeld, string(StateNotDispatched), string(StateFailedConfirmed))
	if err != nil {
		return nil, err
	}

	type terminalCandidate struct {
		reservationID string
		accountRef    string
		orderID       string
		market        string
		tradingDay    string
		symbol        string
		side          string
		intentOwners  int
	}
	rows, err := tx.QueryContext(ctx, allFillSnapshotsCTE+`
		SELECT r.id, i.account_ref, f.order_id, i.market, i.trading_day, i.symbol, i.side,
		       (SELECT count(DISTINCT owner_intent.id)
		          FROM mutation_attempts owner_attempt
		          JOIN intents owner_intent ON owner_intent.id = owner_attempt.intent_id
		         WHERE owner_attempt.state = ? AND owner_attempt.kind IN ('PLACE','AMEND')
		           AND owner_attempt.settled_at < f.committed_at
		           AND owner_attempt.broker_order_id = f.order_id
		           AND TRIM(owner_intent.account_ref) = TRIM(f.account_ref)
		           AND LOWER(TRIM(owner_intent.market)) = LOWER(TRIM(f.market))
		           AND TRIM(owner_intent.trading_day) = TRIM(f.trading_day)
		           AND UPPER(TRIM(owner_intent.symbol)) = UPPER(TRIM(f.symbol))
		           AND UPPER(TRIM(owner_intent.side)) = UPPER(TRIM(f.side)))
		  FROM risk_reservations r
		  JOIN mutation_attempts a ON a.id = r.attempt_id
		  JOIN intents i ON i.id = a.intent_id
		  JOIN all_fill_snapshots f ON f.order_id = a.broker_order_id
		 WHERE r.state = ? AND a.state = ? AND a.kind IN ('PLACE','AMEND')
		   AND a.settled_at < f.committed_at
		   AND f.terminal = 1 AND f.fail_closed = 0
		   AND r.decision_id = a.decision_id
		   AND TRIM(r.account_ref) = TRIM(i.account_ref)
		   AND TRIM(f.account_ref) = TRIM(i.account_ref)
		   AND LOWER(TRIM(f.market)) = LOWER(TRIM(i.market))
		   AND TRIM(f.trading_day) = TRIM(i.trading_day)
		   AND UPPER(TRIM(f.symbol)) = UPPER(TRIM(i.symbol))
		   AND UPPER(TRIM(f.side)) = UPPER(TRIM(i.side))
		 ORDER BY r.id`,
		string(StateConfirmed), ReservationHeld, string(StateConfirmed))
	if err != nil {
		return nil, fmt.Errorf("journal: finding exactly scoped terminal reservations: %w", err)
	}
	var candidates []terminalCandidate
	for rows.Next() {
		var candidate terminalCandidate
		if err := rows.Scan(&candidate.reservationID, &candidate.accountRef, &candidate.orderID,
			&candidate.market, &candidate.tradingDay, &candidate.symbol, &candidate.side,
			&candidate.intentOwners); err != nil {
			rows.Close()
			return nil, fmt.Errorf("journal: finding exactly scoped terminal reservations: %w", err)
		}
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("journal: finding exactly scoped terminal reservations: %w", err)
	}
	rows.Close()

	var byOrderState []ReservationRelease
	applyTx := &ApplyTx{tx: tx, now: nowText}
	defer applyTx.invalidate()
	for _, candidate := range candidates {
		if candidate.intentOwners != 1 {
			evidence := fmt.Sprintf(
				"startup reservation sweep found order %s owned by %d intents in canonical scope %s/%s/%s/%s/%s; reservations remain held",
				candidate.orderID, candidate.intentOwners, candidate.accountRef, candidate.market,
				candidate.tradingDay, candidate.symbol, candidate.side)
			if err := enterReconcileScopeInTx(ctx, applyTx, candidate.accountRef, "",
				ReconcileCauseIdentifierConflict, evidence, nowText); err != nil {
				return nil, err
			}
			continue
		}
		released, err := releaseWhere(ctx, tx, ReleaseReasonBrokerTerminal,
			"recovered at startup: the order was already observed in a derived terminal state",
			nowText, "id = ? AND state = ?", candidate.reservationID, ReservationHeld)
		if err != nil {
			return nil, err
		}
		byOrderState = append(byOrderState, released...)
	}
	return append(byAttemptState, byOrderState...), nil
}

// mergeAlerts appends the alerts that are not already reported, keyed by
// reservation id. A hold can qualify twice (an expiry whose nonce was spent on
// an attempt that is also unresolved) and an operator should see it once.
func mergeAlerts(existing, extra []ReservationAlert) []ReservationAlert {
	seen := make(map[string]bool, len(existing))
	for _, a := range existing {
		seen[a.Reservation.ID] = true
	}
	for _, a := range extra {
		if seen[a.Reservation.ID] {
			continue
		}
		seen[a.Reservation.ID] = true
		existing = append(existing, a)
	}
	return existing
}

// --- (c) the operator's exit -------------------------------------------------

// OperatorReleaseRequest is a human releasing a hold by hand.
type OperatorReleaseRequest struct {
	ReservationID string
	// Operator identifies the person. Required: this is a human overriding a
	// fail-closed judgement about a live account.
	Operator string
	// Reason is why the release is correct — "the order was cancelled at the
	// broker", not "unblocking".
	Reason string
	// Evidence is what they checked to establish it.
	Evidence string
	// Auditor writes the operator-visible record. Required: a release nobody
	// can find afterwards is indistinguishable from one that never happened.
	Auditor ReservationAuditor
}

// OperatorReleaseReservation is the only exit for a reservation held by an
// UNKNOWN_BROKER_STATE or UNRESOLVED_IN_DOUBT attempt.
//
// It refuses an already-released reservation rather than reporting success,
// because "release it again" usually means the operator is looking at a stale
// screen, and a silent no-op would confirm the wrong belief.
func (j *Journal) OperatorReleaseReservation(ctx context.Context, req OperatorReleaseRequest) (ReservationRelease, error) {
	id := strings.TrimSpace(req.ReservationID)
	operator := strings.TrimSpace(req.Operator)
	reason := strings.TrimSpace(req.Reason)
	evidence := strings.TrimSpace(req.Evidence)
	switch {
	case id == "":
		return ReservationRelease{}, fmt.Errorf("%w: which reservation?", ErrInvalidRequest)
	case operator == "":
		return ReservationRelease{}, fmt.Errorf(
			"%w: an operator release requires the operator's identity", ErrInvalidRequest)
	case reason == "":
		return ReservationRelease{}, fmt.Errorf(
			"%w: an operator release requires a reason; a fail-closed hold is not cleared by clicking", ErrInvalidRequest)
	case evidence == "":
		return ReservationRelease{}, fmt.Errorf(
			"%w: an operator release requires the evidence that was checked", ErrInvalidRequest)
	case req.Auditor == nil:
		return ReservationRelease{}, fmt.Errorf(
			"%w: an operator release requires an audit log; an unrecorded override is not an override anybody can review",
			ErrInvalidRequest)
	}

	now := j.nowString()
	detail := fmt.Sprintf("operator %s: %s (evidence: %s)", operator, reason, evidence)

	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return ReservationRelease{}, fmt.Errorf("journal: starting the operator release transaction: %w", err)
	}
	defer tx.Rollback()

	existing, err := scanReservation(tx.QueryRowContext(ctx, reservationSelect+" WHERE id = ?", id))
	if err != nil {
		return ReservationRelease{}, err
	}
	if !existing.Held() {
		return ReservationRelease{}, fmt.Errorf(
			"%w: reservation %s was already released at %s (%s)",
			ErrInvalidRequest, id, formatJournalTime(existing.ReleasedAt), existing.ReleaseReason)
	}

	released, err := releaseWhere(ctx, tx, ReleaseReasonOperator, detail, now,
		"id = ? AND state = ?", id, ReservationHeld)
	if err != nil {
		return ReservationRelease{}, err
	}
	if len(released) != 1 {
		return ReservationRelease{}, fmt.Errorf("%w: reservation %s", ErrReservationNotFound, id)
	}

	// The audit line is written before the commit: a released hold whose record
	// is missing is exactly the state this path exists to make impossible, and
	// a failed audit write must abort the release rather than outlive it.
	if err := req.Auditor.RecordAction(AuditActionReservationRelease, "reservation:"+id,
		released[0].Kind+" "+released[0].Amount+" "+released[0].Currency, detail); err != nil {
		return ReservationRelease{}, fmt.Errorf(
			"journal: recording the operator release of %s in the audit log (nothing was released): %w", id, err)
	}
	if err := tx.Commit(); err != nil {
		return ReservationRelease{}, fmt.Errorf("journal: committing the operator release of %s: %w", id, err)
	}
	return released[0], nil
}

// ReservationsAwaitingOperator lists every hold whose only exit is a human.
//
// Two conditions put a reservation here, and both mean the same thing: the
// engine cannot establish what happened to the order the hold was taken for.
//   - the attempt is parked as UNRESOLVED_IN_DOUBT;
//   - the last observation of its broker order failed closed, so nothing
//     derived it as terminal (the CLOSED + nothing filled + no cancellation
//     case the derivation refuses to read as an expiry).
func (j *Journal) ReservationsAwaitingOperator(ctx context.Context) ([]ReservationAlert, error) {
	rows, err := j.db.QueryContext(ctx,
		allFillSnapshotsCTE+` SELECT r.id, r.decision_id, r.attempt_id, r.account_ref, r.kind, r.amount,
		        r.currency, r.trading_day, r.snapshot_as_of, r.state, r.released_at,
		        r.release_reason, a.state, f.fail_closed, f.reason_code, f.detail
		 FROM risk_reservations r
		 JOIN mutation_attempts a ON a.id = r.attempt_id
		 JOIN intents i ON i.id = a.intent_id
		 LEFT JOIN all_fill_snapshots f
		   ON f.order_id = a.broker_order_id
			  AND a.state = 'CONFIRMED' AND a.kind IN ('PLACE','AMEND')
			  AND a.settled_at < f.committed_at
		  AND TRIM(f.account_ref) = TRIM(i.account_ref)
		  AND LOWER(TRIM(f.market)) = LOWER(TRIM(i.market))
		  AND TRIM(f.trading_day) = TRIM(i.trading_day)
		  AND UPPER(TRIM(f.symbol)) = UPPER(TRIM(i.symbol))
		  AND UPPER(TRIM(f.side)) = UPPER(TRIM(i.side))
		  AND 1 = (SELECT COUNT(DISTINCT owner.intent_id)
		             FROM mutation_attempts owner JOIN intents owned ON owned.id = owner.intent_id
			            WHERE owner.state = 'CONFIRMED' AND owner.kind IN ('PLACE','AMEND')
			              AND owner.settled_at < f.committed_at
		              AND owner.broker_order_id = f.order_id
		              AND TRIM(owned.account_ref) = TRIM(f.account_ref)
		              AND LOWER(TRIM(owned.market)) = LOWER(TRIM(f.market))
		              AND TRIM(owned.trading_day) = TRIM(f.trading_day)
		              AND UPPER(TRIM(owned.symbol)) = UPPER(TRIM(f.symbol))
		              AND UPPER(TRIM(owned.side)) = UPPER(TRIM(f.side)))
		 WHERE r.state = ? AND (a.state = ? OR f.fail_closed = 1)
		   AND r.decision_id = a.decision_id
		 ORDER BY r.rowid`,
		ReservationHeld, string(StateUnresolvedInDoubt))
	if err != nil {
		return nil, fmt.Errorf("journal: listing the reservations awaiting an operator: %w", err)
	}
	defer rows.Close()

	var out []ReservationAlert
	for rows.Next() {
		var (
			res                       Reservation
			attemptID, tradingDay     sql.NullString
			releasedAt, releaseReason sql.NullString
			asOf, attemptState        string
			failClosed                sql.NullInt64
			reasonCode, detail        sql.NullString
		)
		if err := rows.Scan(&res.ID, &res.DecisionID, &attemptID, &res.AccountRef, &res.Kind,
			&res.Amount, &res.Currency, &tradingDay, &asOf, &res.State, &releasedAt,
			&releaseReason, &attemptState, &failClosed, &reasonCode, &detail); err != nil {
			return nil, fmt.Errorf("journal: listing the reservations awaiting an operator: %w", err)
		}
		res.AttemptID = attemptID.String
		res.TradingDay = tradingDay.String
		res.ReleaseReason = releaseReason.String
		parsedAsOf, err := parseJournalTime(asOf)
		if err != nil {
			return nil, fmt.Errorf("journal: reservation %s snapshot_as_of: %w", res.ID, err)
		}
		res.SnapshotAsOf = parsedAsOf

		alert := ReservationAlert{
			Reservation:  res,
			AttemptID:    res.AttemptID,
			AttemptState: AttemptState(attemptState),
		}
		switch {
		case AttemptState(attemptState) == StateUnresolvedInDoubt:
			alert.Cause = AlertCauseUnresolved
			alert.Detail = fmt.Sprintf(
				"attempt %s is unresolved; only an operator can say what happened to it", res.AttemptID)
		default:
			alert.Cause = AlertCauseBrokerStateUnknown
			alert.Detail = fmt.Sprintf("the last observation of the order failed closed (%s): %s",
				reasonCode.String, detail.String)
		}
		out = append(out, alert)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing the reservations awaiting an operator: %w", err)
	}
	return out, nil
}
