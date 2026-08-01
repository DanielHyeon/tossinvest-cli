package protection

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const SchemaVersion = 3

const schemaDDL = `
CREATE TABLE IF NOT EXISTS protection_sagas (
  saga_id TEXT PRIMARY KEY,
  account_ref TEXT NOT NULL,
  profile TEXT NOT NULL,
  market TEXT NOT NULL,
  symbol TEXT NOT NULL,
  generation INTEGER NOT NULL CHECK (generation >= 1),
  revision INTEGER NOT NULL CHECK (revision >= 1),
  state TEXT NOT NULL,
  trigger INTEGER NOT NULL CHECK (trigger >= 1),
  quantity INTEGER NOT NULL CHECK (quantity >= 1),
  pending_trigger INTEGER NOT NULL DEFAULT 0,
  pending_quantity INTEGER NOT NULL DEFAULT 0,
  client_order_id TEXT NOT NULL UNIQUE,
  attempt_id TEXT NOT NULL DEFAULT '',
  broker_id TEXT NOT NULL DEFAULT '',
  previous_broker_id TEXT NOT NULL DEFAULT '',
  reconcile_reason TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL,
  last_event_kind TEXT NOT NULL DEFAULT '',
  last_event_fingerprint TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS protection_sagas_account_symbol
  ON protection_sagas(account_ref, symbol, state);
CREATE UNIQUE INDEX IF NOT EXISTS protection_sagas_live_claim
  ON protection_sagas(account_ref,profile,market,symbol)
  WHERE state NOT IN ('TRIGGERED','CLOSED');
CREATE TABLE IF NOT EXISTS protection_mutation_attempts (
  attempt_id TEXT PRIMARY KEY,
  saga_id TEXT NOT NULL REFERENCES protection_sagas(saga_id),
  generation INTEGER NOT NULL CHECK (generation >= 1),
  kind TEXT NOT NULL CHECK (kind IN ('CREATE','REPLACE','CANCEL')),
  state TEXT NOT NULL CHECK (state IN ('PLANNED','DISPATCHED','ACKNOWLEDGED','IN_DOUBT','CLOSED')),
  serializer_version INTEGER NOT NULL,
  canonical_body TEXT NOT NULL,
  target_broker_id TEXT NOT NULL DEFAULT '',
  result_broker_id TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS protection_attempts_saga
  ON protection_mutation_attempts(saga_id, generation, created_at);
`

type MutationKind string

const (
	MutationCreate  MutationKind = "CREATE"
	MutationReplace MutationKind = "REPLACE"
	MutationCancel  MutationKind = "CANCEL"
)

type MutationState string

const (
	MutationPlanned      MutationState = "PLANNED"
	MutationDispatched   MutationState = "DISPATCHED"
	MutationAcknowledged MutationState = "ACKNOWLEDGED"
	MutationInDoubt      MutationState = "IN_DOUBT"
	MutationClosed       MutationState = "CLOSED"
)

type MutationAttempt struct {
	ID                string
	SagaID            string
	Generation        int64
	Kind              MutationKind
	State             MutationState
	SerializerVersion int
	CanonicalBody     string
	TargetBrokerID    string
	ResultBrokerID    string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) (*Repository, error) {
	if db == nil {
		return nil, errors.New("protection: nil database")
	}
	if _, err := db.Exec(schemaDDL); err != nil {
		return nil, fmt.Errorf("protection: creating additive schema v%d: %w", SchemaVersion, err)
	}
	if err := ensureEventIdentityColumns(db); err != nil {
		return nil, fmt.Errorf("protection: migrating additive schema v%d: %w", SchemaVersion, err)
	}
	return &Repository{db: db}, nil
}

