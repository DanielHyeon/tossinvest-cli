package performance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"testing"
	"time"
)

func TestMaximumAttributionGenerationIntegrityScanMeetsP95Target(t *testing.T) {
	if testing.Short() {
		t.Skip("maximum attribution generation performance gate")
	}
	if raceEnabled {
		t.Skip("wall-clock p95 is verified by the non-instrumented suite")
	}
	store := openTestStore(t)
	rows := make([]Attribution, MaxAttributionGenerationRows)
	for index := range rows {
		rows[index] = unavailableFixture("US", fmt.Sprintf("T%05d", index), fmt.Sprintf("position-%05d", index))
	}
	if err := store.PersistAttributionRebuild(context.Background(), AttributionRebuild{
		ID: "maximum-generation", AccountRef: "acct-maximum", CalculatedAt: attributionAt, Unavailable: rows,
	}); err != nil {
		t.Fatal(err)
	}
	durations := make([]time.Duration, 7)
	for index := range durations {
		started := time.Now()
		got, err := store.AttributionRows(context.Background(), "acct-maximum", AttributionQuery{
			Market: "US", Ticker: "T09999", IncludeLinkMissing: true,
		}, 1)
		if err != nil || len(got) != 1 || got[0].Key.PositionID != "position-09999" {
			t.Fatalf("maximum generation query=%+v err=%v", got, err)
		}
		durations[index] = time.Since(started)
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	p95 := durations[(len(durations)*95+99)/100-1]
	t.Logf("10k-row attribution integrity scan p95=%s runs=%v", p95, durations)
	if p95 > 2*time.Second {
		t.Fatalf("10k-row attribution integrity scan p95=%s target<=2s runs=%v", p95, durations)
	}
}

func TestPersistAttributionRebuildRejectsDuplicateUnavailableAttributionKey(t *testing.T) {
	store := openTestStore(t)
	row := unavailableFixture("US", "AAPL", "position-1")
	err := store.PersistAttributionRebuild(context.Background(), AttributionRebuild{
		ID: "duplicate", AccountRef: "acct-1", CalculatedAt: attributionAt,
		Unavailable: []Attribution{row, row},
	})
	if !errors.Is(err, ErrDuplicateAttribution) {
		t.Fatalf("duplicate attribution err=%v", err)
	}
}

func TestPersistAttributionRebuildRejectsCompleteUnavailableAttributionKeyCollision(t *testing.T) {
	store := openTestStore(t)
	position, events := stagedFixture()
	derived, err := BuildDerivedAttributionStore([]PositionEvidence{position}, events)
	if err != nil {
		t.Fatal(err)
	}
	complete, err := derived.QueryAttribution(AttributionQuery{Market: "US", IncludeLinkMissing: true})
	if err != nil || len(complete) != 1 {
		t.Fatalf("complete=%+v err=%v", complete, err)
	}
	unavailable := unavailableFixture(complete[0].Key.Market, complete[0].Key.Ticker, complete[0].Key.PositionID)
	unavailable.Key = complete[0].Key
	unavailable.ObservedLineage[0].LaneID, unavailable.ObservedLineage[0].LaneVersion = complete[0].Key.LaneID, complete[0].Key.LaneVersion
	unavailable.ObservedLineage[0].CampaignID, unavailable.ObservedLineage[0].LegID = complete[0].Key.CampaignID, complete[0].Key.LegID
	unavailable.ObservedLineage[0].PolicyID, unavailable.ObservedLineage[0].PolicyVersion = complete[0].Key.PolicyID, complete[0].Key.PolicyVersion
	unavailable.MissingLineage = missingObservedLineage(unavailable.ObservedLineage[0])
	err = store.PersistAttributionRebuild(context.Background(), AttributionRebuild{
		ID: "collision", AccountRef: "acct-1", CalculatedAt: attributionAt,
		Positions: []PositionEvidence{position}, FillDeltas: events, Unavailable: []Attribution{unavailable},
	})
	if !errors.Is(err, ErrDuplicateAttribution) {
		t.Fatalf("complete/unavailable collision err=%v", err)
	}
}

func TestUnavailableAttributionRejectsObservedLineageScopeMismatch(t *testing.T) {
	base := unavailableFixture("KR", "005930", "position-kr")
	base.Key.CampaignID, base.Key.LegID = "campaign-kr", "leg-kr"
	base.ObservedLineage = []CompositeLineage{{Market: "KR", Ticker: "005930", LaneID: "lane", LaneVersion: "lane/v1",
		CampaignID: "campaign-kr", LegID: "leg-kr", PositionID: "position-kr", PolicyID: "policy", PolicyVersion: "policy/v1"}}
	base.MissingLineage = missingObservedLineage(base.ObservedLineage[0])
	cases := map[string]func(*CompositeLineage){
		"market":         func(lineage *CompositeLineage) { lineage.Market = "US" },
		"ticker":         func(lineage *CompositeLineage) { lineage.Ticker = "AAPL" },
		"lane":           func(lineage *CompositeLineage) { lineage.LaneID = "other" },
		"lane-version":   func(lineage *CompositeLineage) { lineage.LaneVersion = "lane/v2" },
		"campaign":       func(lineage *CompositeLineage) { lineage.CampaignID = "other" },
		"leg":            func(lineage *CompositeLineage) { lineage.LegID = "other" },
		"position":       func(lineage *CompositeLineage) { lineage.PositionID = "position-us" },
		"policy":         func(lineage *CompositeLineage) { lineage.PolicyID = "other" },
		"policy-version": func(lineage *CompositeLineage) { lineage.PolicyVersion = "policy/v2" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			row := cloneAttribution(base)
			mutate(&row.ObservedLineage[0])
			if err := validateUnavailableAttribution(row); !errors.Is(err, ErrLineageConflict) {
				t.Fatalf("scope mismatch err=%v", err)
			}
		})
	}
}

