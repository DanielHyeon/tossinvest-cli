package optimization

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
)

const (
	DatabaseFileName     = "optimization-control.db"
	previewTTL           = 5 * time.Minute
	riskWait             = 3 * time.Second
	controlSchemaVersion = 3
	readHistoryLimit     = 1000
	readAuditLimit       = 1000
)

type Options struct {
	Path        string
	Registry    Registry
	Evidence    EvidenceProvider
	Clock       clock.Clock
	Actor       string
	Random      io.Reader
	BusyTimeout time.Duration
}

type Store struct {
	db       *sql.DB
	path     string
	registry Registry
	evidence EvidenceProvider
	clock    clock.Clock
	actor    string
	random   io.Reader
	closeMu  sync.Mutex
	closed   bool
}

// candidatePayload is the complete immutable candidate contract that a preview
// authenticates with the opaque capability as its HMAC key. Its JSON form is
// deliberately field-ordered and contains no mutable database-only values.
type candidatePayload struct {
	CandidateID         string          `json:"candidate_id"`
	Actor               string          `json:"actor"`
	BaseVersion         uint64          `json:"base_version"`
	Category            Category        `json:"category"`
	Source              CandidateSource `json:"source"`
	Reason              ReasonCode      `json:"reason"`
	ChangesJSON         string          `json:"changes_json"`
	EvidenceJSON        string          `json:"evidence_json"`
	NotBefore           time.Time       `json:"not_before"`
	ExpiresAt           time.Time       `json:"expires_at"`
	RiskRequired        int             `json:"risk_required"`
	RestartRequired     int             `json:"restart_required"`
	EffectiveEntryAfter int             `json:"effective_entry_after"`
	CreatedAt           time.Time       `json:"created_at"`
}

type appendOnlyTrigger struct {
	name string
	sql  string
}

var appendOnlyTriggers = []appendOnlyTrigger{
	{"optimization_snapshots_no_update", `CREATE TRIGGER optimization_snapshots_no_update BEFORE UPDATE ON optimization_snapshots BEGIN SELECT RAISE(ABORT, 'optimization snapshots are append-only'); END`},
	{"optimization_snapshots_no_delete", `CREATE TRIGGER optimization_snapshots_no_delete BEFORE DELETE ON optimization_snapshots BEGIN SELECT RAISE(ABORT, 'optimization snapshots are append-only'); END`},
	{"optimization_candidates_no_update", `CREATE TRIGGER optimization_candidates_no_update BEFORE UPDATE ON optimization_candidates BEGIN SELECT RAISE(ABORT, 'optimization candidates are append-only'); END`},
	{"optimization_candidates_no_delete", `CREATE TRIGGER optimization_candidates_no_delete BEFORE DELETE ON optimization_candidates BEGIN SELECT RAISE(ABORT, 'optimization candidates are append-only'); END`},
	{"optimization_applications_no_update", `CREATE TRIGGER optimization_applications_no_update BEFORE UPDATE ON optimization_applications BEGIN SELECT RAISE(ABORT, 'optimization applications are append-only'); END`},
	{"optimization_applications_no_delete", `CREATE TRIGGER optimization_applications_no_delete BEFORE DELETE ON optimization_applications BEGIN SELECT RAISE(ABORT, 'optimization applications are append-only'); END`},
	{"optimization_audit_no_update", `CREATE TRIGGER optimization_audit_no_update BEFORE UPDATE ON optimization_audit BEGIN SELECT RAISE(ABORT, 'optimization audit is append-only'); END`},
	{"optimization_audit_no_delete", `CREATE TRIGGER optimization_audit_no_delete BEFORE DELETE ON optimization_audit BEGIN SELECT RAISE(ABORT, 'optimization audit is append-only'); END`},
}

func Open(ctx context.Context, opts Options) (*Store, error) {
	path := strings.TrimSpace(opts.Path)
	if path == "" {
		return nil, errors.New("optimization: database path is required")
	}
	if strings.TrimSpace(opts.Actor) == "" {
		return nil, errors.New("optimization: actor is required")
	}
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, fmt.Errorf("optimization: creating control directory: %w", err)
	}
	if err := os.Chmod(directory, 0o700); err != nil {
		return nil, fmt.Errorf("optimization: securing control directory: %w", err)
	}
	busy := opts.BusyTimeout
	if busy <= 0 {
		busy = 5 * time.Second
	}
	q := url.Values{}
	q.Add("_pragma", "foreign_keys(1)")
	q.Add("_pragma", "journal_mode(WAL)")
	q.Add("_pragma", "synchronous(FULL)")
	q.Add("_pragma", "busy_timeout("+strconv.FormatInt(busy.Milliseconds(), 10)+")")
	db, err := sql.Open("sqlite", "file:"+filepath.Clean(path)+"?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("optimization: opening control database: %w", err)
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("optimization: connecting control database: %w", err)
	}
	if err := secureSQLiteFiles(path); err != nil {
		db.Close()
		return nil, fmt.Errorf("optimization: securing control database: %w", err)
	}
	store := &Store{db: db, path: path, registry: opts.Registry, evidence: opts.Evidence,
		clock: opts.Clock, actor: strings.TrimSpace(opts.Actor), random: opts.Random}
	if store.clock == nil {
		store.clock = clock.System()
	}
	if store.random == nil {
		store.random = rand.Reader
	}
	if err := store.init(ctx); err != nil {
		db.Close()
		return nil, err
	}
	if err := secureSQLiteFiles(path); err != nil {
		db.Close()
		return nil, fmt.Errorf("optimization: securing control database sidecars: %w", err)
	}
	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	s.closeMu.Lock()
	defer s.closeMu.Unlock()
	if s.closed {
		return nil
	}
	_ = secureSQLiteFiles(s.path)
	err := s.db.Close()
	s.closed = true
	return err
}

func (s *Store) init(ctx context.Context) error {
	if err := s.migrate(ctx); err != nil {
		return err
	}
	return s.ensureInitialSnapshot(ctx)
}

