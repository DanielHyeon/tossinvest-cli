package journal

import (
	"context"
	"errors"
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

func TestMigrationV19NeverBindsLegacyEvidenceToAFutureReusedOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 18)
	insertLegacyMigrationSnapshot(t, old, "future-reuse", "2026-03-29T20:00:00Z")
	if _, err := old.db.Exec(`INSERT INTO fill_events
		(order_id, account_ref, symbol, market, trading_day, side, delta_quantity,
		 cumulative_quantity, average_price, broker_visible_at, committed_at)
		VALUES ('future-reuse','','AAPL','us','','','2','2','200','',
		        '2026-03-29T20:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := old.db.Exec(`INSERT INTO execution_corrections
		(id, account_ref, order_id, prev_avg_price, new_avg_price, prev_filled_amount,
		 new_filled_amount, cumulative_qty, observed_at)
		VALUES ('old-correction','','future-reuse','200','201','','','2',
		        '2026-03-29T20:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	insertConfirmedMigrationOwnerAt(t, old, "future-reuse", "future-reuse",
		"2026-03-29T19:00:00Z", "2026-03-29T20:00:00Z")
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openJournalAtSchema(t, path, 19)
	defer j.Close()
	scope := FillSnapshotScope{
		OrderID: "future-reuse", AccountRef: "acct-1", Market: "us",
		TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY",
	}
	events, err := j.FillEventsScoped(context.Background(), scope)
	if err != nil || len(events) != 0 {
		t.Fatalf("future scope events=%+v err=%v, want no legacy attribution", events, err)
	}
	corrections, err := j.ExecutionCorrectionsScoped(context.Background(), scope)
	if err != nil || len(corrections) != 0 {
		t.Fatalf("future scope corrections=%+v err=%v, want no legacy attribution", corrections, err)
	}
	if _, err := j.FillEvents(context.Background(), "future-reuse"); !errors.Is(err, ErrFillScopeAmbiguous) {
		t.Fatalf("unscoped events err=%v, want ambiguity for unbound legacy evidence", err)
	}
	if _, err := j.ExecutionCorrections(context.Background(), "future-reuse"); !errors.Is(err, ErrFillScopeAmbiguous) {
		t.Fatalf("unscoped corrections err=%v, want ambiguity for unbound legacy evidence", err)
	}
}

