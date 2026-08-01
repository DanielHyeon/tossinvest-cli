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
  lineage_status TEXT NOT NULL,
  market TEXT NOT NULL,
  side TEXT NOT NULL,
  decision_at TEXT,
  decision_price TEXT,
  entry_at TEXT NOT NULL,
  entry_price TEXT NOT NULL,
  quantity TEXT NOT NULL,
  cost_total TEXT NOT NULL,
  realized_pnl_after_costs TEXT NOT NULL,
  realized_r TEXT NOT NULL,
  closed_at TEXT NOT NULL
);
CREATE INDEX idx_performance_trades_window ON performance_trades(closed_at, lineage_status, market, lane_id, lane_version, policy_id, policy_version);
CREATE INDEX idx_performance_trades_lane ON performance_trades(lane_id, lane_version, policy_id, policy_version, closed_at);

CREATE TABLE price_observations (
  id TEXT PRIMARY KEY,
  position_id TEXT NOT NULL,
  observed_at TEXT NOT NULL,
  price TEXT NOT NULL,
  source TEXT NOT NULL,
  source_version TEXT NOT NULL
);
CREATE INDEX idx_price_observations_position_at ON price_observations(position_id, observed_at, id);
CREATE INDEX idx_price_observations_at ON price_observations(observed_at, id);

CREATE TABLE measurement_snapshots (
  id INTEGER PRIMARY KEY,
  trade_id TEXT NOT NULL REFERENCES performance_trades(id),
  calculated_at TEXT NOT NULL,
  semantics_version TEXT NOT NULL,
  lineage_status TEXT NOT NULL
);
CREATE INDEX idx_measurement_snapshots_trade ON measurement_snapshots(trade_id, id);

CREATE TABLE metric_observations (
  snapshot_id INTEGER NOT NULL REFERENCES measurement_snapshots(id),
  metric_key TEXT NOT NULL,
  status TEXT NOT NULL,
  value TEXT,
  gross_value TEXT,
  cost_adjusted_value TEXT,
  observation_id TEXT,
  observed_at TEXT,
  source TEXT,
  source_version TEXT,
  PRIMARY KEY(snapshot_id, metric_key)
);

CREATE TABLE maintenance_state (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

