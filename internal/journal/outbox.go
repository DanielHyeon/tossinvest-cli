package journal

// outbox.go is the durable alert outbox (harden-execution-base task 4.3,
// engine-safety "등급화된 알림").
//
// # Why an outbox and not just "send the notification"
//
// The events that matter most are the ones raised while something is already
// wrong: an IN_DOUBT mutation, a broker state we cannot read, a reconciliation
// that will not settle. Those are exactly the moments when the network is also
// likely to be misbehaving — and an alert that is lost because the send failed is
// an incident nobody is told about.
//
// So a critical alert is written here, durably, *before* anyone tries to send it,
// and it is not marked delivered until a send actually succeeds. The engine can
// then answer a question it otherwise could not: "is there something I decided
// you needed to know that you have not been told?" If the answer is yes, new
// entries stop (spec: "전달 실패가 지속되면 신규 진입을 차단한다"), because
// trading on while the operator is deaf is the failure mode the grading exists to
// prevent.
//
// # Why delivery failure does not delete the row
//
// A row that exhausted its retries stays PENDING. It is not abandoned, not
// dropped and not summarised — the spec requires it to be "outbox에 미전달
// 상태로 보존", and release is an explicit operator acknowledgement, because "the
// network came back" is not evidence that anybody read the alert.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// schemaV3 adds the alert outbox. Additive, per the rules in schema.go: no
// existing table or column changes, and an older binary refuses the database
// rather than misreading it.
const schemaV3 = `
-- Durable queue of alerts that must not be lost. Critical events are written
-- here before any delivery attempt; ordinary ones never reach this table at all
-- (they are best-effort by definition).
CREATE TABLE alert_outbox (
	id              INTEGER PRIMARY KEY AUTOINCREMENT,
	-- event_key deduplicates: the same condition observed on three consecutive
	-- polls is one alert, not three. Callers build it from the condition, not
	-- from the clock.
	event_key       TEXT NOT NULL UNIQUE,
	event_type      TEXT NOT NULL,
	severity        TEXT NOT NULL,
	title           TEXT NOT NULL DEFAULT '',
	body            TEXT NOT NULL DEFAULT '',
	-- payload is JSON context for the operator (symbol, attempt id, reason code).
	payload         TEXT NOT NULL DEFAULT '',
	state           TEXT NOT NULL,            -- 'PENDING' | 'DELIVERED' | 'ACKNOWLEDGED'
	attempts        INTEGER NOT NULL DEFAULT 0,
	created_at      TEXT NOT NULL,
	last_attempt_at TEXT,
	delivered_at    TEXT,
	acknowledged_at TEXT,
	acknowledged_by TEXT NOT NULL DEFAULT '',
	last_error      TEXT NOT NULL DEFAULT ''
) STRICT;

CREATE INDEX idx_outbox_state ON alert_outbox(state, id);
`

// Alert states.
const (
	// AlertPending is written but not yet delivered. It stays PENDING however
	// many times delivery fails.
	AlertPending = "PENDING"
	// AlertDelivered means a send succeeded.
	AlertDelivered = "DELIVERED"
	// AlertAcknowledged means an operator confirmed they saw it. Only an
	// acknowledged (or delivered) outbox permits new entries again.
	AlertAcknowledged = "ACKNOWLEDGED"
)

// ErrAlertNotFound means no outbox row has that id.
var ErrAlertNotFound = errors.New("journal: no such alert")

// Alert is one queued notification.
type Alert struct {
	ID       int64
	EventKey string
	Type     string
	Severity string
	Title    string
	Body     string
	Payload  string
	State    string
	Attempts int

	CreatedAt      time.Time
	LastAttemptAt  *time.Time
	DeliveredAt    *time.Time
	AcknowledgedAt *time.Time
	AcknowledgedBy string
	LastError      string

	// The lease, projected for the operator: an undelivered row that nobody is
	// sending and one that a sender is holding look identical without it, and the
	// operator triaging a backlog needs to tell "stuck" from "in flight".
	//
	// The token is deliberately absent. It is the proof of ownership, and a
	// listing that showed it would let anyone who read the screen settle
	// somebody else's send.
	ClaimedBy      string
	ClaimedAt      *time.Time
	ClaimExpiresAt *time.Time
}

