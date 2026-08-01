package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var (
	ErrCapabilitySpent       = errors.New("httpapi: capability nonce has already been consumed")
	ErrIdempotencyConflict   = errors.New("httpapi: idempotency key was used with another body")
	ErrIdempotencyInProgress = errors.New("httpapi: idempotent mutation outcome is still pending")
)

type MutationLedgerRequest struct {
	Identity        MutationIdentity
	Method          string
	Resource        string
	BodyDigest      string
	IdempotencyKey  string
	IfMatch         string
	CapabilityNonce string
	At              time.Time
}

type StoredMutationResponse struct {
	Status  int
	Version string
	Body    []byte
}

type MutationReservation struct {
	ID     int64
	Replay *StoredMutationResponse
}

type MutationLedger interface {
	Reserve(context.Context, MutationLedgerRequest) (MutationReservation, error)
	Complete(context.Context, int64, StoredMutationResponse) error
}

// SecurityStore owns only remote API authorization state. It is deliberately
// separate from the trading journal and has no broker or engine writer.
type SecurityStore struct {
	db *sql.DB
}

func OpenSecurityStore(path string) (*SecurityStore, error) {
	path = filepath.Clean(strings.TrimSpace(path))
	if path == "." || path == "" {
		return nil, errors.New("httpapi: security database path is required")
	}
	if err := validateSecurityStoreDirectory(filepath.Dir(path)); err != nil {
		return nil, fmt.Errorf("httpapi: insecure security database directory: %w", err)
	}
	if info, err := os.Lstat(path); err == nil {
		if err := validateSecurityStoreFile(info); err != nil {
			return nil, fmt.Errorf("httpapi: insecure existing security database: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("httpapi: inspecting security database: %w", err)
	}
	query := url.Values{}
	query.Set("_pragma", "busy_timeout(5000)")
	query.Add("_pragma", "foreign_keys(1)")
	query.Add("_pragma", "journal_mode(WAL)")
	query.Add("_pragma", "synchronous(FULL)")
	db, err := sql.Open("sqlite", "file:"+path+"?"+query.Encode())
	if err != nil {
		return nil, fmt.Errorf("httpapi: opening security database: %w", err)
	}
	db.SetMaxOpenConns(1)
	store := &SecurityStore{db: db}
	if err := store.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("httpapi: securing database permissions: %w", err)
	}
	return store, nil
}

func (s *SecurityStore) migrate(ctx context.Context) error {
	const schema = `
CREATE TABLE IF NOT EXISTS consumed_capability_nonces (
 nonce_digest TEXT PRIMARY KEY CHECK(length(nonce_digest)=64),
 consumed_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS mutation_idempotency (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 actor TEXT NOT NULL,
 client TEXT NOT NULL,
 method TEXT NOT NULL,
 resource TEXT NOT NULL,
 idempotency_key TEXT NOT NULL,
	 body_digest TEXT NOT NULL CHECK(length(body_digest)=64),
	 auth_mode TEXT NOT NULL,
	 if_match TEXT NOT NULL,
 state TEXT NOT NULL CHECK(state IN ('pending','complete')),
 response_status INTEGER,
 response_version TEXT,
 response_body BLOB,
 created_at TEXT NOT NULL,
 completed_at TEXT,
 UNIQUE(actor,client,method,resource,idempotency_key)
);
CREATE TABLE IF NOT EXISTS mutation_audit (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 reservation_id INTEGER NOT NULL REFERENCES mutation_idempotency(id),
 stage TEXT NOT NULL CHECK(stage IN ('authorized','replayed','completed')),
 actor TEXT NOT NULL,
 client TEXT NOT NULL,
 auth_mode TEXT NOT NULL,
 method TEXT NOT NULL,
 resource TEXT NOT NULL,
 body_digest TEXT NOT NULL,
 idempotency_key TEXT NOT NULL,
 if_match TEXT NOT NULL,
 status INTEGER,
 recorded_at TEXT NOT NULL
);
CREATE TRIGGER IF NOT EXISTS mutation_audit_no_update
 BEFORE UPDATE ON mutation_audit BEGIN SELECT RAISE(ABORT,'mutation audit is append-only'); END;
CREATE TRIGGER IF NOT EXISTS mutation_audit_no_delete
 BEFORE DELETE ON mutation_audit BEGIN SELECT RAISE(ABORT,'mutation audit is append-only'); END;
CREATE TRIGGER IF NOT EXISTS capability_nonce_no_update
 BEFORE UPDATE ON consumed_capability_nonces BEGIN SELECT RAISE(ABORT,'capability nonce is immutable'); END;
CREATE TRIGGER IF NOT EXISTS capability_nonce_no_delete
 BEFORE DELETE ON consumed_capability_nonces BEGIN SELECT RAISE(ABORT,'capability nonce is immutable'); END;`
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("httpapi: migrating security database: %w", err)
	}
	return nil
}