func ensureEventIdentityColumns(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info(protection_sagas)`)
	if err != nil {
		return err
	}
	columns := map[string]bool{}
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			_ = rows.Close()
			return err
		}
		columns[name] = true
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, migration := range []struct {
		name string
		sql  string
	}{
		{"last_event_kind", `ALTER TABLE protection_sagas ADD COLUMN last_event_kind TEXT NOT NULL DEFAULT ''`},
		{"last_event_fingerprint", `ALTER TABLE protection_sagas ADD COLUMN last_event_fingerprint TEXT NOT NULL DEFAULT ''`},
	} {
		if !columns[migration.name] {
			if _, err := db.Exec(migration.sql); err != nil {
				return err
			}
		}
	}
	return validateStoredSagaRows(db)
}

func validateStoredSagaRows(db *sql.DB) error {
	rows, err := db.Query(`SELECT
 saga_id,account_ref,profile,market,symbol,generation,revision,state,trigger,quantity,
 pending_trigger,pending_quantity,client_order_id,attempt_id,broker_id,previous_broker_id,reconcile_reason,updated_at,
 last_event_kind,last_event_fingerprint
 FROM protection_sagas`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if _, err := scanStoredSaga(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (r *Repository) Insert(ctx context.Context, saga Saga) error {
	if saga.Revision == 0 {
		saga.Revision = 1
	}
	if saga.State != StatePlanned || saga.Revision != 1 {
		return fmt.Errorf("%w: repository insert requires revision-1 PLANNED saga", ErrInvalidTransition)
	}
	if err := saga.Validate(); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO protection_sagas (
 saga_id,account_ref,profile,market,symbol,generation,revision,state,trigger,quantity,
 pending_trigger,pending_quantity,client_order_id,attempt_id,broker_id,previous_broker_id,reconcile_reason,updated_at,
 last_event_kind,last_event_fingerprint
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, append(sagaValues(saga), "", "")...)
	if err != nil {
		return fmt.Errorf("protection: insert saga %s: %w", saga.ID, err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, id string) (Saga, error) {
	stored, err := r.getStored(ctx, id)
	return stored.Saga, err
}

func (r *Repository) recordAttempt(ctx context.Context, attempt MutationAttempt) error {
	if err := attempt.validate(); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO protection_mutation_attempts
 (attempt_id,saga_id,generation,kind,state,serializer_version,canonical_body,target_broker_id,result_broker_id,created_at,updated_at)
 VALUES (?,?,?,?,?,?,?,?,?,?,?)`, attempt.ID, attempt.SagaID, attempt.Generation, attempt.Kind, attempt.State,
		attempt.SerializerVersion, attempt.CanonicalBody, attempt.TargetBrokerID, attempt.ResultBrokerID,
		attempt.CreatedAt.UTC().Format(time.RFC3339Nano), attempt.UpdatedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("protection: record mutation attempt %s: %w", attempt.ID, err)
	}
	return nil
}

func (r *Repository) markAttempt(ctx context.Context, id string, from, to MutationState, at time.Time, resultBrokerID string) error {
	if id == "" || !validMutationState(from) || !validMutationState(to) || at.IsZero() || !validMutationStep(from, to) {
		return fmt.Errorf("protection: invalid mutation attempt transition %s -> %s", from, to)
	}
	result, err := r.db.ExecContext(ctx, `UPDATE protection_mutation_attempts
 SET state=?,result_broker_id=?,updated_at=? WHERE attempt_id=? AND state=?`,
		to, resultBrokerID, at.UTC().Format(time.RFC3339Nano), id, from)
	if err != nil {
		return fmt.Errorf("protection: mark mutation attempt %s: %w", id, err)
	}
	n, err := result.RowsAffected()
	if err != nil || n != 1 {
		return fmt.Errorf("%w: mutation attempt %s state %s", ErrConcurrentUpdate, id, from)
	}
	return nil
}

