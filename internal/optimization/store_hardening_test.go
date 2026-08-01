package optimization_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
)

type mutableEvidenceProvider struct {
	mu       sync.Mutex
	evidence optimization.Evidence
	err      error
}

func (p *mutableEvidenceProvider) ReadEvidence(context.Context) (optimization.Evidence, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.evidence, p.err
}

func (p *mutableEvidenceProvider) set(evidence optimization.Evidence, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.evidence = evidence
	p.err = err
}

func hardeningPreview(t *testing.T, store *optimization.Store) optimization.Preview {
	t.Helper()
	view, err := store.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	preview, err := store.Preview(context.Background(), optimization.PreviewRequest{
		BaseVersion: view.Snapshot.Version, Category: optimization.CategoryExitProtection,
		Changes: map[string]string{"exit.common-policy": "SAFE"},
		Source:  optimization.SourceServerPreset, Reason: optimization.ReasonServerPreset,
	})
	if err != nil {
		t.Fatal(err)
	}
	return preview
}

func hardeningActorPreview(t *testing.T, commander optimization.Commander) optimization.Preview {
	t.Helper()
	view, err := commander.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	preview, err := commander.Preview(context.Background(), optimization.PreviewRequest{
		BaseVersion: view.Snapshot.Version, Category: optimization.CategoryExitProtection,
		Changes: map[string]string{"exit.common-policy": "SAFE"},
		Source:  optimization.SourceServerPreset, Reason: optimization.ReasonServerPreset,
	})
	if err != nil {
		t.Fatal(err)
	}
	return preview
}

func rawDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func dropCandidateTrigger(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec(`DROP TRIGGER optimization_candidates_no_update`); err != nil {
		t.Fatal(err)
	}
}

func TestApplyFailsClosedForTamperedCandidateMetadata(t *testing.T) {
	tests := []struct {
		name  string
		query string
		args  []any
	}{
		{"category", `UPDATE optimization_candidates SET category='overview' WHERE capability_hash=?`, nil},
		{"source", `UPDATE optimization_candidates SET source='forged' WHERE capability_hash=?`, nil},
		{"reason", `UPDATE optimization_candidates SET reason='forged' WHERE capability_hash=?`, nil},
		{"actor", `UPDATE optimization_candidates SET actor='verified:mallory' WHERE capability_hash=?`, nil},
		{"before", `UPDATE optimization_candidates SET changes_json=replace(changes_json, '"after_option_id":"SAFE"', '"before_option_id":"forged","after_option_id":"SAFE"') WHERE capability_hash=?`, nil},
		{"timing", `UPDATE optimization_candidates SET changes_json=replace(changes_json, 'new-position-only', 'immediate') WHERE capability_hash=?`, nil},
		{"safety", `UPDATE optimization_candidates SET changes_json=replace(changes_json, 'contextual', 'neutral') WHERE capability_hash=?`, nil},
		{"risk", `UPDATE optimization_candidates SET risk_required=0 WHERE capability_hash=?`, nil},
		{"restart", `UPDATE optimization_candidates SET restart_required=0 WHERE capability_hash=?`, nil},
		{"effective-entry", `UPDATE optimization_candidates SET effective_entry_after=1 WHERE capability_hash=?`, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
			path := filepath.Join(t.TempDir(), "optimization.db")
			store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
			preview := hardeningPreview(t, store)
			db := rawDB(t, path)
			dropCandidateTrigger(t, db)
			if _, err := db.Exec(tc.query, optimizationHash(preview.Capability)); err != nil {
				t.Fatal(err)
			}
			clk.Advance(3 * time.Second)
			if _, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); !errors.Is(err, optimization.ErrCapabilityInvalid) {
				t.Fatalf("Apply error = %v, want invalid capability", err)
			}
		})
	}
}

