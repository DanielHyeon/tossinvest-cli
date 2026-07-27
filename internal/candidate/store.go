package candidate

// store.go is the discovery store: its own SQLite file, in the same data
// directory as the ledger and sharing nothing else with it.
//
// # Separate file, separate lock (D2)
//
//	                     journal.db              candidates.db
//	authority            orders, fills, positions derived observations
//	writer               one per account, flock   the scan process
//	if lost              recovery procedure       a shorter history
//	write rate           per event                tens to hundreds per scan
//
// A scan that took the ledger's lock would stop `tossctl engine run` from coming
// up for as long as it ran. An observation feature blocking an execution feature
// is the wrong direction, and it is the direction that happens by accident when
// two things share a file. TestAScanningStoreDoesNotBlockTheEngineFromStarting is
// that stated as something an operator would notice.
//
// # Why SQLite and not JSONL
//
// Every query this change makes is a time-series query — the instant a symbol
// first crossed, the trading value delta over the last N minutes, the actual
// spacing between two observations. JSONL would mean replaying the whole file per
// scan, and that cost would end up deciding the scan interval.
//
// # Retention is two layers and only one of them is prunable (D11)
//
// KR at 15s is roughly 5,760 scans a day; five sources at 100-150 rows each puts
// the raw table in the millions of rows per day, tens of millions per month. That
// does not fit a local SQLite queried every 15 seconds, so raw rows are prunable.
//
// The candidate summary is not. PruneObservations deletes from `observations` and
// nothing else, and TestPruningRawObservationsLeavesTheCandidateSummary fails if
// that ever changes — because the tidy-up that takes first_seen_at with it
// deletes this package's entire claim without failing anything.
//
// # Expiry is read, not swept
//
// Candidates carries the instant to judge against and derives the state from the
// stored timestamps. Nothing has to have run in between: a candidate that cooled
// before a restart reads as expired after it. A state that only advances when a
// sweeper runs is a state that lies on a process which has just come up, which is
// precisely when it is asked.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	// modernc.org/sqlite is the pure-Go driver, the same one the ledger uses: no
	// cgo, so the binary still cross-compiles. It registers itself as "sqlite".
	_ "modernc.org/sqlite"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

const (
	// DBFileName is the discovery store's file name inside the data directory.
	// It is deliberately not the ledger's — see the file header.
	DBFileName = "candidates.db"

	// EnvDataDir and EnvXDGDataHome are the ledger's own variables. The store
	// lives beside the ledger, so it must answer to the same override or an
	// operator who moved their data would end up with two directories.
	EnvDataDir     = "TOSSOS_DATA_DIR"
	EnvXDGDataHome = "XDG_DATA_HOME"
	// AppDirName is the per-application subdirectory under an XDG data home.
	AppDirName = "tossos"

	// dataDirPerm keeps the directory private to its owner. The store holds no
	// account identifiers, but it sits beside a file that does.
	dataDirPerm = 0o700

	// defaultBusyTimeout bounds how long a writer waits for another writer. This
	// store tolerates a second writer far better than the ledger does — the rows
	// are derived — but a scan that blocks forever is still a scan that stopped
	// observing.
	defaultBusyTimeout = 5 * time.Second

	// DefaultRawRetention is how long raw observations are kept before they may
	// be pruned. Two trading days: long enough that yesterday's session can still
	// be re-derived, short enough that the table stays queryable at scan cadence.
	DefaultRawRetention = 48 * time.Hour
)

// ErrSchemaTooNew means the store on disk was written by a newer TossOS. An older
// binary reading it would see columns it does not know about as absent.
var ErrSchemaTooNew = errors.New("candidate: store schema is newer than this build")

// SchemaVersion is the schema this build writes and understands.
//
//	1  observations, candidates, store_meta
//	2  candidates.first_price / first_price_at / first_price_source (D17)
const SchemaVersion = 2

const schema = `
-- Raw observations. One row per (source, symbol, instant). Prunable (D11).
--
-- Every reported decimal is TEXT and NULL means the source did not carry it.
-- A numeric column with a 0 default would erase the difference between "the
-- source said zero" and "we never measured this", and D10 is built on that
-- difference surviving all the way to the veto.
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

-- Candidate summaries. One row per (market, symbol). NOT prunable by
-- PruneObservations — that is the whole point of the split.
CREATE TABLE IF NOT EXISTS candidates (
	market             TEXT NOT NULL,
	symbol             TEXT NOT NULL,
	first_seen_at      TEXT NOT NULL,
	last_seen_at       TEXT NOT NULL,
	-- NULL while active. Non-NULL is what starts the cooling clock, and the
	-- expiry is derived from it at read time rather than written here.
	cooled_at          TEXT,
	sources            TEXT NOT NULL DEFAULT '',
	sources_attempted  INTEGER NOT NULL DEFAULT 0,
	sources_responded  INTEGER NOT NULL DEFAULT 0,
	degraded           INTEGER NOT NULL DEFAULT 0,
	-- The expansion baseline (D17). It is a fact of the same grade as
	-- first_seen_at and it follows the same lifecycle: preserved across cooling
	-- and re-entry, reset on expiry.
	--
	-- It is a column rather than a query over the observations table because that
	-- table is prunable (D11, DefaultRawRetention 48h) and because
	-- SourceObservations' documented "since" window narrows it further. Deriving
	-- the baseline from either produced a re-based one — a candidate that had
	-- tripled reported 50% expansion with Measured true and no reason — and the
	-- extended veto then read "not extended" on exactly the longest-running
	-- candidates.
	--
	-- NULL means no observation has carried a price yet. It is not "0", the same
	-- rule every other decimal in this package follows.
	first_price        TEXT,
	first_price_at     TEXT,
	first_price_source TEXT,

	PRIMARY KEY (market, symbol),
	CHECK (degraded IN (0,1))
) STRICT;

CREATE TABLE IF NOT EXISTS store_meta (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL
) STRICT;
`