func (s *Store) migrate(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("optimization: beginning schema migration: %w", err)
	}
	defer tx.Rollback()
	var version int
	if err := tx.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&version); err != nil {
		return fmt.Errorf("optimization: reading schema version: %w", err)
	}
	if version > controlSchemaVersion {
		return fmt.Errorf("optimization: database schema version %d is newer than supported %d", version, controlSchemaVersion)
	}
	const schema = `
CREATE TABLE IF NOT EXISTS optimization_control (
  singleton INTEGER PRIMARY KEY CHECK(singleton=1), current_version INTEGER NOT NULL,
  control_digest TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS optimization_snapshots (
  version INTEGER PRIMARY KEY, effective_version INTEGER NOT NULL,
  desired_json TEXT NOT NULL, effective_json TEXT NOT NULL,
  settings_digest TEXT NOT NULL UNIQUE, evidence_digest TEXT NOT NULL,
  activation_manifest_digest TEXT NOT NULL, effective_entry INTEGER NOT NULL,
  restart_required INTEGER NOT NULL, actor TEXT NOT NULL, reason TEXT NOT NULL,
  audit_id TEXT NOT NULL UNIQUE, created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS optimization_candidates (
  candidate_id TEXT PRIMARY KEY, base_version INTEGER NOT NULL, category TEXT NOT NULL,
  source TEXT NOT NULL, reason TEXT NOT NULL, changes_json TEXT NOT NULL,
  evidence_json TEXT NOT NULL, capability_hash TEXT NOT NULL UNIQUE, payload_mac TEXT NOT NULL,
  actor TEXT NOT NULL,
  not_before TEXT NOT NULL, expires_at TEXT NOT NULL, risk_required INTEGER NOT NULL,
  restart_required INTEGER NOT NULL, effective_entry_after INTEGER NOT NULL,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS optimization_applications (
  candidate_id TEXT PRIMARY KEY REFERENCES optimization_candidates(candidate_id),
  result_version INTEGER NOT NULL REFERENCES optimization_snapshots(version),
  applied_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS optimization_audit (
  id INTEGER PRIMARY KEY AUTOINCREMENT, audit_id TEXT NOT NULL, version INTEGER NOT NULL,
  candidate_id TEXT NOT NULL, setting_key TEXT NOT NULL, before_option_id TEXT NOT NULL,
  after_option_id TEXT NOT NULL, actor TEXT NOT NULL, reason TEXT NOT NULL, created_at TEXT NOT NULL,
  event_digest TEXT NOT NULL
);`
	if _, err := tx.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("optimization: creating lifecycle schema: %w", err)
	}
	var payloadMACColumn int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM pragma_table_info('optimization_candidates') WHERE name='payload_mac'`).Scan(&payloadMACColumn); err != nil {
		return fmt.Errorf("optimization: checking candidate payload MAC schema: %w", err)
	}
	if payloadMACColumn == 0 {
		if _, err := tx.ExecContext(ctx, `ALTER TABLE optimization_candidates ADD COLUMN payload_mac TEXT NOT NULL DEFAULT ''`); err != nil {
			return fmt.Errorf("optimization: migrating candidate payload MAC schema: %w", err)
		}
	}
	var actorColumn int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM pragma_table_info('optimization_candidates') WHERE name='actor'`).Scan(&actorColumn); err != nil {
		return fmt.Errorf("optimization: checking candidate actor schema: %w", err)
	}
	if actorColumn == 0 {
		if _, err := tx.ExecContext(ctx, `ALTER TABLE optimization_candidates ADD COLUMN actor TEXT NOT NULL DEFAULT ''`); err != nil {
			return fmt.Errorf("optimization: migrating candidate actor schema: %w", err)
		}
	}
	var controlDigestColumn int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM pragma_table_info('optimization_control') WHERE name='control_digest'`).Scan(&controlDigestColumn); err != nil {
		return fmt.Errorf("optimization: checking control digest schema: %w", err)
	}
	if controlDigestColumn == 0 {
		if _, err := tx.ExecContext(ctx, `ALTER TABLE optimization_control ADD COLUMN control_digest TEXT NOT NULL DEFAULT ''`); err != nil {
			return fmt.Errorf("optimization: migrating control digest schema: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO optimization_control(singleton,current_version,control_digest) VALUES(1,0,'')`); err != nil {
		return fmt.Errorf("optimization: initializing control pointer: %w", err)
	}
	var auditDigestColumn int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM pragma_table_info('optimization_audit') WHERE name='event_digest'`).Scan(&auditDigestColumn); err != nil {
		return fmt.Errorf("optimization: checking audit digest schema: %w", err)
	}
	if auditDigestColumn == 0 {
		if _, err := tx.ExecContext(ctx, `ALTER TABLE optimization_audit ADD COLUMN event_digest TEXT NOT NULL DEFAULT ''`); err != nil {
			return fmt.Errorf("optimization: migrating audit digest schema: %w", err)
		}
	}
	if err := verifyAppendOnlyTriggers(ctx, tx, true); err != nil {
		return err
	}
	if version < 2 {
		if _, err := tx.ExecContext(ctx, `DROP TRIGGER IF EXISTS optimization_snapshots_no_update`); err != nil {
			return fmt.Errorf("optimization: preparing snapshot digest migration: %w", err)
		}
		if err := migrateSnapshotDigests(ctx, tx); err != nil {
			return err
		}
	}
	if version < 3 {
		if err := migrateControlPointerDigest(ctx, tx); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DROP TRIGGER IF EXISTS optimization_audit_no_update`); err != nil {
			return fmt.Errorf("optimization: preparing audit digest migration: %w", err)
		}
		if err := migrateAuditDigests(ctx, tx); err != nil {
			return err
		}
	}
	if err := installAppendOnlyTriggers(ctx, tx); err != nil {
		return err
	}
	if err := verifyAppendOnlyTriggers(ctx, tx, false); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `PRAGMA user_version = 3`); err != nil {
		return fmt.Errorf("optimization: recording schema version: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("optimization: committing schema migration: %w", err)
	}
	return nil
}

func installAppendOnlyTriggers(ctx context.Context, tx *sql.Tx) error {
	for _, trigger := range appendOnlyTriggers {
		statement := strings.Replace(trigger.sql, "CREATE TRIGGER ", "CREATE TRIGGER IF NOT EXISTS ", 1)
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("optimization: installing append-only trigger %s: %w", trigger.name, err)
		}
	}
	return nil
}

func verifyAppendOnlyTriggers(ctx context.Context, tx *sql.Tx, allowMissing bool) error {
	for _, trigger := range appendOnlyTriggers {
		var actual string
		err := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_schema WHERE type='trigger' AND name=?`, trigger.name).Scan(&actual)
		if errors.Is(err, sql.ErrNoRows) && allowMissing {
			continue
		}
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("optimization: required append-only trigger %s is missing", trigger.name)
		}
		if err != nil {
			return fmt.Errorf("optimization: reading append-only trigger %s: %w", trigger.name, err)
		}
		if canonicalTriggerSQL(actual) != canonicalTriggerSQL(trigger.sql) {
			return fmt.Errorf("optimization: append-only trigger %s has an unexpected definition", trigger.name)
		}
	}
	return nil
}

func canonicalTriggerSQL(statement string) string {
	canonical := strings.Join(strings.Fields(strings.TrimSpace(strings.TrimSuffix(statement, ";"))), " ")
	return strings.Replace(canonical, "CREATE TRIGGER IF NOT EXISTS ", "CREATE TRIGGER ", 1)
}

func migrateSnapshotDigests(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `SELECT version,effective_version,desired_json,effective_json,settings_digest,evidence_digest,
		activation_manifest_digest,effective_entry,restart_required,actor,reason,audit_id,created_at
		FROM optimization_snapshots ORDER BY version`)
	if err != nil {
		return fmt.Errorf("optimization: reading snapshots for digest migration: %w", err)
	}
	var snapshots []Snapshot
	for rows.Next() {
		snapshot, err := scanSnapshotUnchecked(rows)
		if err != nil {
			_ = rows.Close()
			return fmt.Errorf("optimization: validating snapshot digest migration: %w", err)
		}
		snapshots = append(snapshots, snapshot)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("optimization: closing snapshot digest migration rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("optimization: iterating snapshot digest migration: %w", err)
	}
	for _, snapshot := range snapshots {
		if snapshot.SettingsDigest == "" || snapshot.SettingsDigest != digestSnapshotV1(snapshot) {
			return fmt.Errorf("optimization: snapshot %d has an invalid legacy digest", snapshot.Version)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE optimization_snapshots SET settings_digest=? WHERE version=?`,
			digestSnapshot(snapshot), snapshot.Version); err != nil {
			return fmt.Errorf("optimization: migrating snapshot %d digest: %w", snapshot.Version, err)
		}
	}
	return nil
}