func TestPreviewRejectsMixedCategoryChangesWithoutPersistingCandidate(t *testing.T) {
	ctx := context.Background()
	registry, err := optimization.BuildRegistry(ctx,
		optimization.ProviderBinding{Category: optimization.CategoryExitProtection,
			Provider: optimization.StaticProvider{Owner: "a041", Fields: []settingmeta.FieldDescriptor{descriptor("a041", "exit.common-policy")}}},
		optimization.ProviderBinding{Category: optimization.CategoryPositionManagement,
			Provider: optimization.StaticProvider{Owner: "a042", Fields: []settingmeta.FieldDescriptor{descriptor("a042", "position.common-policy")}}},
	)
	if err != nil {
		t.Fatal(err)
	}
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	path := filepath.Join(t.TempDir(), "optimization.db")
	store, err := optimization.Open(ctx, optimization.Options{Path: path, Registry: registry, Clock: clk,
		Actor: "operator:test", Evidence: evidenceProvider{evidence: optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"}}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	view, err := store.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Preview(ctx, optimization.PreviewRequest{BaseVersion: view.Snapshot.Version,
		Category: optimization.CategoryExitProtection,
		Changes:  map[string]string{"exit.common-policy": "SAFE", "position.common-policy": "SAFE"},
		Source:   optimization.SourceServerPreset, Reason: optimization.ReasonServerPreset})
	if !errors.Is(err, optimization.ErrInvalidCandidate) {
		t.Fatalf("Preview error = %v, want invalid candidate", err)
	}
	var candidates int
	if err := rawDB(t, path).QueryRow(`SELECT COUNT(*) FROM optimization_candidates`).Scan(&candidates); err != nil || candidates != 0 {
		t.Fatalf("candidate rows=%d err=%v, want 0", candidates, err)
	}
}

func TestPreviewRejectsMalformedReadOnlyRegistryFieldWithoutPersistingCandidate(t *testing.T) {
	ctx := context.Background()
	bad := descriptor("a041", "exit.malformed-policy")
	bad.Options = nil // BuildRegistry exposes invalid descriptors as safe read-only fields.
	registry, err := optimization.BuildRegistry(ctx, optimization.ProviderBinding{
		Category: optimization.CategoryExitProtection,
		Provider: optimization.StaticProvider{Owner: "a041", Fields: []settingmeta.FieldDescriptor{bad}},
	})
	if err != nil {
		t.Fatal(err)
	}
	field, ok := registry.Field("exit.malformed-policy")
	if !ok || field.ConfigurationError == "" || field.Descriptor.Control != settingmeta.ControlReadOnly || len(field.Descriptor.Options) != 0 {
		t.Fatalf("malformed field was not safely exposed: %+v / found=%v", field, ok)
	}
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	path := filepath.Join(t.TempDir(), "optimization.db")
	store, err := optimization.Open(ctx, optimization.Options{Path: path, Registry: registry, Clock: clk,
		Actor: "operator:test", Evidence: evidenceProvider{evidence: optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"}}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	view, err := store.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Preview(ctx, optimization.PreviewRequest{BaseVersion: view.Snapshot.Version,
		Category: optimization.CategoryExitProtection, Changes: map[string]string{"exit.malformed-policy": "SAFE"},
		Source: optimization.SourceServerPreset, Reason: optimization.ReasonServerPreset})
	if !errors.Is(err, optimization.ErrInvalidCandidate) {
		t.Fatalf("Preview error = %v, want invalid candidate", err)
	}
	var candidates int
	if err := rawDB(t, path).QueryRow(`SELECT COUNT(*) FROM optimization_candidates`).Scan(&candidates); err != nil || candidates != 0 {
		t.Fatalf("candidate rows=%d err=%v, want 0", candidates, err)
	}
}

func TestApplyRejectsCorruptCandidateTimes(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	path := filepath.Join(t.TempDir(), "optimization.db")
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	preview := hardeningPreview(t, store)
	db := rawDB(t, path)
	dropCandidateTrigger(t, db)
	if _, err := db.Exec(`UPDATE optimization_candidates SET not_before='not-a-time' WHERE capability_hash=?`, optimizationHash(preview.Capability)); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); !errors.Is(err, optimization.ErrCapabilityInvalid) {
		t.Fatalf("Apply error = %v, want invalid capability", err)
	}
}

func TestApplyRejectsCandidateScheduleThatBypassesRiskWait(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	path := filepath.Join(t.TempDir(), "optimization.db")
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	preview := hardeningPreview(t, store)
	db := rawDB(t, path)
	dropCandidateTrigger(t, db)
	if _, err := db.Exec(`UPDATE optimization_candidates SET not_before=created_at WHERE capability_hash=?`, optimizationHash(preview.Capability)); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); !errors.Is(err, optimization.ErrCapabilityInvalid) {
		t.Fatalf("Apply error = %v, want invalid capability", err)
	}
}

func TestApplyRejectsSelfConsistentCandidatePayloadTamperWithoutMAC(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	path := filepath.Join(t.TempDir(), "optimization.db")
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	preview := hardeningPreview(t, store)
	db := rawDB(t, path)
	dropCandidateTrigger(t, db)
	// This keeps category, reason, evidence, changes and all derived timing/risk
	// fields valid. Only a holder of the opaque capability can recompute its MAC.
	if _, err := db.Exec(`UPDATE optimization_candidates SET source='evidence-backed' WHERE capability_hash=?`, optimizationHash(preview.Capability)); err != nil {
		t.Fatal(err)
	}
	clk.Advance(3 * time.Second)
	if _, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); !errors.Is(err, optimization.ErrCapabilityInvalid) {
		t.Fatalf("Apply error = %v, want invalid capability", err)
	}
}