// Options configures Open.
type Options struct {
	// Path is the SQLite file. Empty resolves DefaultPath() from the environment,
	// the same precedence the ledger uses.
	Path string
	// Clock is the time source. Defaults to clock.System(). Every instant this
	// package writes comes from a caller's argument or from here; nothing calls
	// time.Now() directly, because the time axis is what is being tested.
	Clock clock.Clock
	// FSProber inspects the hosting filesystem. Defaults to SystemFSProber().
	FSProber FSProber
	// BusyTimeout overrides how long a blocked writer waits. Zero uses
	// defaultBusyTimeout.
	BusyTimeout time.Duration
	// CoolingTTL overrides how long a cooled candidate stays re-entrant. Zero
	// uses DefaultCoolingTTL.
	CoolingTTL time.Duration
	// StalenessTTL overrides how long an un-refreshed candidate stays active
	// before it cools implicitly. Zero uses DefaultStalenessTTL. See stateAt.
	StalenessTTL time.Duration
}

// Store is an open discovery store.
type Store struct {
	db        *sql.DB
	path      string
	clk       clock.Clock
	fs        FSInfo
	cooling   time.Duration
	staleness time.Duration
}

// DataDir resolves the directory that holds the store.
//
// It is the ledger's directory and the ledger's precedence — $TOSSOS_DATA_DIR >
// $XDG_DATA_HOME/tossos > ~/.local/share/tossos — resolved here rather than
// imported for the reason fsguard.go gives. The two files sit side by side; only
// the locks are separate.
func DataDir() (string, error) {
	if override := strings.TrimSpace(os.Getenv(EnvDataDir)); override != "" {
		if !filepath.IsAbs(override) {
			return "", fmt.Errorf("candidate: %s must be an absolute path (got %q)", EnvDataDir, override)
		}
		return filepath.Clean(override), nil
	}
	if xdg := strings.TrimSpace(os.Getenv(EnvXDGDataHome)); xdg != "" && filepath.IsAbs(xdg) {
		return filepath.Join(filepath.Clean(xdg), AppDirName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "", fmt.Errorf("candidate: cannot determine the data directory (%v)", err)
	}
	return filepath.Join(home, ".local", "share", AppDirName), nil
}

// DefaultPath returns the store path inside the resolved data directory.
func DefaultPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, DBFileName), nil
}

// Open resolves the path, checks the filesystem, creates the directory if needed
// and brings the schema up.
//
// It refuses — returns an error and creates nothing — when the filesystem is
// outside the allowlist, so a rejected mount is left untouched.
func Open(ctx context.Context, opts Options) (*Store, error) {
	path := opts.Path
	if path == "" {
		resolved, err := DefaultPath()
		if err != nil {
			return nil, err
		}
		path = resolved
	}
	path = filepath.Clean(path)
	dir := filepath.Dir(path)

	prober := opts.FSProber
	if prober == nil {
		prober = SystemFSProber()
	}
	// Before anything is created: a refused mount must be left as it was found.
	fs, err := CheckFilesystem(prober, dir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, dataDirPerm); err != nil {
		return nil, fmt.Errorf("candidate: creating %s: %w", dir, err)
	}

	busy := opts.BusyTimeout
	if busy <= 0 {
		busy = defaultBusyTimeout
	}
	db, err := sql.Open("sqlite", dsn(path, busy))
	if err != nil {
		return nil, fmt.Errorf("candidate: opening %s: %w", path, err)
	}
	// One connection. The store is written by a single scan loop and read by the
	// console; more connections would buy contention on a file whose whole design
	// assumes there is not any.
	db.SetMaxOpenConns(1)

	clk := opts.Clock
	if clk == nil {
		clk = clock.System()
	}
	cooling := opts.CoolingTTL
	if cooling <= 0 {
		cooling = DefaultCoolingTTL
	}
	staleness := opts.StalenessTTL
	if staleness <= 0 {
		staleness = DefaultStalenessTTL
	}
	s := &Store{
		db: db, path: path, clk: clk, fs: fs,
		cooling: cooling, staleness: staleness,
	}

	if err := s.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// dsn builds the driver connection string.
//
// The path is escaped rather than concatenated. SQLite opens this with
// SQLITE_OPEN_URI, so an unescaped `#` in a data directory silently opens a
// different file — `/data/team#1/candidates.db` becomes `/data/team` — while
// Store.Path() keeps reporting the directory the operator asked for. That is the
// one string somebody debugging a missing first_seen_at will trust.
//
// _txlock=immediate is the ledger's, and it is load-bearing here for the same
// reason: Promote reads the candidate and then writes it, and under Go's default
// DEFERRED transaction a second process that commits in between makes the write
// fail with SQLITE_BUSY_SNAPSHOT — which busy_timeout does not retry. IMMEDIATE
// takes the write lock up front, so the second writer waits on the busy handler
// instead of failing.
func dsn(path string, busy time.Duration) string {
	// The remaining pragmas are the ledger's minus the paranoia it needs and this
	// does not: WAL so a reading console never blocks a writing scan, NORMAL
	// rather than FULL because losing the last few observations to a power cut
	// costs a shorter history, not a lost order.
	q := url.Values{}
	q.Add("_pragma", "busy_timeout("+fmt.Sprint(busy.Milliseconds())+")")
	q.Add("_pragma", "journal_mode(WAL)")
	q.Add("_pragma", "synchronous(NORMAL)")
	q.Add("_pragma", "foreign_keys(1)")
	q.Set("_txlock", "immediate")

	uri := url.URL{Scheme: "file", Opaque: (&url.URL{Path: path}).EscapedPath()}
	return uri.String() + "?" + q.Encode()
}

// migrations[v] brings a store stamped at version v up to v+1.
//
// The CREATE TABLE statements above are IF NOT EXISTS, so they build the current
// shape on a new file and do nothing at all to an existing one. Everything that
// has to happen to a store somebody already has is here, and it is a ladder rather
// than one step, so a store two versions behind climbs both rungs in order.
//
// Additive and nullable, which is WORKFLOW §0.6's preference for a schema change
// and is what makes the step reversible: an older binary opening a v2 file reads
// columns it does not select, and the rows it does read are unchanged.
var migrations = map[int][]string{
	1: {
		`ALTER TABLE candidates ADD COLUMN first_price TEXT`,
		`ALTER TABLE candidates ADD COLUMN first_price_at TEXT`,
		`ALTER TABLE candidates ADD COLUMN first_price_source TEXT`,
		// Backfill from the raw rows that are still there.
		//
		// The alternative is to leave the column NULL and let the next priced
		// observation fill it, and that is strictly worse: the next observation is
		// newer than every row already in the table, so it sets a later baseline
		// and understates every expansion measured against it. The earliest
		// surviving raw row is the oldest evidence that exists at this moment.
		//
		// It can still be later than the candidate's true first price, if the rows
		// that carried it were already pruned. That is why first_price_at is
		// written with it and why D17 asks for the instant: a baseline three days
		// old and one twenty minutes old are different events, and a reader who has
		// only the value cannot tell which they are looking at.
		`UPDATE candidates SET
		   first_price = (SELECT o.price FROM observations o
		                   WHERE o.market = candidates.market AND o.symbol = candidates.symbol
		                     AND o.price IS NOT NULL
		                   ORDER BY o.observed_at, o.id LIMIT 1),
		   first_price_at = (SELECT o.observed_at FROM observations o
		                   WHERE o.market = candidates.market AND o.symbol = candidates.symbol
		                     AND o.price IS NOT NULL
		                   ORDER BY o.observed_at, o.id LIMIT 1),
		   first_price_source = (SELECT o.source FROM observations o
		                   WHERE o.market = candidates.market AND o.symbol = candidates.symbol
		                     AND o.price IS NOT NULL
		                   ORDER BY o.observed_at, o.id LIMIT 1)`,
	},
}

func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("candidate: creating the schema in %s: %w", s.path, err)
	}
	var found string
	err := s.db.QueryRowContext(ctx,
		`SELECT value FROM store_meta WHERE key = 'schema_version'`).Scan(&found)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		// No stamp means the statements above just built the file, so it is already
		// at the current version and there is nothing to climb.
		_, err = s.db.ExecContext(ctx,
			`INSERT INTO store_meta (key, value) VALUES ('schema_version', ?)`,
			fmt.Sprint(SchemaVersion))
		if err != nil {
			return fmt.Errorf("candidate: stamping the schema version: %w", err)
		}
		return nil
	case err != nil:
		return fmt.Errorf("candidate: reading the schema version: %w", err)
	}
	var version int
	if _, scanErr := fmt.Sscanf(found, "%d", &version); scanErr != nil {
		return fmt.Errorf("candidate: unreadable schema version %q in %s", found, s.path)
	}
	if version > SchemaVersion {
		return fmt.Errorf("%w (found %d, this build understands %d)",
			ErrSchemaTooNew, version, SchemaVersion)
	}
	if version == SchemaVersion {
		return nil
	}
	return s.climb(ctx, version)
}