func migrateControlPointerDigest(ctx context.Context, tx *sql.Tx) error {
	var version uint64
	if err := tx.QueryRowContext(ctx, `SELECT current_version FROM optimization_control WHERE singleton=1`).Scan(&version); err != nil {
		return fmt.Errorf("optimization: reading control pointer for digest migration: %w", err)
	}
	snapshotDigest := ""
	if version != 0 {
		snapshot, err := scanSnapshot(tx.QueryRowContext(ctx, `SELECT version,effective_version,desired_json,effective_json,settings_digest,evidence_digest,
			activation_manifest_digest,effective_entry,restart_required,actor,reason,audit_id,created_at
			FROM optimization_snapshots WHERE version=?`, version))
		if err != nil {
			return fmt.Errorf("optimization: validating control pointer migration: %w", err)
		}
		snapshotDigest = snapshot.SettingsDigest
	}
	if _, err := tx.ExecContext(ctx, `UPDATE optimization_control SET control_digest=? WHERE singleton=1`,
		digestControlPointer(version, snapshotDigest)); err != nil {
		return fmt.Errorf("optimization: migrating control pointer digest: %w", err)
	}
	return nil
}

func migrateAuditDigests(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `SELECT id,audit_id,version,candidate_id,setting_key,before_option_id,
		after_option_id,actor,reason,created_at,event_digest FROM optimization_audit ORDER BY id`)
	if err != nil {
		return fmt.Errorf("optimization: reading audit rows for digest migration: %w", err)
	}
	var events []AuditEvent
	for rows.Next() {
		event, _, err := scanAuditEventUnchecked(rows)
		if err != nil {
			_ = rows.Close()
			return fmt.Errorf("optimization: validating audit digest migration: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("optimization: closing audit digest migration rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("optimization: iterating audit digest migration: %w", err)
	}
	for _, event := range events {
		if err := validateLegacyAuditEvent(ctx, tx, event); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE optimization_audit SET event_digest=? WHERE id=?`,
			digestAuditEvent(event), event.ID); err != nil {
			return fmt.Errorf("optimization: migrating audit event %d digest: %w", event.ID, err)
		}
	}
	return nil
}

func validateLegacyAuditEvent(ctx context.Context, tx *sql.Tx, event AuditEvent) error {
	snapshot, err := scanSnapshot(tx.QueryRowContext(ctx, `SELECT version,effective_version,desired_json,effective_json,settings_digest,evidence_digest,
		activation_manifest_digest,effective_entry,restart_required,actor,reason,audit_id,created_at
		FROM optimization_snapshots WHERE version=?`, event.Version))
	if err != nil {
		return fmt.Errorf("optimization: audit event %d snapshot validation: %w", event.ID, err)
	}
	if snapshot.AuditID != event.AuditID || snapshot.Actor != event.Actor || snapshot.Reason != event.Reason ||
		!snapshot.CreatedAt.Equal(event.CreatedAt) {
		return fmt.Errorf("optimization: audit event %d does not match its snapshot", event.ID)
	}
	var changesJSON string
	if err := tx.QueryRowContext(ctx, `SELECT changes_json FROM optimization_candidates WHERE candidate_id=?`, event.CandidateID).Scan(&changesJSON); err != nil {
		return fmt.Errorf("optimization: audit event %d candidate validation: %w", event.ID, err)
	}
	var changes []OptionChange
	if err := json.Unmarshal([]byte(changesJSON), &changes); err != nil {
		return fmt.Errorf("optimization: audit event %d candidate changes are invalid", event.ID)
	}
	for _, change := range changes {
		if change.Key == event.Key && change.BeforeOptionID == event.BeforeOptionID && change.AfterOptionID == event.AfterOptionID {
			return nil
		}
	}
	return fmt.Errorf("optimization: audit event %d does not match its candidate", event.ID)
}

func (s *Store) ensureInitialSnapshot(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var version uint64
	var controlDigest string
	if err := tx.QueryRowContext(ctx, `SELECT current_version,control_digest FROM optimization_control WHERE singleton=1`).Scan(&version, &controlDigest); err != nil {
		return err
	}
	if version != 0 {
		snapshot, err := s.snapshot(ctx, tx, version)
		if err != nil {
			return err
		}
		if controlDigest != digestControlPointer(version, snapshot.SettingsDigest) {
			return errors.New("optimization: corrupt control pointer digest")
		}
		return tx.Commit()
	}
	if controlDigest != digestControlPointer(0, "") {
		return errors.New("optimization: corrupt empty control pointer digest")
	}
	desired := map[string]string{}
	effective := map[string]string{}
	for _, field := range s.registry.All() {
		if field.Descriptor.Default.Kind == settingmeta.StateValue {
			desired[field.Descriptor.Key] = field.Descriptor.Default.OptionID
		}
		if field.Descriptor.Effective.Kind == settingmeta.StateValue {
			effective[field.Descriptor.Key] = field.Descriptor.Effective.OptionID
		}
	}
	now := s.clock.Now().UTC()
	snapshot := Snapshot{Version: 1, EffectiveVersion: 1, Desired: desired, Effective: effective, EffectiveEntry: false,
		Actor: "system", Reason: ReasonServerPreset, AuditID: "initial", CreatedAt: now}
	snapshot.SettingsDigest = digestSnapshot(snapshot)
	if err := insertSnapshot(ctx, tx, snapshot); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE optimization_control SET current_version=1,control_digest=?
		WHERE singleton=1 AND current_version=0 AND control_digest=?`,
		digestControlPointer(snapshot.Version, snapshot.SettingsDigest), digestControlPointer(0, ""))
	if err != nil {
		return err
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		return tx.Rollback()
	}
	return tx.Commit()
}

func (s *Store) Read(ctx context.Context) (View, error) {
	snapshot, err := s.currentSnapshot(ctx, s.db)
	if err != nil {
		return View{}, err
	}
	history, err := s.history(ctx)
	if err != nil {
		return View{}, err
	}
	audit, err := s.audit(ctx)
	if err != nil {
		return View{}, err
	}
	evidence, _ := s.readEvidence(ctx)
	return View{Registry: s.registry, Snapshot: snapshot, History: history, Audit: audit, Evidence: evidence}, nil
}

func (s *Store) Preview(ctx context.Context, req PreviewRequest) (Preview, error) {
	return s.preview(ctx, req, "", s.actor)
}

func (s *Store) PreviewRollback(ctx context.Context, req RollbackPreviewRequest) (Preview, error) {
	return s.previewRollback(ctx, req, s.actor)
}

func (s *Store) previewRollback(ctx context.Context, req RollbackPreviewRequest, actor string) (Preview, error) {
	current, err := s.currentSnapshot(ctx, s.db)
	if err != nil {
		return Preview{}, err
	}
	if req.BaseVersion != current.Version {
		return Preview{}, ErrVersionConflict
	}
	target, err := s.snapshot(ctx, s.db, req.TargetVersion)
	if err != nil {
		return Preview{}, fmt.Errorf("%w: version %d", ErrInvalidCandidate, req.TargetVersion)
	}
	keys := make(map[string]struct{}, len(current.Desired)+len(target.Desired))
	for key := range current.Desired {
		keys[key] = struct{}{}
	}
	for key := range target.Desired {
		keys[key] = struct{}{}
	}
	changes := map[string]string{}
	for key := range keys {
		field, ok := s.registry.Field(key)
		if !ok {
			return Preview{}, fmt.Errorf("%w: %s", ErrHistoricalKeyInactive, key)
		}
		optionID := target.Desired[key]
		if field.Category != req.Category || current.Desired[key] == optionID {
			continue
		}
		changes[key] = optionID
	}
	return s.preview(ctx, PreviewRequest{BaseVersion: req.BaseVersion, Category: req.Category,
		Changes: changes, Source: SourceRollback, Reason: ReasonRollback}, "rollback:"+strconv.FormatUint(req.TargetVersion, 10), actor)
}

func (s *Store) preview(ctx context.Context, req PreviewRequest, salt, actor string) (Preview, error) {
	actor = strings.TrimSpace(actor)
	if actor == "" {
		return Preview{}, ErrCapabilityInvalid
	}
	current, err := s.currentSnapshot(ctx, s.db)
	if err != nil {
		return Preview{}, err
	}
	if req.BaseVersion != current.Version {
		return Preview{}, ErrVersionConflict
	}
	if _, ok := ParseCategory(string(req.Category)); !ok || req.Category == CategoryOverview ||
		req.Category == CategoryPerformanceHistory || len(req.Changes) == 0 {
		return Preview{}, ErrInvalidCandidate
	}
	if req.Source != SourceServerPreset && req.Source != SourceEvidence && req.Source != SourceRollback {
		return Preview{}, ErrInvalidCandidate
	}
	if req.Reason != ReasonServerPreset && req.Reason != ReasonRollback {
		return Preview{}, ErrInvalidCandidate
	}
	evidence, evidenceErr := s.readEvidence(ctx)
	if req.Source == SourceEvidence && (evidenceErr != nil || evidence.Status != EvidenceComplete || evidence.Digest == "") {
		return Preview{}, fmt.Errorf("%w: %s", ErrInsufficientEvidence, strings.Join(evidence.Missing, ","))
	}
	keys := make([]string, 0, len(req.Changes))
	for key := range req.Changes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	changes := make([]OptionChange, 0, len(keys))
	riskRequired := false
	restartRequired := false
	for _, key := range keys {
		field, ok := s.registry.Field(key)
		if !ok || field.Category != req.Category || field.Descriptor.Control == settingmeta.ControlReadOnly {
			return Preview{}, fmt.Errorf("%w: field %q is not writable in %s", ErrInvalidCandidate, key, req.Category)
		}
		optionID := strings.TrimSpace(req.Changes[key])
		rollbackToUnapproved := req.Source == SourceRollback && optionID == "" &&
			field.Descriptor.Default.Kind != settingmeta.StateValue
		if err := field.Descriptor.ValidateOption(optionID); err != nil && !rollbackToUnapproved {
			return Preview{}, fmt.Errorf("%w: %v", ErrInvalidCandidate, err)
		}
		if current.Desired[key] == optionID {
			continue
		}
		changes = append(changes, OptionChange{Key: key, BeforeOptionID: current.Desired[key],
			AfterOptionID: optionID, Category: field.Category, ApplyTiming: field.Descriptor.ApplyTiming,
			Safety: field.Descriptor.SafetyDirection})
		if field.Descriptor.SafetyDirection != settingmeta.SafetyNeutral {
			riskRequired = true
		}
		if field.Descriptor.ApplyTiming == settingmeta.ApplyNextEngineStart ||
			field.Descriptor.ApplyTiming == settingmeta.ApplyNewPositionOnly {
			restartRequired = true
		}
	}
	if len(changes) == 0 {
		return Preview{}, ErrInvalidCandidate
	}
	token, err := s.randomToken(32)
	if err != nil {
		return Preview{}, err
	}
	candidateID, err := s.randomToken(18)
	if err != nil {
		return Preview{}, err
	}
	if salt != "" {
		candidateID = candidateID + ":" + salt
	}
	now := s.clock.Now().UTC()
	notBefore := now
	if riskRequired {
		notBefore = now.Add(riskWait)
	}
	preview := Preview{CandidateID: candidateID, BaseVersion: req.BaseVersion, Category: req.Category,
		Changes: changes, Evidence: evidence, Capability: token, NotBefore: notBefore,
		ExpiresAt: now.Add(previewTTL), RestartRequired: restartRequired,
		RiskConfirmationRequired: riskRequired, ExistingPositionsUnchanged: true,
		LiveStateUnchanged: true, EffectiveEntryAfterApply: current.EffectiveEntry && !riskRequired}
	changesJSON, _ := json.Marshal(changes)
	evidenceJSON, _ := json.Marshal(evidence)
	payload := candidatePayload{CandidateID: candidateID, Actor: actor, BaseVersion: req.BaseVersion, Category: req.Category, Source: req.Source, Reason: req.Reason,
		ChangesJSON: string(changesJSON), EvidenceJSON: string(evidenceJSON), NotBefore: notBefore, ExpiresAt: preview.ExpiresAt,
		RiskRequired: boolInt(riskRequired), RestartRequired: boolInt(restartRequired),
		EffectiveEntryAfter: boolInt(preview.EffectiveEntryAfterApply), CreatedAt: now}
	payloadMAC, err := signCandidatePayload(token, payload)
	if err != nil {
		return Preview{}, err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO optimization_candidates
		(candidate_id,base_version,category,source,reason,changes_json,evidence_json,capability_hash,payload_mac,actor,
		not_before,expires_at,risk_required,restart_required,effective_entry_after,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, candidateID, req.BaseVersion, req.Category, req.Source, req.Reason,
		string(changesJSON), string(evidenceJSON), hashCapability(token), payloadMAC, actor, notBefore.Format(time.RFC3339Nano),
		preview.ExpiresAt.Format(time.RFC3339Nano), boolInt(riskRequired), boolInt(restartRequired),
		boolInt(preview.EffectiveEntryAfterApply), now.Format(time.RFC3339Nano))
	if err != nil {
		return Preview{}, fmt.Errorf("optimization: storing immutable candidate: %w", err)
	}
	return preview, nil
}

func (s *Store) Apply(ctx context.Context, req ApplyRequest) (ApplyResult, error) {
	return s.apply(ctx, req, s.actor)
}

// ForActor binds a verified server-side actor to a narrow, immutable Commander.
// It never changes Store state, so independently authenticated requests cannot
// race and relabel one another's audit records.
func (s *Store) ForActor(actor string) (Commander, error) {
	actor = strings.TrimSpace(actor)
	if s == nil || s.db == nil || actor == "" {
		return nil, errors.New("optimization: actor is required")
	}
	return actorCommander{store: s, actor: actor}, nil
}

type actorCommander struct {
	store *Store
	actor string
}

func (c actorCommander) Read(ctx context.Context) (View, error) { return c.store.Read(ctx) }
func (c actorCommander) Preview(ctx context.Context, req PreviewRequest) (Preview, error) {
	return c.store.preview(ctx, req, "", c.actor)
}
func (c actorCommander) PreviewRollback(ctx context.Context, req RollbackPreviewRequest) (Preview, error) {
	return c.store.previewRollback(ctx, req, c.actor)
}
func (c actorCommander) Apply(ctx context.Context, req ApplyRequest) (ApplyResult, error) {
	return c.store.apply(ctx, req, c.actor)
}
func (c actorCommander) RecoverConflict(ctx context.Context, capability string) (ConflictView, error) {
	return c.store.recoverConflict(ctx, capability, c.actor)
}

func (s *Store) apply(ctx context.Context, req ApplyRequest, actor string) (ApplyResult, error) {
	token := strings.TrimSpace(req.Capability)
	actor = strings.TrimSpace(actor)
	if token == "" || actor == "" {
		return ApplyResult{}, ErrCapabilityInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ApplyResult{}, err
	}
	defer tx.Rollback()
	row := tx.QueryRowContext(ctx, `SELECT c.candidate_id,c.base_version,c.category,c.source,c.reason,c.changes_json,c.evidence_json,
		c.actor,c.not_before,c.expires_at,c.created_at,c.risk_required,c.restart_required,c.effective_entry_after,c.payload_mac,a.result_version
		FROM optimization_candidates c LEFT JOIN optimization_applications a ON a.candidate_id=c.candidate_id
		WHERE c.capability_hash=?`, hashCapability(token))
	var candidateID, categoryRaw, sourceRaw, reasonRaw, changesJSON, evidenceJSON, storedActor, notBeforeRaw, expiresRaw, createdRaw, payloadMAC string
	var baseVersion uint64
	var riskRequired, restartRequired, effectiveEntryAfter int
	var resultVersion sql.NullInt64
	if err := row.Scan(&candidateID, &baseVersion, &categoryRaw, &sourceRaw, &reasonRaw, &changesJSON, &evidenceJSON,
		&storedActor, &notBeforeRaw, &expiresRaw, &createdRaw, &riskRequired, &restartRequired, &effectiveEntryAfter, &payloadMAC, &resultVersion); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ApplyResult{}, ErrCapabilityInvalid
		}
		return ApplyResult{}, err
	}
	var changes []OptionChange
	if err := json.Unmarshal([]byte(changesJSON), &changes); err != nil {
		return ApplyResult{}, ErrCapabilityInvalid
	}
	var evidence Evidence
	if err := json.Unmarshal([]byte(evidenceJSON), &evidence); err != nil {
		return ApplyResult{}, ErrCapabilityInvalid
	}
	notBefore, err := parseStoredTime(notBeforeRaw)
	if err != nil {
		return ApplyResult{}, ErrCapabilityInvalid
	}
	expiresAt, err := parseStoredTime(expiresRaw)
	if err != nil || !expiresAt.After(notBefore) {
		return ApplyResult{}, ErrCapabilityInvalid
	}
	createdAt, err := parseStoredTime(createdRaw)
	if err != nil {
		return ApplyResult{}, ErrCapabilityInvalid
	}
	if (riskRequired != 0 && riskRequired != 1) || (restartRequired != 0 && restartRequired != 1) ||
		(effectiveEntryAfter != 0 && effectiveEntryAfter != 1) {
		return ApplyResult{}, ErrCapabilityInvalid
	}
	payload := candidatePayload{CandidateID: candidateID, Actor: storedActor, BaseVersion: baseVersion, Category: Category(categoryRaw), Source: CandidateSource(sourceRaw),
		Reason: ReasonCode(reasonRaw), ChangesJSON: changesJSON, EvidenceJSON: evidenceJSON, NotBefore: notBefore, ExpiresAt: expiresAt,
		RiskRequired: riskRequired, RestartRequired: restartRequired, EffectiveEntryAfter: effectiveEntryAfter, CreatedAt: createdAt}
	if storedActor != actor || !verifyCandidatePayloadMAC(token, payload, payloadMAC) {
		return ApplyResult{}, ErrCapabilityInvalid
	}
	current, err := s.currentSnapshot(ctx, tx)
	if err != nil {
		return ApplyResult{}, err
	}
	if resultVersion.Valid {
		snapshot, err := s.snapshot(ctx, tx, uint64(resultVersion.Int64))
		if err != nil {
			return ApplyResult{}, err
		}
		return ApplyResult{Snapshot: snapshot, Replayed: true}, tx.Commit()
	}
	expectedNotBefore := createdAt
	if riskRequired != 0 {
		expectedNotBefore = expectedNotBefore.Add(riskWait)
	}
	if !notBefore.Equal(expectedNotBefore) || !expiresAt.Equal(createdAt.Add(previewTTL)) {
		return ApplyResult{}, ErrCapabilityInvalid
	}
	now := s.clock.Now().UTC()
	if now.Before(notBefore) {
		return ApplyResult{}, ErrCapabilityTooEarly
	}
	if now.After(expiresAt) {
		return ApplyResult{}, ErrCapabilityExpired
	}
	if riskRequired != 0 && !req.Confirmed {
		return ApplyResult{}, ErrConfirmationRequired
	}
	if current.Version != baseVersion {
		return ApplyResult{}, ErrVersionConflict
	}
	category, knownCategory := ParseCategory(categoryRaw)
	source := CandidateSource(sourceRaw)
	reason := ReasonCode(reasonRaw)
	if !knownCategory || category == CategoryOverview || category == CategoryPerformanceHistory ||
		!validCandidateOrigin(source, reason) {
		return ApplyResult{}, ErrCapabilityInvalid
	}
	if len(changes) == 0 {
		return ApplyResult{}, ErrCapabilityInvalid
	}
	if source == SourceEvidence {
		currentEvidence, evidenceErr := s.readEvidence(ctx)
		if evidenceErr != nil || currentEvidence.Status != EvidenceComplete || strings.TrimSpace(currentEvidence.Digest) == "" ||
			currentEvidence.Digest != evidence.Digest {
			return ApplyResult{}, fmt.Errorf("%w: evidence changed or is no longer complete", ErrInsufficientEvidence)
		}
	}
	riskDerived, restartDerived := false, false
	seen := make(map[string]struct{}, len(changes))
	for _, change := range changes {
		field, ok := s.registry.Field(change.Key)
		rollbackToUnapproved := change.AfterOptionID == "" && source == SourceRollback && reason == ReasonRollback &&
			ok && field.Descriptor.Default.Kind != settingmeta.StateValue
		if _, duplicate := seen[change.Key]; duplicate || !ok || field.Category != category || change.Category != category ||
			field.Descriptor.Control == settingmeta.ControlReadOnly || change.BeforeOptionID != current.Desired[change.Key] ||
			change.ApplyTiming != field.Descriptor.ApplyTiming || change.Safety != field.Descriptor.SafetyDirection ||
			(field.Descriptor.ValidateOption(change.AfterOptionID) != nil && !rollbackToUnapproved) {
			return ApplyResult{}, ErrCapabilityInvalid
		}
		seen[change.Key] = struct{}{}
		if field.Descriptor.SafetyDirection != settingmeta.SafetyNeutral {
			riskDerived = true
		}
		if field.Descriptor.ApplyTiming == settingmeta.ApplyNextEngineStart || field.Descriptor.ApplyTiming == settingmeta.ApplyNewPositionOnly {
			restartDerived = true
		}
	}
	if riskDerived != (riskRequired != 0) ||
		restartDerived != (restartRequired != 0) || (current.EffectiveEntry && !riskDerived) != (effectiveEntryAfter != 0) {
		return ApplyResult{}, ErrCapabilityInvalid
	}
	desired := cloneValues(current.Desired)
	effective := cloneValues(current.Effective)
	for _, change := range changes {
		field, _ := s.registry.Field(change.Key)
		rollbackToUnapproved := change.AfterOptionID == "" && source == SourceRollback && reason == ReasonRollback &&
			field.Descriptor.Default.Kind != settingmeta.StateValue
		if rollbackToUnapproved {
			delete(desired, change.Key)
		} else {
			desired[change.Key] = change.AfterOptionID
		}
		if change.ApplyTiming == settingmeta.ApplyImmediate && change.Safety == settingmeta.SafetyNeutral {
			if rollbackToUnapproved && field.Descriptor.Effective.Kind != settingmeta.StateValue {
				delete(effective, change.Key)
			} else {
				effective[change.Key] = change.AfterOptionID
			}
		}
	}
	auditID, err := s.randomToken(18)
	if err != nil {
		return ApplyResult{}, err
	}
	next := Snapshot{Version: current.Version + 1, EffectiveVersion: current.EffectiveVersion,
		Desired: desired, Effective: effective,
		EvidenceDigest: evidence.Digest, ActivationManifestDigest: current.ActivationManifestDigest,
		EffectiveEntry: effectiveEntryAfter != 0, RestartRequired: restartRequired != 0,
		Actor: actor, Reason: reason, AuditID: auditID, CreatedAt: now}
	if riskRequired != 0 {
		next.ActivationManifestDigest = ""
		next.EffectiveEntry = false
	}
	allImmediateNeutral := true
	for _, change := range changes {
		if change.ApplyTiming != settingmeta.ApplyImmediate || change.Safety != settingmeta.SafetyNeutral {
			allImmediateNeutral = false
			break
		}
	}
	if allImmediateNeutral {
		next.EffectiveVersion = next.Version
	}
	next.SettingsDigest = digestSnapshot(next)
	result, err := tx.ExecContext(ctx, `UPDATE optimization_control SET current_version=?,control_digest=?
		WHERE singleton=1 AND current_version=? AND control_digest=?`, next.Version,
		digestControlPointer(next.Version, next.SettingsDigest), current.Version,
		digestControlPointer(current.Version, current.SettingsDigest))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "locked") || strings.Contains(strings.ToLower(err.Error()), "busy") {
			return ApplyResult{}, ErrVersionConflict
		}
		return ApplyResult{}, err
	}
	changed, _ := result.RowsAffected()
	if changed != 1 {
		return ApplyResult{}, ErrVersionConflict
	}
	if err := insertSnapshot(ctx, tx, next); err != nil {
		return ApplyResult{}, err
	}
	var nextAuditRowID int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(id),0)+1 FROM optimization_audit`).Scan(&nextAuditRowID); err != nil {
		return ApplyResult{}, err
	}
	for _, change := range changes {
		event := AuditEvent{ID: nextAuditRowID, AuditID: auditID, Version: next.Version, CandidateID: candidateID,
			Key: change.Key, BeforeOptionID: change.BeforeOptionID, AfterOptionID: change.AfterOptionID,
			Actor: actor, Reason: reason, CreatedAt: now}
		if _, err := tx.ExecContext(ctx, `INSERT INTO optimization_audit
			(id,audit_id,version,candidate_id,setting_key,before_option_id,after_option_id,actor,reason,created_at,event_digest)
			VALUES(?,?,?,?,?,?,?,?,?,?,?)`, event.ID, event.AuditID, event.Version, event.CandidateID, event.Key,
			event.BeforeOptionID, event.AfterOptionID, event.Actor, event.Reason,
			event.CreatedAt.Format(time.RFC3339Nano), digestAuditEvent(event)); err != nil {
			return ApplyResult{}, err
		}
		nextAuditRowID++
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO optimization_applications(candidate_id,result_version,applied_at)
		VALUES(?,?,?)`, candidateID, next.Version, now.Format(time.RFC3339Nano)); err != nil {
		return ApplyResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return ApplyResult{}, err
	}
	return ApplyResult{Snapshot: cloneSnapshot(next)}, nil
}

