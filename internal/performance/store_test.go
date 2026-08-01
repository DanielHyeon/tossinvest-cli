package performance

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(filepath.Join(t.TempDir(), "performance.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})
	return store
}

func TestStoreSchemaIsSeparateAppendOnlyAndVersioned(t *testing.T) {
	store := openTestStore(t)
	if got, err := store.SchemaVersion(context.Background()); err != nil || got != 1 {
		t.Fatalf("schema version = %d err=%v", got, err)
	}
	for _, table := range []string{"performance_trades", "price_observations", "measurement_snapshots", "maintenance_state"} {
		var count int
		if err := store.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count); err != nil || count != 1 {
			t.Fatalf("table %s count=%d err=%v", table, count, err)
		}
	}
	if filepath.Base(store.Path()) != "performance.db" {
		t.Fatalf("path = %q", store.Path())
	}
}

func TestCollectAppendsExistingObservationsAndLatestMeasurement(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	trade := measuredTrade(at)
	rows := []Observation{
		{ID: "o-5", PositionID: "position-1", At: at.Add(5 * time.Minute), Price: "105", Source: "exit-observation", SourceVersion: "v1"},
		{ID: "o-15", PositionID: "position-1", At: at.Add(15 * time.Minute), Price: "106", Source: "exit-observation", SourceVersion: "v1"},
		{ID: "o-30", PositionID: "position-1", At: at.Add(30 * time.Minute), Price: "107", Source: "exit-observation", SourceVersion: "v1"},
	}
	got, err := store.Collect(ctx, trade, rows, at.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if got.Markout(30).GrossPct != "7" {
		t.Fatalf("snapshot = %+v", got)
	}
	var observations, snapshots int
	if err := store.db.QueryRow(`SELECT count(*) FROM price_observations`).Scan(&observations); err != nil {
		t.Fatal(err)
	}
	if err := store.db.QueryRow(`SELECT count(*) FROM measurement_snapshots`).Scan(&snapshots); err != nil {
		t.Fatal(err)
	}
	if observations != 3 || snapshots != 1 {
		t.Fatalf("rows observations=%d snapshots=%d", observations, snapshots)
	}
	if _, err := store.Collect(ctx, trade, rows, at.Add(2*time.Hour)); err == nil {
		t.Fatal("duplicate immutable trade/observation IDs were accepted")
	}
}

func TestDashboardUsesFixedCompleteLineageFilterAndFirstClassStates(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		trade := measuredTrade(now.AddDate(0, 0, -i-1))
		trade.ID = fmt.Sprintf("trade-%d", i)
		trade.Lineage.PositionID = fmt.Sprintf("position-%d", i)
		trade.Lineage.CloseID = trade.Lineage.PositionID
		trade.Lineage.FillID = fmt.Sprintf("fill-%d", i)
		trade.RealizedPnLAfterCosts = []string{"10", "-4", "6"}[i]
		trade.RealizedR = []string{"1", "-0.4", "0.6"}[i]
		trade.ClosedAt = now.AddDate(0, 0, -i-1)
		if _, err := store.Collect(ctx, trade, nil, now); err != nil {
			t.Fatal(err)
		}
	}
	missing := measuredTrade(now.AddDate(0, 0, -2))
	missing.ID = "trade-link-missing"
	missing.Lineage.PositionID = "position-missing"
	missing.Lineage.CloseID = "position-missing"
	missing.Lineage.FillID = ""
	if _, err := store.Collect(ctx, missing, nil, now); err != nil {
		t.Fatal(err)
	}

	query := DefaultQuery(now)
	query.MinimumSample = 5
	got, err := store.Dashboard(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	if query.PeriodDays != 30 || query.Market != AllMarkets || query.Lane != AllLanes || !query.CompleteOnly {
		t.Fatalf("default query = %+v", query)
	}
	if len(got.Aggregates) != 1 || got.Aggregates[0].Samples != 3 || got.Aggregates[0].Status != StatusInsufficientSample {
		t.Fatalf("aggregates = %+v", got.Aggregates)
	}
	if got.Aggregates[0].NetPnL != "12" || got.Aggregates[0].WinRate != "0.666666666667" || got.Aggregates[0].ProfitFactor != "4" {
		t.Fatalf("cost-adjusted aggregate = %+v", got.Aggregates[0])
	}
	if got.Aggregates[0].Metrics[0].Provenance != "journal-outcome@"+SemanticsVersion {
		t.Fatalf("outcome provenance = %+v", got.Aggregates[0].Metrics[0])
	}
	if got.States.LinkMissing != 1 || got.States.NotMeasured == 0 || got.States.InsufficientSample != 1 {
		t.Fatalf("states = %+v", got.States)
	}
	for _, code := range []Status{StatusLinkMissing, StatusNotMeasured, StatusInsufficientSample} {
		if !strings.Contains(got.Explanation(code), string(code)) {
			t.Errorf("explanation %s = %q", code, got.Explanation(code))
		}
	}
}

func TestPruneIsDueEvery24HoursAndDeletesAtMost500RowsPerTransaction(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	insertObservationFixture(t, store.db, 1200, now.Add(-91*24*time.Hour))

	first, err := store.PruneDue(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if first.Deleted != MaxPruneRows || first.Deleted > 500 || (!raceEnabled && first.LockDuration >= 100*time.Millisecond) {
		t.Fatalf("first prune = %+v", first)
	}
	second, err := store.PruneDue(ctx, now.Add(23*time.Hour))
	if err != nil || second.Deleted != 0 || !second.Skipped {
		t.Fatalf("23h prune = %+v err=%v", second, err)
	}
	third, err := store.PruneDue(ctx, now.Add(24*time.Hour))
	if err != nil || third.Deleted != MaxPruneRows {
		t.Fatalf("24h prune = %+v err=%v", third, err)
	}
}

func TestConcurrentObservationIDsAreCompareAndAppend(t *testing.T) {
	store := openTestStore(t)
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	row := Observation{ID: "same-id", PositionID: "position-1", At: at, Price: "100", Source: "existing", SourceVersion: "v1"}
	start := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- store.AppendObservations(context.Background(), []Observation{row})
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	var succeeded, refused int
	for err := range errs {
		if err == nil {
			succeeded++
		} else {
			refused++
		}
	}
	if succeeded != 1 || refused != 1 {
		t.Fatalf("concurrent append succeeded=%d refused=%d", succeeded, refused)
	}
}

func insertObservationFixture(t *testing.T, db *sql.DB, count int, at time.Time) {
	t.Helper()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := tx.Prepare(`INSERT INTO price_observations
		(id, position_id, observed_at, price, source, source_version) VALUES (?,?,?,?,?,?)`)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < count; i++ {
		if _, err := stmt.Exec(fmt.Sprintf("fixture-%09d", i), "position-fixture", at.Add(time.Duration(i)*time.Second).Format(time.RFC3339Nano), "100", "fixture", "v1"); err != nil {
			t.Fatal(err)
		}
	}
	if err := stmt.Close(); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}