func (s *SecurityStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *SecurityStore) Reserve(ctx context.Context, req MutationLedgerRequest) (MutationReservation, error) {
	if s == nil || s.db == nil || !req.Identity.valid() || req.Method != "POST" || !validResource(req.Resource) ||
		!validDigest(req.BodyDigest) || !validIdempotencyKey(req.IdempotencyKey) || !validIfMatch(req.IfMatch) || req.At.IsZero() {
		return MutationReservation{}, errors.New("httpapi: invalid mutation ledger request")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return MutationReservation{}, fmt.Errorf("httpapi: beginning mutation reservation: %w", err)
	}
	defer tx.Rollback()
	now := req.At.UTC().Format(time.RFC3339Nano)
	if req.CapabilityNonce != "" {
		digest := sha256.Sum256([]byte(req.CapabilityNonce))
		if _, err := tx.ExecContext(ctx, `INSERT INTO consumed_capability_nonces(nonce_digest,consumed_at) VALUES(?,?)`,
			hex.EncodeToString(digest[:]), now); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "unique") {
				return MutationReservation{}, ErrCapabilitySpent
			}
			return MutationReservation{}, fmt.Errorf("httpapi: consuming capability nonce: %w", err)
		}
	}
	var id int64
	var bodyDigest, state string
	var responseStatus sql.NullInt64
	var responseVersion sql.NullString
	var responseBody []byte
	err = tx.QueryRowContext(ctx, `SELECT id,body_digest,state,response_status,response_version,response_body
 FROM mutation_idempotency WHERE actor=? AND client=? AND method=? AND resource=? AND idempotency_key=?`,
		req.Identity.Actor, req.Identity.Client, req.Method, req.Resource, req.IdempotencyKey,
	).Scan(&id, &bodyDigest, &state, &responseStatus, &responseVersion, &responseBody)
	if err == nil {
		if bodyDigest != req.BodyDigest {
			return MutationReservation{}, ErrIdempotencyConflict
		}
		if state != "complete" {
			return MutationReservation{}, ErrIdempotencyInProgress
		}
		if err := insertAudit(ctx, tx, id, "replayed", req, int(responseStatus.Int64)); err != nil {
			return MutationReservation{}, err
		}
		if err := tx.Commit(); err != nil {
			return MutationReservation{}, fmt.Errorf("httpapi: committing idempotent replay: %w", err)
		}
		return MutationReservation{ID: id, Replay: &StoredMutationResponse{
			Status: int(responseStatus.Int64), Version: responseVersion.String, Body: append([]byte(nil), responseBody...),
		}}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return MutationReservation{}, fmt.Errorf("httpapi: reading idempotency reservation: %w", err)
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO mutation_idempotency(
 actor,client,method,resource,idempotency_key,body_digest,auth_mode,if_match,state,created_at)
 VALUES(?,?,?,?,?,?,?,?,'pending',?)`, req.Identity.Actor, req.Identity.Client, req.Method, req.Resource,
		req.IdempotencyKey, req.BodyDigest, req.Identity.Mode, req.IfMatch, now)
	if err != nil {
		return MutationReservation{}, fmt.Errorf("httpapi: inserting idempotency reservation: %w", err)
	}
	id, err = result.LastInsertId()
	if err != nil {
		return MutationReservation{}, fmt.Errorf("httpapi: reading reservation id: %w", err)
	}
	if err := insertAudit(ctx, tx, id, "authorized", req, 0); err != nil {
		return MutationReservation{}, err
	}
	if err := tx.Commit(); err != nil {
		return MutationReservation{}, fmt.Errorf("httpapi: committing mutation reservation: %w", err)
	}
	return MutationReservation{ID: id}, nil
}

func insertAudit(ctx context.Context, tx *sql.Tx, id int64, stage string, req MutationLedgerRequest, status int) error {
	var value any
	if status != 0 {
		value = status
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO mutation_audit(
 reservation_id,stage,actor,client,auth_mode,method,resource,body_digest,idempotency_key,if_match,status,recorded_at)
 VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, id, stage, req.Identity.Actor, req.Identity.Client, req.Identity.Mode,
		req.Method, req.Resource, req.BodyDigest, req.IdempotencyKey, req.IfMatch, value,
		req.At.UTC().Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("httpapi: recording pre-command mutation audit: %w", err)
	}
	return nil
}

func (s *SecurityStore) Complete(ctx context.Context, id int64, response StoredMutationResponse) error {
	if id <= 0 || response.Status < 100 || response.Status > 599 || int64(len(response.Body)) > MaxRequestBodyBytes {
		return errors.New("httpapi: invalid mutation completion")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("httpapi: beginning mutation completion: %w", err)
	}
	defer tx.Rollback()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := tx.ExecContext(ctx, `UPDATE mutation_idempotency SET state='complete',response_status=?,
 response_version=?,response_body=?,completed_at=? WHERE id=? AND state='pending'`,
		response.Status, response.Version, response.Body, now, id)
	if err != nil {
		return fmt.Errorf("httpapi: completing mutation: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("httpapi: reading mutation completion result: %w", err)
	}
	if rows != 1 {
		var state string
		var status int
		var version string
		var body []byte
		readErr := tx.QueryRowContext(ctx, `SELECT state,response_status,response_version,response_body
 FROM mutation_idempotency WHERE id=?`, id).Scan(&state, &status, &version, &body)
		if readErr == nil && state == "complete" && status == response.Status && version == response.Version && bytes.Equal(body, response.Body) {
			return nil
		}
		return errors.New("httpapi: mutation reservation is not pending")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO mutation_audit(
 reservation_id,stage,actor,client,auth_mode,method,resource,body_digest,idempotency_key,if_match,status,recorded_at)
 SELECT id,'completed',actor,client,auth_mode,method,resource,body_digest,idempotency_key,if_match,?,?
 FROM mutation_idempotency WHERE id=?`, response.Status, now, id); err != nil {
		return fmt.Errorf("httpapi: recording mutation completion audit: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("httpapi: committing mutation completion: %w", err)
	}
	return nil
}

func (s *SecurityStore) PendingCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM mutation_idempotency WHERE state='pending'`).Scan(&count)
	return count, err
}

func validIfMatch(value string) bool {
	if len(value) < 3 || len(value) > 22 || value[0] != '"' || value[len(value)-1] != '"' {
		return false
	}
	for _, r := range value[1 : len(value)-1] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