func (s *Store) RecoverConflict(ctx context.Context, capability string) (ConflictView, error) {
	return s.recoverConflict(ctx, capability, s.actor)
}

func (s *Store) recoverConflict(ctx context.Context, capability, actor string) (ConflictView, error) {
	token := strings.TrimSpace(capability)
	actor = strings.TrimSpace(actor)
	if token == "" || actor == "" {
		return ConflictView{}, ErrCapabilityInvalid
	}
	var baseVersion uint64
	var candidateID, storedActor, categoryRaw, sourceRaw, reasonRaw, changesJSON, evidenceJSON string
	var notBeforeRaw, expiresRaw, createdRaw, payloadMAC string
	var riskRequired, restartRequired, effectiveEntryAfter int
	err := s.db.QueryRowContext(ctx, `SELECT candidate_id,actor,base_version,category,source,reason,changes_json,evidence_json,
		not_before,expires_at,created_at,risk_required,restart_required,effective_entry_after,payload_mac
		FROM optimization_candidates WHERE capability_hash=?`, hashCapability(token)).
		Scan(&candidateID, &storedActor, &baseVersion, &categoryRaw, &sourceRaw, &reasonRaw, &changesJSON, &evidenceJSON,
			&notBeforeRaw, &expiresRaw, &createdRaw, &riskRequired, &restartRequired, &effectiveEntryAfter, &payloadMAC)
	if errors.Is(err, sql.ErrNoRows) {
		return ConflictView{}, ErrCapabilityInvalid
	}
	if err != nil {
		return ConflictView{}, err
	}
	notBefore, notBeforeErr := parseStoredTime(notBeforeRaw)
	expiresAt, expiresErr := parseStoredTime(expiresRaw)
	createdAt, createdErr := parseStoredTime(createdRaw)
	payload := candidatePayload{CandidateID: candidateID, Actor: storedActor, BaseVersion: baseVersion,
		Category: Category(categoryRaw), Source: CandidateSource(sourceRaw), Reason: ReasonCode(reasonRaw),
		ChangesJSON: changesJSON, EvidenceJSON: evidenceJSON, NotBefore: notBefore, ExpiresAt: expiresAt,
		RiskRequired: riskRequired, RestartRequired: restartRequired, EffectiveEntryAfter: effectiveEntryAfter, CreatedAt: createdAt}
	if storedActor != actor || notBeforeErr != nil || expiresErr != nil || createdErr != nil ||
		(riskRequired != 0 && riskRequired != 1) || (restartRequired != 0 && restartRequired != 1) ||
		(effectiveEntryAfter != 0 && effectiveEntryAfter != 1) || !verifyCandidatePayloadMAC(token, payload, payloadMAC) {
		return ConflictView{}, ErrCapabilityInvalid
	}
	category, known := ParseCategory(categoryRaw)
	if !known || category == CategoryOverview || category == CategoryPerformanceHistory ||
		!validCandidateOrigin(CandidateSource(sourceRaw), ReasonCode(reasonRaw)) {
		return ConflictView{}, ErrCapabilityInvalid
	}
	var attempted []OptionChange
	if err := json.Unmarshal([]byte(changesJSON), &attempted); err != nil || len(attempted) == 0 {
		return ConflictView{}, ErrCapabilityInvalid
	}
	for _, change := range attempted {
		field, ok := s.registry.Field(change.Key)
		rollbackToUnapproved := change.AfterOptionID == "" && ReasonCode(reasonRaw) == ReasonRollback &&
			ok && field.Descriptor.Default.Kind != settingmeta.StateValue
		if !ok || field.Category != category || change.Category != category ||
			field.Descriptor.Control == settingmeta.ControlReadOnly || change.ApplyTiming != field.Descriptor.ApplyTiming ||
			change.Safety != field.Descriptor.SafetyDirection ||
			(field.Descriptor.ValidateOption(change.AfterOptionID) != nil && !rollbackToUnapproved) {
			return ConflictView{}, ErrCapabilityInvalid
		}
	}
	latest, err := s.currentSnapshot(ctx, s.db)
	if err != nil {
		return ConflictView{}, err
	}
	return ConflictView{BaseVersion: baseVersion, Category: category,
		Attempted: append([]OptionChange(nil), attempted...), Latest: latest, Registry: s.registry}, nil
}