func TestUnavailableAttributionMissingListMatchesObservedLineage(t *testing.T) {
	row := unavailableFixture("US", "AAPL", "position-1")
	row.ObservedLineage = []CompositeLineage{{Market: "US", Ticker: "AAPL", LaneID: "lane", LaneVersion: "lane/v1",
		PositionID: "position-1", PolicyID: "policy", PolicyVersion: "policy/v1"}}
	row.MissingLineage = []string{"campaign_id"}
	if err := validateUnavailableAttribution(row); err == nil {
		t.Fatal("incomplete missing-lineage list accepted")
	}
}

func TestUnavailableAttributionRequiresObservedLineage(t *testing.T) {
	row := unavailableFixture("US", "AAPL", "position-1")
	row.ObservedLineage = nil
	if err := validateUnavailableAttribution(row); err == nil {
		t.Fatal("unavailable attribution without observed lineage accepted")
	}
}

func TestAttributionRowsRejectsShadowColumnTampering(t *testing.T) {
	store := openTestStore(t)
	rebuild := AttributionRebuild{ID: "tamper", AccountRef: "acct-1", CalculatedAt: attributionAt,
		Unavailable: []Attribution{unavailableFixture("KR", "005930", "position-1")}}
	if err := store.PersistAttributionRebuild(context.Background(), rebuild); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`UPDATE attribution_rows SET market='US' WHERE account_ref='acct-1'`); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AttributionRows(context.Background(), "acct-1", AttributionQuery{Market: "US", IncludeLinkMissing: true}, 10); !errors.Is(err, ErrImmutableConflict) {
		t.Fatalf("shadow tamper err=%v", err)
	}
}

func TestAttributionRowsRejectsDeletedGenerationRows(t *testing.T) {
	store := openTestStore(t)
	rebuild := AttributionRebuild{ID: "delete", AccountRef: "acct-1", CalculatedAt: attributionAt,
		Unavailable: []Attribution{unavailableFixture("KR", "005930", "position-1")}}
	if err := store.PersistAttributionRebuild(context.Background(), rebuild); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`DELETE FROM attribution_rows WHERE account_ref='acct-1'`); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AttributionRows(context.Background(), "acct-1", AttributionQuery{Market: "KR", IncludeLinkMissing: true}, 10); !errors.Is(err, ErrImmutableConflict) {
		t.Fatalf("deleted generation err=%v", err)
	}
}

