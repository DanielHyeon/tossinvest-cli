package journal

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMigrationV15ToV16AddsEmptyOrderScopeWithoutRewritingSnapshots(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 15)
	if _, err := old.db.Exec(`INSERT INTO fill_snapshots
		(order_id, symbol, market, state, terminal, filled_quantity, committed_at)
		VALUES ('legacy-order','AAPL','us','FILLED',1,'2','2026-03-29T20:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openJournalAtSchema(t, path, 16)
	defer j.Close()
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != 16 {
		t.Fatalf("version=%d err=%v", version, err)
	}
	var account, day, side, quantity string
	if err := j.db.QueryRow(`SELECT account_ref, trading_day, side, filled_quantity
		FROM fill_snapshots WHERE order_id='legacy-order'`).
		Scan(&account, &day, &side, &quantity); err != nil {
		t.Fatal(err)
	}
	if account != "" || day != "" || side != "" || quantity != "2" {
		t.Fatalf("migrated legacy scope=(%q,%q,%q) quantity=%q, want empty scope and preserved 2",
			account, day, side, quantity)
	}
	var indexes int
	if err := j.db.QueryRow(`SELECT count(*) FROM sqlite_master
		WHERE type='index' AND name='idx_fill_snapshots_scope'`).Scan(&indexes); err != nil || indexes != 1 {
		t.Fatalf("scope index count=%d err=%v", indexes, err)
	}
	var scopedLineageTables int
	if err := j.db.QueryRow(`SELECT count(*) FROM sqlite_master
		WHERE type='table' AND name='scoped_lineage_edges'`).Scan(&scopedLineageTables); err != nil || scopedLineageTables != 1 {
		t.Fatalf("scoped lineage table count=%d err=%v", scopedLineageTables, err)
	}
}