func insertSnapshot(ctx context.Context, exec interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, snapshot Snapshot) error {
	desiredJSON, _ := json.Marshal(snapshot.Desired)
	effectiveJSON, _ := json.Marshal(snapshot.Effective)
	_, err := exec.ExecContext(ctx, `INSERT INTO optimization_snapshots
		(version,effective_version,desired_json,effective_json,settings_digest,evidence_digest,activation_manifest_digest,
		effective_entry,restart_required,actor,reason,audit_id,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		snapshot.Version, snapshot.EffectiveVersion, string(desiredJSON), string(effectiveJSON), snapshot.SettingsDigest,
		snapshot.EvidenceDigest, snapshot.ActivationManifestDigest, boolInt(snapshot.EffectiveEntry),
		boolInt(snapshot.RestartRequired), snapshot.Actor, snapshot.Reason, snapshot.AuditID,
		snapshot.CreatedAt.UTC().Format(time.RFC3339Nano))
	return err
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (s *Store) currentSnapshot(ctx context.Context, q queryer) (Snapshot, error) {
	var version uint64
	var controlDigest string
	if err := q.QueryRowContext(ctx, `SELECT current_version,control_digest FROM optimization_control WHERE singleton=1`).Scan(&version, &controlDigest); err != nil {
		return Snapshot{}, err
	}
	snapshot, err := s.snapshot(ctx, q, version)
	if err != nil {
		return Snapshot{}, err
	}
	if controlDigest != digestControlPointer(version, snapshot.SettingsDigest) {
		return Snapshot{}, errors.New("optimization: corrupt control pointer digest")
	}
	return snapshot, nil
}

func (s *Store) snapshot(ctx context.Context, q queryer, version uint64) (Snapshot, error) {
	out, err := scanSnapshot(q.QueryRowContext(ctx, `SELECT version,effective_version,desired_json,effective_json,settings_digest,evidence_digest,
		activation_manifest_digest,effective_entry,restart_required,actor,reason,audit_id,created_at
		FROM optimization_snapshots WHERE version=?`, version))
	if err != nil {
		return Snapshot{}, err
	}
	return out, nil
}

func (s *Store) history(ctx context.Context) ([]Snapshot, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT version,effective_version,desired_json,effective_json,settings_digest,evidence_digest,
		activation_manifest_digest,effective_entry,restart_required,actor,reason,audit_id,created_at
		FROM optimization_snapshots ORDER BY version DESC LIMIT ?`, readHistoryLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Snapshot, 0)
	for rows.Next() {
		snapshot, err := scanSnapshot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, snapshot)
	}
	return out, rows.Err()
}