func TestApplyMACBindsCandidateIDRawPayloadAndReplayIntegers(t *testing.T) {
	for _, tc := range []struct{ name, statement string }{
		{"candidate-id", `UPDATE optimization_candidates SET candidate_id='forged' WHERE capability_hash=?`},
		{"unknown-json-field", `UPDATE optimization_candidates SET changes_json=json_set(changes_json, '$[0].forged', 'x') WHERE capability_hash=?`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
			path := filepath.Join(t.TempDir(), "optimization.db")
			store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
			preview := hardeningPreview(t, store)
			db := rawDB(t, path)
			dropCandidateTrigger(t, db)
			if _, err := db.Exec(tc.statement, optimizationHash(preview.Capability)); err != nil {
				t.Fatal(err)
			}
			clk.Advance(3 * time.Second)
			if _, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); !errors.Is(err, optimization.ErrCapabilityInvalid) {
				t.Fatalf("Apply error = %v, want invalid capability", err)
			}
		})
	}

	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	path := filepath.Join(t.TempDir(), "optimization.db")
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	preview := hardeningPreview(t, store)
	clk.Advance(3 * time.Second)
	if _, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); err != nil {
		t.Fatal(err)
	}
	db := rawDB(t, path)
	dropCandidateTrigger(t, db)
	if _, err := db.Exec(`UPDATE optimization_candidates SET risk_required=2 WHERE capability_hash=?`, optimizationHash(preview.Capability)); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); !errors.Is(err, optimization.ErrCapabilityInvalid) {
		t.Fatalf("replay error = %v, want invalid capability", err)
	}
}

func TestMigrationAddsActorAndPayloadMACToEmptyLegacyCandidateSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "optimization.db")
	db := rawDB(t, path)
	if _, err := db.Exec(`CREATE TABLE optimization_candidates (
		candidate_id TEXT PRIMARY KEY, base_version INTEGER NOT NULL, category TEXT NOT NULL, source TEXT NOT NULL,
		reason TEXT NOT NULL, changes_json TEXT NOT NULL, evidence_json TEXT NOT NULL, capability_hash TEXT NOT NULL UNIQUE,
		not_before TEXT NOT NULL, expires_at TEXT NOT NULL, risk_required INTEGER NOT NULL, restart_required INTEGER NOT NULL,
		effective_entry_after INTEGER NOT NULL, created_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	if _, err := store.Read(context.Background()); err != nil {
		t.Fatal(err)
	}
	check, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer check.Close()
	var columns int
	if err := check.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('optimization_candidates') WHERE name IN ('actor','payload_mac')`).Scan(&columns); err != nil || columns != 2 {
		t.Fatalf("hardened candidate columns=%d err=%v", columns, err)
	}
}