// EnqueueAlert records an alert that must be delivered.
//
// Enqueuing the same event key twice is not an error and does not create a
// second row: the caller observing the same condition again is the normal case,
// and duplicating it would turn one problem into a pager storm. The existing
// row's id is returned so the caller can still drive its delivery.
//
// Use it when the alert only has to be *recorded*. A caller that is about to
// send needs ClaimAlertForDelivery instead, because the id alone does not say
// whether that send is still owed.
//
// It takes no delivery lease, and that asymmetry is deliberate. A recorder that
// claimed would leave every row it wrote behind a lease nobody will ever
// release — it does not deliver, so it never settles — and the row would sit
// unsendable until the lease expired. replay.go writes rows through here.
func (j *Journal) EnqueueAlert(ctx context.Context, a Alert) (int64, error) {
	key, err := alertKey(a)
	if err != nil {
		return 0, err
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("journal: enqueueing alert %s: %w", key, err)
	}
	defer tx.Rollback()

	// Zero remindAfter: this caller does not deliver, so it has no reminder
	// policy to express and must not re-arm a settled row on somebody else's
	// behalf. recordAlertTx may still recover an unrecognised state to PENDING;
	// that is fail-safe state repair, not a timed reminder.
	id, _, err := j.recordAlertTx(ctx, tx, a, 0)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("journal: committing alert %s: %w", key, err)
	}
	return id, nil
}

// alertKey validates an alert and returns the key it deduplicates on. It exists
// so the two entry points can refuse a malformed alert before opening a write
// transaction — this journal's transactions take the write lock up front, and
// waiting for it only to reject the input would be a real wait.
func alertKey(a Alert) (string, error) {
	key := strings.TrimSpace(a.EventKey)
	if key == "" {
		return "", errors.New("journal: an alert needs an event key, or it cannot be deduplicated")
	}
	if strings.TrimSpace(a.Type) == "" {
		return "", errors.New("journal: an alert needs an event type")
	}
	return key, nil
}

