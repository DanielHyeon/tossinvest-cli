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
	SchemaVersion         = 1
	RawRetention          = 90 * 24 * time.Hour
	PruneCadence          = 24 * time.Hour
	MaxPruneRows          = 500
	MaxPruneBatchesPerRun = 4
)

var (
	ErrSchemaTooNew      = errors.New("performance: database schema is newer than this build")
	ErrImmutableConflict = errors.New("performance: immutable identity has divergent bytes")
	transactionPhaseHook = func(string) {}
)

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
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
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
			if err := os.Chmod(name, 0o600); err != nil {
				db.Close()
				return nil, fmt.Errorf("performance: securing %s: %w", filepath.Base(name), err)
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			db.Close()
			return nil, fmt.Errorf("performance: checking %s: %w", filepath.Base(name), err)
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
	transactionPhaseHook("migration_after_schema")
	if _, err := tx.ExecContext(ctx, "PRAGMA user_version = "+strconv.Itoa(targetVersion)); err != nil {
		return fmt.Errorf("performance: recording schema v1: %w", err)
	}
	transactionPhaseHook("migration_after_version")
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("performance: committing schema v1: %w", err)
	}
	return nil
}

func (s *Store) AppendTrade(ctx context.Context, trade Trade) error {
	if err := trade.validate(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("performance: starting trade append: %w", err)
	}
	defer tx.Rollback()
	if err := compareAndAppendTrade(ctx, tx, trade); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("performance: committing trade append: %w", err)
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
	transactionPhaseHook("observations_after_rows")
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
		if observation.PositionID != trade.Lineage.PositionID {
			return Snapshot{}, fmt.Errorf("performance: observation %s position does not match trade", observation.ID)
		}
	}
	if calculatedAt.IsZero() {
		return Snapshot{}, errors.New("performance: calculated-at is required")
	}
	snapshot := Measure(trade, observations, calculatedAt)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Snapshot{}, fmt.Errorf("performance: starting collection: %w", err)
	}
	defer tx.Rollback()
	if err := compareAndAppendTrade(ctx, tx, trade); err != nil {
		return Snapshot{}, err
	}
	transactionPhaseHook("collect_after_trade")
	if err := appendObservations(ctx, tx, observations); err != nil {
		return Snapshot{}, err
	}
	transactionPhaseHook("collect_after_observations")
	if err := appendSnapshot(ctx, tx, trade.ID, snapshot); err != nil {
		return Snapshot{}, err
	}
	transactionPhaseHook("collect_after_snapshot")
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
		wanted := []any{observation.ID, observation.PositionID, timestamp(observation.At),
			strings.TrimSpace(observation.Price), strings.TrimSpace(observation.Source), strings.TrimSpace(observation.SourceVersion)}
		equal, exists, err := immutableRowEqual(ctx, tx,
			`SELECT id,position_id,observed_at,price,source,source_version FROM price_observations WHERE id=?`,
			[]any{observation.ID}, wanted)
		if err != nil {
			return fmt.Errorf("performance: reading observation %s: %w", observation.ID, err)
		}
		if exists {
			if !equal {
				return fmt.Errorf("%w: observation %s", ErrImmutableConflict, observation.ID)
			}
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO price_observations
			(id, position_id, observed_at, price, source, source_version) VALUES (?,?,?,?,?,?)`, wanted...); err != nil {
			return fmt.Errorf("performance: appending observation %s: %w", observation.ID, err)
		}
		// A late backfill older than the most recently completed cutoff makes
		// retention due again immediately. Keeping this in the append transaction
		// prevents a successful prune marker from hiding newly arrived backlog for
		// another 24 hours.
		if _, err := tx.ExecContext(ctx, `DELETE FROM maintenance_state
			WHERE key='last_pruned_at' AND value > ?`, timestamp(observation.At.UTC().Add(RawRetention))); err != nil {
			return fmt.Errorf("performance: rescheduling retention for observation %s: %w", observation.ID, err)
		}
	}
	return nil
}

func appendSnapshot(ctx context.Context, tx *sql.Tx, tradeID string, snapshot Snapshot) error {
	records, err := snapshotMetricRecords(snapshot)
	if err != nil {
		return err
	}
	var snapshotID int64
	var persistedStatus string
	err = tx.QueryRowContext(ctx, `SELECT id,lineage_status FROM measurement_snapshots
		WHERE trade_id=? AND calculated_at=? AND semantics_version=?`,
		tradeID, timestamp(snapshot.CalculatedAt), snapshot.SemanticsVersion).Scan(&snapshotID, &persistedStatus)
	if err == nil {
		if persistedStatus != string(snapshot.LineageStatus) {
			return fmt.Errorf("%w: measurement snapshot %s@%s", ErrImmutableConflict, tradeID, timestamp(snapshot.CalculatedAt))
		}
		return compareSnapshotMetrics(ctx, tx, snapshotID, records)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("performance: reading measurement snapshot: %w", err)
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO measurement_snapshots
		(trade_id, calculated_at, semantics_version, lineage_status) VALUES (?,?,?,?)`,
		tradeID, timestamp(snapshot.CalculatedAt), snapshot.SemanticsVersion, snapshot.LineageStatus)
	if err != nil {
		return fmt.Errorf("performance: appending measurement snapshot: %w", err)
	}
	snapshotID, err = result.LastInsertId()
	if err != nil {
		return fmt.Errorf("performance: reading measurement identity: %w", err)
	}
	for _, record := range records {
		if _, err := tx.ExecContext(ctx, insertMetricSQL, append([]any{snapshotID}, record.values()...)...); err != nil {
			return fmt.Errorf("performance: appending %s metric: %w", record.key, err)
		}
	}
	return nil
}

