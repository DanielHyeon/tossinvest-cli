package performance

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestMillionRowRecentQueryUsesCoveringIndexAndMeetsP95Target(t *testing.T) {
	if testing.Short() {
		t.Skip("million-row on-disk query contract")
	}
	if raceEnabled {
		t.Skip("wall-clock p95 is verified by the non-instrumented suite")
	}
	store := openTestStore(t)
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	insertMillionObservationFixture(t, store, now)

	plan, err := store.RecentObservationQueryPlan(context.Background(), now.Add(-30*24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(plan), "covering index idx_price_observations_at") {
		t.Fatalf("query plan does not use the retention/range covering index: %s", plan)
	}

	durations := make([]time.Duration, 25)
	for i := range durations {
		started := time.Now()
		count, err := store.RecentObservationCount(context.Background(), now.Add(-30*24*time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		if count != 300_000 {
			t.Fatalf("recent count = %d, want 300000", count)
		}
		durations[i] = time.Since(started)
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	p95 := durations[(len(durations)*95+99)/100-1]
	if p95 > 250*time.Millisecond {
		t.Fatalf("1M-row recent 30d query p95 = %s, target <=250ms (runs=%v)", p95, durations)
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