func (s *Store) audit(ctx context.Context) ([]AuditEvent, error) {
	if err := s.validateAuditCoverage(ctx); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,audit_id,version,candidate_id,setting_key,before_option_id,
		after_option_id,actor,reason,created_at,event_digest FROM optimization_audit ORDER BY id DESC LIMIT ?`, readAuditLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AuditEvent
	for rows.Next() {
		event, eventDigest, err := scanAuditEventUnchecked(rows)
		if err != nil {
			return nil, err
		}
		if eventDigest == "" || eventDigest != digestAuditEvent(event) {
			return nil, errors.New("optimization: corrupt audit event digest")
		}
		out = append(out, event)
	}
	return out, rows.Err()
}

// validateAuditCoverage checks the entire durable audit ledger in one bounded
// aggregate query before audit() returns its newest-page projection. It makes a
// removed application or audit row visible even when the affected event is
// older than the read limit, while append-only triggers remain the primary
// write-time protection.
func (s *Store) validateAuditCoverage(ctx context.Context) error {
	var snapshotCount, applicationCount, invalidGroups int64
	err := s.db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM optimization_snapshots WHERE version > 1),
		(SELECT COUNT(*) FROM optimization_applications),
		EXISTS (
			SELECT 1 FROM optimization_applications ap
			JOIN optimization_candidates c ON c.candidate_id=ap.candidate_id
			JOIN optimization_snapshots s ON s.version=ap.result_version
			LEFT JOIN optimization_audit a ON a.version=ap.result_version AND a.candidate_id=ap.candidate_id
			GROUP BY ap.candidate_id,ap.result_version,c.changes_json,s.audit_id,s.actor,s.reason,s.created_at
			HAVING json_valid(c.changes_json)=0 OR json_type(c.changes_json)!='array'
				OR COUNT(a.id)!=json_array_length(c.changes_json)
				OR COALESCE(SUM(CASE WHEN a.id IS NULL THEN 0
					WHEN a.audit_id=s.audit_id AND a.actor=s.actor AND a.reason=s.reason AND a.created_at=s.created_at
					THEN 0 ELSE 1 END),0)!=0
		)
	`).Scan(&snapshotCount, &applicationCount, &invalidGroups)
	if err != nil {
		return fmt.Errorf("optimization: validating audit coverage: %w", err)
	}
	if snapshotCount != applicationCount || invalidGroups != 0 {
		return errors.New("optimization: corrupt audit coverage")
	}
	return nil
}

