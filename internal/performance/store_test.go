package performance

import (
	"context"
	"database/sql"
	"errors"
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
	if _, err := store.Collect(ctx, trade, rows, at.Add(2*time.Hour)); err != nil {
		t.Fatalf("same immutable trade/observation bytes with a new snapshot were not appendable: %v", err)
	}
}

func TestCollectPersistsUnknownCostAsSQLNullWithoutInventingZero(t *testing.T) {
	store := openTestStore(t)
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	trade := measuredTrade(at)
	trade.CostTotal = ""
	if _, err := store.Collect(context.Background(), trade, []Observation{{
		ID: "unknown-cost-5", PositionID: trade.Lineage.PositionID, At: at.Add(5 * time.Minute),
		Price: "105", Source: "existing-position", SourceVersion: "v1",
	}}, at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	var cost sql.NullString
	if err := store.db.QueryRow(`SELECT cost_total FROM performance_trades WHERE id=?`, trade.ID).Scan(&cost); err != nil {
		t.Fatal(err)
	}
	if cost.Valid {
		t.Fatalf("unknown cost persisted as authoritative bytes %q", cost.String)
	}
	var status string
	var gross, adjusted sql.NullString
	if err := store.db.QueryRow(`SELECT status,gross_value,cost_adjusted_value
		FROM metric_observations WHERE metric_key='markout_5'`).Scan(&status, &gross, &adjusted); err != nil {
		t.Fatal(err)
	}
	if status != string(StatusNotMeasured) || gross.String != "5" || adjusted.Valid {
		t.Fatalf("persisted unknown-cost metric status=%q gross=%+v adjusted=%+v", status, gross, adjusted)
	}
	query := DefaultQuery(at.Add(time.Hour))
	query.MinimumSample = 1
	view, err := store.Dashboard(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}
	if view.States.NotMeasured == 0 || len(view.Aggregates) != 1 {
		t.Fatalf("unknown cost was not surfaced as missing measurement: %+v", view)
	}
	for _, metric := range view.Aggregates[0].Metrics {
		if metric.Key == "markout_5" && (metric.Status != StatusNotMeasured || metric.Value != "") {
			t.Fatalf("unknown-cost dashboard metric = %+v", metric)
		}
	}
}

func TestDashboardRejectsCorruptPersistedDecimalsInsteadOfUsingZero(t *testing.T) {
	for _, test := range []struct {
		name    string
		corrupt func(*testing.T, *Store)
	}{
		{name: "outcome pnl", corrupt: func(t *testing.T, store *Store) {
			if _, err := store.db.Exec(`UPDATE performance_trades SET realized_pnl_after_costs='broken'`); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "outcome r", corrupt: func(t *testing.T, store *Store) {
			if _, err := store.db.Exec(`UPDATE performance_trades SET realized_r='broken'`); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "complete metric", corrupt: func(t *testing.T, store *Store) {
			if _, err := store.db.Exec(`UPDATE metric_observations SET value='broken'
				WHERE metric_key='slippage' AND status='complete'`); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := openTestStore(t)
			now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
			trade := measuredTrade(now.Add(-time.Hour))
			trade.ClosedAt = now.Add(-time.Minute)
			if _, err := store.Collect(context.Background(), trade, []Observation{{
				ID: "corrupt-5", PositionID: trade.Lineage.PositionID, At: trade.EntryAt.Add(5 * time.Minute),
				Price: "105", Source: "existing-position", SourceVersion: "v1",
			}}, now); err != nil {
				t.Fatal(err)
			}
			test.corrupt(t, store)
			if _, err := store.Dashboard(context.Background(), DefaultQuery(now)); err == nil ||
				!strings.Contains(err.Error(), "invalid persisted decimal") {
				t.Fatalf("Dashboard corrupt value error = %v", err)
			}
		})
	}
}

func TestCollectExactReplayIsIdempotentAcrossRestartAndConcurrency(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "performance.db")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	trade := measuredTrade(at)
	rows := []Observation{{ID: "replay-5", PositionID: trade.Lineage.PositionID, At: at.Add(5 * time.Minute), Price: "105", Source: "cache", SourceVersion: "v1"}}
	calculatedAt := at.Add(time.Hour)
	if _, err := store.Collect(ctx, trade, rows, calculatedAt); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Collect(ctx, trade, rows, calculatedAt); err != nil {
		t.Fatalf("in-process exact replay: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if _, err := store.Collect(ctx, trade, rows, calculatedAt); err != nil {
		t.Fatalf("restart exact replay: %v", err)
	}

	start := make(chan struct{})
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			<-start
			_, err := store.Collect(ctx, trade, rows, calculatedAt)
			errs <- err
		}()
	}
	close(start)
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent exact replay: %v", err)
		}
	}
	for table, want := range map[string]int{"performance_trades": 1, "price_observations": 1, "measurement_snapshots": 1, "metric_observations": 6} {
		var got int
		if err := store.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&got); err != nil || got != want {
			t.Fatalf("%s rows=%d want=%d err=%v", table, got, want, err)
		}
	}
}

