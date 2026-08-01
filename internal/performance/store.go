package performance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	SchemaVersion = 1
	RawRetention  = 90 * 24 * time.Hour
	PruneCadence  = 24 * time.Hour
	MaxPruneRows  = 500
)

var ErrSchemaTooNew = errors.New("performance: database schema is newer than this build")

type Store struct {
	db   *sql.DB
	path string
}

func Open(path string) (*Store, error) {
	return openWithSchema(path, schemaV1, SchemaVersion)
}

func openWithSchema(path, schema string, targetVersion int) (*Store, error) {
	path = filepath.Clean(path)
	if strings.TrimSpace(path) == "" || path == "." {
		return nil, errors.New("performance: database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("performance: creating data directory: %w", err)
	}
	q := url.Values{}
	q.Add("_pragma", "journal_mode(wal)")
	q.Add("_pragma", "synchronous(full)")
	q.Add("_pragma", "foreign_keys(on)")
	q.Add("_pragma", "busy_timeout(5000)")
	q.Set("_txlock", "immediate")
	db, err := sql.Open("sqlite", "file:"+path+"?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("performance: opening database: %w", err)
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	store := &Store{db: db, path: path}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("performance: connecting database: %w", err)
	}
	if err := store.migrate(context.Background(), schema, targetVersion); err != nil {
		db.Close()
		return nil, err
	}
	for _, name := range []string{path, path + "-wal", path + "-shm"} {
		if _, err := os.Stat(name); err == nil {
			_ = os.Chmod(name, 0o600)
		}
	}
	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Path() string { return s.path }

func (s *Store) SchemaVersion(ctx context.Context) (int, error) {
	var version int
	if err := s.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return 0, fmt.Errorf("performance: reading schema version: %w", err)
	}
	return version, nil
}

func (s *Store) migrate(ctx context.Context, schema string, targetVersion int) error {
	version, err := s.SchemaVersion(ctx)
	if err != nil {
		return err
	}
	if version > targetVersion {
		return fmt.Errorf("%w: found %d, understand %d", ErrSchemaTooNew, version, targetVersion)
	}
	if version == targetVersion {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("performance: starting migration: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("performance: applying schema v1: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "PRAGMA user_version = "+strconv.Itoa(targetVersion)); err != nil {
		return fmt.Errorf("performance: recording schema v1: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("performance: committing schema v1: %w", err)
	}
	return nil
}

func (s *Store) AppendTrade(ctx context.Context, trade Trade) error {
	if err := trade.validate(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, insertTradeSQL, tradeArgs(trade)...)
	if err != nil {
		return fmt.Errorf("performance: appending trade %s: %w", trade.ID, err)
	}
	return nil
}

func (s *Store) AppendObservations(ctx context.Context, observations []Observation) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("performance: starting observation append: %w", err)
	}
	defer tx.Rollback()
	if err := appendObservations(ctx, tx, observations); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("performance: committing observations: %w", err)
	}
	return nil
}

func (s *Store) Collect(ctx context.Context, trade Trade, observations []Observation, calculatedAt time.Time) (Snapshot, error) {
	if err := trade.validate(); err != nil {
		return Snapshot{}, err
	}
	for _, observation := range observations {
		if err := observation.validate(); err != nil {
			return Snapshot{}, err
		}
	}
	snapshot := Measure(trade, observations, calculatedAt)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Snapshot{}, fmt.Errorf("performance: starting collection: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, insertTradeSQL, tradeArgs(trade)...); err != nil {
		return Snapshot{}, fmt.Errorf("performance: appending trade %s: %w", trade.ID, err)
	}
	if err := appendObservations(ctx, tx, observations); err != nil {
		return Snapshot{}, err
	}
	if err := appendSnapshot(ctx, tx, trade.ID, snapshot); err != nil {
		return Snapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return Snapshot{}, fmt.Errorf("performance: committing collection: %w", err)
	}
	return snapshot, nil
}