func scanAuditEventUnchecked(row snapshotScanner) (AuditEvent, string, error) {
	var event AuditEvent
	var created, eventDigest string
	if err := row.Scan(&event.ID, &event.AuditID, &event.Version, &event.CandidateID, &event.Key,
		&event.BeforeOptionID, &event.AfterOptionID, &event.Actor, &event.Reason, &created, &eventDigest); err != nil {
		return AuditEvent{}, "", err
	}
	createdAt, err := parseStoredTime(created)
	if err != nil || event.ID <= 0 || event.Version == 0 || strings.TrimSpace(event.AuditID) == "" ||
		strings.TrimSpace(event.CandidateID) == "" || strings.TrimSpace(event.Key) == "" || strings.TrimSpace(event.Actor) == "" ||
		(event.Reason != ReasonServerPreset && event.Reason != ReasonRollback) {
		return AuditEvent{}, "", errors.New("optimization: corrupt audit event")
	}
	event.CreatedAt = createdAt
	return event, eventDigest, nil
}

type snapshotScanner interface {
	Scan(...any) error
}

func scanSnapshot(row snapshotScanner) (Snapshot, error) {
	out, err := scanSnapshotUnchecked(row)
	if err != nil {
		return Snapshot{}, err
	}
	if out.SettingsDigest == "" || out.SettingsDigest != digestSnapshot(out) {
		return Snapshot{}, errors.New("optimization: corrupt snapshot digest")
	}
	return cloneSnapshot(out), nil
}

