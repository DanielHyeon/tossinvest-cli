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
}

// EnqueueAlert records an alert that must be delivered.
//
// Enqueuing the same event key twice is not an error and does not create a
// second row: the caller observing the same condition again is the normal case,
// and duplicating it would turn one problem into a pager storm. The existing
// row's id is returned so the caller can still drive its delivery.
func (j *Journal) EnqueueAlert(ctx context.Context, a Alert) (int64, error) {
	key := strings.TrimSpace(a.EventKey)
	if key == "" {
		return 0, errors.New("journal: an alert needs an event key, or it cannot be deduplicated")
	}
	if strings.TrimSpace(a.Type) == "" {
		return 0, errors.New("journal: an alert needs an event type")
	}
	now := RFC3339(j.clk.Now())

	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("journal: enqueueing alert %s: %w", key, err)
	}
	defer tx.Rollback()

	var existing int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM alert_outbox WHERE event_key = ?`, key).Scan(&existing)
	switch {
	case err == nil:
		return existing, tx.Commit()
	case !errors.Is(err, sql.ErrNoRows):
		return 0, fmt.Errorf("journal: looking up alert %s: %w", key, err)
	}

	res, err := tx.ExecContext(ctx,
		`INSERT INTO alert_outbox(event_key, event_type, severity, title, body, payload, state, created_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		key, a.Type, a.Severity, a.Title, a.Body, a.Payload, AlertPending, now)
	if err != nil {
		return 0, fmt.Errorf("journal: enqueueing alert %s: %w", key, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("journal: reading the id of alert %s: %w", key, err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("journal: committing alert %s: %w", key, err)
	}
	return id, nil
}

// MarkAlertDelivered records a successful send.
func (j *Journal) MarkAlertDelivered(ctx context.Context, id int64) error {
	now := RFC3339(j.clk.Now())
	res, err := j.db.ExecContext(ctx,
		`UPDATE alert_outbox
		    SET state = ?, delivered_at = ?, last_attempt_at = ?, attempts = attempts + 1, last_error = ''
		  WHERE id = ? AND state = ?`,
		AlertDelivered, now, now, id, AlertPending)
	if err != nil {
		return fmt.Errorf("journal: marking alert %d delivered: %w", id, err)
	}
	return requireOneRow(res, id)
}

// MarkAlertAttemptFailed records a failed send. The row stays PENDING: a critical
// alert is not discarded because the network was down.
func (j *Journal) MarkAlertAttemptFailed(ctx context.Context, id int64, cause string) error {
	now := RFC3339(j.clk.Now())
	res, err := j.db.ExecContext(ctx,
		`UPDATE alert_outbox
		    SET attempts = attempts + 1, last_attempt_at = ?, last_error = ?
		  WHERE id = ? AND state = ?`,
		now, cause, id, AlertPending)
	if err != nil {
		return fmt.Errorf("journal: recording a failed alert attempt for %d: %w", id, err)
	}
	return requireOneRow(res, id)
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
	res, err := j.db.ExecContext(ctx,
		`UPDATE alert_outbox
		    SET state = ?, acknowledged_at = ?, acknowledged_by = ?
		  WHERE id = ? AND state = ?`,
		AlertAcknowledged, now, strings.TrimSpace(operator), id, AlertPending)
	if err != nil {
		return fmt.Errorf("journal: acknowledging alert %d: %w", id, err)
	}
	return requireOneRow(res, id)
}

const alertSelect = `SELECT id, event_key, event_type, severity, title, body, payload, state,
       attempts, created_at, last_attempt_at, delivered_at, acknowledged_at, acknowledged_by, last_error
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
			a          Alert
			created    string
			lastTry    sql.NullString
			delivered  sql.NullString
			acked      sql.NullString
			ackedBy    string
			lastErrStr string
		)
		if err := rows.Scan(&a.ID, &a.EventKey, &a.Type, &a.Severity, &a.Title, &a.Body,
			&a.Payload, &a.State, &a.Attempts, &created, &lastTry, &delivered, &acked,
			&ackedBy, &lastErrStr); err != nil {
			return nil, fmt.Errorf("journal: scanning an alert: %w", err)
		}
		a.CreatedAt = parseStamp(created)
		a.LastAttemptAt = parseOptionalStamp(lastTry)
		a.DeliveredAt = parseOptionalStamp(delivered)
		a.AcknowledgedAt = parseOptionalStamp(acked)
		a.AcknowledgedBy = ackedBy
		a.LastError = lastErrStr
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