// climb runs the migration ladder from `version` to SchemaVersion in one
// transaction, so a store is never left half way up it. A rung with no statements
// is refused rather than skipped: a gap in the ladder means this build believes it
// understands a shape nobody wrote the step for.
func (s *Store) climb(ctx context.Context, version int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("candidate: beginning the migration of %s: %w", s.path, err)
	}
	defer func() { _ = tx.Rollback() }()

	for v := version; v < SchemaVersion; v++ {
		stmts, ok := migrations[v]
		if !ok {
			return fmt.Errorf("candidate: %s is at schema %d and this build has no step to %d",
				s.path, v, v+1)
		}
		for _, stmt := range stmts {
			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("candidate: migrating %s from schema %d to %d: %w",
					s.path, v, v+1, err)
			}
		}
	}
	_, err = tx.ExecContext(ctx,
		`UPDATE store_meta SET value = ? WHERE key = 'schema_version'`, fmt.Sprint(SchemaVersion))
	if err != nil {
		return fmt.Errorf("candidate: stamping schema %d on %s: %w", SchemaVersion, s.path, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("candidate: committing the migration of %s: %w", s.path, err)
	}
	return nil
}

// Close releases the database handle.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// Path is the file this store is open on.
func (s *Store) Path() string { return s.path }

// Filesystem is what the guard found.
func (s *Store) Filesystem() FSInfo { return s.fs }

// CoolingTTL is how long this store keeps a cooled candidate re-entrant.
func (s *Store) CoolingTTL() time.Duration { return s.cooling }

// Now is this store's clock.
//
// Callers that need "now" — a scan stamping its readings, a retention sweep
// deciding what is old — take it from here rather than from time.Now(), so a test
// can drive a whole trading day through the store without waiting for one. The
// time axis is what this package measures; a second, unmockable time source would
// be a hole in the only thing being tested.
func (s *Store) Now() time.Time { return s.clk.Now() }

// PruneStale drops raw observations older than the retention window, measured
// back from this store's own clock, and reports how many went.
//
// It is PruneObservations with the policy applied — same statement, same reach:
// the candidate summaries are outside it, so a retention sweep can never take a
// first_seen_at with it (D11).
func (s *Store) PruneStale(ctx context.Context, retention time.Duration) (int64, error) {
	if retention <= 0 {
		retention = DefaultRawRetention
	}
	return s.PruneObservations(ctx, s.clk.Now().Add(-retention))
}

// --- observations -------------------------------------------------------------

