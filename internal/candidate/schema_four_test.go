package candidate

// schema_four_test.go is the fourth rung of the ladder on its own, and the columns
// it adds named out loud.
//
// # What was not pinned
//
// The four columns this change introduced — observations.newly_listed_state,
// observations.rank_requested, candidates.first_rank_newly_listed and
// candidates.first_rank_requested — were not named in any assertion anywhere. The
// two candidateColumns checks stop at first_rank_source, and there was no v3 fixture,
// so the 3 → 4 rung was only ever reached transitively by the v1 test climbing all
// three. Stripping the CHECK text from the four ALTER statements would have failed
// nothing at all, and a store migrated without them ends up with different
// constraints from a store created fresh — which is the drift the candidates table's
// first_rank columns were spelled per column to avoid.
//
// # Why a v3 fixture and not just more assertions on the v1 one
//
// v1 climbs three rungs in one transaction, which is the case that cannot catch a
// rung that only works when the one below it has just run beside it. Most stores in
// the field are at whatever the last release wrote, and 3 → 4 is the step this
// change's own upgrade takes.

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// v3Schema is the schema exactly as SchemaVersion 3 wrote it: v2 plus D20's four
// first-rank columns, and none of the schema-4 ones. Spelled out rather than derived
// for v1Schema's reason — a ladder test that builds its "old" database from the
// current source migrates a current file and calls it an old one.
const v3Schema = `
CREATE TABLE IF NOT EXISTS observations (
	id             INTEGER PRIMARY KEY AUTOINCREMENT,
	market         TEXT    NOT NULL,
	symbol         TEXT    NOT NULL,
	source         TEXT    NOT NULL,
	observed_at    TEXT    NOT NULL,
	rank           INTEGER NOT NULL DEFAULT 0,
	rank_total     INTEGER NOT NULL DEFAULT 0,
	newly_listed   INTEGER NOT NULL DEFAULT 0,
	trading_value  TEXT,
	trading_volume TEXT,
	price          TEXT,
	day_high       TEXT,
	day_low        TEXT,
	via_fallback   INTEGER NOT NULL DEFAULT 0,
	CHECK (newly_listed IN (0,1)),
	CHECK (via_fallback IN (0,1)),
	CHECK (rank >= 0 AND rank_total >= 0)
) STRICT;
CREATE INDEX IF NOT EXISTS idx_observations_symbol_time
	ON observations(market, symbol, observed_at);
CREATE INDEX IF NOT EXISTS idx_observations_time ON observations(observed_at);
CREATE TABLE IF NOT EXISTS candidates (
	market             TEXT NOT NULL,
	symbol             TEXT NOT NULL,
	first_seen_at      TEXT NOT NULL,
	last_seen_at       TEXT NOT NULL,
	cooled_at          TEXT,
	sources            TEXT NOT NULL DEFAULT '',
	sources_attempted  INTEGER NOT NULL DEFAULT 0,
	sources_responded  INTEGER NOT NULL DEFAULT 0,
	degraded           INTEGER NOT NULL DEFAULT 0,
	first_price        TEXT,
	first_price_at     TEXT,
	first_price_source TEXT,
	first_rank         INTEGER CHECK (first_rank IS NULL OR first_rank > 0),
	first_rank_total   INTEGER CHECK (first_rank_total IS NULL OR first_rank_total > 0),
	first_rank_at      TEXT,
	first_rank_source  TEXT,
	PRIMARY KEY (market, symbol),
	CHECK (degraded IN (0,1))
) STRICT;
CREATE TABLE IF NOT EXISTS store_meta (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL
) STRICT;
INSERT INTO store_meta (key, value) VALUES ('schema_version', '3');
`