func (r *Repository) Attempts(ctx context.Context, sagaID string) ([]MutationAttempt, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT attempt_id,saga_id,generation,kind,state,serializer_version,
 canonical_body,target_broker_id,result_broker_id,created_at,updated_at
	FROM protection_mutation_attempts WHERE saga_id=? ORDER BY rowid`, sagaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MutationAttempt
	for rows.Next() {
		var a MutationAttempt
		var created, updated string
		if err := rows.Scan(&a.ID, &a.SagaID, &a.Generation, &a.Kind, &a.State, &a.SerializerVersion,
			&a.CanonicalBody, &a.TargetBrokerID, &a.ResultBrokerID, &created, &updated); err != nil {
			return nil, err
		}
		a.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
		if err != nil {
			return nil, err
		}
		a.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
		if err != nil {
			return nil, err
		}
		if err := a.validate(); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (a MutationAttempt) validate() error {
	if strings.TrimSpace(a.ID) == "" || strings.TrimSpace(a.SagaID) == "" || a.Generation < 1 || !validMutationKind(a.Kind) || !validMutationState(a.State) || a.SerializerVersion != SerializerVersion || strings.TrimSpace(a.CanonicalBody) == "" || a.CreatedAt.IsZero() || a.UpdatedAt.Before(a.CreatedAt) {
		return fmt.Errorf("protection: invalid mutation attempt")
	}
	if a.Kind == MutationCreate && a.TargetBrokerID != "" {
		return fmt.Errorf("protection: create attempt carries target broker id")
	}
	if a.Kind != MutationCreate && strings.TrimSpace(a.TargetBrokerID) == "" {
		return fmt.Errorf("protection: %s attempt lacks target broker id", a.Kind)
	}
	return nil
}

func validMutationKind(kind MutationKind) bool {
	return kind == MutationCreate || kind == MutationReplace || kind == MutationCancel
}

func validMutationState(state MutationState) bool {
	switch state {
	case MutationPlanned, MutationDispatched, MutationAcknowledged, MutationInDoubt, MutationClosed:
		return true
	default:
		return false
	}
}

func validMutationStep(from, to MutationState) bool {
	return (from == MutationPlanned && to == MutationDispatched) ||
		(from == MutationDispatched && (to == MutationAcknowledged || to == MutationInDoubt || to == MutationClosed)) ||
		(from == MutationInDoubt && (to == MutationAcknowledged || to == MutationClosed))
}

type storedSaga struct {
	Saga
	lastEventKind        EventKind
	lastEventFingerprint string
}

func (r *Repository) getStored(ctx context.Context, id string) (storedSaga, error) {
	row := r.db.QueryRowContext(ctx, `SELECT
 saga_id,account_ref,profile,market,symbol,generation,revision,state,trigger,quantity,
 pending_trigger,pending_quantity,client_order_id,attempt_id,broker_id,previous_broker_id,reconcile_reason,updated_at,
 last_event_kind,last_event_fingerprint
 FROM protection_sagas WHERE saga_id=?`, id)
	return scanStoredSaga(row)
}

func (r *Repository) BeginRegistration(ctx context.Context, id string, expectedRevision int64, at time.Time, attemptID string) (Saga, error) {
	return r.apply(ctx, id, expectedRevision, Event{Kind: EventBeginRegistration, At: at, AttemptID: attemptID})
}

func (r *Repository) MarkRegistrationActive(ctx context.Context, id string, expectedRevision int64, at time.Time, attemptID, brokerID string) (Saga, error) {
	return r.apply(ctx, id, expectedRevision, Event{Kind: EventRegistrationActive, At: at, AttemptID: attemptID, BrokerID: brokerID})
}

func (r *Repository) MarkMutationUnknown(ctx context.Context, id string, expectedRevision int64, at time.Time, attemptID string) (Saga, error) {
	return r.apply(ctx, id, expectedRevision, Event{Kind: EventMutationUnknown, At: at, AttemptID: attemptID})
}

func (r *Repository) BeginReplace(ctx context.Context, id string, expectedRevision int64, at time.Time, attemptID string, trigger, quantity int64) (Saga, error) {
	return r.apply(ctx, id, expectedRevision, Event{Kind: EventBeginReplace, At: at, AttemptID: attemptID, Trigger: trigger, Quantity: quantity})
}

func (r *Repository) MarkReplaceActive(ctx context.Context, id string, expectedRevision int64, at time.Time, attemptID, brokerID string) (Saga, error) {
	return r.apply(ctx, id, expectedRevision, Event{Kind: EventReplaceActive, At: at, AttemptID: attemptID, BrokerID: brokerID})
}

func (r *Repository) BeginCancel(ctx context.Context, id string, expectedRevision int64, at time.Time, attemptID, brokerID string) (Saga, error) {
	return r.apply(ctx, id, expectedRevision, Event{Kind: EventBeginCancel, At: at, AttemptID: attemptID, BrokerID: brokerID})
}

func (r *Repository) MarkCancelClosed(ctx context.Context, id string, expectedRevision int64, at time.Time, attemptID, brokerID string) (Saga, error) {
	return r.apply(ctx, id, expectedRevision, Event{Kind: EventCancelClosed, At: at, AttemptID: attemptID, BrokerID: brokerID})
}

func (r *Repository) RecoverActive(ctx context.Context, id string, expectedRevision int64, at time.Time, brokerID string, trigger, quantity int64) (Saga, error) {
	return r.apply(ctx, id, expectedRevision, Event{Kind: EventRecoveryActive, At: at, BrokerID: brokerID, Trigger: trigger, Quantity: quantity})
}

func (r *Repository) RecoverTriggered(ctx context.Context, id string, expectedRevision int64, at time.Time, brokerID string) (Saga, error) {
	return r.apply(ctx, id, expectedRevision, Event{Kind: EventRecoveryTriggered, At: at, BrokerID: brokerID})
}

func (r *Repository) RecoverClosed(ctx context.Context, id string, expectedRevision int64, at time.Time, brokerID string) (Saga, error) {
	return r.apply(ctx, id, expectedRevision, Event{Kind: EventRecoveryClosed, At: at, BrokerID: brokerID})
}

func (r *Repository) List(ctx context.Context, scope Scope) ([]Saga, error) {
	if err := scope.Validate(); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT
 saga_id,account_ref,profile,market,symbol,generation,revision,state,trigger,quantity,
 pending_trigger,pending_quantity,client_order_id,attempt_id,broker_id,previous_broker_id,reconcile_reason,updated_at,
 last_event_kind,last_event_fingerprint
 FROM protection_sagas WHERE account_ref=? AND profile=? AND market=? AND symbol=? ORDER BY generation,saga_id`,
		scope.AccountRef, scope.Profile, scope.Market, scope.Symbol)
	if err != nil {
		return nil, fmt.Errorf("protection: list sagas: %w", err)
	}
	defer rows.Close()
	var out []Saga
	for rows.Next() {
		stored, scanErr := scanStoredSaga(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, stored.Saga)
	}
	return out, rows.Err()
}