// ClaimAlertForDelivery records an alert and reports whether a send is owed. It
// deduplicates exactly as EnqueueAlert does — same row, same id.
//
// The second return is the point. Deduplicating the row while sending on every
// observation is not "one condition, one alert": the row folds and the
// condition does not, so a caller holding only an id arrives with the same id
// every cycle and sends every time. Then the deduplication is what *sustains*
// the storm rather than preventing it. On 2026-08-08 that was one outbox row
// and sixty pushes, one every 5.6 seconds, for as long as the market stayed
// shut (a096).
//
// # Owed
//
// True for a row this call inserted, and for any row still PENDING. A send that
// failed is unfinished rather than duplicate, so it keeps its retry budget and
// the entry block that follows a spent one.
//
// True again for a settled row — DELIVERED, or ACKNOWLEDGED by an operator —
// once remindAfter has elapsed since it settled. Such a row is *re-armed* to
// PENDING here, in this transaction, so everything downstream is the ordinary
// first-delivery path: the retry budget, the gate latch and the operating-mode
// escalation all apply to the reminder exactly as they applied to the original.
//
// Permanent suppression was the first design and it was wrong twice. It removed
// the only thing that periodically proved the transport still worked, so an
// engine could trade on with a dead alert channel; and because an event key
// carries the condition and not its cause — `exit.proposal_refused` is keyed by
// position, action and rung, never by the refusal — it silenced a later, unlike
// occurrence on the strength of an earlier one. A window costs one push per
// interval per condition and returns both.
//
// remindAfter <= 0 disables re-arming: a settled row is never owed. That is for
// callers that record without delivering and therefore hold no reminder policy.
//
// # Unknown states
//
// A state this build does not recognise is treated as owed. The column has no
// CHECK constraint, so an unrecognised value is a row written by something this
// code does not understand, and the safe reading of "I do not know whether the
// operator got this" is to send.
//
// The state is read inside the transaction that resolves the id, so the two are
// one instant's answer and nothing can settle between them. The lease is taken
// in that same transaction, which is what makes "is a send owed" and "is this
// send mine" one answer instead of two.
//
// # Exclusion
//
// Until a099 this function claimed nothing. It read the state and told every
// caller that a send was owed, and the only exclusion anywhere was
// obs.Notifier's delivery mutex — which excludes senders inside one Notifier and
// nothing whatsoever between two. Two senders reaching one PENDING row both left
// believing the send was theirs, and both published.
//
// The row now carries the lease and the ledger decides. What that buys and what
// it does not is written down on ClaimDisposition and in ClaimResult: a
// settlement can never land twice, a *publish* still can if a sender stalls past
// its own lease, and no expiry can call back a push already in flight. The
// ledger's guarantee is about who settles, not about how many times a phone
// buzzes.
func (j *Journal) ClaimAlertForDelivery(
	ctx context.Context, a Alert, remindAfter time.Duration, claimant string,
) (ClaimResult, error) {
	key, err := alertKey(a)
	if err != nil {
		return ClaimResult{}, err
	}
	if strings.TrimSpace(claimant) == "" {
		return ClaimResult{}, errors.New("journal: claiming an alert for delivery requires the claimant's name")
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return ClaimResult{}, fmt.Errorf("journal: enqueueing alert %s: %w", key, err)
	}
	defer tx.Rollback()

	id, owed, err := j.recordAlertTx(ctx, tx, a, remindAfter)
	if err != nil {
		return ClaimResult{}, err
	}
	if !owed {
		if err := tx.Commit(); err != nil {
			return ClaimResult{}, fmt.Errorf("journal: committing alert %s: %w", key, err)
		}
		// Nothing is owed, so nothing is claimed. Taking a lease here would lock
		// a row that needs no sending.
		return ClaimResult{Disposition: ClaimSettled, ID: id}, nil
	}
	out, err := acquireAlertClaimTx(ctx, tx, id, claimant, j.clk.Now(), j.alertLease)
	if err != nil {
		return ClaimResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return ClaimResult{}, fmt.Errorf("journal: committing the claim on alert %s: %w", key, err)
	}
	return out, nil
}