// writeV3Store builds a schema-3 database: a candidate with a first sighting
// position recorded and nothing recorded about the reading it came from, which is
// what every store in the field looks like the moment before this rung runs.
func writeV3Store(t *testing.T, path string) {
	t.Helper()
	db, err := sql.Open("sqlite", dsn(path, time.Second))
	if err != nil {
		t.Fatalf("opening a v3 store: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(v3Schema); err != nil {
		t.Fatalf("creating the v3 schema: %v", err)
	}
	_, err = db.Exec(`INSERT INTO candidates
		(market, symbol, first_seen_at, last_seen_at, cooled_at, sources,
		 sources_attempted, sources_responded, degraded,
		 first_price, first_price_at, first_price_source,
		 first_rank, first_rank_total, first_rank_at, first_rank_source)
		VALUES ('KR','005930',?,?,NULL,'official_rankings_trading_value',5,5,0,
		        '10000',?,'official_prices',12,100,?,'official_rankings_trading_value')`,
		stamp(t0), stamp(t0.Add(time.Minute)), stamp(t0), stamp(t0))
	if err != nil {
		t.Fatalf("seeding the v3 candidate: %v", err)
	}
	_, err = db.Exec(`INSERT INTO observations
		(market, symbol, source, observed_at, rank, rank_total, newly_listed, price, via_fallback)
		VALUES ('KR','005930','official_rankings_trading_value',?,12,100,0,'10000',0)`,
		stamp(t0))
	if err != nil {
		t.Fatalf("seeding the v3 observation: %v", err)
	}
}

// observationColumns is candidateColumns for the other table.
func observationColumns(t *testing.T, s *Store) map[string]bool {
	t.Helper()
	rows, err := s.db.QueryContext(context.Background(),
		`SELECT name FROM pragma_table_info('observations')`)
	if err != nil {
		t.Fatalf("pragma_table_info: %v", err)
	}
	defer func() { _ = rows.Close() }()
	cols := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		cols[name] = true
	}
	return cols
}

// TestAStoreAtSchemaThreeGainsTheFourColumnsAndTheirConstraints.
func TestAStoreAtSchemaThreeGainsTheFourColumnsAndTheirConstraints(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "candidates.db")
	writeV3Store(t, path)

	s, err := Open(ctx, Options{Path: path, FSProber: FixedFSProber(FSInfo{Name: "ext4"})})
	if err != nil {
		t.Fatalf("Open on a v3 store: %v", err)
	}
	defer func() { _ = s.Close() }()

	obs := observationColumns(t, s)
	for _, want := range []string{"newly_listed_state", "rank_requested"} {
		if !obs[want] {
			t.Errorf("after the migration observations still has no %s", want)
		}
	}
	// And the schema-1 boolean is still there. Forward-only means the column is not
	// dropped and not renamed, so a build from before this rung finds the shape it
	// expects.
	if !obs["newly_listed"] {
		t.Error("the schema-1 newly_listed boolean was dropped; the rung is additive")
	}
	cands := candidateColumns(t, s)
	for _, want := range []string{"first_rank_newly_listed", "first_rank_requested"} {
		if !cands[want] {
			t.Errorf("after the migration candidates still has no %s", want)
		}
	}

	// Nothing was backfilled, which is the rung's own decision and the one most
	// likely to be "fixed": the schema-1 boolean is 0 in every row ever written,
	// because no production source assigned the field it came from, so copying it
	// would turn "nobody measured this" into "the source said it was already listed"
	// for the entire stored history in one statement.
	first, found, err := s.FirstRank(ctx, "KR", "005930")
	if err != nil || !found {
		t.Fatalf("FirstRank: %v found=%v", err, found)
	}
	if first.Rank != 12 || first.Total != 100 {
		t.Errorf("the position across the migration = %d of %d, want 12 of 100",
			first.Rank, first.Total)
	}
	if first.NewlyListed.Known() || first.Requested != 0 {
		t.Errorf("the rung backfilled %s / %d onto a position recorded before either fact "+
			"existed", first.NewlyListed, first.Requested)
	}
	rows, err := s.Observations(ctx, "KR", "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("%d raw rows survived the migration, want 1", len(rows))
	}
	if rows[0].Reported.NewlyListed.Known() || rows[0].Reported.RankRequested != 0 {
		t.Errorf("a migrated raw row reports %s / %d; nobody measured either of these when "+
			"it was written", rows[0].Reported.NewlyListed, rows[0].Reported.RankRequested)
	}

	// The constraints came with the columns. A migrated store and a fresh one must
	// end up the same shape, which is why the CHECKs are spelled per column in the
	// CREATE TABLE and repeated verbatim in the ALTERs — SQLite cannot add a table
	// constraint to an existing table, so a table-level CHECK would exist on new
	// files and not on migrated ones and nothing would say so.
	assertMigratedConstraints(t, s)
}