func (r *Repository) MarkTriggerObserved(ctx context.Context, id string, expectedRevision int64, at time.Time, brokerID string) (Saga, error) {
	return r.apply(ctx, id, expectedRevision, Event{Kind: EventTriggerObserved, At: at, BrokerID: brokerID})
}

func (r *Repository) Close(ctx context.Context, id string, expectedRevision int64, at time.Time, brokerID string) (Saga, error) {
	return r.apply(ctx, id, expectedRevision, Event{Kind: EventClose, At: at, BrokerID: brokerID})
}

func (r *Repository) MarkDiscrepancy(ctx context.Context, id string, expectedRevision int64, at time.Time, reason string) (Saga, error) {
	return r.apply(ctx, id, expectedRevision, Event{Kind: EventDiscrepancy, At: at, Reason: reason})
}

// apply loads the durable saga and applies an event constructed by one of the
// event-specific public methods. Callers cannot submit a replacement Saga or
// populate fields that are irrelevant to that event kind.
func (r *Repository) apply(ctx context.Context, id string, expectedRevision int64, event Event) (Saga, error) {
	if id == "" {
		return Saga{}, fmt.Errorf("%w: saga id is absent", ErrConcurrentUpdate)
	}
	if expectedRevision < 1 {
		return Saga{}, fmt.Errorf("%w: expected revision must be positive", ErrConcurrentUpdate)
	}
	stored, err := r.getStored(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Saga{}, fmt.Errorf("%w: saga %s not found", ErrConcurrentUpdate, id)
		}
		return Saga{}, err
	}
	current := stored.Saga
	if current.Revision != expectedRevision {
		if current.Revision == expectedRevision+1 && exactEventResult(stored, event) {
			return current, nil
		}
		return Saga{}, fmt.Errorf("%w: saga %s revision %d", ErrConcurrentUpdate, id, expectedRevision)
	}
	next, err := Transition(current, event)
	if err != nil {
		return Saga{}, err
	}
	next.Revision = expectedRevision + 1
	fingerprint := eventFingerprint(event)
	values := sagaValues(next)
	args := append(values[1:], string(event.Kind), fingerprint, id, expectedRevision)
	result, err := r.db.ExecContext(ctx, `UPDATE protection_sagas SET
 account_ref=?,profile=?,market=?,symbol=?,generation=?,revision=?,state=?,trigger=?,quantity=?,
 pending_trigger=?,pending_quantity=?,client_order_id=?,attempt_id=?,broker_id=?,previous_broker_id=?,reconcile_reason=?,updated_at=?,
 last_event_kind=?,last_event_fingerprint=?
 WHERE saga_id=? AND revision=?`, args...)
	if err != nil {
		return Saga{}, fmt.Errorf("protection: apply saga %s event %s: %w", id, event.Kind, err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return Saga{}, fmt.Errorf("protection: apply saga %s result: %w", id, err)
	}
	if n != 1 {
		latest, getErr := r.getStored(ctx, id)
		if getErr == nil && latest.Revision == expectedRevision+1 && exactEventResult(latest, event) {
			return latest.Saga, nil
		}
		return Saga{}, fmt.Errorf("%w: saga %s revision %d", ErrConcurrentUpdate, id, expectedRevision)
	}
	return next, nil
}