func compareAndAppendTrade(ctx context.Context, tx *sql.Tx, trade Trade) error {
	wanted := tradeArgs(trade)
	equal, exists, err := immutableRowEqual(ctx, tx, `SELECT
		id, candidate_life_id, threshold_version, threshold_set_digest, evidence_digest,
		lane_id, lane_version, decision_id, risk_intent_id, attempt_id, mutation_attempt_id, order_id, fill_id, position_id,
		close_id, policy_id, policy_version, lineage_status, market, side, decision_at,
		decision_price, entry_at, entry_price, quantity, cost_total,
		realized_pnl_after_costs, realized_r, closed_at
		FROM performance_trades WHERE id=?`, []any{trade.ID}, wanted)
	if err != nil {
		return fmt.Errorf("performance: reading trade %s: %w", trade.ID, err)
	}
	if exists {
		if !equal {
			return fmt.Errorf("%w: trade %s", ErrImmutableConflict, trade.ID)
		}
		return nil
	}
	if _, err := tx.ExecContext(ctx, insertTradeSQL, wanted...); err != nil {
		return fmt.Errorf("performance: appending trade %s: %w", trade.ID, err)
	}
	return nil
}

func immutableRowEqual(ctx context.Context, tx *sql.Tx, query string, args, wanted []any) (bool, bool, error) {
	persisted := make([]sql.NullString, len(wanted))
	targets := make([]any, len(persisted))
	for i := range persisted {
		targets[i] = &persisted[i]
	}
	err := tx.QueryRowContext(ctx, query, args...).Scan(targets...)
	if errors.Is(err, sql.ErrNoRows) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	for i, value := range wanted {
		if value == nil {
			if persisted[i].Valid {
				return false, true, nil
			}
			continue
		}
		if !persisted[i].Valid || persisted[i].String != fmt.Sprint(value) {
			return false, true, nil
		}
	}
	return true, true, nil
}

type snapshotMetricRecord struct {
	key, value, gross, costAdjusted, observationID, observedAt, source, sourceVersion string
	status                                                                            Status
}

func (r snapshotMetricRecord) values() []any {
	return []any{r.key, r.status, nullable(r.value), nullable(r.gross), nullable(r.costAdjusted),
		nullable(r.observationID), nullable(r.observedAt), nullable(r.source), nullable(r.sourceVersion)}
}