// RecordObservations appends a scan's readings.
//
// Appends, never upserts: two readings of the same symbol at two instants are two
// rows, and the difference between them is the only thing this package computes.
// A store that collapsed them would be a point-in-time query with extra steps.
func (s *Store) RecordObservations(ctx context.Context, in []Observation) error {
	if len(in) == 0 {
		return nil
	}
	rows := make([]Observation, 0, len(in))
	for i, o := range in {
		valid, err := o.validate()
		if err != nil {
			return fmt.Errorf("observation %d of %d: %w", i+1, len(in), err)
		}
		rows = append(rows, valid)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("candidate: beginning an observation write: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO observations
		  (market, symbol, source, observed_at, rank, rank_total, newly_listed,
		   trading_value, trading_volume, price, day_high, day_low, via_fallback)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("candidate: preparing the observation insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, o := range rows {
		// Normalised on the way in as well as on the way out, so the two spellings
		// cannot disagree: a padded id written verbatim and looked up trimmed would
		// answer with an empty series and no error.
		id, err := normaliseSource(o.Source)
		if err != nil {
			return fmt.Errorf("candidate: recording %s: %w", o.Key(), err)
		}
		_, err = stmt.ExecContext(ctx,
			o.Market, o.Symbol, string(id), stamp(o.ObservedAt),
			o.Reported.Rank, o.Reported.RankTotal, boolToInt(o.Reported.NewlyListed),
			nullable(o.Reported.TradingValue), nullable(o.Reported.TradingVolume),
			nullable(o.Reported.Price), nullable(o.Reported.DayHigh), nullable(o.Reported.DayLow),
			boolToInt(o.ViaFallback))
		if err != nil {
			return fmt.Errorf("candidate: recording %s from %s: %w", o.Key(), o.Source, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("candidate: committing %d observations: %w", len(rows), err)
	}
	return nil
}

// Observations returns one symbol's readings at or after `since`, oldest first.
//
// A zero `since` returns everything retained. Oldest first because every consumer
// walks the series forward looking for the first crossing.
func (s *Store) Observations(ctx context.Context, market, symbol string, since time.Time) ([]Observation, error) {
	market = strings.ToUpper(strings.TrimSpace(market))
	symbol = strings.TrimSpace(symbol)

	query := `
		SELECT market, symbol, source, observed_at, rank, rank_total, newly_listed,
		       trading_value, trading_volume, price, day_high, day_low, via_fallback
		  FROM observations
		 WHERE market = ? AND symbol = ?`
	args := []any{market, symbol}
	if !since.IsZero() {
		query += ` AND observed_at >= ?`
		args = append(args, stamp(since))
	}
	query += ` ORDER BY observed_at, id`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("candidate: reading observations of %s:%s: %w", market, symbol, err)
	}
	defer func() { _ = rows.Close() }()

	var out []Observation
	for rows.Next() {
		o, err := scanObservation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("candidate: reading observations of %s:%s: %w", market, symbol, err)
	}
	return out, nil
}

// SourceObservations returns one source's readings of one symbol at or after
// `since`, oldest first.
//
// This is the read D9's last section requires, and it exists because the obvious
// one is wrong. Observations above interleaves every source in time order, and
// TradingValue is a different cumulative per source — different unit, different
// aggregation window, different fallback status. A rate differenced across that
// mixture measures the gap between two sources' conventions, and it compiles, and
// the number it produces is plausible. So the series a metric is computed from is
// keyed by (market, symbol, source), and this is the only read that produces one.
//
// No index is added for it. idx_observations_symbol_time already seeks straight to
// (market, symbol) and walks in observed_at order, so the source predicate is a
// filter on rows this query was going to visit anyway and nothing has to be
// sorted. A fourth column would trim that walk and cost another index on the table
// D16 identifies as the one that can fill the ledger's filesystem.
func (s *Store) SourceObservations(ctx context.Context, market, symbol string,
	source SourceID, since time.Time) ([]Observation, error) {

	market = strings.ToUpper(strings.TrimSpace(market))
	symbol = strings.TrimSpace(symbol)
	source, err := normaliseSource(source)
	if err != nil {
		return nil, fmt.Errorf("candidate: reading observations of %s:%s: %w", market, symbol, err)
	}

	query := `
		SELECT market, symbol, source, observed_at, rank, rank_total, newly_listed,
		       trading_value, trading_volume, price, day_high, day_low, via_fallback
		  FROM observations
		 WHERE market = ? AND symbol = ? AND source = ?`
	args := []any{market, symbol, string(source)}
	if !since.IsZero() {
		query += ` AND observed_at >= ?`
		args = append(args, stamp(since))
	}
	query += ` ORDER BY observed_at, id`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("candidate: reading %s observations of %s:%s: %w",
			source, market, symbol, err)
	}
	defer func() { _ = rows.Close() }()

	var out []Observation
	for rows.Next() {
		o, err := scanObservation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("candidate: reading %s observations of %s:%s: %w",
			source, market, symbol, err)
	}
	return out, nil
}

// SourceSeries returns one source's series for one symbol, ready to measure.
//
// A symbol this source has never reported comes back as an empty series rather
// than an error: not having looked yet is an ordinary state, and every metric
// answers WARMING_UP for it. NewSourceSeries refuses an empty input because it
// could not tell whose series it was; here the identity is in the arguments.
func (s *Store) SourceSeries(ctx context.Context, market, symbol string,
	source SourceID, since time.Time) (SourceSeries, error) {

	rows, err := s.SourceObservations(ctx, market, symbol, source, since)
	if err != nil {
		return SourceSeries{}, err
	}
	if len(rows) == 0 {
		// The id is normalised here too, so an empty series names the source the
		// store actually looked for rather than the spelling the caller used.
		id, err := normaliseSource(source)
		if err != nil {
			return SourceSeries{}, err
		}
		return SourceSeries{
			Key: Key{
				Market: strings.ToUpper(strings.TrimSpace(market)),
				Symbol: strings.TrimSpace(symbol),
			},
			Source: id,
		}, nil
	}
	return NewSourceSeries(rows)
}

// ObservationsSince returns every retained reading at or after `since`, oldest
// first, across all symbols. It is what a scan uses to compute metrics for a
// whole market without one query per symbol.
func (s *Store) ObservationsSince(ctx context.Context, market string, since time.Time) ([]Observation, error) {
	market = strings.ToUpper(strings.TrimSpace(market))
	rows, err := s.db.QueryContext(ctx, `
		SELECT market, symbol, source, observed_at, rank, rank_total, newly_listed,
		       trading_value, trading_volume, price, day_high, day_low, via_fallback
		  FROM observations
		 WHERE market = ? AND observed_at >= ?
		 ORDER BY symbol, observed_at, id`, market, stamp(since))
	if err != nil {
		return nil, fmt.Errorf("candidate: reading %s observations: %w", market, err)
	}
	defer func() { _ = rows.Close() }()

	var out []Observation
	for rows.Next() {
		o, err := scanObservation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("candidate: reading %s observations: %w", market, err)
	}
	return out, nil
}

// PruneObservations deletes raw rows observed before `before` and reports how
// many went.
//
// It touches `observations` and nothing else. The candidate summaries — and with
// them every first_seen_at — are outside its reach by construction, which is D11
// expressed as the shape of the statement rather than as a warning above it.
func (s *Store) PruneObservations(ctx context.Context, before time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM observations WHERE observed_at < ?`, stamp(before))
	if err != nil {
		return 0, fmt.Errorf("candidate: pruning observations before %s: %w", stamp(before), err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("candidate: counting pruned observations: %w", err)
	}
	return n, nil
}

// --- candidate lifecycle --------------------------------------------------------

// Promote records that (market, symbol) is above the threshold at `at`.
//
// The rule that matters is the one that does nothing visible: a candidate that is
// active or cooled keeps its original FirstSeenAt. D1 calls this the single path
// by which the chase veto could be bypassed — a symbol that has already doubled
// leaves the ranking list for one scan, returns, and is "just discovered", so
// seen_late answers no for it forever after.
//
// Only an expired candidate starts over.
func (s *Store) Promote(ctx context.Context, market, symbol string, at time.Time) (Candidate, error) {
	market = strings.ToUpper(strings.TrimSpace(market))
	symbol = strings.TrimSpace(symbol)
	if market == "" || symbol == "" {
		return Candidate{}, fmt.Errorf("candidate: Promote needs a market and a symbol (got %q:%q)",
			market, symbol)
	}
	if at.IsZero() {
		return Candidate{}, fmt.Errorf("candidate: Promote of %s:%s has no instant", market, symbol)
	}
	at = at.UTC()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Candidate{}, fmt.Errorf("candidate: beginning a promotion: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	existing, found, err := readCandidate(ctx, tx, market, symbol)
	if err != nil {
		return Candidate{}, err
	}
	// A promotion that runs backwards is refused rather than applied. The
	// lifecycle is entirely derived from these instants, so an `at` older than
	// what is already recorded — an NTP step, a replayed reading, a source-supplied
	// timestamp — would move first_seen_at or the cooling clock in a direction time
	// does not go. first_seen_at is the one field this package exists to protect;
	// it must not be reachable by an unchecked argument.
	if found && at.Before(existing.LastSeenAt) {
		return Candidate{}, fmt.Errorf(
			"candidate: Promote of %s:%s at %s is older than its last reading %s; the lifecycle only moves forward",
			market, symbol, stamp(at), stamp(existing.LastSeenAt))
	}

	expired := found && stateAt(existing, s.cooling, s.staleness, at) == StateExpired
	firstSeen := at
	if found && !expired {
		firstSeen = existing.FirstSeenAt
	}

	// An expired candidate starts over completely, provenance included. Leaving
	// the previous life's sources and completeness in place would report the
	// source mix of a scan that ended hours ago against a candidate one second
	// old, which is D4's accounting describing the wrong scan.
	// Every placeholder is positional. Mixing `?1` with a bare `?` renumbers the
	// bare ones around it, which binds the arguments to the wrong columns — the
	// first version of this statement wrote market="0", symbol="KR".
	reset := boolToInt(expired)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO candidates (market, symbol, first_seen_at, last_seen_at, cooled_at)
		VALUES (?,?,?,?,NULL)
		ON CONFLICT(market, symbol) DO UPDATE SET
		  first_seen_at     = excluded.first_seen_at,
		  last_seen_at      = excluded.last_seen_at,
		  cooled_at         = NULL,
		  sources           = CASE WHEN ? THEN '' ELSE candidates.sources           END,
		  sources_attempted = CASE WHEN ? THEN 0  ELSE candidates.sources_attempted END,
		  sources_responded = CASE WHEN ? THEN 0  ELSE candidates.sources_responded END,
		  degraded          = CASE WHEN ? THEN 0  ELSE candidates.degraded          END,
		  -- The baseline is first_seen_at's twin (D17): kept through cooling and
		  -- re-entry, cleared only when the candidate expires and the next pass is
		  -- a different candidate. Carrying it across expiry would measure a new
		  -- life's expansion from a dead one's price; dropping it on re-entry would
		  -- reopen the D1 bypass one field along — a symbol that had already
		  -- doubled would leave the list for one scan and come back with a fresh
		  -- baseline, and extended would answer "no" for it from then on.
		  first_price        = CASE WHEN ? THEN NULL ELSE candidates.first_price        END,
		  first_price_at     = CASE WHEN ? THEN NULL ELSE candidates.first_price_at     END,
		  first_price_source = CASE WHEN ? THEN NULL ELSE candidates.first_price_source END`,
		market, symbol, stamp(firstSeen), stamp(at),
		reset, reset, reset, reset, reset, reset, reset)
	if err != nil {
		return Candidate{}, fmt.Errorf("candidate: promoting %s:%s: %w", market, symbol, err)
	}
	if err := tx.Commit(); err != nil {
		return Candidate{}, fmt.Errorf("candidate: committing the promotion of %s:%s: %w",
			market, symbol, err)
	}

	out := Candidate{
		Key:         Key{Market: market, Symbol: symbol},
		State:       StateActive,
		FirstSeenAt: firstSeen,
		LastSeenAt:  at,
	}
	if found && !expired {
		out.Sources = existing.Sources
		out.SourcesAttempted, out.SourcesResponded = existing.SourcesAttempted, existing.SourcesResponded
		out.Degraded = existing.Degraded
	}
	return out, nil
}

// Cool records that a candidate dropped below the threshold at `at`.
//
// It does not delete. The cooled row is what a re-entry inside the TTL joins, and
// deleting it here would recreate exactly the "newly discovered" bypass Promote
// exists to close. Cooling an unknown or already-cooled candidate is a no-op.
//
// `at` may not predate the candidate's last reading, for the reason Promote gives:
// a cooling instant from the past starts the TTL in the past, so a candidate that
// has been running all day reads as expired and the next promotion mints a fresh
// first_seen_at for a symbol that has already doubled. That is the D1 bypass
// arriving through the back door, and nothing about it looks like a failure.
func (s *Store) Cool(ctx context.Context, market, symbol string, at time.Time) error {
	market = strings.ToUpper(strings.TrimSpace(market))
	symbol = strings.TrimSpace(symbol)
	if at.IsZero() {
		return fmt.Errorf("candidate: Cool of %s:%s has no instant", market, symbol)
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE candidates SET cooled_at = ?
		 WHERE market = ? AND symbol = ? AND cooled_at IS NULL AND last_seen_at <= ?`,
		stamp(at.UTC()), market, symbol, stamp(at.UTC()))
	if err != nil {
		return fmt.Errorf("candidate: cooling %s:%s: %w", market, symbol, err)
	}
	// Nothing updated is ambiguous between three harmless cases (unknown, already
	// cooled, and the row moved on) and one that is not: a backwards instant. The
	// row is re-read only in that case, so the common path stays one statement.
	if n, _ := res.RowsAffected(); n == 0 {
		current, found, readErr := readCandidate(ctx, s.db, market, symbol)
		if readErr == nil && found && current.CooledAt.IsZero() && at.UTC().Before(current.LastSeenAt) {
			return fmt.Errorf(
				"candidate: Cool of %s:%s at %s is older than its last reading %s; "+
					"a cooling instant from the past expires a live candidate",
				market, symbol, stamp(at), stamp(current.LastSeenAt))
		}
	}
	return nil
}

// ErrNoCandidate means the named candidate is not in the store.
//
// NoteSources returns it rather than succeeding silently: D4's whole claim is
// that a reduced source panel is never hidden, and a provenance write that lands
// nowhere hides it more completely than a failed scan would.
var ErrNoCandidate = errors.New("candidate: no such candidate")

// NoteSources records which sources supported this candidate and how complete the
// scan that touched it was (D4).
//
// It is separate from Promote because the two answer different questions and a
// caller may know one without the other: Promote is "it crossed", this is "here
// is who said so and how much of the panel answered". A candidate produced from
// two of five sources is not wrong; a reader who does not know it was two of five
// is judging on less than they think.
//
// `sources` is merged into what the candidate already carries, not substituted for
// it. Candidate.Sources is "the sources that have supported this candidate" over
// its life (D1: one symbol raised by two sources is one candidate two sources
// supported), so a caller recording one source at a time must not erase the
// others. The completeness counters do replace — they describe the most recent
// scan, and a stale ratio would be a worse answer than the current one.
func (s *Store) NoteSources(ctx context.Context, market, symbol string,
	sources []SourceID, attempted, responded int, degraded bool) error {

	market = strings.ToUpper(strings.TrimSpace(market))
	symbol = strings.TrimSpace(symbol)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("candidate: beginning a source write: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	existing, found, err := readCandidate(ctx, tx, market, symbol)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("%w: %s:%s has no candidate to record sources against",
			ErrNoCandidate, market, symbol)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE candidates
		   SET sources = ?, sources_attempted = ?, sources_responded = ?, degraded = ?
		 WHERE market = ? AND symbol = ?`,
		encodeSources(append(append([]SourceID{}, existing.Sources...), sources...)),
		attempted, responded, boolToInt(degraded), market, symbol)
	if err != nil {
		return fmt.Errorf("candidate: recording the sources of %s:%s: %w", market, symbol, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("candidate: committing the sources of %s:%s: %w", market, symbol, err)
	}
	return nil
}

// --- the expansion baseline (D17) ------------------------------------------------

// Baseline is the first price anyone reported for a candidate.
//
// It is the denominator of the `extended` veto, and it is a stored fact rather
// than a derived one for the reason the schema comment gives: derived from the
// observations table it moves whenever the table is pruned or windowed, and it
// moves in the direction that reports a candidate which has tripled as having
// gained 50%.
//
// The instant and the source travel with the value because the value alone cannot
// be read. "Two times the baseline" is a different event when the baseline is
// twenty minutes old and when it is three days old, and a reader with only the
// number cannot tell which one they have (D17) — the same reason D9 stores both
// windows' actual seconds beside the ratio built from them.
type Baseline struct {
	// Price is the source's own string, verbatim (D6). Empty means no observation
	// has carried a price yet — which is not the same fact as a price of zero, and
	// the column is NULL rather than 0 so the two cannot merge.
	Price string
	// At is when that price was observed, and Source is who reported it.
	At     time.Time
	Source SourceID
}

// Recorded reports that a baseline exists. It is deliberately not "Price != 0":
// the absent case has no number at all.
func (b Baseline) Recorded() bool { return strings.TrimSpace(b.Price) != "" }

// NoteFirstPrice records a candidate's baseline price, once.
//
// Once is the whole contract. The first priced observation of a candidate's life
// sets it and nothing overwrites it while that life lasts — not a later scan, not
// a session boundary, not a re-entry after cooling. D17 decided against a session
// reset explicitly: a tripling spread over three days is still chasing, and the
// lifecycle already requires the candidate to have stayed on the lists throughout
// (forty minutes of silence expires it).
//
// It returns the baseline in force after the call, which for every call but the
// first is the one that was already there. Callers can therefore write on every
// priced reading without checking first, and the store is what makes it idempotent.
//
// An unknown candidate is an error, for NoteSources' reason: a provenance write
// that lands nowhere and returns nil hides the gap more completely than a failure
// would.
func (s *Store) NoteFirstPrice(ctx context.Context, market, symbol, price string,
	at time.Time, source SourceID) (Baseline, error) {

	market = strings.ToUpper(strings.TrimSpace(market))
	symbol = strings.TrimSpace(symbol)
	price = strings.TrimSpace(price)
	if price == "" {
		return Baseline{}, fmt.Errorf(
			"candidate: NoteFirstPrice of %s:%s has no price; absent is not a baseline", market, symbol)
	}
	if at.IsZero() {
		return Baseline{}, fmt.Errorf(
			"candidate: NoteFirstPrice of %s:%s has no instant; a baseline whose age is unknown "+
				"cannot be read", market, symbol)
	}
	id, err := normaliseSource(source)
	if err != nil {
		return Baseline{}, fmt.Errorf("candidate: NoteFirstPrice of %s:%s: %w", market, symbol, err)
	}
	at = at.UTC()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Baseline{}, fmt.Errorf("candidate: beginning a baseline write: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var (
		stored   sql.NullString
		storedAt sql.NullString
		storedBy sql.NullString
	)
	row := tx.QueryRowContext(ctx,
		`SELECT first_price, first_price_at, first_price_source FROM candidates
		  WHERE market = ? AND symbol = ?`, market, symbol)
	switch err := row.Scan(&stored, &storedAt, &storedBy); {
	case errors.Is(err, sql.ErrNoRows):
		return Baseline{}, fmt.Errorf("%w: %s:%s has no candidate to record a baseline against",
			ErrNoCandidate, market, symbol)
	case err != nil:
		return Baseline{}, fmt.Errorf("candidate: reading the baseline of %s:%s: %w", market, symbol, err)
	}
	if strings.TrimSpace(stored.String) != "" {
		return decodeBaseline(stored, storedAt, storedBy, market, symbol)
	}

	// The WHERE clause repeats the guard the read just made. Two writers racing on
	// the same candidate would otherwise both see NULL and the later one would
	// overwrite the earlier baseline with a newer price — the exact re-basing this
	// column exists to stop, arriving through concurrency instead of through
	// retention. IMMEDIATE transactions make that unlikely; the predicate makes it
	// impossible.
	_, err = tx.ExecContext(ctx, `
		UPDATE candidates
		   SET first_price = ?, first_price_at = ?, first_price_source = ?
		 WHERE market = ? AND symbol = ? AND first_price IS NULL`,
		price, stamp(at), string(id), market, symbol)
	if err != nil {
		return Baseline{}, fmt.Errorf("candidate: recording the baseline of %s:%s: %w",
			market, symbol, err)
	}
	if err := tx.Commit(); err != nil {
		return Baseline{}, fmt.Errorf("candidate: committing the baseline of %s:%s: %w",
			market, symbol, err)
	}
	return Baseline{Price: price, At: at, Source: id}, nil
}

// Baseline returns a candidate's stored first price.
//
// The second return distinguishes "no such candidate" from "a candidate with no
// baseline yet". Both produce an unrecorded Baseline, and only one of them is a
// question about the candidate lifecycle.
func (s *Store) Baseline(ctx context.Context, market, symbol string) (Baseline, bool, error) {
	market = strings.ToUpper(strings.TrimSpace(market))
	symbol = strings.TrimSpace(symbol)

	var price, at, source sql.NullString
	row := s.db.QueryRowContext(ctx,
		`SELECT first_price, first_price_at, first_price_source FROM candidates
		  WHERE market = ? AND symbol = ?`, market, symbol)
	switch err := row.Scan(&price, &at, &source); {
	case errors.Is(err, sql.ErrNoRows):
		return Baseline{}, false, nil
	case err != nil:
		return Baseline{}, false, fmt.Errorf("candidate: reading the baseline of %s:%s: %w",
			market, symbol, err)
	}
	b, err := decodeBaseline(price, at, source, market, symbol)
	return b, true, err
}

func decodeBaseline(price, at, source sql.NullString, market, symbol string) (Baseline, error) {
	if strings.TrimSpace(price.String) == "" {
		return Baseline{}, nil
	}
	b := Baseline{Price: price.String, Source: SourceID(source.String)}
	if at.Valid {
		parsed, err := parseStamp(at.String)
		if err != nil {
			return Baseline{}, fmt.Errorf(
				"candidate %s:%s has an unreadable first_price_at: %w", market, symbol, err)
		}
		b.At = parsed
	}
	return b, nil
}

// Candidates returns every stored candidate with its state as of `at`.
//
// `at` is an argument rather than the clock because the state is derived, not
// stored: a candidate that cooled before the process restarted has to read as
// expired afterwards without a sweeper having run. Expired rows are returned
// rather than hidden — the count of candidates we saw and let go is part of the
// evidence this system is early.
func (s *Store) Candidates(ctx context.Context, at time.Time) ([]Candidate, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT market, symbol, first_seen_at, last_seen_at, cooled_at,
		       sources, sources_attempted, sources_responded, degraded
		  FROM candidates
		 ORDER BY market, symbol`)
	if err != nil {
		return nil, fmt.Errorf("candidate: reading candidates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Candidate
	for rows.Next() {
		c, err := scanCandidate(rows)
		if err != nil {
			return nil, err
		}
		c.State = stateAt(c, s.cooling, s.staleness, at)
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("candidate: reading candidates: %w", err)
	}
	return out, nil
}

// Candidate returns one candidate with its state as of `at`.
func (s *Store) Candidate(ctx context.Context, market, symbol string, at time.Time) (Candidate, bool, error) {
	market = strings.ToUpper(strings.TrimSpace(market))
	symbol = strings.TrimSpace(symbol)
	c, found, err := readCandidate(ctx, s.db, market, symbol)
	if err != nil || !found {
		return Candidate{}, found, err
	}
	c.State = stateAt(c, s.cooling, s.staleness, at)
	return c, true, nil
}

// stateAt derives the lifecycle state from the stored timestamps.
//
// # Cooling is not only explicit
//
// The obvious reading — cooled_at is set, therefore cooling started — needs a
// writer, and the writer is the scan process. When that process is killed, or
// backs off, or simply never sees the symbol on a list again, no Cool is written.
// Deriving the state from cooled_at alone would leave such a candidate active
// forever, still holding a first_seen_at from a previous session, so seen_late
// would answer "we have known about this since Monday" about a symbol nobody has
// looked at since. That is the same "keeps working while quietly changing what it
// means" shape as the re-entry bypass, in the opposite direction.
//
// So a candidate nobody has re-promoted within `staleness` cools implicitly, at
// last_seen_at + staleness, and the cooling TTL runs from there. `staleness` has
// to sit above the longest planned backoff (300s), or an ordinary rate-limit
// retreat would expire the candidates it was protecting.
//
// # The boundary
//
// Closed at the far end: cooled_at + ttl reads as expired, not as the last
// instant of cooling. A candidate is re-entrant for strictly less than the TTL,
// which a test can pin without an off-by-one argument.
func stateAt(c Candidate, ttl, staleness time.Duration, at time.Time) State {
	cooledAt := c.CooledAt
	if cooledAt.IsZero() {
		if at.Sub(c.LastSeenAt) < staleness {
			return StateActive
		}
		cooledAt = c.LastSeenAt.Add(staleness)
	}
	if !at.Before(cooledAt.Add(ttl)) {
		return StateExpired
	}
	return StateCooled
}

// --- row plumbing ---------------------------------------------------------------

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface{ Scan(dest ...any) error }

// querier is satisfied by *sql.DB and *sql.Tx, so readCandidate serves both the
// plain read and the read inside Promote's transaction.
type querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func readCandidate(ctx context.Context, q querier, market, symbol string) (Candidate, bool, error) {
	row := q.QueryRowContext(ctx, `
		SELECT market, symbol, first_seen_at, last_seen_at, cooled_at,
		       sources, sources_attempted, sources_responded, degraded
		  FROM candidates WHERE market = ? AND symbol = ?`, market, symbol)
	c, err := scanCandidate(row)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Candidate{}, false, nil
	case err != nil:
		return Candidate{}, false, err
	}
	return c, true, nil
}

func scanCandidate(row rowScanner) (Candidate, error) {
	var (
		c        Candidate
		first    string
		last     string
		cooled   sql.NullString
		sources  string
		degraded int
	)
	err := row.Scan(&c.Market, &c.Symbol, &first, &last, &cooled,
		&sources, &c.SourcesAttempted, &c.SourcesResponded, &degraded)
	if err != nil {
		return Candidate{}, err
	}
	if c.FirstSeenAt, err = parseStamp(first); err != nil {
		return Candidate{}, fmt.Errorf("candidate %s:%s has an unreadable first_seen_at: %w",
			c.Market, c.Symbol, err)
	}
	if c.LastSeenAt, err = parseStamp(last); err != nil {
		return Candidate{}, fmt.Errorf("candidate %s:%s has an unreadable last_seen_at: %w",
			c.Market, c.Symbol, err)
	}
	if cooled.Valid {
		if c.CooledAt, err = parseStamp(cooled.String); err != nil {
			return Candidate{}, fmt.Errorf("candidate %s:%s has an unreadable cooled_at: %w",
				c.Market, c.Symbol, err)
		}
	}
	c.Sources = decodeSources(sources)
	c.Degraded = degraded == 1
	return c, nil
}

func scanObservation(row rowScanner) (Observation, error) {
	var (
		o                    Observation
		observed             string
		newly, fallback      int
		value, volume, price sql.NullString
		dayHigh, dayLow      sql.NullString
		source               string
	)
	err := row.Scan(&o.Market, &o.Symbol, &source, &observed,
		&o.Reported.Rank, &o.Reported.RankTotal, &newly,
		&value, &volume, &price, &dayHigh, &dayLow, &fallback)
	if err != nil {
		return Observation{}, fmt.Errorf("candidate: scanning an observation: %w", err)
	}
	if o.ObservedAt, err = parseStamp(observed); err != nil {
		return Observation{}, fmt.Errorf("candidate: observation of %s:%s has an unreadable instant: %w",
			o.Market, o.Symbol, err)
	}
	o.Source = SourceID(source)
	o.Reported.NewlyListed = newly == 1
	o.ViaFallback = fallback == 1
	o.Reported.TradingValue = value.String
	o.Reported.TradingVolume = volume.String
	o.Reported.Price = price.String
	o.Reported.DayHigh = dayHigh.String
	o.Reported.DayLow = dayLow.String
	return o, nil
}

// stampLayout is fixed width, and that is the whole point.
//
// observed_at is a TEXT column, so every `<`, `>=` and `ORDER BY` in this file is
// a BINARY string comparison. time.RFC3339Nano — the obvious choice, and what
// this was until the section-1 review — strips trailing zeros from the fraction,
// which makes text order stop being time order:
//
//	"2026-07-28T09:00:00Z"    is 09:00:00.0
//	"2026-07-28T09:00:00.5Z"  is 09:00:00.5, half a second LATER
//	'Z' is 0x5A and '.' is 0x2E, so the later instant sorts FIRST.
//
// Three things broke on that and not one of them failed: PruneObservations
// deleted rows newer than the cutoff it was given, Observations(since) dropped
// rows from inside the window it was asked for, and "oldest first" was not.
// Sections 3 and 4 compute the acceleration and the veto from exactly those two
// queries. Every test in this package used whole-second instants, which is the
// one branch where all stamps have the same (empty) fraction and the defect
// cannot appear — see TestTheStoredInstantOrdersLikeTime.
//
// A fixed-width fraction restores the invariant that makes a text timestamp
// usable as a range key at all: the same length for every instant, so byte order
// is chronological order. Text rather than an integer so the file stays readable
// with sqlite3 when somebody is working out why a candidate has the
// first_seen_at it has.
const stampLayout = "2006-01-02T15:04:05.000000000Z07:00"

func stamp(t time.Time) string { return t.UTC().Format(stampLayout) }

// parseStamp reads a stored instant. It accepts any RFC3339 fraction width
// rather than only the one stamp writes, so a store left by an older build is
// still readable instead of failing every row.
func parseStamp(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

// nullable maps an absent value to SQL NULL and trims what it keeps.
//
// Trimming matters because "present" and "usable" have to be the same predicate:
// a padded " 0 " that survives to the veto layer is a value that reads as
// measured and then fails to parse, which is a third state nothing accounts for.
func nullable(s string) any {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

// normaliseSource puts a source id into the one spelling the table stores.
//
// Market and symbol were normalised on every read from the beginning and the
// source was not, so a padded or mis-cased id — from a configuration file, from a
// hand-written query, from a caller building the id by concatenation — selected no
// rows and returned no error. Downstream that is an empty series, and an empty
// series is WARMING_UP: "the scan has just started, nothing needs fixing", for
// something that will never start.
//
// Lowercasing is safe rather than lossy: every SourceID in this package is lower
// snake case, so the map is injective over the ids that exist, and no spelling can
// be folded onto a different source.
//
// An id that is empty after trimming is refused. That one cannot be a typo in the
// case of a real source — it is a caller that has no source at all — and answering
// it with silence is the failure this function exists to stop, in its most
// complete form.
//
// It deliberately does not check membership of the known set. That list lives in
// candidate.go beside the constants, and a second copy here would be a list that
// goes stale the first time a source is added — refusing a legitimate source is a
// worse failure than the one being fixed.
func normaliseSource(id SourceID) (SourceID, error) {
	trimmed := strings.ToLower(strings.TrimSpace(string(id)))
	if trimmed == "" {
		return "", fmt.Errorf("candidate: %q is not a source id; a read with no source "+
			"would answer with an empty series and no error", string(id))
	}
	return SourceID(trimmed), nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// encodeSources stores the supporting sources as a sorted comma-separated list.
// Sorted so the column is comparable between two readings, which is what makes
// "the source mix changed" visible rather than being a reordering.
func encodeSources(in []SourceID) string {
	if len(in) == 0 {
		return ""
	}
	seen := make(map[SourceID]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, string(s))
	}
	sort.Strings(out)
	return strings.Join(out, ",")
}

func decodeSources(s string) []SourceID {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]SourceID, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, SourceID(p))
		}
	}
	return out
}