func exactEventResult(stored storedSaga, event Event) bool {
	if stored.lastEventKind != event.Kind || stored.lastEventFingerprint != eventFingerprint(event) {
		return false
	}
	saga := stored.Saga
	if !saga.UpdatedAt.Equal(event.At) {
		return false
	}
	switch event.Kind {
	case EventBeginRegistration:
		return saga.State == StateRegistering && saga.AttemptID == event.AttemptID
	case EventRegistrationActive:
		return saga.State == StateActive && saga.AttemptID == event.AttemptID && saga.BrokerID == event.BrokerID
	case EventMutationUnknown:
		return saga.State == StateInDoubt && saga.AttemptID == event.AttemptID && saga.ReconcileReason == "MUTATION_RESULT_UNKNOWN"
	case EventBeginReplace:
		quantity := event.Quantity
		if quantity == 0 {
			quantity = saga.Quantity
		}
		return saga.State == StateReplacing && saga.AttemptID == event.AttemptID && saga.PendingTrigger == event.Trigger && saga.PendingQuantity == quantity && saga.PreviousBrokerID == saga.BrokerID
	case EventReplaceActive:
		return saga.State == StateActive && saga.AttemptID == event.AttemptID && saga.BrokerID == event.BrokerID
	case EventBeginCancel:
		return saga.State == StateCancelling && saga.AttemptID == event.AttemptID && saga.BrokerID == event.BrokerID
	case EventCancelClosed:
		return saga.State == StateClosed && saga.AttemptID == event.AttemptID && saga.BrokerID == event.BrokerID
	case EventRecoveryActive:
		return saga.State == StateActive && saga.BrokerID == event.BrokerID && saga.Trigger == event.Trigger && saga.Quantity == event.Quantity
	case EventRecoveryTriggered:
		return saga.State == StateTriggered && saga.BrokerID == event.BrokerID
	case EventRecoveryClosed:
		return saga.State == StateClosed && saga.BrokerID == event.BrokerID
	case EventTriggerObserved:
		return saga.State == StateTriggered && saga.BrokerID == event.BrokerID
	case EventClose:
		return saga.State == StateClosed && saga.BrokerID == event.BrokerID
	case EventDiscrepancy:
		return saga.State == StateReconcile && saga.ReconcileReason == event.Reason
	default:
		return false
	}
}