// recordAlertTx is the recording half of the outbox: deduplicate on the event
// key, insert or re-arm, and say whether a send is owed. It takes no lease.
//
// It is a separate function rather than a flag on the claiming one because the
// difference between the two callers is whether a lease follows, and a flag
// would have made "record without claiming" a mode — the mode that mattered
// being the one a caller could forget to pass.
func (j *Journal) recordAlertTx(
	ctx context.Context, tx *sql.Tx, a Alert, remindAfter time.Duration,
) (int64, bool, error) {
	key, err := alertKey(a)
	if err != nil {
		return 0, false, err
	}
	nowAt := j.clk.Now()
	now := RFC3339(nowAt)

	var existing int64
	var state string
	var deliveredAt, acknowledgedAt sql.NullString
	err = tx.QueryRowContext(ctx,
		`SELECT id, state, delivered_at, acknowledged_at FROM alert_outbox WHERE event_key = ?`,
		key).Scan(&existing, &state, &deliveredAt, &acknowledgedAt)
	switch {
	case err == nil:
		owed, rearm := claimOwed(state, deliveredAt, acknowledgedAt, nowAt, remindAfter)
		if rearm {
			// Back to PENDING so the reminder walks the same path the original
			// did — and every other column with it, because re-arming is the
			// statement that this row is now a *different episode* of the same
			// condition.
			//
			// The acknowledgement cannot stay. acknowledged_by exists to record
			// that a named human saw this alert, and this is a new episode they
			// have not seen; leaving it produces a row that reads "undelivered"
			// and "acknowledged by daniel" at the same time. An operator triaging
			// a backlog after an incident skips a live undelivered critical alert
			// on the strength of their own name on it (a096 round 2).
			//
			// The content cannot stay either, for the same reason one step out.
			// An event key holds the condition and not its cause —
			// exit.proposal_refused is keyed by position, action and rung, never
			// by the refusal — so the second episode of one key is routinely a
			// different reason. Flush builds what it sends from this row rather
			// than from the live event, and an operator reading the outbox after
			// an incident reads this row (a097).
			//
			// attempts, last_attempt_at and delivered_at go with it. a097's first
			// draft kept delivered_at, calling it the record that the previous
			// episode really was delivered; that was wrong. Once the body is
			// overwritten, a surviving delivered_at makes the row claim that *this*
			// content was delivered then, and nothing anywhere holds the content
			// that actually was. A row is evidence only when every column points
			// at one event. What happened before belongs to the structured log;
			// this table is a delivery queue.
			//
			// None of it is load-bearing for the decision above: latestStamp is
			// reached only from the settled case, and this row is PENDING now.
			//
			// The lease goes with the rest of it, and for exactly the same reason.
			// A new episode is a new send; whoever was holding the previous one is
			// not sending this. Leaving the four columns behind would hand the new
			// episode to a sender that claimed the old one — possibly one that has
			// since died — and no rule anywhere says the lease is the exception to
			// "every column points at one event".
			if _, uerr := tx.ExecContext(ctx,
				`UPDATE alert_outbox
				    SET state = ?, title = ?, body = ?, payload = ?,
				        attempts = 0, last_error = '',
				        last_attempt_at = NULL, delivered_at = NULL,
				        acknowledged_at = NULL, acknowledged_by = '',
				        `+alertClaimCleared+`
				  WHERE id = ?`,
				AlertPending, a.Title, a.Body, a.Payload, existing); uerr != nil {
				return 0, false, fmt.Errorf("journal: re-arming alert %s: %w", key, uerr)
			}
		}
		return existing, owed, nil
	case !errors.Is(err, sql.ErrNoRows):
		return 0, false, fmt.Errorf("journal: looking up alert %s: %w", key, err)
	}

	res, err := tx.ExecContext(ctx,
		`INSERT INTO alert_outbox(event_key, event_type, severity, title, body, payload, state, created_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		key, a.Type, a.Severity, a.Title, a.Body, a.Payload, AlertPending, now)
	if err != nil {
		return 0, false, fmt.Errorf("journal: enqueueing alert %s: %w", key, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, false, fmt.Errorf("journal: reading the id of alert %s: %w", key, err)
	}
	// The INSERT above wrote AlertPending. Nothing has been sent.
	return id, true, nil
}

// claimOwed decides whether a send is owed for an existing row, and whether the
// row has to be re-armed to PENDING first.
//
// It is split out because it is the whole of a096's judgement and nothing else
// in the transaction is: given a state and when that state was reached, is the
// operator still owed this alert?
func claimOwed(
	state string,
	deliveredAt, acknowledgedAt sql.NullString,
	now time.Time,
	remindAfter time.Duration,
) (owed bool, rearm bool) {
	switch state {
	case AlertPending:
		// Never sent, or sent and failed. Either way it is unfinished.
		return true, false
	case AlertDelivered, AlertAcknowledged:
		if remindAfter <= 0 {
			return false, false
		}
		settled, ok := latestStamp(deliveredAt, acknowledgedAt)
		if !ok {
			// Settled but with no timestamp to measure from. That is a row this
			// code cannot date, so it cannot claim the window has not elapsed.
			return true, true
		}
		elapsed := now.Sub(settled)
		if elapsed < 0 {
			// Settled in the future, which means the clock moved backwards after
			// the delivery landed — an NTP step-back, a restored snapshot, an RTC
			// that was ahead at boot. Suppressing here would be the worst of the
			// two errors: every negative duration is less than the window, so the
			// key stays silent for the whole skew *plus* the window, nothing
			// publishes, and therefore neither the gate latch nor the escalation
			// ever runs.
			//
			// This row has exactly the epistemic standing of the undateable row
			// three lines above, and that one fails open. So does this one
			// (a096 round 2).
			return true, true
		}
		if elapsed < remindAfter {
			return false, false
		}
		return true, true
	default:
		// See the doc comment: an unrecognised state is not evidence of delivery.
		// Re-arm it as well as declaring the send owed. Otherwise Publish succeeds
		// but MarkAlertDelivered cannot move the row out of its unknown state, and
		// every later observation publishes it again.
		return true, true
	}
}

// latestStamp returns the most recent of the parsable timestamps.
func latestStamp(values ...sql.NullString) (time.Time, bool) {
	var latest time.Time
	var found bool
	for _, v := range values {
		if !v.Valid || strings.TrimSpace(v.String) == "" {
			continue
		}
		ts, err := time.Parse(time.RFC3339, v.String)
		if err != nil {
			continue
		}
		if !found || ts.After(latest) {
			latest, found = ts, true
		}
	}
	return latest, found
}

// MarkAlertDelivered records a successful send and hands the lease back.
//
// The token is the sender's proof that this send was its own. Without it the id
// and the state alone let a sender that lost its lease — because it stalled long
// enough for the claim to expire and somebody else to take the row — settle the
// send the new holder is in the middle of making.
//
// Releasing here is safe in a way that releasing after a failed settlement is
// not: the row leaves PENDING in the same statement, so nothing will look at it
// again.
func (j *Journal) MarkAlertDelivered(ctx context.Context, id int64, token string) (SettleResult, error) {
	now := RFC3339(j.clk.Now())
	return j.settleUnderClaim(ctx, id, token, "marking delivered",
		`UPDATE alert_outbox
		    SET state = ?, delivered_at = ?, last_attempt_at = ?, attempts = attempts + 1, last_error = '',
		        `+alertClaimCleared+`
		  WHERE id = ? AND state = ? AND claim_token = ?`,
		[]any{AlertDelivered, now, now, id, AlertPending, token})
}

// MarkAlertAttemptFailed records a failed send. The row stays PENDING: a critical
// alert is not discarded because the network was down.
//
// It keeps the lease, and that is the difference between recording a failure and
// giving up. The sender still has retries left and is about to use them; letting
// go here would open the row to a second sender in the gap between two attempts
// of the first, which is the double delivery this change exists to prevent.
// Giving up is ReleaseAlertClaim, and it is a separate act.
func (j *Journal) MarkAlertAttemptFailed(ctx context.Context, id int64, token, cause string) (SettleResult, error) {
	now := RFC3339(j.clk.Now())
	return j.settleUnderClaim(ctx, id, token, "recording a failed attempt on",
		`UPDATE alert_outbox
		    SET attempts = attempts + 1, last_attempt_at = ?, last_error = ?
		  WHERE id = ? AND state = ? AND claim_token = ?`,
		[]any{now, cause, id, AlertPending, token})
}

// AcknowledgeAlert is the operator's release.
//
// It requires an identity for the same reason the journal's operator resolution
// does: this is a human asserting they have seen something the machine could not
// prove was seen, and the audit trail is the point.
func (j *Journal) AcknowledgeAlert(ctx context.Context, id int64, operator string) error {
	if strings.TrimSpace(operator) == "" {
		return errors.New("journal: acknowledging an alert requires the operator's identity")
	}
	now := RFC3339(j.clk.Now())
	// 임차도 같이 푼다. 승인은 PENDING 을 떠나는 네 경로 중 하나이고, 나머지 셋
	// (배달·반납·재무장)은 전부 alertClaimCleared 를 쓴다. 이것만 빠져 있었다.
	//
	// 배제가 깨져서가 아니다 — 임차를 읽는 술어는 전부 state = PENDING 안에서만
	// 돈다. 토큰 때문이다. 토큰은 이 원장을 떠나면 안 되는 유일한 열인데
	// (alertSelect 의 주석), 정산된 행이 그것을 계속 들고 있으면 backup.go 의
	// VACUUM INTO 가 뜨는 복사본마다 그 토큰이 따라 나간다. 승인된 행은 아무도
	// 발송 중이 아니므로 들고 있을 이유가 없다.
	res, err := j.db.ExecContext(ctx,
		`UPDATE alert_outbox
		    SET state = ?, acknowledged_at = ?, acknowledged_by = ?,
		        `+alertClaimCleared+`
		  WHERE id = ? AND state = ?`,
		AlertAcknowledged, now, strings.TrimSpace(operator), id, AlertPending)
	if err != nil {
		return fmt.Errorf("journal: acknowledging alert %d: %w", id, err)
	}
	return requireOneRow(res, id)
}

// alertSelect projects the lease alongside the row, minus the token: the
// operator's backlog has to distinguish a row nobody is sending from one a
// sender is holding, and the token is the one column that must never leave the
// ledger.
const alertSelect = `SELECT id, event_key, event_type, severity, title, body, payload, state,
       attempts, created_at, last_attempt_at, delivered_at, acknowledged_at, acknowledged_by, last_error,
       claimed_by, claimed_at, claim_expires_at
  FROM alert_outbox`

// PendingAlerts lists undelivered alerts, oldest first. A limit of zero means all
// of them.
func (j *Journal) PendingAlerts(ctx context.Context, limit int) ([]Alert, error) {
	query := alertSelect + ` WHERE state = ? ORDER BY id`
	args := []any{AlertPending}
	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	}
	rows, err := j.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("journal: listing pending alerts: %w", err)
	}
	defer rows.Close()
	return scanAlerts(rows)
}

// UndeliveredCount is the number the entry gate reacts to.
func (j *Journal) UndeliveredCount(ctx context.Context) (int, error) {
	var n int
	if err := j.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM alert_outbox WHERE state = ?`, AlertPending).Scan(&n); err != nil {
		return 0, fmt.Errorf("journal: counting undelivered alerts: %w", err)
	}
	return n, nil
}