func snapshotMetricRecords(snapshot Snapshot) ([]snapshotMetricRecord, error) {
	if snapshot.CalculatedAt.IsZero() || snapshot.SemanticsVersion == "" {
		return nil, errors.New("performance: snapshot identity is incomplete")
	}
	markouts := make(map[int]MarkoutMetric, len(snapshot.Markouts))
	for _, metric := range snapshot.Markouts {
		if _, exists := markouts[metric.Minutes]; exists {
			return nil, fmt.Errorf("performance: duplicate %dm markout", metric.Minutes)
		}
		markouts[metric.Minutes] = metric
	}
	records := make([]snapshotMetricRecord, 0, 6)
	for _, minutes := range []int{5, 15, 30} {
		metric, exists := markouts[minutes]
		if !exists {
			return nil, fmt.Errorf("performance: missing %dm markout", minutes)
		}
		records = append(records, snapshotMetricRecord{
			key: "markout_" + strconv.Itoa(minutes), status: metric.Status,
			gross: metric.GrossPct, costAdjusted: metric.CostAdjustedPct,
			observationID: metric.ObservationID, observedAt: nullableTimestamp(metric.ObservedAt),
			source: metric.Source, sourceVersion: metric.SourceVersion,
		})
	}
	for _, item := range []struct {
		key    string
		metric Metric
	}{{"slippage", snapshot.Slippage}, {"mfe", snapshot.MFE}, {"mae", snapshot.MAE}} {
		records = append(records, snapshotMetricRecord{
			key: item.key, status: item.metric.Status, value: item.metric.Value,
			observationID: item.metric.ObservationID, observedAt: nullableTimestamp(item.metric.ObservedAt),
			source: item.metric.Source, sourceVersion: item.metric.SourceVersion,
		})
	}
	return records, nil
}

func compareSnapshotMetrics(ctx context.Context, tx *sql.Tx, snapshotID int64, records []snapshotMetricRecord) error {
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM metric_observations WHERE snapshot_id=?`, snapshotID).Scan(&count); err != nil {
		return fmt.Errorf("performance: counting measurement metrics: %w", err)
	}
	if count != len(records) {
		return fmt.Errorf("%w: measurement snapshot %d metric count", ErrImmutableConflict, snapshotID)
	}
	for _, record := range records {
		equal, exists, err := immutableRowEqual(ctx, tx, `SELECT
			metric_key,status,value,gross_value,cost_adjusted_value,observation_id,observed_at,source,source_version
			FROM metric_observations WHERE snapshot_id=? AND metric_key=?`,
			[]any{snapshotID, record.key}, record.values())
		if err != nil {
			return fmt.Errorf("performance: reading %s metric: %w", record.key, err)
		}
		if !exists || !equal {
			return fmt.Errorf("%w: measurement snapshot %d metric %s", ErrImmutableConflict, snapshotID, record.key)
		}
	}
	return nil
}

func nullableTimestamp(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return timestamp(value)
}

type PruneResult struct {
	Deleted              int
	Skipped              bool
	Transactions         int
	MaxBatchDeleted      int
	MaxBatchLockDuration time.Duration
	BacklogRemaining     bool
}

func (s *Store) PruneDue(ctx context.Context, now time.Time) (PruneResult, error) {
	var out PruneResult
	cutoff := timestamp(now.UTC().Add(-RawRetention))
	for batch := 0; batch < MaxPruneBatchesPerRun; batch++ {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return PruneResult{}, fmt.Errorf("performance: starting prune: %w", err)
		}
		batchStarted := time.Now()
		if batch == 0 {
			var lastRaw sql.NullString
			if err := tx.QueryRowContext(ctx, `SELECT value FROM maintenance_state WHERE key='last_pruned_at'`).Scan(&lastRaw); err != nil && !errors.Is(err, sql.ErrNoRows) {
				_ = tx.Rollback()
				return PruneResult{}, fmt.Errorf("performance: reading prune cadence: %w", err)
			}
			if lastRaw.Valid {
				last, err := time.Parse(time.RFC3339Nano, lastRaw.String)
				if err != nil {
					_ = tx.Rollback()
					return PruneResult{}, fmt.Errorf("performance: invalid last prune instant: %w", err)
				}
				if now.UTC().Before(last.Add(PruneCadence)) {
					if err := tx.Commit(); err != nil {
						return PruneResult{}, fmt.Errorf("performance: committing cadence read: %w", err)
					}
					out.Skipped = true
					out.MaxBatchLockDuration = time.Since(batchStarted)
					return out, nil
				}
			}
		}
		result, err := tx.ExecContext(ctx, `DELETE FROM price_observations WHERE id IN (
			SELECT id FROM price_observations WHERE observed_at < ? ORDER BY observed_at, id LIMIT ?
		)`, cutoff, MaxPruneRows)
		if err != nil {
			_ = tx.Rollback()
			return PruneResult{}, fmt.Errorf("performance: pruning observations: %w", err)
		}
		deleted, err := result.RowsAffected()
		if err != nil {
			_ = tx.Rollback()
			return PruneResult{}, fmt.Errorf("performance: reading pruned row count: %w", err)
		}
		var backlog bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM price_observations WHERE observed_at < ? LIMIT 1)`, cutoff).Scan(&backlog); err != nil {
			_ = tx.Rollback()
			return PruneResult{}, fmt.Errorf("performance: checking prune backlog: %w", err)
		}
		if !backlog {
			if _, err := tx.ExecContext(ctx, `INSERT INTO maintenance_state(key,value) VALUES('last_pruned_at',?)
				ON CONFLICT(key) DO UPDATE SET value=excluded.value`, timestamp(now)); err != nil {
				_ = tx.Rollback()
				return PruneResult{}, fmt.Errorf("performance: recording prune cadence: %w", err)
			}
		}
		if err := tx.Commit(); err != nil {
			return PruneResult{}, fmt.Errorf("performance: committing prune: %w", err)
		}
		duration := time.Since(batchStarted)
		out.Transactions++
		out.Deleted += int(deleted)
		if int(deleted) > out.MaxBatchDeleted {
			out.MaxBatchDeleted = int(deleted)
		}
		if duration > out.MaxBatchLockDuration {
			out.MaxBatchLockDuration = duration
		}
		out.BacklogRemaining = backlog
		if !backlog {
			return out, nil
		}
	}
	return out, nil
}