func eventFingerprint(event Event) string {
	canonical := struct {
		Kind      EventKind `json:"kind"`
		At        string    `json:"at"`
		AttemptID string    `json:"attempt_id"`
		BrokerID  string    `json:"broker_id"`
		Trigger   int64     `json:"trigger"`
		Quantity  int64     `json:"quantity"`
		Reason    string    `json:"reason"`
	}{event.Kind, event.At.UTC().Format(time.RFC3339Nano), event.AttemptID, event.BrokerID, event.Trigger, event.Quantity, event.Reason}
	payload, _ := json.Marshal(canonical)
	return fmt.Sprintf("%x", sha256.Sum256(payload))
}

func sagaValues(s Saga) []any {
	return []any{s.ID, s.AccountRef, s.Profile, s.Market, s.Symbol, s.Generation, s.Revision, string(s.State), s.Trigger, s.Quantity, s.PendingTrigger, s.PendingQuantity, s.ClientOrderID, s.AttemptID, s.BrokerID, s.PreviousBrokerID, s.ReconcileReason, s.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}

type rowScanner interface{ Scan(...any) error }

func scanStoredSaga(row rowScanner) (storedSaga, error) {
	var stored storedSaga
	var state, updated, lastEventKind string
	s := &stored.Saga
	err := row.Scan(&s.ID, &s.AccountRef, &s.Profile, &s.Market, &s.Symbol, &s.Generation, &s.Revision, &state, &s.Trigger, &s.Quantity, &s.PendingTrigger, &s.PendingQuantity, &s.ClientOrderID, &s.AttemptID, &s.BrokerID, &s.PreviousBrokerID, &s.ReconcileReason, &updated, &lastEventKind, &stored.lastEventFingerprint)
	if err != nil {
		return storedSaga{}, err
	}
	s.State = State(state)
	s.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return storedSaga{}, fmt.Errorf("protection: invalid persisted updated_at: %w", err)
	}
	if err := s.Validate(); err != nil {
		return storedSaga{}, err
	}
	stored.lastEventKind = EventKind(lastEventKind)
	blankEventIdentity := stored.lastEventKind == "" && stored.lastEventFingerprint == ""
	if (blankEventIdentity && (s.Revision != 1 || s.State != StatePlanned)) ||
		(!blankEventIdentity && (!validEventKind(stored.lastEventKind) || !validEventFingerprint(stored.lastEventFingerprint))) {
		return storedSaga{}, fmt.Errorf("protection: invalid persisted event identity")
	}
	return stored, nil
}

func validEventKind(kind EventKind) bool {
	switch kind {
	case EventBeginRegistration, EventRegistrationActive, EventMutationUnknown, EventBeginReplace, EventReplaceActive, EventBeginCancel, EventCancelClosed, EventRecoveryActive, EventRecoveryTriggered, EventRecoveryClosed, EventTriggerObserved, EventClose, EventDiscrepancy:
		return true
	default:
		return false
	}
}

func validEventFingerprint(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	for i := range value {
		if !((value[i] >= '0' && value[i] <= '9') || (value[i] >= 'a' && value[i] <= 'f')) {
			return false
		}
	}
	return true
}
