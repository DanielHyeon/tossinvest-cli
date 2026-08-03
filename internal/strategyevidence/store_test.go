package strategyevidence

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestStoreAppendIsIdempotentAndQuarantinesRevisionConflict(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := openTestStore(t)
	h := validHeader(marketclock.MarketUS, KindUSParticipation, "ev-1", "rev-1")
	first := mustEnvelope(t, h, `{"value":1}`)
	result, err := store.Append(ctx, first)
	if err != nil || result.Idempotent {
		t.Fatalf("first append = %+v, %v", result, err)
	}
	result, err = store.Append(ctx, mustEnvelope(t, h, `{ "value": 1 }`))
	if err != nil || !result.Idempotent || result.Evidence.PayloadDigest() != first.PayloadDigest() {
		t.Fatalf("idempotent append = %+v, %v", result, err)
	}

	conflictHeader := h
	conflictHeader.EvidenceID = "ev-conflict"
	_, err = store.Append(ctx, mustEnvelope(t, conflictHeader, `{"value":2}`))
	if !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("want ErrRevisionConflict, got %v", err)
	}
	if got, err := store.QuarantineCount(ctx); err != nil || got != 1 {
		t.Fatalf("quarantine count = %d, %v", got, err)
	}
	if got, err := store.EvidenceCount(ctx); err != nil || got != 1 {
		t.Fatalf("evidence count = %d, %v", got, err)
	}
	var quarantined []byte
	if err := store.db.QueryRowContext(ctx, `SELECT attempted_payload FROM revision_conflicts`).Scan(&quarantined); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(quarantined), `"value":2`) || string(quarantined) != `{"redacted":true}` {
		t.Fatalf("quarantine retained attempted payload: %s", quarantined)
	}
}

