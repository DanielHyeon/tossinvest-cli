package optimization

import (
	"context"
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
	"time"

	_ "modernc.org/sqlite"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
)

const (
	DatabaseFileName = "optimization-control.db"
	previewTTL       = 5 * time.Minute
	riskWait         = 3 * time.Second
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
	registry Registry
	evidence EvidenceProvider
	clock    clock.Clock
	actor    string
	random   io.Reader
}

func Open(ctx context.Context, opts Options) (*Store, error) {
	path := strings.TrimSpace(opts.Path)
	if path == "" {
		return nil, errors.New("optimization: database path is required")
	}
	if strings.TrimSpace(opts.Actor) == "" {
		return nil, errors.New("optimization: actor is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("optimization: creating control directory: %w", err)
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
	if err := os.Chmod(path, 0o600); err != nil {
		db.Close()
		return nil, fmt.Errorf("optimization: securing control database: %w", err)
	}
	store := &Store{db: db, registry: opts.Registry, evidence: opts.Evidence,
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
	return store, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) init(ctx context.Context) error {
	const schema = `
CREATE TABLE IF NOT EXISTS optimization_control (
  singleton INTEGER PRIMARY KEY CHECK(singleton=1), current_version INTEGER NOT NULL
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
  evidence_json TEXT NOT NULL, capability_hash TEXT NOT NULL UNIQUE,
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
  after_option_id TEXT NOT NULL, actor TEXT NOT NULL, reason TEXT NOT NULL, created_at TEXT NOT NULL
);
INSERT OR IGNORE INTO optimization_control(singleton,current_version) VALUES(1,0);`
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("optimization: creating lifecycle schema: %w", err)
	}
	return s.ensureInitialSnapshot(ctx)
}

func (s *Store) ensureInitialSnapshot(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var version uint64
	if err := tx.QueryRowContext(ctx, `SELECT current_version FROM optimization_control WHERE singleton=1`).Scan(&version); err != nil {
		return err
	}
	if version != 0 {
		return tx.Commit()
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
	result, err := tx.ExecContext(ctx, `UPDATE optimization_control SET current_version=1 WHERE singleton=1 AND current_version=0`)
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
	return s.preview(ctx, req, "")
}

func (s *Store) PreviewRollback(ctx context.Context, req RollbackPreviewRequest) (Preview, error) {
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
		Changes: changes, Source: SourceRollback, Reason: ReasonRollback}, "rollback:"+strconv.FormatUint(req.TargetVersion, 10))
}

func (s *Store) preview(ctx context.Context, req PreviewRequest, salt string) (Preview, error) {
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
	_, err = s.db.ExecContext(ctx, `INSERT INTO optimization_candidates
		(candidate_id,base_version,category,source,reason,changes_json,evidence_json,capability_hash,
		not_before,expires_at,risk_required,restart_required,effective_entry_after,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, candidateID, req.BaseVersion, req.Category, req.Source, req.Reason,
		string(changesJSON), string(evidenceJSON), hashCapability(token), notBefore.Format(time.RFC3339Nano),
		preview.ExpiresAt.Format(time.RFC3339Nano), boolInt(riskRequired), boolInt(restartRequired),
		boolInt(preview.EffectiveEntryAfterApply), now.Format(time.RFC3339Nano))
	if err != nil {
		return Preview{}, fmt.Errorf("optimization: storing immutable candidate: %w", err)
	}
	return preview, nil
}

func (s *Store) Apply(ctx context.Context, req ApplyRequest) (ApplyResult, error) {
	token := strings.TrimSpace(req.Capability)
	if token == "" {
		return ApplyResult{}, ErrCapabilityInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ApplyResult{}, err
	}
	defer tx.Rollback()
	row := tx.QueryRowContext(ctx, `SELECT c.candidate_id,c.base_version,c.reason,c.changes_json,c.evidence_json,
		c.not_before,c.expires_at,c.risk_required,c.restart_required,c.effective_entry_after,a.result_version
		FROM optimization_candidates c LEFT JOIN optimization_applications a ON a.candidate_id=c.candidate_id
		WHERE c.capability_hash=?`, hashCapability(token))
	var candidateID, reasonRaw, changesJSON, evidenceJSON, notBeforeRaw, expiresRaw string
	var baseVersion uint64
	var riskRequired, restartRequired, effectiveEntryAfter int
	var resultVersion sql.NullInt64
	if err := row.Scan(&candidateID, &baseVersion, &reasonRaw, &changesJSON, &evidenceJSON,
		&notBeforeRaw, &expiresRaw, &riskRequired, &restartRequired, &effectiveEntryAfter, &resultVersion); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ApplyResult{}, ErrCapabilityInvalid
		}
		return ApplyResult{}, err
	}
	if resultVersion.Valid {
		snapshot, err := s.snapshot(ctx, tx, uint64(resultVersion.Int64))
		if err != nil {
			return ApplyResult{}, err
		}
		return ApplyResult{Snapshot: snapshot, Replayed: true}, tx.Commit()
	}
	notBefore, _ := time.Parse(time.RFC3339Nano, notBeforeRaw)
	expiresAt, _ := time.Parse(time.RFC3339Nano, expiresRaw)
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
	current, err := s.currentSnapshot(ctx, tx)
	if err != nil {
		return ApplyResult{}, err
	}
	if current.Version != baseVersion {
		return ApplyResult{}, ErrVersionConflict
	}
	var changes []OptionChange
	if err := json.Unmarshal([]byte(changesJSON), &changes); err != nil {
		return ApplyResult{}, ErrCapabilityInvalid
	}
	var evidence Evidence
	if err := json.Unmarshal([]byte(evidenceJSON), &evidence); err != nil {
		return ApplyResult{}, ErrCapabilityInvalid
	}
	desired := cloneValues(current.Desired)
	effective := cloneValues(current.Effective)
	for _, change := range changes {
		field, ok := s.registry.Field(change.Key)
		rollbackToUnapproved := change.AfterOptionID == "" && ReasonCode(reasonRaw) == ReasonRollback &&
			ok && field.Descriptor.Default.Kind != settingmeta.StateValue
		if !ok || field.Category != change.Category ||
			(field.Descriptor.ValidateOption(change.AfterOptionID) != nil && !rollbackToUnapproved) {
			return ApplyResult{}, ErrCapabilityInvalid
		}
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
		Actor: s.actor, Reason: ReasonCode(reasonRaw), AuditID: auditID, CreatedAt: now}
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
	result, err := tx.ExecContext(ctx, `UPDATE optimization_control SET current_version=?
		WHERE singleton=1 AND current_version=?`, next.Version, current.Version)
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
	for _, change := range changes {
		if _, err := tx.ExecContext(ctx, `INSERT INTO optimization_audit
			(audit_id,version,candidate_id,setting_key,before_option_id,after_option_id,actor,reason,created_at)
			VALUES(?,?,?,?,?,?,?,?,?)`, auditID, next.Version, candidateID, change.Key, change.BeforeOptionID,
			change.AfterOptionID, s.actor, reasonRaw, now.Format(time.RFC3339Nano)); err != nil {
			return ApplyResult{}, err
		}
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
	token := strings.TrimSpace(capability)
	if token == "" {
		return ConflictView{}, ErrCapabilityInvalid
	}
	var baseVersion uint64
	var categoryRaw, reasonRaw, changesJSON string
	err := s.db.QueryRowContext(ctx, `SELECT base_version,category,reason,changes_json
		FROM optimization_candidates WHERE capability_hash=?`, hashCapability(token)).
		Scan(&baseVersion, &categoryRaw, &reasonRaw, &changesJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return ConflictView{}, ErrCapabilityInvalid
	}
	if err != nil {
		return ConflictView{}, err
	}
	category, known := ParseCategory(categoryRaw)
	if !known {
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
	if err := q.QueryRowContext(ctx, `SELECT current_version FROM optimization_control WHERE singleton=1`).Scan(&version); err != nil {
		return Snapshot{}, err
	}
	return s.snapshot(ctx, q, version)
}

func (s *Store) snapshot(ctx context.Context, q queryer, version uint64) (Snapshot, error) {
	var out Snapshot
	var desiredJSON, effectiveJSON, createdRaw string
	var effectiveEntry, restartRequired int
	err := q.QueryRowContext(ctx, `SELECT version,effective_version,desired_json,effective_json,settings_digest,evidence_digest,
		activation_manifest_digest,effective_entry,restart_required,actor,reason,audit_id,created_at
		FROM optimization_snapshots WHERE version=?`, version).Scan(&out.Version, &out.EffectiveVersion, &desiredJSON, &effectiveJSON,
		&out.SettingsDigest, &out.EvidenceDigest, &out.ActivationManifestDigest, &effectiveEntry,
		&restartRequired, &out.Actor, &out.Reason, &out.AuditID, &createdRaw)
	if err != nil {
		return Snapshot{}, err
	}
	if json.Unmarshal([]byte(desiredJSON), &out.Desired) != nil || json.Unmarshal([]byte(effectiveJSON), &out.Effective) != nil {
		return Snapshot{}, errors.New("optimization: corrupt snapshot")
	}
	out.EffectiveEntry = effectiveEntry != 0
	out.RestartRequired = restartRequired != 0
	out.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdRaw)
	return cloneSnapshot(out), nil
}

func (s *Store) history(ctx context.Context) ([]Snapshot, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT version FROM optimization_snapshots ORDER BY version DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var versions []uint64
	for rows.Next() {
		var version uint64
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]Snapshot, 0, len(versions))
	for _, version := range versions {
		snapshot, err := s.snapshot(ctx, s.db, version)
		if err != nil {
			return nil, err
		}
		out = append(out, snapshot)
	}
	return out, nil
}

func (s *Store) audit(ctx context.Context) ([]AuditEvent, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,audit_id,version,candidate_id,setting_key,before_option_id,
		after_option_id,actor,reason,created_at FROM optimization_audit ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AuditEvent
	for rows.Next() {
		var event AuditEvent
		var created string
		if err := rows.Scan(&event.ID, &event.AuditID, &event.Version, &event.CandidateID, &event.Key,
			&event.BeforeOptionID, &event.AfterOptionID, &event.Actor, &event.Reason, &created); err != nil {
			return nil, err
		}
		event.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		out = append(out, event)
	}
	return out, rows.Err()
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

func digestSnapshot(snapshot Snapshot) string {
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