func scanSnapshotUnchecked(row snapshotScanner) (Snapshot, error) {
	var out Snapshot
	var desiredJSON, effectiveJSON, createdRaw string
	var effectiveEntry, restartRequired int
	if err := row.Scan(&out.Version, &out.EffectiveVersion, &desiredJSON, &effectiveJSON, &out.SettingsDigest,
		&out.EvidenceDigest, &out.ActivationManifestDigest, &effectiveEntry, &restartRequired, &out.Actor,
		&out.Reason, &out.AuditID, &createdRaw); err != nil {
		return Snapshot{}, err
	}
	if json.Unmarshal([]byte(desiredJSON), &out.Desired) != nil || json.Unmarshal([]byte(effectiveJSON), &out.Effective) != nil ||
		out.Desired == nil || out.Effective == nil || (effectiveEntry != 0 && effectiveEntry != 1) ||
		(restartRequired != 0 && restartRequired != 1) || out.Version == 0 || out.EffectiveVersion == 0 ||
		out.EffectiveVersion > out.Version || strings.TrimSpace(out.Actor) == "" || strings.TrimSpace(out.AuditID) == "" {
		return Snapshot{}, errors.New("optimization: corrupt snapshot")
	}
	createdAt, err := parseStoredTime(createdRaw)
	if err != nil {
		return Snapshot{}, errors.New("optimization: corrupt snapshot timestamp")
	}
	out.EffectiveEntry = effectiveEntry != 0
	out.RestartRequired = restartRequired != 0
	out.CreatedAt = createdAt
	return out, nil
}

func (s *Store) readEvidence(ctx context.Context) (Evidence, error) {
	if s.evidence == nil {
		return Evidence{Status: EvidenceUnavailable, Missing: []string{"a049-provider-unavailable"}}, ErrInsufficientEvidence
	}
	evidence, err := s.evidence.ReadEvidence(ctx)
	if err != nil {
		return Evidence{Status: EvidenceUnavailable, Missing: []string{"evidence-read-failed"}}, err
	}
	evidence.Missing = append([]string(nil), evidence.Missing...)
	return evidence, nil
}

func (s *Store) randomToken(bytes int) (string, error) {
	raw := make([]byte, bytes)
	if _, err := io.ReadFull(s.random, raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashCapability(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func signCandidatePayload(capability string, payload candidatePayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(capability))
	_, _ = mac.Write(raw)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func verifyCandidatePayloadMAC(capability string, payload candidatePayload, encoded string) bool {
	expected, err := signCandidatePayload(capability, payload)
	if err != nil {
		return false
	}
	actualBytes, err := hex.DecodeString(encoded)
	if err != nil {
		return false
	}
	expectedBytes, err := hex.DecodeString(expected)
	return err == nil && hmac.Equal(actualBytes, expectedBytes)
}

func validCandidateOrigin(source CandidateSource, reason ReasonCode) bool {
	switch source {
	case SourceServerPreset, SourceEvidence:
		return reason == ReasonServerPreset
	case SourceRollback:
		return reason == ReasonRollback
	default:
		return false
	}
}

func parseStoredTime(raw string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil || parsed.IsZero() || parsed.UTC().Format(time.RFC3339Nano) != raw {
		return time.Time{}, errors.New("invalid stored timestamp")
	}
	return parsed.UTC(), nil
}

func secureSQLiteFiles(path string) error {
	for _, file := range []string{path, path + "-wal", path + "-shm"} {
		if err := os.Chmod(file, 0o600); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func digestControlPointer(version uint64, snapshotDigest string) string {
	payload := struct {
		Domain         string `json:"domain"`
		CurrentVersion uint64 `json:"current_version"`
		SnapshotDigest string `json:"snapshot_digest"`
	}{Domain: "optimization-control-v1", CurrentVersion: version, SnapshotDigest: snapshotDigest}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func digestAuditEvent(event AuditEvent) string {
	payload := struct {
		Domain         string     `json:"domain"`
		ID             int64      `json:"id"`
		AuditID        string     `json:"audit_id"`
		Version        uint64     `json:"version"`
		CandidateID    string     `json:"candidate_id"`
		Key            string     `json:"setting_key"`
		BeforeOptionID string     `json:"before_option_id"`
		AfterOptionID  string     `json:"after_option_id"`
		Actor          string     `json:"actor"`
		Reason         ReasonCode `json:"reason"`
		CreatedAt      time.Time  `json:"created_at"`
	}{
		Domain: "optimization-audit-event-v1", ID: event.ID, AuditID: event.AuditID, Version: event.Version,
		CandidateID: event.CandidateID, Key: event.Key, BeforeOptionID: event.BeforeOptionID,
		AfterOptionID: event.AfterOptionID, Actor: event.Actor, Reason: event.Reason, CreatedAt: event.CreatedAt.UTC(),
	}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func digestSnapshot(snapshot Snapshot) string {
	payload := struct {
		Version                  uint64            `json:"version"`
		EffectiveVersion         uint64            `json:"effective_version"`
		Desired                  map[string]string `json:"desired"`
		Effective                map[string]string `json:"effective"`
		EvidenceDigest           string            `json:"evidence_digest"`
		ActivationManifestDigest string            `json:"activation_manifest_digest"`
		EffectiveEntry           bool              `json:"effective_entry"`
		RestartRequired          bool              `json:"restart_required"`
		Actor                    string            `json:"actor"`
		Reason                   ReasonCode        `json:"reason"`
		AuditID                  string            `json:"audit_id"`
		CreatedAt                time.Time         `json:"created_at"`
	}{
		Version: snapshot.Version, EffectiveVersion: snapshot.EffectiveVersion,
		Desired: snapshot.Desired, Effective: snapshot.Effective,
		EvidenceDigest: snapshot.EvidenceDigest, ActivationManifestDigest: snapshot.ActivationManifestDigest,
		EffectiveEntry: snapshot.EffectiveEntry, RestartRequired: snapshot.RestartRequired,
		Actor: snapshot.Actor, Reason: snapshot.Reason, AuditID: snapshot.AuditID, CreatedAt: snapshot.CreatedAt.UTC()}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func digestSnapshotV1(snapshot Snapshot) string {
	payload := struct {
		Version            uint64 `json:"version"`
		EffectiveVersion   uint64 `json:"effective_version"`
		Desired, Effective map[string]string
		EffectiveEntry     bool   `json:"effective_entry"`
		Manifest           string `json:"manifest"`
	}{
		Version: snapshot.Version, EffectiveVersion: snapshot.EffectiveVersion,
		Desired: snapshot.Desired, Effective: snapshot.Effective,
		EffectiveEntry: snapshot.EffectiveEntry, Manifest: snapshot.ActivationManifestDigest}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func cloneValues(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
func cloneSnapshot(in Snapshot) Snapshot {
	in.Desired = cloneValues(in.Desired)
	in.Effective = cloneValues(in.Effective)
	return in
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
