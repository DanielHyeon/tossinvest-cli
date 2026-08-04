package journal

import (
	"context"
	"database/sql"
	"testing"
)

func TestWeeklyReservationLifecycleConsumesPositiveAndReleasesTerminalZeroExactlyOnce(t *testing.T) {
	for _, test := range []struct {
		name, cumulative, status string
		terminal, increment      bool
	}{
		{name: "positive-partial", cumulative: "1", status: WeeklyReservationConsumed, increment: true},
		{name: "terminal-zero", cumulative: "0", status: WeeklyReservationReleased, terminal: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			db := weeklyLifecycleDB(t)
			tx, err := db.BeginTx(context.Background(), nil)
			if err != nil {
				t.Fatal(err)
			}
			fill := AppliedFill{OrderID: "order-1", AccountRef: "acct", Market: "kr", Symbol: "005930", Side: "BUY",
				CumulativeQuantity: test.cumulative, Delta: test.cumulative, Terminal: test.terminal,
				CommittedAt: "2026-08-04T12:00:00Z"}
			if err := applyWeeklyReservationLifecycleInTx(context.Background(), tx, fill); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
			var status string
			var version, positive int
			if err := db.QueryRow(`SELECT status FROM strategy_weekly_market_reservations WHERE reservation_id='reservation-1'`).Scan(&status); err != nil {
				t.Fatal(err)
			}
			if err := db.QueryRow(`SELECT version,positive_leg_count FROM strategy_weekly_reservation_scopes WHERE campaign_id='campaign-1' AND market='KR'`).Scan(&version, &positive); err != nil {
				t.Fatal(err)
			}
			wantPositive := 0
			if test.increment {
				wantPositive = 1
			}
			if status != test.status || version != 2 || positive != wantPositive {
				t.Fatalf("status/version/positive=%s/%d/%d", status, version, positive)
			}
			// A later observation cannot increment or transition the same reservation again.
			tx, _ = db.BeginTx(context.Background(), nil)
			fill.CumulativeQuantity = "2"
			fill.Delta = "1"
			fill.CommittedAt = "2026-08-04T12:01:00Z"
			if err := applyWeeklyReservationLifecycleInTx(context.Background(), tx, fill); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
			if err := db.QueryRow(`SELECT version,positive_leg_count FROM strategy_weekly_reservation_scopes`).Scan(&version, &positive); err != nil {
				t.Fatal(err)
			}
			if version != 2 || positive != wantPositive {
				t.Fatalf("replay advanced scope=%d/%d", version, positive)
			}
		})
	}
}

func weeklyLifecycleDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	schema := `
	CREATE TABLE strategy_weekly_first_leg_bindings(decision_id TEXT,reservation_id TEXT,campaign_id TEXT,market TEXT);
	CREATE TABLE strategy_weekly_market_reservations(reservation_id TEXT PRIMARY KEY,status TEXT,updated_at TEXT);
	CREATE TABLE strategy_weekly_reservation_scopes(campaign_id TEXT,market TEXT,version INTEGER,positive_leg_count INTEGER,updated_at TEXT);
	CREATE TABLE strategy_dispatch_leases(guardian_decision_id TEXT,broker_order_id TEXT,account_ref TEXT,market TEXT,symbol TEXT);
	CREATE TABLE strategy_weekly_reservation_lifecycle_receipts(event_id TEXT PRIMARY KEY,reservation_id TEXT,from_status TEXT,to_status TEXT,scope_version INTEGER,cumulative_quantity TEXT,record_digest TEXT,observed_at TEXT);
	INSERT INTO strategy_weekly_first_leg_bindings VALUES('decision-1','reservation-1','campaign-1','KR');
	INSERT INTO strategy_weekly_market_reservations VALUES('reservation-1','ACTIVE','2026-08-04T11:00:00Z');
	INSERT INTO strategy_weekly_reservation_scopes VALUES('campaign-1','KR',1,0,'2026-08-04T11:00:00Z');
	INSERT INTO strategy_dispatch_leases VALUES('decision-1','order-1','acct','KR','005930');`
	if _, err := db.Exec(schema); err != nil {
		t.Fatal(err)
	}
	return db
}