// TestAFreshStoreCarriesTheSameFourConstraints is the other end of the comparison.
func TestAFreshStoreCarriesTheSameFourConstraints(t *testing.T) {
	s, _ := openStore(t)
	assertMigratedConstraints(t, s)
}

// assertMigratedConstraints drives each of the four columns' CHECK clauses with a
// value it must refuse.
//
// It writes through the raw handle rather than through the store's own methods on
// purpose: those validate on the way in, and what is under test here is what the
// file itself refuses when something else — a repair script, a future writer, a
// migration — puts a value in.
func assertMigratedConstraints(t *testing.T, s *Store) {
	t.Helper()
	ctx := context.Background()
	if _, err := s.db.ExecContext(ctx, `INSERT INTO candidates
		(market, symbol, first_seen_at, last_seen_at, sources)
		VALUES ('KR','CHK',?,?,'official_rankings_trading_value')`,
		stamp(t0), stamp(t0)); err != nil {
		t.Fatalf("seeding the constraint probe: %v", err)
	}

	for _, tc := range []struct {
		what string
		stmt string
		args []any
		why  string
	}{
		{
			what: "observations.newly_listed_state accepts a third word",
			stmt: `INSERT INTO observations
				(market, symbol, source, observed_at, rank, rank_total, newly_listed_state)
				VALUES ('KR','CHK','official_rankings_trading_value',?,1,100,'maybe')`,
			args: []any{stamp(t0)},
			why: "the column holds a three-state fact whose third state is NULL. A fourth " +
				"spelling reads back as unknown through newlyListedFromStore, so nothing " +
				"downstream would report it and the row would simply stop being measurable",
		},
		{
			what: "observations.rank_requested accepts a request for no rows",
			stmt: `INSERT INTO observations
				(market, symbol, source, observed_at, rank, rank_total, rank_requested)
				VALUES ('KR','CHK','official_rankings_trading_value',?,1,100,0)`,
			args: []any{stamp(t0)},
			why: "0 is what a row written before this fact existed carries, so it has to " +
				"mean unknown. A stored 0 and an absent value would then be the same state " +
				"spelled two ways, and truncationOf reads one of them as `asked for nothing`",
		},
		{
			what: "candidates.first_rank_newly_listed accepts a third word",
			stmt: `UPDATE candidates SET first_rank_newly_listed = 'probably'
			        WHERE market = 'KR' AND symbol = 'CHK'`,
			why: "same three-state fact, same reason, on the column seen_late actually reads",
		},
		{
			what: "candidates.first_rank_requested accepts a negative",
			stmt: `UPDATE candidates SET first_rank_requested = -1
			        WHERE market = 'KR' AND symbol = 'CHK'`,
			why: "NoteFirstRank refuses a negative request at the boundary and the column " +
				"has to refuse it too, or the boundary is the only thing holding it",
		},
	} {
		_, err := s.db.ExecContext(ctx, tc.stmt, tc.args...)
		if err == nil {
			t.Errorf("%s.\n\n%s", tc.what, tc.why)
			continue
		}
		if !strings.Contains(strings.ToUpper(err.Error()), "CONSTRAINT") {
			t.Errorf("%s was refused by %v, which is not the CHECK; the write may be "+
				"failing for an unrelated reason and this assertion would pass on a file "+
				"with no constraint at all", tc.what, err)
		}
	}
}