func TestMigrationFromV1VerifiesAndExpandsSnapshotDigest(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "optimization.db")
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	view, err := store.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	legacyDigest := legacySnapshotDigest(view.Snapshot)
	db := rawDB(t, path)
	if _, err := db.Exec(`DROP TRIGGER optimization_snapshots_no_update`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE optimization_snapshots SET settings_digest=? WHERE version=1`, legacyDigest); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`PRAGMA user_version=1`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	migrated, err := optimization.Open(ctx, optimization.Options{Path: path, Registry: testRegistry(t), Clock: clk,
		Actor: "operator:test", Evidence: evidenceProvider{evidence: optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"}}})
	if err != nil {
		t.Fatal(err)
	}
	defer migrated.Close()
	migratedView, err := migrated.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if migratedView.Snapshot.SettingsDigest == legacyDigest {
		t.Fatal("v1 snapshot digest was not expanded during migration")
	}
	var version int
	if err := rawDB(t, path).QueryRow(`PRAGMA user_version`).Scan(&version); err != nil || version != 3 {
		t.Fatalf("user_version=%d err=%v", version, err)
	}
}

func TestMigrationRejectsCorruptV1SnapshotDigest(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "optimization.db")
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	db := rawDB(t, path)
	if _, err := db.Exec(`DROP TRIGGER optimization_snapshots_no_update`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE optimization_snapshots SET settings_digest='forged' WHERE version=1`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`PRAGMA user_version=1`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	opened, err := optimization.Open(ctx, optimization.Options{Path: path, Registry: testRegistry(t), Clock: clk,
		Actor: "operator:test", Evidence: evidenceProvider{evidence: optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"}}})
	if opened != nil {
		_ = opened.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "legacy digest") {
		t.Fatalf("Open error=%v, want corrupt legacy digest rejection", err)
	}
}

func TestMigrationFromV2BindsControlPointerAndLegacyAuditRows(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "optimization.db")
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	preview := hardeningPreview(t, store)
	clk.Advance(3 * time.Second)
	if _, err := store.Apply(ctx, optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	db := rawDB(t, path)
	if _, err := db.Exec(`ALTER TABLE optimization_control DROP COLUMN control_digest`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`ALTER TABLE optimization_audit DROP COLUMN event_digest`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`PRAGMA user_version=2`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	migrated, err := optimization.Open(ctx, optimization.Options{Path: path, Registry: testRegistry(t), Clock: clk,
		Actor: "operator:test", Evidence: evidenceProvider{evidence: optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"}}})
	if err != nil {
		t.Fatal(err)
	}
	defer migrated.Close()
	view, err := migrated.Read(ctx)
	if err != nil || view.Snapshot.Version != 2 || len(view.Audit) != 1 {
		t.Fatalf("migrated view=%+v err=%v", view, err)
	}
	check := rawDB(t, path)
	var controlDigest, eventDigest string
	if err := check.QueryRow(`SELECT control_digest FROM optimization_control WHERE singleton=1`).Scan(&controlDigest); err != nil || controlDigest == "" {
		t.Fatalf("control digest=%q err=%v", controlDigest, err)
	}
	if err := check.QueryRow(`SELECT event_digest FROM optimization_audit`).Scan(&eventDigest); err != nil || eventDigest == "" {
		t.Fatalf("event digest=%q err=%v", eventDigest, err)
	}
}

func TestMigrationRejectsCorruptLegacyAuditRow(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "optimization.db")
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	preview := hardeningPreview(t, store)
	clk.Advance(3 * time.Second)
	if _, err := store.Apply(ctx, optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	db := rawDB(t, path)
	if _, err := db.Exec(`DROP TRIGGER optimization_audit_no_update`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE optimization_audit SET actor='verified:mallory'`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`ALTER TABLE optimization_control DROP COLUMN control_digest`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`ALTER TABLE optimization_audit DROP COLUMN event_digest`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`PRAGMA user_version=2`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	opened, err := optimization.Open(ctx, optimization.Options{Path: path, Registry: testRegistry(t), Clock: clk,
		Actor: "operator:test", Evidence: evidenceProvider{evidence: optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"}}})
	if opened != nil {
		_ = opened.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "does not match its snapshot") {
		t.Fatalf("Open error=%v, want corrupt legacy audit rejection", err)
	}
}

func TestSnapshotAndAuditCorruptionFailClosed(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	path := filepath.Join(t.TempDir(), "optimization.db")
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	preview := hardeningPreview(t, store)
	clk.Advance(3 * time.Second)
	if _, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); err != nil {
		t.Fatal(err)
	}
	db := rawDB(t, path)
	for _, trigger := range []string{"optimization_snapshots_no_update", "optimization_audit_no_update"} {
		if _, err := db.Exec(`DROP TRIGGER ` + trigger); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`UPDATE optimization_snapshots SET settings_digest='forged' WHERE version=2`); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Read(context.Background()); err == nil {
		t.Fatal("Read accepted corrupt snapshot digest")
	}
}

func TestAuditTimestampCorruptionFailsClosed(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	path := filepath.Join(t.TempDir(), "optimization.db")
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	preview := hardeningPreview(t, store)
	clk.Advance(3 * time.Second)
	if _, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); err != nil {
		t.Fatal(err)
	}
	db := rawDB(t, path)
	if _, err := db.Exec(`DROP TRIGGER optimization_audit_no_update`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE optimization_audit SET created_at='not-a-time'`); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Read(context.Background()); err == nil {
		t.Fatal("Read accepted corrupt audit timestamp")
	}
}

func TestSchemaIsVersionedAndAppendOnly(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	path := filepath.Join(t.TempDir(), "optimization.db")
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	preview := hardeningPreview(t, store)
	clk.Advance(3 * time.Second)
	if _, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); err != nil {
		t.Fatal(err)
	}
	db := rawDB(t, path)
	var version int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil || version != 3 {
		t.Fatalf("user_version=%d err=%v, want 3", version, err)
	}
	var hardenedColumns int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('optimization_candidates') WHERE name IN ('actor','payload_mac')`).Scan(&hardenedColumns); err != nil || hardenedColumns != 2 {
		t.Fatalf("hardened candidate columns=%d err=%v", hardenedColumns, err)
	}
	var controlDigestColumns, auditDigestColumns int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('optimization_control') WHERE name='control_digest'`).Scan(&controlDigestColumns); err != nil || controlDigestColumns != 1 {
		t.Fatalf("control digest columns=%d err=%v", controlDigestColumns, err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('optimization_audit') WHERE name='event_digest'`).Scan(&auditDigestColumns); err != nil || auditDigestColumns != 1 {
		t.Fatalf("audit digest columns=%d err=%v", auditDigestColumns, err)
	}
	for _, statement := range []string{
		`UPDATE optimization_snapshots SET actor='forged' WHERE version=1`,
		`DELETE FROM optimization_snapshots WHERE version=1`,
		`UPDATE optimization_candidates SET reason='forged'`,
		`DELETE FROM optimization_candidates`,
		`UPDATE optimization_applications SET result_version=1`,
		`DELETE FROM optimization_applications`,
		`UPDATE optimization_audit SET actor='forged'`,
		`DELETE FROM optimization_audit`,
	} {
		if _, err := db.Exec(statement); err == nil || !strings.Contains(err.Error(), "append-only") {
			t.Fatalf("tamper statement succeeded or had wrong error: %q / %v", statement, err)
		}
	}
}

func TestOpenRefusesNewerSchemaAndSecuresFiles(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "control")
	if err := os.Mkdir(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(parent, "optimization.db")
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	for _, file := range []string{parent, path, path + "-wal", path + "-shm"} {
		info, err := os.Stat(file)
		if errors.Is(err, os.ErrNotExist) && (strings.HasSuffix(file, "-wal") || strings.HasSuffix(file, "-shm")) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		want := os.FileMode(0o600)
		if file == parent {
			want = 0o700
		}
		if info.Mode().Perm() != want {
			t.Fatalf("%s mode=%#o want %#o", file, info.Mode().Perm(), want)
		}
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	db := rawDB(t, path)
	if _, err := db.Exec(`PRAGMA user_version=999`); err != nil {
		t.Fatal(err)
	}
	if _, err := optimization.Open(context.Background(), optimization.Options{Path: path, Registry: testRegistry(t), Clock: clk, Actor: "operator:test"}); err == nil {
		t.Fatal("Open accepted a newer schema")
	}
}

func TestCloseIsNilSafeAndForActorDoesNotMutateStoreActor(t *testing.T) {
	var nilStore *optimization.Store
	if err := nilStore.Close(); err != nil {
		t.Fatalf("nil Close = %v", err)
	}
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	store := openStore(t, filepath.Join(t.TempDir(), "optimization.db"), clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	first, err := store.ForActor("verified:alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ForActor("   "); err == nil {
		t.Fatal("ForActor accepted blank actor")
	}
	preview := hardeningActorPreview(t, first)
	clk.Advance(3 * time.Second)
	if _, err := first.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); err != nil {
		t.Fatal(err)
	}
	view, err := store.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if view.Snapshot.Actor != "verified:alice" || view.Audit[0].Actor != "verified:alice" {
		t.Fatalf("actor was not bound to wrapper: %+v / %+v", view.Snapshot, view.Audit)
	}
}

func TestCandidateCapabilityIsBoundToPreviewActorAcrossApplyRecoveryAndReplay(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	store := openStore(t, filepath.Join(t.TempDir(), "optimization.db"), clk,
		optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	alice, err := store.ForActor("verified:alice")
	if err != nil {
		t.Fatal(err)
	}
	bob, err := store.ForActor("verified:bob")
	if err != nil {
		t.Fatal(err)
	}

	alicePreview := hardeningActorPreview(t, alice)
	clk.Advance(3 * time.Second)
	if _, err := bob.Apply(ctx, optimization.ApplyRequest{Capability: alicePreview.Capability, Confirmed: true}); !errors.Is(err, optimization.ErrCapabilityInvalid) {
		t.Fatalf("cross-actor Apply error=%v, want invalid capability", err)
	}
	if _, err := bob.RecoverConflict(ctx, alicePreview.Capability); !errors.Is(err, optimization.ErrCapabilityInvalid) {
		t.Fatalf("cross-actor RecoverConflict error=%v, want invalid capability", err)
	}
	applied, err := alice.Apply(ctx, optimization.ApplyRequest{Capability: alicePreview.Capability, Confirmed: true})
	if err != nil {
		t.Fatal(err)
	}
	if applied.Snapshot.Actor != "verified:alice" {
		t.Fatalf("applied actor=%q", applied.Snapshot.Actor)
	}
	if _, err := bob.Apply(ctx, optimization.ApplyRequest{Capability: alicePreview.Capability, Confirmed: true}); !errors.Is(err, optimization.ErrCapabilityInvalid) {
		t.Fatalf("cross-actor replay error=%v, want invalid capability", err)
	}
	replay, err := alice.Apply(ctx, optimization.ApplyRequest{Capability: alicePreview.Capability, Confirmed: true})
	if err != nil || !replay.Replayed || replay.Snapshot.Actor != "verified:alice" {
		t.Fatalf("original-actor replay=%+v err=%v", replay, err)
	}
}

func TestRollbackPreviewAndConflictRecoveryUseTheSameActor(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	store := openStore(t, filepath.Join(t.TempDir(), "optimization.db"), clk,
		optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	alice, _ := store.ForActor("verified:alice")
	bob, _ := store.ForActor("verified:bob")

	first := hardeningActorPreview(t, alice)
	clk.Advance(3 * time.Second)
	firstResult, err := alice.Apply(ctx, optimization.ApplyRequest{Capability: first.Capability, Confirmed: true})
	if err != nil {
		t.Fatal(err)
	}
	rollback, err := alice.PreviewRollback(ctx, optimization.RollbackPreviewRequest{
		BaseVersion: firstResult.Snapshot.Version, TargetVersion: 1, Category: optimization.CategoryExitProtection,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bob.Apply(ctx, optimization.ApplyRequest{Capability: rollback.Capability, Confirmed: true}); !errors.Is(err, optimization.ErrCapabilityInvalid) {
		t.Fatalf("cross-actor rollback Apply error=%v", err)
	}

	// Move the current version so the rollback capability has a conflict to recover.
	bobPreview, err := bob.Preview(ctx, optimization.PreviewRequest{
		BaseVersion: firstResult.Snapshot.Version, Category: optimization.CategoryExitProtection,
		Changes: map[string]string{"exit.common-policy": "BALANCED"},
		Source:  optimization.SourceServerPreset, Reason: optimization.ReasonServerPreset,
	})
	if err != nil {
		t.Fatal(err)
	}
	clk.Advance(3 * time.Second)
	if _, err := bob.Apply(ctx, optimization.ApplyRequest{Capability: bobPreview.Capability, Confirmed: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := alice.Apply(ctx, optimization.ApplyRequest{Capability: rollback.Capability, Confirmed: true}); !errors.Is(err, optimization.ErrVersionConflict) {
		t.Fatalf("rollback Apply error=%v, want version conflict", err)
	}
	if _, err := bob.RecoverConflict(ctx, rollback.Capability); !errors.Is(err, optimization.ErrCapabilityInvalid) {
		t.Fatalf("cross-actor rollback recovery error=%v", err)
	}
	conflict, err := alice.RecoverConflict(ctx, rollback.Capability)
	if err != nil || conflict.BaseVersion != firstResult.Snapshot.Version {
		t.Fatalf("original-actor rollback recovery=%+v err=%v", conflict, err)
	}
}

func TestDirectStoreCandidatesRemainBoundToTheStoreActor(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	store := openStore(t, filepath.Join(t.TempDir(), "optimization.db"), clk,
		optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	alice, _ := store.ForActor("verified:alice")
	directPreview := hardeningPreview(t, store)
	clk.Advance(3 * time.Second)
	if _, err := alice.Apply(ctx, optimization.ApplyRequest{Capability: directPreview.Capability, Confirmed: true}); !errors.Is(err, optimization.ErrCapabilityInvalid) {
		t.Fatalf("wrapper used direct-store capability: %v", err)
	}
	if _, err := store.Apply(ctx, optimization.ApplyRequest{Capability: directPreview.Capability, Confirmed: true}); err != nil {
		t.Fatal(err)
	}
}

func TestEvidenceBackedApplyRevalidatesCurrentEvidenceBeforeMutation(t *testing.T) {
	tests := []struct {
		name     string
		evidence optimization.Evidence
		err      error
	}{
		{name: "digest changed", evidence: optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v2"}},
		{name: "link missing", evidence: optimization.Evidence{Status: optimization.EvidenceInsufficient, Digest: "evidence-v1", Missing: []string{"link_missing"}}},
		{name: "stale", evidence: optimization.Evidence{Status: optimization.EvidenceStale, Digest: "evidence-v1", Missing: []string{"stale"}}},
		{name: "provider error", evidence: optimization.Evidence{Status: optimization.EvidenceUnavailable}, err: errors.New("provider unavailable")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
			provider := &mutableEvidenceProvider{evidence: optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"}}
			store, err := optimization.Open(ctx, optimization.Options{Path: filepath.Join(t.TempDir(), "optimization.db"),
				Registry: testRegistry(t), Evidence: provider, Clock: clk, Actor: "operator:test"})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = store.Close() })
			view, err := store.Read(ctx)
			if err != nil {
				t.Fatal(err)
			}
			preview, err := store.Preview(ctx, optimization.PreviewRequest{BaseVersion: view.Snapshot.Version,
				Category: optimization.CategoryExitProtection, Changes: map[string]string{"exit.common-policy": "SAFE"},
				Source: optimization.SourceEvidence, Reason: optimization.ReasonServerPreset})
			if err != nil {
				t.Fatal(err)
			}
			provider.set(tc.evidence, tc.err)
			clk.Advance(3 * time.Second)
			if _, err := store.Apply(ctx, optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); !errors.Is(err, optimization.ErrInsufficientEvidence) {
				t.Fatalf("Apply error=%v, want insufficient evidence", err)
			}
			current, err := store.Read(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if current.Snapshot.Version != view.Snapshot.Version {
				t.Fatalf("snapshot mutated from %d to %d", view.Snapshot.Version, current.Snapshot.Version)
			}
		})
	}
}

func TestSnapshotDigestCoversEveryPersistedImmutableMetadataField(t *testing.T) {
	tests := []struct {
		name, assignment string
	}{
		{"effective-version", `effective_version=2`},
		{"desired", `desired_json=json_set(desired_json, '$.forged', 'x')`},
		{"effective", `effective_json=json_set(effective_json, '$.forged', 'x')`},
		{"evidence", `evidence_digest='forged'`},
		{"manifest", `activation_manifest_digest='forged'`},
		{"effective-entry", `effective_entry=1`},
		{"restart-required", `restart_required=0`},
		{"actor", `actor='forged'`},
		{"reason", `reason='operator-rollback'`},
		{"audit-id", `audit_id='forged'`},
		{"created-at", `created_at='2026-08-01T00:00:04Z'`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
			path := filepath.Join(t.TempDir(), "optimization.db")
			store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
			preview := hardeningPreview(t, store)
			clk.Advance(3 * time.Second)
			if _, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); err != nil {
				t.Fatal(err)
			}
			db := rawDB(t, path)
			if _, err := db.Exec(`DROP TRIGGER optimization_snapshots_no_update`); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`UPDATE optimization_snapshots SET ` + tc.assignment + ` WHERE version=2`); err != nil {
				t.Fatal(err)
			}
			if _, err := store.Read(context.Background()); err == nil || !strings.Contains(err.Error(), "snapshot") {
				t.Fatalf("Read error=%v, want corrupt snapshot", err)
			}
		})
	}
}

func TestOpenRejectsSameNameNoOpOrDriftedAppendOnlyTrigger(t *testing.T) {
	for _, tc := range []struct {
		name, definition string
	}{
		{"no-op", `CREATE TRIGGER optimization_candidates_no_update BEFORE UPDATE ON optimization_candidates BEGIN SELECT 1; END`},
		{"drifted-event", `CREATE TRIGGER optimization_candidates_no_update BEFORE DELETE ON optimization_candidates BEGIN SELECT RAISE(ABORT, 'optimization candidates are append-only'); END`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			path := filepath.Join(t.TempDir(), "optimization.db")
			clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
			store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
			if err := store.Close(); err != nil {
				t.Fatal(err)
			}
			db := rawDB(t, path)
			if _, err := db.Exec(`DROP TRIGGER optimization_candidates_no_update`); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(tc.definition); err != nil {
				t.Fatal(err)
			}
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			opened, err := optimization.Open(ctx, optimization.Options{Path: path, Registry: testRegistry(t), Clock: clk,
				Actor: "operator:test", Evidence: evidenceProvider{evidence: optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"}}})
			if opened != nil {
				_ = opened.Close()
			}
			if err == nil || !strings.Contains(err.Error(), "trigger") {
				t.Fatalf("Open error=%v, want trigger definition rejection", err)
			}
		})
	}
}

func TestControlPointerRollbackTamperFailsReadPreviewAndApply(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	path := filepath.Join(t.TempDir(), "optimization.db")
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	first := hardeningPreview(t, store)
	clk.Advance(3 * time.Second)
	firstResult, err := store.Apply(ctx, optimization.ApplyRequest{Capability: first.Capability, Confirmed: true})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Preview(ctx, optimization.PreviewRequest{
		BaseVersion: firstResult.Snapshot.Version, Category: optimization.CategoryExitProtection,
		Changes: map[string]string{"exit.common-policy": "BALANCED"},
		Source:  optimization.SourceServerPreset, Reason: optimization.ReasonServerPreset,
	})
	if err != nil {
		t.Fatal(err)
	}
	db := rawDB(t, path)
	if _, err := db.Exec(`UPDATE optimization_control SET current_version=1 WHERE singleton=1`); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Read(ctx); err == nil {
		t.Fatal("Read accepted rolled-back control pointer")
	}
	if _, err := store.Apply(ctx, optimization.ApplyRequest{Capability: first.Capability, Confirmed: true}); err == nil {
		t.Fatal("idempotent replay accepted rolled-back control pointer")
	}
	if _, err := store.Preview(ctx, optimization.PreviewRequest{BaseVersion: 1,
		Category: optimization.CategoryExitProtection, Changes: map[string]string{"exit.common-policy": "RUNNER"},
		Source: optimization.SourceServerPreset, Reason: optimization.ReasonServerPreset}); err == nil {
		t.Fatal("Preview accepted rolled-back control pointer")
	}
	clk.Advance(3 * time.Second)
	if _, err := store.Apply(ctx, optimization.ApplyRequest{Capability: second.Capability, Confirmed: true}); err == nil {
		t.Fatal("Apply accepted rolled-back control pointer")
	}
}

func TestReadRejectsDeletedAuditRowsIncludingPartialMultiChangeDeletion(t *testing.T) {
	tests := []struct {
		name       string
		fieldCount int
		deleteSQL  string
	}{
		{name: "complete deletion", fieldCount: 1, deleteSQL: `DELETE FROM optimization_audit`},
		{name: "partial multi-change deletion", fieldCount: 2, deleteSQL: `DELETE FROM optimization_audit WHERE setting_key='exit.secondary-policy'`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			fields := []settingmeta.FieldDescriptor{descriptor("a041", "exit.common-policy")}
			changes := map[string]string{"exit.common-policy": "SAFE"}
			if tc.fieldCount == 2 {
				fields = append(fields, descriptor("a041", "exit.secondary-policy"))
				changes["exit.secondary-policy"] = "SAFE"
			}
			registry, err := optimization.BuildRegistry(ctx, optimization.ProviderBinding{
				Category: optimization.CategoryExitProtection,
				Provider: optimization.StaticProvider{Owner: "a041", Fields: fields},
			})
			if err != nil {
				t.Fatal(err)
			}
			clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
			path := filepath.Join(t.TempDir(), "optimization.db")
			store, err := optimization.Open(ctx, optimization.Options{Path: path, Registry: registry, Clock: clk,
				Actor: "operator:test", Evidence: evidenceProvider{evidence: optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"}}})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = store.Close() })
			view, err := store.Read(ctx)
			if err != nil {
				t.Fatal(err)
			}
			preview, err := store.Preview(ctx, optimization.PreviewRequest{BaseVersion: view.Snapshot.Version,
				Category: optimization.CategoryExitProtection, Changes: changes,
				Source: optimization.SourceServerPreset, Reason: optimization.ReasonServerPreset})
			if err != nil {
				t.Fatal(err)
			}
			clk.Advance(3 * time.Second)
			if _, err := store.Apply(ctx, optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); err != nil {
				t.Fatal(err)
			}
			db := rawDB(t, path)
			if _, err := db.Exec(`DROP TRIGGER optimization_audit_no_delete`); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(tc.deleteSQL); err != nil {
				t.Fatal(err)
			}
			if _, err := store.Read(ctx); err == nil || !strings.Contains(err.Error(), "audit") {
				t.Fatalf("Read error=%v, want deleted audit rejection", err)
			}
		})
	}
}

func TestAuditEventDigestRejectsValidLookingPersistedTampering(t *testing.T) {
	for _, tc := range []struct {
		name, assignment string
	}{
		{"actor", `actor='verified:mallory'`},
		{"reason", `reason='operator-rollback'`},
		{"before", `before_option_id='RUNNER'`},
		{"after", `after_option_id='BALANCED'`},
		{"candidate", `candidate_id='forged-candidate'`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
			path := filepath.Join(t.TempDir(), "optimization.db")
			store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
			preview := hardeningPreview(t, store)
			clk.Advance(3 * time.Second)
			if _, err := store.Apply(ctx, optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); err != nil {
				t.Fatal(err)
			}
			db := rawDB(t, path)
			if _, err := db.Exec(`DROP TRIGGER optimization_audit_no_update`); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`UPDATE optimization_audit SET ` + tc.assignment); err != nil {
				t.Fatal(err)
			}
			if _, err := store.Read(ctx); err == nil || !strings.Contains(err.Error(), "audit") {
				t.Fatalf("Read error=%v, want corrupt audit rejection", err)
			}
		})
	}
}

func optimizationHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func legacySnapshotDigest(snapshot optimization.Snapshot) string {
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
