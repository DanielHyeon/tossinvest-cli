package journal

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMigrationV16ToV17AddsScopedSnapshotsWithoutRewritingLegacyRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 16)
	if _, err := old.db.Exec(`INSERT INTO fill_snapshots
		(order_id, account_ref, symbol, market, trading_day, side, state, terminal,
		 filled_quantity, committed_at)
		VALUES ('v16-order','acct-1','AAPL','us','2026-03-30','BUY','FILLED',1,'2',
		        '2026-03-30T20:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openJournalAtSchema(t, path, 17)
	defer j.Close()
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != 17 {
		t.Fatalf("version=%d err=%v", version, err)
	}
	var account, day, quantity string
	if err := j.db.QueryRow(`SELECT account_ref, trading_day, filled_quantity
		FROM fill_snapshots WHERE order_id='v16-order'`).Scan(&account, &day, &quantity); err != nil {
		t.Fatal(err)
	}
	if account != "acct-1" || day != "2026-03-30" || quantity != "2" {
		t.Fatalf("legacy row=(%q,%q,%q), want byte-preserved scope and quantity", account, day, quantity)
	}
	var scopedRows int
	if err := j.db.QueryRow(`SELECT count(*) FROM scoped_fill_snapshots`).Scan(&scopedRows); err != nil {
		t.Fatal(err)
	}
	if scopedRows != 0 {
		t.Fatalf("migration copied or rewrote history: scoped rows=%d", scopedRows)
	}
	stored, err := j.LookupFillScoped(context.Background(), FillSnapshotScope{
		OrderID: "v16-order", AccountRef: "acct-1", Market: "us",
		TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY",
	})
	if err != nil || stored.FilledQuantity != "2" || !stored.Terminal {
		t.Fatalf("LookupFillScoped(v16 fallback) = %+v, %v", stored, err)
	}
}

func TestScopedFillSnapshotPrimaryKeyContainsTheWholeCanonicalScope(t *testing.T) {
	j := openTestJournal(t)
	rows, err := j.db.Query(`PRAGMA table_info(scoped_fill_snapshots)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	pk := map[int]string{}
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, typ string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatal(err)
		}
		if primaryKey > 0 {
			pk[primaryKey] = name
		}
	}
	want := []string{"account_ref", "market", "trading_day", "symbol", "side", "order_id"}
	for i, column := range want {
		if got := pk[i+1]; got != column {
			t.Fatalf("primary key position %d = %q, want %q (all=%v)", i+1, got, column, pk)
		}
	}
}