func tradeArgs(trade Trade) []any {
	l := trade.Lineage
	return []any{
		trade.ID, nullable(l.CandidateLifeID), nullable(l.ThresholdVersion), nullable(l.ThresholdSetDigest), nullable(l.EvidenceDigest),
		nullable(l.LaneID), nullable(l.LaneVersion), nullable(l.DecisionID), nullable(l.RiskIntentID), nullable(l.AttemptID), nullable(l.MutationAttemptID), nullable(l.OrderID), nullable(l.FillID),
		l.PositionID, nullable(l.CloseID), nullable(l.PolicyID), nullable(l.PolicyVersion), l.Status(), strings.ToLower(strings.TrimSpace(trade.Market)), trade.Side,
		nullableTime(trade.DecisionAt), nullable(trade.DecisionPrice), timestamp(trade.EntryAt), trade.EntryPrice, trade.Quantity, nullable(trade.CostTotal),
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
	lane_id, lane_version, decision_id, risk_intent_id, attempt_id, mutation_attempt_id, order_id, fill_id, position_id,
	close_id, policy_id, policy_version, lineage_status, market, side, decision_at,
	decision_price, entry_at, entry_price, quantity, cost_total,
	realized_pnl_after_costs, realized_r, closed_at
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

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
	risk_intent_id TEXT,
	attempt_id TEXT,
	mutation_attempt_id TEXT,
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
	cost_total TEXT,
	realized_pnl_after_costs TEXT NOT NULL,
	realized_r TEXT NOT NULL,
	closed_at TEXT NOT NULL
) STRICT;
CREATE INDEX idx_performance_trades_window
	ON performance_trades(closed_at, lineage_status, market, lane_id, lane_version, policy_id, policy_version);
CREATE INDEX idx_performance_trades_lane
	ON performance_trades(lane_id, lane_version, policy_id, policy_version, closed_at);

CREATE TABLE performance_scope (
	singleton INTEGER PRIMARY KEY CHECK(singleton = 1),
	account_ref TEXT NOT NULL
) STRICT;

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
	lineage_status TEXT NOT NULL CHECK(lineage_status IN ('complete','link_missing')),
	UNIQUE(trade_id, calculated_at, semantics_version)
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