func TestAttributionRowsRejectsDuplicatedGenerationRows(t *testing.T) {
	store := openTestStore(t)
	rebuild := AttributionRebuild{ID: "duplicate-row", AccountRef: "acct-1", CalculatedAt: attributionAt,
		Unavailable: []Attribution{unavailableFixture("KR", "005930", "position-1")}}
	if err := store.PersistAttributionRebuild(context.Background(), rebuild); err != nil {
		t.Fatal(err)
	}
	var envelope attributionRowEnvelope
	if err := store.db.QueryRow(`SELECT account_ref,rebuild_id,ordinal,market,ticker,
		COALESCE(lane_id,''),COALESCE(lane_version,''),COALESCE(campaign_id,''),COALESCE(leg_id,''),
		position_id,COALESCE(policy_id,''),COALESCE(policy_version,''),lineage_status,row_json
		FROM attribution_rows WHERE account_ref='acct-1'`).Scan(
		&envelope.AccountRef, &envelope.RebuildID, &envelope.Ordinal, &envelope.Market, &envelope.Ticker,
		&envelope.LaneID, &envelope.LaneVersion, &envelope.CampaignID, &envelope.LegID,
		&envelope.PositionID, &envelope.PolicyID, &envelope.PolicyVersion, &envelope.LineageStatus,
		&envelope.RowJSON); err != nil {
		t.Fatal(err)
	}
	envelope.Ordinal, envelope.Ticker = 1, "000660"
	if _, err := store.db.Exec(`INSERT INTO attribution_rows
		(account_ref,rebuild_id,ordinal,market,ticker,lane_id,lane_version,campaign_id,leg_id,
		position_id,policy_id,policy_version,lineage_status,row_digest,row_json)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, envelope.AccountRef, envelope.RebuildID, envelope.Ordinal,
		envelope.Market, envelope.Ticker, nullable(envelope.LaneID), nullable(envelope.LaneVersion),
		nullable(envelope.CampaignID), nullable(envelope.LegID), envelope.PositionID,
		nullable(envelope.PolicyID), nullable(envelope.PolicyVersion), envelope.LineageStatus,
		attributionRowDigest(envelope), envelope.RowJSON); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AttributionRows(context.Background(), "acct-1", AttributionQuery{Market: "KR", IncludeLinkMissing: true}, 10); !errors.Is(err, ErrImmutableConflict) {
		t.Fatalf("duplicated generation err=%v", err)
	}
}

func TestSameIDReplayRevalidatesPersistedGeneration(t *testing.T) {
	store := openTestStore(t)
	rebuild := AttributionRebuild{ID: "replay", AccountRef: "acct-1", CalculatedAt: attributionAt,
		Unavailable: []Attribution{unavailableFixture("KR", "005930", "position-1")}}
	if err := store.PersistAttributionRebuild(context.Background(), rebuild); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`UPDATE attribution_rows SET ticker='000660' WHERE account_ref='acct-1'`); err != nil {
		t.Fatal(err)
	}
	if err := store.PersistAttributionRebuild(context.Background(), rebuild); !errors.Is(err, ErrImmutableConflict) {
		t.Fatalf("same-ID replay err=%v", err)
	}
}

func TestAttributionEvidenceRejectsHeadEnvelopeMismatch(t *testing.T) {
	store := openTestStore(t)
	rebuild := AttributionRebuild{ID: "head", AccountRef: "acct-1", CalculatedAt: attributionAt,
		Unavailable: []Attribution{unavailableFixture("US", "AAPL", "position-1")}}
	if err := store.PersistAttributionRebuild(context.Background(), rebuild); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`UPDATE attribution_rebuilds SET calculated_at='2099-01-01T00:00:00Z' WHERE account_ref='acct-1'`); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AttributionEvidence(context.Background(), "acct-1"); !errors.Is(err, ErrImmutableConflict) {
		t.Fatalf("head envelope err=%v", err)
	}
}

func TestAttributionRebuildCanonicalReplayIsOrderIndependent(t *testing.T) {
	store := openTestStore(t)
	position, events := stagedFixture()
	first := AttributionRebuild{ID: "canonical", AccountRef: " acct-1 ", CalculatedAt: attributionAt,
		Positions: []PositionEvidence{position}, FillDeltas: events}
	if err := store.PersistAttributionRebuild(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	slices.Reverse(events)
	second := AttributionRebuild{ID: " canonical ", AccountRef: "acct-1", CalculatedAt: attributionAt,
		Positions: []PositionEvidence{position}, FillDeltas: events}
	if err := store.PersistAttributionRebuild(context.Background(), second); err != nil {
		t.Fatalf("canonical replay err=%v", err)
	}
}

func TestAttributionRebuildCanonicalReplayOfDuplicateFillIDsIsOrderIndependent(t *testing.T) {
	store := openTestStore(t)
	position, events := stagedFixture()
	duplicate := cloneFillDelta(events[0])
	duplicate.EventID = "z-duplicate-event"
	events = append(events, duplicate)
	first := AttributionRebuild{ID: "canonical-duplicate", AccountRef: "acct-1", CalculatedAt: attributionAt,
		Positions: []PositionEvidence{position}, FillDeltas: events}
	if err := store.PersistAttributionRebuild(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	slices.Reverse(events)
	second := AttributionRebuild{ID: "canonical-duplicate", AccountRef: "acct-1", CalculatedAt: attributionAt,
		Positions: []PositionEvidence{position}, FillDeltas: events}
	if err := store.PersistAttributionRebuild(context.Background(), second); err != nil {
		t.Fatalf("duplicate fill canonical replay err=%v", err)
	}
}

func TestV1MigratesForwardWithoutChangingReleasedRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "performance.db")
	v1, err := openWithSchema(path, schemaV1, 1)
	if err != nil {
		t.Fatal(err)
	}
	trade := measuredTrade(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	if err := v1.AppendTrade(context.Background(), trade); err != nil {
		t.Fatal(err)
	}
	if err := v1.Close(); err != nil {
		t.Fatal(err)
	}

	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	version, err := store.SchemaVersion(context.Background())
	if err != nil || version != SchemaVersion {
		t.Fatalf("version=%d err=%v", version, err)
	}
	var got string
	if err := store.db.QueryRow(`SELECT realized_pnl_after_costs FROM performance_trades WHERE id=?`, trade.ID).Scan(&got); err != nil || got != trade.RealizedPnLAfterCosts {
		t.Fatalf("released v1 row=%q err=%v", got, err)
	}
}

func TestAttributionRebuildPersistsQueriesAndIsolatesSameTickerAcrossMarkets(t *testing.T) {
	store := openTestStore(t)
	now := time.Date(2026, 8, 4, 6, 0, 0, 0, time.UTC)
	kr := unavailableFixture("KR", "005930", "position-kr")
	us := unavailableFixture("US", "005930", "position-us")
	rebuild := AttributionRebuild{ID: "build-1", AccountRef: "acct-1", CalculatedAt: now, Unavailable: []Attribution{kr, us}}
	if err := store.PersistAttributionRebuild(context.Background(), rebuild); err != nil {
		t.Fatal(err)
	}
	got, err := store.AttributionRows(context.Background(), "acct-1", AttributionQuery{
		Market: "KR", Ticker: "005930", IncludeLinkMissing: true,
	}, 100)
	if err != nil || len(got) != 1 || got[0].Key.PositionID != "position-kr" {
		t.Fatalf("KR query=%+v err=%v", got, err)
	}
	got, err = store.AttributionRows(context.Background(), "acct-1", AttributionQuery{
		Market: "US", Ticker: "005930", IncludeLinkMissing: true,
	}, 100)
	if err != nil || len(got) != 1 || got[0].Key.PositionID != "position-us" {
		t.Fatalf("US query=%+v err=%v", got, err)
	}
	if _, err := store.AttributionRows(context.Background(), "acct-1", AttributionQuery{Ticker: "005930"}, 100); !errors.Is(err, ErrMarketRequired) {
		t.Fatalf("ticker-only query err=%v", err)
	}
}

func TestCompleteAttributionEvidenceRoundTripsAndRebuildsExactly(t *testing.T) {
	store := openTestStore(t)
	position, events := stagedFixture()
	rebuild := AttributionRebuild{ID: "complete-build", AccountRef: "acct-1", CalculatedAt: attributionAt,
		Positions: []PositionEvidence{position}, FillDeltas: events}
	if err := store.PersistAttributionRebuild(context.Background(), rebuild); err != nil {
		t.Fatal(err)
	}
	evidence, err := store.AttributionEvidence(context.Background(), "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicalAttributionRebuild(rebuild)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(evidence, canonical) {
		t.Fatalf("evidence round trip\ngot:  %#v\nwant: %#v", evidence, canonical)
	}
	derived, err := BuildDerivedAttributionStore(evidence.Positions, evidence.FillDeltas)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := derived.QueryAttribution(AttributionQuery{Market: "US"})
	got, err := store.AttributionRows(context.Background(), "acct-1", AttributionQuery{Market: "US"}, 10)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("rebuilt rows differ\ngot:  %#v\nwant: %#v\nerr: %v", got, want, err)
	}
}

func TestAttributionRebuildReplayDivergenceAndCrashAreAtomic(t *testing.T) {
	store := openTestStore(t)
	now := time.Date(2026, 8, 4, 6, 0, 0, 0, time.UTC)
	first := AttributionRebuild{ID: "build-1", AccountRef: "acct-1", CalculatedAt: now,
		Unavailable: []Attribution{unavailableFixture("KR", "005930", "position-1")}}
	if err := store.PersistAttributionRebuild(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	if err := store.PersistAttributionRebuild(context.Background(), first); err != nil {
		t.Fatalf("idempotent replay: %v", err)
	}
	divergent := first
	divergent.Unavailable = []Attribution{unavailableFixture("KR", "000660", "position-1")}
	if err := store.PersistAttributionRebuild(context.Background(), divergent); !errors.Is(err, ErrImmutableConflict) {
		t.Fatalf("divergent replay err=%v", err)
	}

	previousHook := transactionPhaseHook
	transactionPhaseHook = func(phase string) {
		if phase == "attribution_after_rows" {
			panic("synthetic crash")
		}
	}
	defer func() { transactionPhaseHook = previousHook }()
	func() {
		defer func() { _ = recover() }()
		_ = store.PersistAttributionRebuild(context.Background(), AttributionRebuild{ID: "build-2", AccountRef: "acct-1", CalculatedAt: now.Add(time.Second),
			Unavailable: []Attribution{unavailableFixture("US", "AAPL", "position-2")}})
	}()
	transactionPhaseHook = previousHook
	got, err := store.AttributionRows(context.Background(), "acct-1", AttributionQuery{Market: "KR", IncludeLinkMissing: true}, 100)
	if err != nil || len(got) != 1 || got[0].Key.PositionID != "position-1" {
		t.Fatalf("crash leaked partial rebuild: %+v err=%v", got, err)
	}
}

func TestUnavailableAttributionRejectsInventedNumbersAndFX(t *testing.T) {
	store := openTestStore(t)
	row := unavailableFixture("US", "AAPL", "position-1")
	row.Reporting.NetPnL = AmountMetric{Status: StatusComplete, Value: "0", Currency: "USD"}
	rebuild := AttributionRebuild{ID: "bad-number", AccountRef: "acct-1", CalculatedAt: attributionAt, Unavailable: []Attribution{row}}
	if err := store.PersistAttributionRebuild(context.Background(), rebuild); err == nil {
		t.Fatal("unavailable evidence accepted an invented zero")
	}
	row = unavailableFixture("US", "AAPL", "position-1")
	row.Reporting.FXSource, row.Reporting.FXSourceVersion, row.Reporting.FXAsOf = "current", "now", attributionAt
	rebuild.ID, rebuild.Unavailable = "bad-fx", []Attribution{row}
	if err := store.PersistAttributionRebuild(context.Background(), rebuild); err == nil {
		t.Fatal("unavailable evidence accepted invented current FX")
	}
}

func TestAttributionReadOnlyQueryAndBounds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "performance.db")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	rebuild := AttributionRebuild{ID: "build-1", AccountRef: "acct-1", CalculatedAt: time.Now().UTC(),
		Unavailable: []Attribution{unavailableFixture("US", "AAPL", "position-1")}}
	if err := store.PersistAttributionRebuild(context.Background(), rebuild); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	reader, err := OpenReadOnly(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	got, err := reader.AttributionRows(context.Background(), "acct-1", AttributionQuery{Market: "US", IncludeLinkMissing: true}, 1)
	if err != nil || len(got) != 1 || got[0].Key.Ticker != "AAPL" {
		t.Fatalf("read-only query=%+v err=%v", got, err)
	}
	if _, err := reader.AttributionRows(context.Background(), "acct-1", AttributionQuery{Market: "US"}, MaxAttributionQueryRows+1); err == nil {
		t.Fatal("unbounded attribution query accepted")
	}
	evidence, err := reader.AttributionEvidence(context.Background(), "acct-1")
	if err != nil || evidence.ID != "build-1" || len(evidence.Unavailable) != 1 {
		t.Fatalf("read-only evidence=%+v err=%v", evidence, err)
	}
}

func TestAttributionRebuildDoesNotEnterRawObservationPruning(t *testing.T) {
	store := openTestStore(t)
	now := time.Date(2026, 8, 4, 6, 0, 0, 0, time.UTC)
	if err := store.PersistAttributionRebuild(context.Background(), AttributionRebuild{ID: "build-1", AccountRef: "acct-1", CalculatedAt: now,
		Unavailable: []Attribution{unavailableFixture("KR", "005930", "position-1")}}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PruneDue(context.Background(), now.Add(200*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := store.db.QueryRow(`SELECT count(*) FROM attribution_rows`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("attribution rows after raw prune=%d err=%v", count, err)
	}
}

func TestFailedV2MigrationRollsBackSchemaAndVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "performance.db")
	v1, err := openWithSchema(path, schemaV1, 1)
	if err != nil {
		t.Fatal(err)
	}
	_ = v1.Close()
	raw, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	store := &Store{db: raw, path: path}
	if err := store.migrateAttributionV2(context.Background(), schemaV2+`INSERT INTO absent_table(x) VALUES(1);`); err == nil {
		t.Fatal("broken v2 schema unexpectedly migrated")
	}
	defer raw.Close()
	var version, tables int
	if err := raw.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='attribution_rows'`).Scan(&tables); err != nil {
		t.Fatal(err)
	}
	if version != 1 || tables != 0 {
		t.Fatalf("failed v2 migration leaked version=%d tables=%d", version, tables)
	}
}

func unavailableFixture(market, ticker, position string) Attribution {
	row := Attribution{
		Key: AttributionKey{Market: market, Ticker: ticker, LaneID: "lane", LaneVersion: "lane/v1", PositionID: position,
			PolicyID: "policy", PolicyVersion: "policy/v1"},
		LineageStatus:       StatusLinkMissing,
		MissingMeasurements: []string{"entry_fill", "close_fill", "fees", "fx"},
		Source:              notMeasuredBreakdown(""), Reporting: notMeasuredBreakdown(""),
	}
	row.ObservedLineage = []CompositeLineage{{Market: market, Ticker: ticker, LaneID: "lane", LaneVersion: "lane/v1",
		PositionID: position, PolicyID: "policy", PolicyVersion: "policy/v1"}}
	row.MissingLineage = missingObservedLineage(row.ObservedLineage[0])
	return row
}