func appendObservations(ctx context.Context, tx *sql.Tx, observations []Observation) error {
	for _, observation := range observations {
		if err := observation.validate(); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO price_observations
			(id, position_id, observed_at, price, source, source_version) VALUES (?,?,?,?,?,?)`,
			observation.ID, observation.PositionID, timestamp(observation.At),
			strings.TrimSpace(observation.Price), observation.Source, observation.SourceVersion); err != nil {
			return fmt.Errorf("performance: appending observation %s: %w", observation.ID, err)
		}
	}
	return nil
}

func appendSnapshot(ctx context.Context, tx *sql.Tx, tradeID string, snapshot Snapshot) error {
	result, err := tx.ExecContext(ctx, `INSERT INTO measurement_snapshots
		(trade_id, calculated_at, semantics_version, lineage_status) VALUES (?,?,?,?)`,
		tradeID, timestamp(snapshot.CalculatedAt), snapshot.SemanticsVersion, snapshot.LineageStatus)
	if err != nil {
		return fmt.Errorf("performance: appending measurement snapshot: %w", err)
	}
	snapshotID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("performance: reading measurement identity: %w", err)
	}
	for _, metric := range snapshot.Markouts {
		if _, err := tx.ExecContext(ctx, insertMetricSQL, snapshotID, "markout_"+strconv.Itoa(metric.Minutes),
			metric.Status, nil, nullable(metric.GrossPct), nullable(metric.CostAdjustedPct),
			nullable(metric.ObservationID), nullableTime(metric.ObservedAt), nullable(metric.Source), nullable(metric.SourceVersion)); err != nil {
			return fmt.Errorf("performance: appending %dm markout: %w", metric.Minutes, err)
		}
	}
	for key, metric := range map[string]Metric{"slippage": snapshot.Slippage, "mfe": snapshot.MFE, "mae": snapshot.MAE} {
		if _, err := tx.ExecContext(ctx, insertMetricSQL, snapshotID, key, metric.Status,
			nullable(metric.Value), nil, nil, nullable(metric.ObservationID), nullableTime(metric.ObservedAt),
			nullable(metric.Source), nullable(metric.SourceVersion)); err != nil {
			return fmt.Errorf("performance: appending %s metric: %w", key, err)
		}
	}
	return nil
}

type PruneResult struct {
	Deleted      int
	Skipped      bool
	LockDuration time.Duration
}

func (s *Store) PruneDue(ctx context.Context, now time.Time) (PruneResult, error) {
	started := time.Now()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PruneResult{}, fmt.Errorf("performance: starting prune: %w", err)
	}
	defer tx.Rollback()
	var lastRaw sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT value FROM maintenance_state WHERE key='last_pruned_at'`).Scan(&lastRaw); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return PruneResult{}, fmt.Errorf("performance: reading prune cadence: %w", err)
	}
	if lastRaw.Valid {
		last, err := time.Parse(time.RFC3339Nano, lastRaw.String)
		if err != nil {
			return PruneResult{}, fmt.Errorf("performance: invalid last prune instant: %w", err)
		}
		if now.UTC().Before(last.Add(PruneCadence)) {
			if err := tx.Commit(); err != nil {
				return PruneResult{}, err
			}
			return PruneResult{Skipped: true, LockDuration: time.Since(started)}, nil
		}
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM price_observations WHERE id IN (
		SELECT id FROM price_observations WHERE observed_at < ? ORDER BY observed_at, id LIMIT 500
	)`, timestamp(now.UTC().Add(-RawRetention)))
	if err != nil {
		return PruneResult{}, fmt.Errorf("performance: pruning observations: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return PruneResult{}, fmt.Errorf("performance: reading pruned row count: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO maintenance_state(key,value) VALUES('last_pruned_at',?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value`, timestamp(now)); err != nil {
		return PruneResult{}, fmt.Errorf("performance: recording prune cadence: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return PruneResult{}, fmt.Errorf("performance: committing prune: %w", err)
	}
	return PruneResult{Deleted: int(deleted), LockDuration: time.Since(started)}, nil
}

