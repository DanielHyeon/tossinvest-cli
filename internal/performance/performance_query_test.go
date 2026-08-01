package performance

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestMillionRawRowsActualDashboardUsesBoundedIndexesAndMeetsP95Target(t *testing.T) {
	if testing.Short() {
		t.Skip("million-row on-disk query contract")
	}
	if raceEnabled {
		t.Skip("wall-clock p95 is verified by the non-instrumented suite")
	}
	store := openTestStore(t)
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	insertMillionObservationFixture(t, store, now)
	trade := measuredTrade(now.Add(-24 * time.Hour))
	if _, err := store.Collect(context.Background(), trade, nil, now); err != nil {
		t.Fatal(err)
	}
	query := DefaultQuery(now)

	plan, err := store.DashboardQueryPlan(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}
	lowerPlan := strings.ToLower(plan)
	for _, index := range []string{"idx_performance_trades_window", "idx_measurement_snapshots_trade"} {
		if !strings.Contains(lowerPlan, index) {
			t.Fatalf("actual dashboard plan misses %s:\n%s", index, plan)
		}
	}
	if strings.Contains(lowerPlan, "scan price_observations") || strings.Contains(lowerPlan, "scan measurement_snapshots") ||
		strings.Contains(lowerPlan, "scan metric_observations") {
		t.Fatalf("actual dashboard performs an unrelated/global scan:\n%s", plan)
	}

	durations := make([]time.Duration, 25)
	for i := range durations {
		started := time.Now()
		view, err := store.Dashboard(context.Background(), query)
		if err != nil {
			t.Fatal(err)
		}
		if len(view.Aggregates) != 1 || view.Aggregates[0].Samples != 1 {
			t.Fatalf("dashboard = %+v", view)
		}
		durations[i] = time.Since(started)
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	p95 := durations[(len(durations)*95+99)/100-1]
	if p95 > 250*time.Millisecond {
		t.Fatalf("1M-row recent 30d query p95 = %s, target <=250ms (runs=%v)", p95, durations)
	}
}

func TestDashboardBoundsPeriodAndTradeRows(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	query := DefaultQuery(now)
	query.PeriodDays = MaxDashboardPeriodDays + 1
	if _, err := store.Dashboard(ctx, query); !errors.Is(err, ErrDashboardPeriod) {
		t.Fatalf("unbounded period error=%v", err)
	}

	tx, err := store.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := tx.Prepare(insertTradeSQL)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= MaxDashboardTrades; i++ {
		trade := measuredTrade(now.Add(-time.Hour))
		trade.ID = fmt.Sprintf("bounded-trade-%05d", i)
		trade.Lineage.PositionID = fmt.Sprintf("bounded-position-%05d", i)
		trade.Lineage.CloseID = trade.Lineage.PositionID
		trade.Lineage.FillID = fmt.Sprintf("bounded-fill-%05d", i)
		if _, err := stmt.Exec(tradeArgs(trade)...); err != nil {
			t.Fatal(err)
		}
	}
	if err := stmt.Close(); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	query = DefaultQuery(now)
	if _, err := store.Dashboard(ctx, query); !errors.Is(err, ErrDashboardRowLimit) {
		t.Fatalf("row bound error=%v", err)
	}
}

func TestDashboardStateCountsUseTheSameMarketLaneAndCompleteFilters(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	appendTrade := func(id, market, lane string, complete bool) {
		t.Helper()
		trade := measuredTrade(now.Add(-24 * time.Hour))
		trade.ID = id
		trade.Market = market
		trade.Lineage.PositionID = id + "-position"
		trade.Lineage.CloseID = trade.Lineage.PositionID
		trade.Lineage.FillID = id + "-fill"
		trade.Lineage.LaneID = lane
		if !complete {
			trade.Lineage.FillID = ""
		}
		if _, err := store.Collect(ctx, trade, nil, now); err != nil {
			t.Fatal(err)
		}
	}
	appendTrade("kr-a-complete", "KR", "lane-a", true)
	appendTrade("kr-a-missing", "KR", "lane-a", false)
	appendTrade("kr-b-complete", "KR", "lane-b", true)
	appendTrade("us-a-missing", "US", "lane-a", false)

	query := DefaultQuery(now)
	query.Market, query.Lane = "KR", "lane-a"
	view, err := store.Dashboard(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	if view.States.Complete != 1 || view.States.LinkMissing != 0 {
		t.Fatalf("complete-only scoped states=%+v", view.States)
	}
	query.CompleteOnly = false
	view, err = store.Dashboard(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	if view.States.Complete != 1 || view.States.LinkMissing != 1 {
		t.Fatalf("all-lineage scoped states=%+v", view.States)
	}
}

func insertMillionObservationFixture(t *testing.T, store *Store, now time.Time) {
	t.Helper()
	tx, err := store.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := tx.Prepare(`INSERT INTO price_observations
		(id, position_id, observed_at, price, source, source_version) VALUES (?,?,?,?,?,?)`)
	if err != nil {
		t.Fatal(err)
	}
	base := now.Add(-100 * 24 * time.Hour)
	for i := 0; i < 1_000_000; i++ {
		// Exactly 10,000 rows per UTC day: the final 30 days contain 300,000 rows.
		at := base.Add(time.Duration(i/10_000) * 24 * time.Hour).Add(time.Duration(i%10_000) * time.Millisecond)
		if _, err := stmt.Exec(fmt.Sprintf("million-%07d", i), fmt.Sprintf("p-%05d", i%10_000),
			at.Format(time.RFC3339Nano), "100", "fixture", "v1"); err != nil {
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
