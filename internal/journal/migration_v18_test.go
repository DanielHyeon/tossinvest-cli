package journal

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMigrationV17ToV18PreservesLegacyEventsAndCorrectionsWithoutInventingScope(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 17)
	if _, err := old.db.Exec(`INSERT INTO fill_snapshots
		(order_id, account_ref, symbol, market, trading_day, side, state, terminal,
		 filled_quantity, committed_at)
		VALUES ('legacy-order','','AAPL','us','','','FILLED',1,'2',
		        '2026-03-30T20:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := old.db.Exec(`INSERT INTO fill_events
		(order_id, symbol, market, delta_quantity, cumulative_quantity, average_price,
		 broker_visible_at, committed_at)
		VALUES ('legacy-order','AAPL','us','2','2','200','',
		        '2026-03-30T20:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := old.db.Exec(`INSERT INTO execution_corrections
		(id, account_ref, order_id, prev_avg_price, new_avg_price, prev_filled_amount,
		 new_filled_amount, cumulative_qty, observed_at)
		VALUES ('legacy-correction','','legacy-order','200','201','','','2',
		        '2026-03-30T20:01:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openJournalAtSchema(t, path, 18)
	defer j.Close()
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != 18 {
		t.Fatalf("version=%d err=%v", version, err)
	}
	var account, day, side, quantity string
	if err := j.db.QueryRow(`SELECT account_ref, trading_day, side, cumulative_quantity
		FROM fill_events WHERE order_id='legacy-order'`).Scan(&account, &day, &side, &quantity); err != nil {
		t.Fatal(err)
	}
	if account != "" || day != "" || side != "" || quantity != "2" {
		t.Fatalf("legacy event scope=(%q,%q,%q) quantity=%q, want blank scope and preserved data",
			account, day, side, quantity)
	}
	var legacyCorrections, scopedCorrections int
	if err := j.db.QueryRow(`SELECT count(*) FROM execution_corrections`).Scan(&legacyCorrections); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM scoped_execution_corrections`).Scan(&scopedCorrections); err != nil {
		t.Fatal(err)
	}
	if legacyCorrections != 1 || scopedCorrections != 0 {
		t.Fatalf("corrections legacy=%d scoped=%d, want 1/0 without migration backfill",
			legacyCorrections, scopedCorrections)
	}
}