func TestCollectDivergentReplayFailsClosed(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	trade := measuredTrade(at)
	rows := []Observation{{ID: "immutable-5", PositionID: trade.Lineage.PositionID, At: at.Add(5 * time.Minute), Price: "105", Source: "cache", SourceVersion: "v1"}}
	calculatedAt := at.Add(time.Hour)
	if _, err := store.Collect(ctx, trade, rows, calculatedAt); err != nil {
		t.Fatal(err)
	}

	changedTrade := trade
	changedTrade.RealizedPnLAfterCosts = "81"
	if _, err := store.Collect(ctx, changedTrade, rows, calculatedAt); !errors.Is(err, ErrImmutableConflict) {
		t.Fatalf("trade divergence error=%v", err)
	}
	changedObservation := append([]Observation(nil), rows...)
	changedObservation[0].Price = "106"
	if _, err := store.Collect(ctx, trade, changedObservation, calculatedAt); !errors.Is(err, ErrImmutableConflict) {
		t.Fatalf("observation divergence error=%v", err)
	}
	if _, err := store.Collect(ctx, trade, nil, calculatedAt); !errors.Is(err, ErrImmutableConflict) {
		t.Fatalf("snapshot divergence error=%v", err)
	}
}

func TestConcurrentDivergentSnapshotsLeaveOneCompleteImmutableCollection(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	at := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	trade := measuredTrade(at)
	calculatedAt := at.Add(time.Hour)
	rows := [][]Observation{
		{{ID: "concurrent-a", PositionID: trade.Lineage.PositionID, At: at.Add(5 * time.Minute), Price: "105", Source: "cache", SourceVersion: "v1"}},
		{{ID: "concurrent-b", PositionID: trade.Lineage.PositionID, At: at.Add(5 * time.Minute), Price: "106", Source: "cache", SourceVersion: "v1"}},
	}
	start := make(chan struct{})
	errs := make(chan error, len(rows))
	for _, observations := range rows {
		observations := observations
		go func() {
			<-start
			_, err := store.Collect(ctx, trade, observations, calculatedAt)
			errs <- err
		}()
	}
	close(start)
	var succeeded, conflicted int
	for range rows {
		err := <-errs
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, ErrImmutableConflict):
			conflicted++
		default:
			t.Fatalf("unexpected concurrent error: %v", err)
		}
	}
	if succeeded != 1 || conflicted != 1 {
		t.Fatalf("concurrent divergent results success=%d conflict=%d", succeeded, conflicted)
	}
	for table, want := range map[string]int{"performance_trades": 1, "price_observations": 1, "measurement_snapshots": 1, "metric_observations": 6} {
		var got int
		if err := store.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&got); err != nil || got != want {
			t.Fatalf("%s rows=%d want=%d err=%v", table, got, want, err)
		}
	}
}

