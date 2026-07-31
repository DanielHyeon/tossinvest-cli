package protection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const SchemaVersion = 1

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
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS protection_sagas_account_symbol
  ON protection_sagas(account_ref, symbol, state);
`

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) (*Repository, error) {
	if db == nil {
		return nil, errors.New("protection: nil database")
	}
	if _, err := db.Exec(schemaDDL); err != nil {
		return nil, fmt.Errorf("protection: creating additive schema v%d: %w", SchemaVersion, err)
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Insert(ctx context.Context, saga Saga) error {
	if err := saga.Validate(); err != nil {
		return err
	}
	if saga.Revision == 0 {
		saga.Revision = 1
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO protection_sagas (
 saga_id,account_ref,profile,market,symbol,generation,revision,state,trigger,quantity,
 pending_trigger,pending_quantity,client_order_id,attempt_id,broker_id,previous_broker_id,reconcile_reason,updated_at
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, sagaValues(saga)...)
	if err != nil {
		return fmt.Errorf("protection: insert saga %s: %w", saga.ID, err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, id string) (Saga, error) {
	row := r.db.QueryRowContext(ctx, `SELECT
 saga_id,account_ref,profile,market,symbol,generation,revision,state,trigger,quantity,
 pending_trigger,pending_quantity,client_order_id,attempt_id,broker_id,previous_broker_id,reconcile_reason,updated_at
 FROM protection_sagas WHERE saga_id=?`, id)
	return scanSaga(row)
}

func (r *Repository) Update(ctx context.Context, expectedRevision int64, saga Saga) error {
	if err := saga.Validate(); err != nil {
		return err
	}
	if expectedRevision < 1 {
		return fmt.Errorf("%w: expected revision must be positive", ErrConcurrentUpdate)
	}
	saga.Revision = expectedRevision + 1
	values := sagaValues(saga)
	args := append(values[1:], saga.ID, expectedRevision)
	result, err := r.db.ExecContext(ctx, `UPDATE protection_sagas SET
 account_ref=?,profile=?,market=?,symbol=?,generation=?,revision=?,state=?,trigger=?,quantity=?,
 pending_trigger=?,pending_quantity=?,client_order_id=?,attempt_id=?,broker_id=?,previous_broker_id=?,reconcile_reason=?,updated_at=?
 WHERE saga_id=? AND revision=?`, args...)
	if err != nil {
		return fmt.Errorf("protection: update saga %s: %w", saga.ID, err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("protection: update saga %s result: %w", saga.ID, err)
	}
	if n != 1 {
		return fmt.Errorf("%w: saga %s revision %d", ErrConcurrentUpdate, saga.ID, expectedRevision)
	}
	return nil
}

func sagaValues(s Saga) []any {
	return []any{s.ID, s.AccountRef, s.Profile, s.Market, s.Symbol, s.Generation, s.Revision, string(s.State), s.Trigger, s.Quantity, s.PendingTrigger, s.PendingQuantity, s.ClientOrderID, s.AttemptID, s.BrokerID, s.PreviousBrokerID, s.ReconcileReason, s.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}

type rowScanner interface{ Scan(...any) error }

func scanSaga(row rowScanner) (Saga, error) {
	var s Saga
	var state, updated string
	err := row.Scan(&s.ID, &s.AccountRef, &s.Profile, &s.Market, &s.Symbol, &s.Generation, &s.Revision, &state, &s.Trigger, &s.Quantity, &s.PendingTrigger, &s.PendingQuantity, &s.ClientOrderID, &s.AttemptID, &s.BrokerID, &s.PreviousBrokerID, &s.ReconcileReason, &updated)
	if err != nil {
		return Saga{}, err
	}
	s.State = State(state)
	s.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return Saga{}, fmt.Errorf("protection: invalid persisted updated_at: %w", err)
	}
	if err := s.Validate(); err != nil {
		return Saga{}, err
	}
	return s, nil
}