// LookupAlert reads one row.
func (j *Journal) LookupAlert(ctx context.Context, id int64) (Alert, error) {
	rows, err := j.db.QueryContext(ctx, alertSelect+` WHERE id = ?`, id)
	if err != nil {
		return Alert{}, fmt.Errorf("journal: reading alert %d: %w", id, err)
	}
	defer rows.Close()
	alerts, err := scanAlerts(rows)
	if err != nil {
		return Alert{}, err
	}
	if len(alerts) == 0 {
		return Alert{}, fmt.Errorf("%w: %d", ErrAlertNotFound, id)
	}
	return alerts[0], nil
}

func scanAlerts(rows *sql.Rows) ([]Alert, error) {
	var out []Alert
	for rows.Next() {
		var (
			a           Alert
			created     string
			lastTry     sql.NullString
			delivered   sql.NullString
			acked       sql.NullString
			ackedBy     string
			lastErrStr  string
			claimedBy   string
			claimedAt   sql.NullString
			claimExpiry sql.NullString
		)
		if err := rows.Scan(&a.ID, &a.EventKey, &a.Type, &a.Severity, &a.Title, &a.Body,
			&a.Payload, &a.State, &a.Attempts, &created, &lastTry, &delivered, &acked,
			&ackedBy, &lastErrStr, &claimedBy, &claimedAt, &claimExpiry); err != nil {
			return nil, fmt.Errorf("journal: scanning an alert: %w", err)
		}
		a.CreatedAt = parseStamp(created)
		a.LastAttemptAt = parseOptionalStamp(lastTry)
		a.DeliveredAt = parseOptionalStamp(delivered)
		a.AcknowledgedAt = parseOptionalStamp(acked)
		a.AcknowledgedBy = ackedBy
		a.LastError = lastErrStr
		a.ClaimedBy = claimedBy
		a.ClaimedAt = parseOptionalStamp(claimedAt)
		a.ClaimExpiresAt = parseOptionalStamp(claimExpiry)
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading alerts: %w", err)
	}
	return out, nil
}

func requireOneRow(res sql.Result, id int64) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("journal: alert %d: %w", id, err)
	}
	if n == 0 {
		return fmt.Errorf("%w: %d (or it is no longer pending)", ErrAlertNotFound, id)
	}
	return nil
}

func parseStamp(s string) time.Time {
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return ts.UTC()
}

func parseOptionalStamp(s sql.NullString) *time.Time {
	if !s.Valid || strings.TrimSpace(s.String) == "" {
		return nil
	}
	ts := parseStamp(s.String)
	if ts.IsZero() {
		return nil
	}
	return &ts
}