func TestObservationCompareAndAppendReplayAndDivergence(t *testing.T) {
	store := openTestStore(t)
	row := Observation{ID: "observation-1", PositionID: "position-1", At: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), Price: "100", Source: "cache", SourceVersion: "v1"}
	if err := store.AppendObservations(context.Background(), []Observation{row}); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendObservations(context.Background(), []Observation{row}); err != nil {
		t.Fatalf("exact replay: %v", err)
	}
	row.SourceVersion = "v2"
	if err := store.AppendObservations(context.Background(), []Observation{row}); !errors.Is(err, ErrImmutableConflict) {
		t.Fatalf("divergent replay error=%v", err)
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
	if got.States.Complete != 3 || got.States.LinkMissing != 0 || got.States.NotMeasured == 0 || got.States.InsufficientSample != 1 {
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
	if first.Deleted != 1200 || first.MaxBatchDeleted > MaxPruneRows || first.Transactions != 3 || first.BacklogRemaining ||
		(!raceEnabled && first.MaxBatchLockDuration >= 100*time.Millisecond) {
		t.Fatalf("first prune = %+v", first)
	}
	second, err := store.PruneDue(ctx, now.Add(23*time.Hour))
	if err != nil || second.Deleted != 0 || !second.Skipped {
		t.Fatalf("23h prune = %+v err=%v", second, err)
	}
	insertObservationFixture(t, store.db, 501, now.Add(-91*24*time.Hour))
	third, err := store.PruneDue(ctx, now.Add(24*time.Hour))
	if err != nil || third.Deleted != 501 || third.Transactions != 2 {
		t.Fatalf("24h prune = %+v err=%v", third, err)
	}
}

func TestPruneKeepsCadenceDueUntilBoundedBacklogDrainsDespiteContinuedInflux(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	insertObservationFixture(t, store.db, MaxPruneRows*MaxPruneBatchesPerRun+100, now.Add(-91*24*time.Hour))

	first, err := store.PruneDue(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if !first.BacklogRemaining || first.Transactions != MaxPruneBatchesPerRun || first.Deleted != MaxPruneRows*MaxPruneBatchesPerRun {
		t.Fatalf("bounded first run = %+v", first)
	}
	var cadenceRows int
	if err := store.db.QueryRow(`SELECT count(*) FROM maintenance_state WHERE key='last_pruned_at'`).Scan(&cadenceRows); err != nil || cadenceRows != 0 {
		t.Fatalf("unfinished backlog locked cadence rows=%d err=%v", cadenceRows, err)
	}
	insertObservationFixtureFrom(t, store.db, "influx", 75, now.Add(-92*24*time.Hour))
	second, err := store.PruneDue(ctx, now.Add(time.Minute))
	if err != nil || second.Skipped || second.BacklogRemaining || second.Deleted != 175 {
		t.Fatalf("immediate reschedule = %+v err=%v", second, err)
	}
	third, err := store.PruneDue(ctx, now.Add(2*time.Minute))
	if err != nil || !third.Skipped {
		t.Fatalf("drained cadence = %+v err=%v", third, err)
	}
	late := Observation{
		ID: "late-overdue", PositionID: "position-late", At: now.Add(-91 * 24 * time.Hour),
		Price: "100", Source: "late-backfill", SourceVersion: "v1",
	}
	if err := store.AppendObservations(ctx, []Observation{late}); err != nil {
		t.Fatal(err)
	}
	fourth, err := store.PruneDue(ctx, now.Add(3*time.Minute))
	if err != nil || fourth.Skipped || fourth.Deleted != 1 || fourth.BacklogRemaining {
		t.Fatalf("late overdue append did not immediately reschedule prune = %+v err=%v", fourth, err)
	}
}

func TestConcurrentObservationIDsAreIdempotentCompareAndAppend(t *testing.T) {
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
	var succeeded int
	for err := range errs {
		if err == nil {
			succeeded++
		} else {
			t.Errorf("exact concurrent append: %v", err)
		}
	}
	if succeeded != 2 {
		t.Fatalf("concurrent exact append succeeded=%d", succeeded)
	}
	var count int
	if err := store.db.QueryRow(`SELECT count(*) FROM price_observations WHERE id=?`, row.ID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("immutable observation rows=%d err=%v", count, err)
	}
}

func insertObservationFixture(t *testing.T, db *sql.DB, count int, at time.Time) {
	insertObservationFixtureFrom(t, db, "fixture", count, at)
}

func insertObservationFixtureFrom(t *testing.T, db *sql.DB, prefix string, count int, at time.Time) {
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
		if _, err := stmt.Exec(fmt.Sprintf("%s-%09d", prefix, i), "position-fixture", at.Add(time.Duration(i)*time.Second).Format(time.RFC3339Nano), "100", "fixture", "v1"); err != nil {
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