func TestMigrationV19BackfillsOnlyAPreexistingConfirmedOwner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 18)
	insertConfirmedMigrationOwner(t, old, "owned-legacy", "2026-03-29T19:00:00Z")
	insertDecision(t, old, "owned-legacy-decision", "owned-legacy-nonce")
	if _, err := old.db.Exec(`UPDATE mutation_attempts SET decision_id='owned-legacy-decision'
		WHERE id='attempt-owned-legacy'`); err != nil {
		t.Fatal(err)
	}
	if _, err := old.db.Exec(`INSERT INTO positions
		(id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
		 quantity, avg_price, opened_at)
		VALUES ('owned-legacy-position','acct-1','us','AAPL',1,
		        'owned-legacy-decision','OPEN','2','200','2026-03-29T20:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	insertLegacyMigrationSnapshot(t, old, "owned-legacy", "2026-03-29T20:00:00Z")
	if _, err := old.db.Exec(`INSERT INTO fill_events
		(order_id, account_ref, symbol, market, trading_day, side, delta_quantity,
		 cumulative_quantity, average_price, broker_visible_at, committed_at)
		VALUES ('owned-legacy','','AAPL','us','','','2','2','200','',
		        '2026-03-29T20:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := old.db.Exec(`INSERT INTO execution_corrections
		(id, account_ref, order_id, prev_avg_price, new_avg_price, prev_filled_amount,
		 new_filled_amount, cumulative_qty, observed_at)
		VALUES ('owned-correction','','owned-legacy','200','201','','','2',
		        '2026-03-29T20:01:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openJournalAtSchema(t, path, 19)
	defer j.Close()
	scope := FillSnapshotScope{
		OrderID: "owned-legacy", AccountRef: "acct-1", Market: "us",
		TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY",
	}
	events, err := j.FillEventsScoped(context.Background(), scope)
	if err != nil || len(events) != 1 || events[0].CumulativeQuantity != "2" {
		t.Fatalf("backfilled events=%+v err=%v", events, err)
	}
	var rawAccount, rawDay, rawSide string
	if err := j.db.QueryRow(`SELECT account_ref, trading_day, side FROM fill_events
		WHERE order_id='owned-legacy'`).Scan(&rawAccount, &rawDay, &rawSide); err != nil {
		t.Fatal(err)
	}
	if rawAccount != "" || rawDay != "" || rawSide != "" {
		t.Fatalf("append-only legacy fill was rewritten to (%q,%q,%q)", rawAccount, rawDay, rawSide)
	}
	var bindings int
	if err := j.db.QueryRow(`SELECT count(*) FROM legacy_fill_event_bindings
		WHERE order_id='owned-legacy'`).Scan(&bindings); err != nil {
		t.Fatal(err)
	}
	if bindings != 1 {
		t.Fatalf("legacy bindings=%d, want 1 additive binding", bindings)
	}
	chain, err := j.PositionProvenance(context.Background(), "owned-legacy-position")
	if err != nil {
		t.Fatal(err)
	}
	if fill := stepOf(t, chain, ProvenanceFill); fill.Ref != "owned-legacy" {
		t.Fatalf("migrated fill provenance=%+v, want companion-bound legacy fill", fill)
	}
	corrections, err := j.ExecutionCorrectionsScoped(context.Background(), scope)
	if err != nil || len(corrections) != 1 || corrections[0].NewAveragePrice != "201" {
		t.Fatalf("backfilled corrections=%+v err=%v", corrections, err)
	}
}

func TestMigrationV19RefusesTwoIntentsClaimingTheSameCanonicalScope(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 18)
	insertConfirmedMigrationOwnerNamed(t, old, "ambiguous-owner", "owner-a", "2026-03-29T18:00:00Z")
	insertConfirmedMigrationOwnerNamed(t, old, "ambiguous-owner", "owner-b", "2026-03-29T19:00:00Z")
	insertLegacyMigrationSnapshot(t, old, "ambiguous-owner", "2026-03-29T20:00:00Z")
	if _, err := old.db.Exec(`INSERT INTO fill_events
		(order_id, account_ref, symbol, market, trading_day, side, delta_quantity,
		 cumulative_quantity, average_price, broker_visible_at, committed_at)
		VALUES ('ambiguous-owner','','AAPL','us','','','2','2','200','',
		        '2026-03-29T20:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openJournalAtSchema(t, path, 19)
	defer j.Close()
	var bindings int
	if err := j.db.QueryRow(`SELECT count(*) FROM legacy_fill_event_bindings
		WHERE order_id='ambiguous-owner'`).Scan(&bindings); err != nil {
		t.Fatal(err)
	}
	if bindings != 0 {
		t.Fatalf("ambiguous legacy event received %d bindings, want none", bindings)
	}
	scope := FillSnapshotScope{OrderID: "ambiguous-owner", AccountRef: "acct-1", Market: "us",
		TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY"}
	if events, err := j.FillEventsScoped(context.Background(), scope); err != nil || len(events) != 0 {
		t.Fatalf("scoped ambiguous events=%+v err=%v, want none", events, err)
	}
	if _, err := j.FillEvents(context.Background(), "ambiguous-owner"); !errors.Is(err, ErrFillScopeAmbiguous) {
		t.Fatalf("unscoped ambiguous events err=%v, want ambiguity", err)
	}
}

func insertConfirmedMigrationOwner(t *testing.T, j *Journal, orderID, recordedAt string) {
	insertConfirmedMigrationOwnerNamed(t, j, orderID, orderID, recordedAt)
}

func insertConfirmedMigrationOwnerNamed(t *testing.T, j *Journal, orderID, ownerID, recordedAt string) {
	insertConfirmedMigrationOwnerAt(t, j, orderID, ownerID, recordedAt, recordedAt)
}

func insertConfirmedMigrationOwnerAt(t *testing.T, j *Journal, orderID, ownerID,
	recordedAt, settledAt string) {
	t.Helper()
	if _, err := j.db.Exec(`INSERT INTO intents
		(id, created_at, market, trading_day, account_ref, symbol, side, order_type,
		 quantity, price, currency, source, fingerprint)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, "intent-"+ownerID, recordedAt, "us", "2026-03-30",
		"acct-1", "AAPL", "BUY", "LIMIT", "2", "200", "USD", "engine", "fp-"+ownerID); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO mutation_attempts
		(id, intent_id, kind, state, attempt_no, broker_order_id, fingerprint,
		 recorded_at, settled_at)
		VALUES (?,?,?,?,?,?,?,?,?)`, "attempt-"+ownerID, "intent-"+ownerID, "PLACE",
		string(StateConfirmed), 1, orderID, "fp-"+ownerID, recordedAt, settledAt); err != nil {
		t.Fatal(err)
	}
}

func insertLegacyMigrationSnapshot(t *testing.T, j *Journal, orderID, committedAt string) {
	t.Helper()
	if _, err := j.db.Exec(`INSERT INTO fill_snapshots
		(order_id, account_ref, symbol, market, trading_day, side, state, terminal,
		 filled_quantity, committed_at)
		VALUES (?, '', 'AAPL', 'us', '', '', 'FILLED', 1, '2', ?)`,
		orderID, committedAt); err != nil {
		t.Fatal(err)
	}
}