func (s *Store) RecentObservationCount(ctx context.Context, since time.Time) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM price_observations WHERE observed_at >= ?`, timestamp(since)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("performance: counting recent observations: %w", err)
	}
	return count, nil
}

func (s *Store) RecentObservationQueryPlan(ctx context.Context, since time.Time) (string, error) {
	rows, err := s.db.QueryContext(ctx, `EXPLAIN QUERY PLAN SELECT count(*) FROM price_observations WHERE observed_at >= ?`, timestamp(since))
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var lines []string
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
			return "", err
		}
		lines = append(lines, detail)
	}
	return strings.Join(lines, "\n"), rows.Err()
}

func tradeArgs(trade Trade) []any {
	l := trade.Lineage
	return []any{
		trade.ID, nullable(l.CandidateLifeID), nullable(l.ThresholdVersion), nullable(l.ThresholdSetDigest), nullable(l.EvidenceDigest),
		nullable(l.LaneID), nullable(l.LaneVersion), nullable(l.DecisionID), nullable(l.AttemptID), nullable(l.OrderID), nullable(l.FillID),
		l.PositionID, nullable(l.CloseID), nullable(l.PolicyID), nullable(l.PolicyVersion), l.Status(), strings.ToLower(strings.TrimSpace(trade.Market)), trade.Side,
		nullableTime(trade.DecisionAt), nullable(trade.DecisionPrice), timestamp(trade.EntryAt), trade.EntryPrice, trade.Quantity, trade.CostTotal,
		trade.RealizedPnLAfterCosts, trade.RealizedR, timestamp(trade.ClosedAt),
	}
}

func nullable(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return timestamp(value)
}

func timestamp(value time.Time) string { return value.UTC().Format("2006-01-02T15:04:05.000000000Z") }

const insertTradeSQL = `INSERT INTO performance_trades (
	id, candidate_life_id, threshold_version, threshold_set_digest, evidence_digest,
	lane_id, lane_version, decision_id, attempt_id, order_id, fill_id, position_id,
	close_id, policy_id, policy_version, lineage_status, market, side, decision_at,
	decision_price, entry_at, entry_price, quantity, cost_total,
	realized_pnl_after_costs, realized_r, closed_at
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

const insertMetricSQL = `INSERT INTO metric_observations (
	snapshot_id, metric_key, status, value, gross_value, cost_adjusted_value,
	observation_id, observed_at, source, source_version
) VALUES (?,?,?,?,?,?,?,?,?,?)`

const schemaV1 = `
CREATE TABLE performance_trades (
	id TEXT PRIMARY KEY,
	candidate_life_id TEXT,
	threshold_version TEXT,
	threshold_set_digest TEXT,
	evidence_digest TEXT,
	lane_id TEXT,
	lane_version TEXT,
	decision_id TEXT,
	attempt_id TEXT,
	order_id TEXT,
	fill_id TEXT,
	position_id TEXT NOT NULL UNIQUE,
	close_id TEXT,
	policy_id TEXT,
	policy_version TEXT,
	lineage_status TEXT NOT NULL CHECK(lineage_status IN ('complete','link_missing')),
	market TEXT NOT NULL,
	side TEXT NOT NULL CHECK(side IN ('BUY','SELL')),
	decision_at TEXT,
	decision_price TEXT,
	entry_at TEXT NOT NULL,
	entry_price TEXT NOT NULL,
	quantity TEXT NOT NULL,
	cost_total TEXT NOT NULL,
	realized_pnl_after_costs TEXT NOT NULL,
	realized_r TEXT NOT NULL,
	closed_at TEXT NOT NULL
) STRICT;
CREATE INDEX idx_performance_trades_window
	ON performance_trades(closed_at, lineage_status, market, lane_id, lane_version, policy_id, policy_version);
CREATE INDEX idx_performance_trades_lane
	ON performance_trades(lane_id, lane_version, policy_id, policy_version, closed_at);

CREATE TABLE price_observations (
	id TEXT PRIMARY KEY,
	position_id TEXT NOT NULL,
	observed_at TEXT NOT NULL,
	price TEXT NOT NULL,
	source TEXT NOT NULL,
	source_version TEXT NOT NULL
) STRICT;
CREATE INDEX idx_price_observations_position_at ON price_observations(position_id, observed_at, id);
CREATE INDEX idx_price_observations_at ON price_observations(observed_at, id);

CREATE TABLE measurement_snapshots (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	trade_id TEXT NOT NULL REFERENCES performance_trades(id),
	calculated_at TEXT NOT NULL,
	semantics_version TEXT NOT NULL,
	lineage_status TEXT NOT NULL CHECK(lineage_status IN ('complete','link_missing'))
) STRICT;
CREATE INDEX idx_measurement_snapshots_trade ON measurement_snapshots(trade_id, id);

CREATE TABLE metric_observations (
	snapshot_id INTEGER NOT NULL REFERENCES measurement_snapshots(id),
	metric_key TEXT NOT NULL CHECK(metric_key IN ('markout_5','markout_15','markout_30','slippage','mfe','mae')),
	status TEXT NOT NULL CHECK(status IN ('complete','not_measured')),
	value TEXT,
	gross_value TEXT,
	cost_adjusted_value TEXT,
	observation_id TEXT,
	observed_at TEXT,
	source TEXT,
	source_version TEXT,
	PRIMARY KEY(snapshot_id, metric_key)
) STRICT;

CREATE TABLE maintenance_state (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
) STRICT;
`