func TestRevisionConflictUsesInjectedClock(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fixed := time.Date(2026, 8, 3, 14, 15, 16, 123, time.UTC)
	store, err := Open(ctx, Options{
		Path:  filepath.Join(t.TempDir(), "evidence.db"),
		Clock: marketclock.NewFake(fixed),
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	h := validHeader(marketclock.MarketUS, KindUSParticipation, "accepted", "rev-1")
	if _, err := store.Append(ctx, mustEnvelope(t, h, `{"value":1}`)); err != nil {
		t.Fatal(err)
	}
	h.EvidenceID = "attempted"
	if _, err := store.Append(ctx, mustEnvelope(t, h, `{"value":2}`)); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("want ErrRevisionConflict, got %v", err)
	}

	var quarantinedAt string
	if err := store.db.QueryRowContext(ctx, `SELECT quarantined_at FROM revision_conflicts`).Scan(&quarantinedAt); err != nil {
		t.Fatal(err)
	}
	if got, want := quarantinedAt, stamp(fixed); got != want {
		t.Fatalf("quarantined_at = %q, want %q", got, want)
	}
}

func TestStoreDualCutoffAndAppendOnlyCorrectionSnapshot(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	oldHeader := validHeader(marketclock.MarketUS, KindUSParticipation, "ev-old", "rev-1")
	clk := marketclock.NewFake(oldHeader.IngestedAt)
	store := openTestStoreWithClock(t, clk)
	old := mustEnvelope(t, oldHeader, `{"value":1}`)
	if _, err := store.Append(ctx, old); err != nil {
		t.Fatal(err)
	}
	newHeader := oldHeader
	newHeader.EvidenceID = "ev-new"
	newHeader.RevisionIdentity = "rev-2"
	newHeader.SupersedesRevisionIdentity = "rev-1"
	newHeader.SourceAvailableAt = oldHeader.SourceAvailableAt.Add(2 * time.Hour)
	newHeader.ObservedAt = newHeader.SourceAvailableAt.Add(time.Minute)
	newHeader.IngestedAt = newHeader.SourceAvailableAt.Add(3 * time.Hour)
	newHeader.SourceEventAt = newHeader.SourceAvailableAt.Add(-time.Minute)
	newHeader.MarketEffectiveDate = "2026-08-03"
	clk.Set(newHeader.IngestedAt)
	newer := mustEnvelope(t, newHeader, `{"value":2}`)
	if _, err := store.Append(ctx, newer); err != nil {
		t.Fatal(err)
	}

	query := SnapshotQuery{Market: marketclock.MarketUS, Symbol: "AAPL", IssuerIdentity: oldHeader.IssuerIdentity, IssuerMappingVersion: oldHeader.IssuerMappingVersion, EvaluationAt: oldHeader.SourceAvailableAt.Add(3 * time.Hour), IngestionCutoff: oldHeader.IngestedAt.Add(time.Hour)}
	beforeIngest, err := store.SealSnapshot(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	assertSnapshotIDs(t, beforeIngest, "ev-old")
	query.IngestionCutoff = newHeader.IngestedAt
	after, err := store.SealSnapshot(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	assertSnapshotIDs(t, after, "ev-new")
	replayed, err := store.SealSnapshot(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	if replayed.ID != after.ID || replayed.Digest != after.Digest {
		t.Fatalf("snapshot not deterministic: %+v != %+v", replayed, after)
	}
}

func TestStoreRejectsCorrectionThatChangesEvidenceScope(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := openTestStore(t)
	originalHeader := validHeader(marketclock.MarketUS, KindUSParticipation, "original", "rev-1")
	if _, err := store.Append(ctx, mustEnvelope(t, originalHeader, `{"value":1}`)); err != nil {
		t.Fatal(err)
	}
	correctionHeader := originalHeader
	correctionHeader.EvidenceID = "invalid-correction"
	correctionHeader.RevisionIdentity = "rev-2"
	correctionHeader.SupersedesRevisionIdentity = "rev-1"
	correctionHeader.Symbol = "MSFT"
	_, err := store.Append(ctx, mustEnvelope(t, correctionHeader, `{"value":2}`))
	var refusal *ValidationError
	if !errors.As(err, &refusal) || refusal.Code != RefusalIdentityMismatch {
		t.Fatalf("want %s, got %v", RefusalIdentityMismatch, err)
	}
	if count, countErr := store.EvidenceCount(ctx); countErr != nil || count != 1 {
		t.Fatalf("evidence count = %d, err = %v", count, countErr)
	}
}

func TestStoreDualCutoffNanosecondBoundaries(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cutoff := time.Date(2026, 8, 3, 13, 0, 0, 123456789, time.UTC)
	store := openTestStoreWithClock(t, marketclock.NewFake(cutoff))
	h := validHeader(marketclock.MarketUS, KindUSParticipation, "boundary", "rev-1")
	h.SourceEventAt = cutoff.Add(-time.Minute)
	h.SourceAvailableAt = cutoff
	h.ObservedAt = cutoff
	h.IngestedAt = cutoff
	if _, err := store.Append(ctx, mustEnvelope(t, h, `{"value":1}`)); err != nil {
		t.Fatal(err)
	}

	exact, err := store.SealSnapshot(ctx, SnapshotQuery{Market: h.Market, Symbol: h.Symbol, IssuerIdentity: h.IssuerIdentity, IssuerMappingVersion: h.IssuerMappingVersion, EvaluationAt: cutoff, IngestionCutoff: cutoff})
	if err != nil {
		t.Fatal(err)
	}
	assertSnapshotIDs(t, exact, "boundary")
	beforeAvailable, err := store.SealSnapshot(ctx, SnapshotQuery{Market: h.Market, Symbol: h.Symbol, IssuerIdentity: h.IssuerIdentity, IssuerMappingVersion: h.IssuerMappingVersion, EvaluationAt: cutoff.Add(-time.Nanosecond), IngestionCutoff: cutoff})
	if err != nil {
		t.Fatal(err)
	}
	assertSnapshotIDs(t, beforeAvailable)
	beforeIngested, err := store.SealSnapshot(ctx, SnapshotQuery{Market: h.Market, Symbol: h.Symbol, IssuerIdentity: h.IssuerIdentity, IssuerMappingVersion: h.IssuerMappingVersion, EvaluationAt: cutoff, IngestionCutoff: cutoff.Add(-time.Nanosecond)})
	if err != nil {
		t.Fatal(err)
	}
	assertSnapshotIDs(t, beforeIngested)
}

func TestStoreExcludesFutureMarketDateEvenWhenIngested(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h := validHeader(marketclock.MarketUS, KindUSParticipation, "future", "rev-1")
	h.MarketEffectiveDate = "2026-08-04"
	h.SourceEventAt = time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	h.SourceAvailableAt = h.SourceEventAt.Add(time.Minute)
	h.ObservedAt = h.SourceAvailableAt.Add(time.Minute)
	h.IngestedAt = h.SourceAvailableAt.Add(2 * time.Minute)
	store := openTestStoreWithClock(t, marketclock.NewFake(h.IngestedAt))
	if _, err := store.Append(ctx, mustEnvelope(t, h, `{"future":true}`)); err != nil {
		t.Fatal(err)
	}
	snap, err := store.SealSnapshot(ctx, SnapshotQuery{
		Market: marketclock.MarketUS, Symbol: "AAPL",
		IssuerIdentity: h.IssuerIdentity, IssuerMappingVersion: h.IssuerMappingVersion,
		EvaluationAt:    time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC),
		IngestionCutoff: time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Items) != 0 {
		t.Fatalf("future evidence leaked: %+v", snap.Items)
	}
}

func TestStoreKeepsKRAndUSSnapshotsIndependent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := openTestStore(t)
	us := validHeader(marketclock.MarketUS, KindUSParticipation, "us-evidence", "us-rev")
	kr := validHeader(marketclock.MarketKR, KindKRNetFlow, "kr-evidence", "kr-rev")
	kr.Authority = AuthorityKRX
	kr.SourceRecordID = "kr-record"
	kr.Currency = "KRW"
	for _, evidence := range []Envelope{mustEnvelope(t, us, `{"market":"us"}`), mustEnvelope(t, kr, `{"market":"kr"}`)} {
		if _, err := store.Append(ctx, evidence); err != nil {
			t.Fatal(err)
		}
	}
	for _, tt := range []struct {
		market marketclock.Market
		want   string
	}{
		{marketclock.MarketKR, "kr-evidence"},
		{marketclock.MarketUS, "us-evidence"},
	} {
		snapshot, err := store.SealSnapshot(ctx, SnapshotQuery{
			Market: tt.market, Symbol: "AAPL",
			IssuerIdentity: us.IssuerIdentity, IssuerMappingVersion: us.IssuerMappingVersion,
			EvaluationAt:    us.SourceAvailableAt.Add(time.Hour),
			IngestionCutoff: us.IngestedAt.Add(time.Hour),
		})
		if err != nil {
			t.Fatal(err)
		}
		assertSnapshotIDs(t, snapshot, tt.want)
	}
}

func TestStoreSchemaTooNewAndAppendOnlyTriggers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := openTestStore(t)
	h := validHeader(marketclock.MarketKR, KindKRNetFlow, "ev", "rev")
	h.Authority = AuthorityKRX
	h.Currency = "KRW"
	if _, err := store.Append(ctx, mustEnvelope(t, h, `{"flow":1}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.ExecContext(ctx, `UPDATE evidence_records SET symbol='MUTATED' WHERE evidence_id='ev'`); err == nil {
		t.Fatal("append-only UPDATE unexpectedly succeeded")
	}
	path := store.Path()
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	db := openRawSQLite(t, path)
	if _, err := db.Exec(`PRAGMA user_version = 999`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	if _, err := Open(ctx, Options{Path: path}); !errors.Is(err, ErrSchemaTooNew) {
		t.Fatalf("want ErrSchemaTooNew, got %v", err)
	}
}

func TestStoreSchemaHasCompositeSnapshotItemKeyAndForeignKeys(t *testing.T) {
	t.Parallel()
	store := openTestStore(t)
	rows, err := store.db.Query(`PRAGMA table_info(snapshot_items)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	primaryKeyColumns := 0
	for rows.Next() {
		var cid, notNull, primaryKeyOrdinal int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKeyOrdinal); err != nil {
			t.Fatal(err)
		}
		if primaryKeyOrdinal > 0 {
			primaryKeyColumns++
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if primaryKeyColumns != 2 {
		t.Fatalf("snapshot_items primary-key columns = %d, want 2", primaryKeyColumns)
	}
	var foreignKeys int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM pragma_foreign_key_list('snapshot_items')`).Scan(&foreignKeys); err != nil {
		t.Fatal(err)
	}
	if foreignKeys != 2 {
		t.Fatalf("snapshot_items foreign keys = %d, want 2", foreignKeys)
	}
}

func openTestStore(t *testing.T) *Store {
	t.Helper()
	return openTestStoreWithClock(t, marketclock.NewFake(time.Date(2026, 8, 3, 13, 2, 0, 0, time.UTC)))
}

func openTestStoreWithClock(t *testing.T, clk marketclock.Clock) *Store {
	t.Helper()
	store, err := Open(context.Background(), Options{Path: filepath.Join(t.TempDir(), "evidence.db"), Clock: clk})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func mustEnvelope(t *testing.T, h Header, payload string) Envelope {
	t.Helper()
	envelope, err := NewEnvelope(h, []byte(payload))
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	return envelope
}

func assertSnapshotIDs(t *testing.T, snapshot Snapshot, want ...string) {
	t.Helper()
	if len(snapshot.Items) != len(want) {
		t.Fatalf("items = %d, want %d: %+v", len(snapshot.Items), len(want), snapshot.Items)
	}
	for i := range want {
		if snapshot.Items[i].Header().EvidenceID != want[i] {
			t.Fatalf("item %d = %s, want %s", i, snapshot.Items[i].Header().EvidenceID, want[i])
		}
	}
}

func openRawSQLite(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("raw sqlite open: %v", err)
	}
	return db
}
