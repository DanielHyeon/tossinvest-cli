package performance

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"
)

const (
	AllMarkets             = "all markets"
	AllLanes               = "all lanes"
	DefaultMinimumSample   = 20
	MaxDashboardPeriodDays = 90
	MaxDashboardTrades     = 10_000
)

var (
	ErrDashboardPeriod   = fmt.Errorf("performance: dashboard period exceeds %d days", MaxDashboardPeriodDays)
	ErrDashboardRowLimit = fmt.Errorf("performance: dashboard exceeds %d trades", MaxDashboardTrades)
)

type Query struct {
	AsOf          time.Time
	PeriodDays    int
	Market        string
	Lane          string
	CompleteOnly  bool
	MinimumSample int
}

func DefaultQuery(now time.Time) Query {
	return Query{AsOf: now.UTC(), PeriodDays: 30, Market: AllMarkets, Lane: AllLanes, CompleteOnly: true, MinimumSample: DefaultMinimumSample}
}

type StateCounts struct {
	Complete           int
	LinkMissing        int
	NotMeasured        int
	InsufficientSample int
}

type Aggregate struct {
	Market                string
	LaneID                string
	LaneVersion           string
	PolicyID              string
	PolicyVersion         string
	Samples               int
	Status                Status
	NetPnL                string
	AverageR              string
	WinRate               string
	ProfitFactor          string
	MaxDrawdown           string
	SlippagePct           string
	MFEPct                string
	MAEPct                string
	Markout5Pct           string
	Markout15Pct          string
	Markout30Pct          string
	SemanticsVersion      string
	ObservationProvenance string
	Metrics               []MetricSummary
}

type MetricSummary struct {
	Key        string
	Label      string
	Help       string
	Unit       string
	Value      string
	Samples    int
	Status     Status
	Provenance string
}

type DashboardView struct {
	Query          Query
	NewestSourceAt time.Time
	Aggregates     []Aggregate
	States         StateCounts
	// Attributions is populated by account-bound transport adapters from the
	// separately persisted attribution generation. Dashboard aggregation itself
	// remains account-agnostic and leaves this slice empty.
	Attributions []Attribution
}

func (d DashboardView) Explanation(status Status) string {
	switch status {
	case StatusLinkMissing:
		return "link_missing · 식별자 연결이 없어 종목/시각으로 추정하지 않았습니다."
	case StatusNotMeasured:
		return "not_measured · 기존 관측이 없어 0으로 계산하지 않았습니다."
	case StatusInsufficientSample:
		return "insufficient_sample · 관측값은 보이지만 추천 근거로 사용할 수 없습니다."
	default:
		return string(status)
	}
}

type queryTrade struct {
	ID, Market, LaneID, LaneVersion, PolicyID, PolicyVersion string
	PnL, RealizedR, ClosedAt                                 string
	SnapshotID                                               sql.NullInt64
	Semantics                                                string
	Metrics                                                  map[string]string
	Statuses                                                 map[string]Status
	Sources                                                  map[string]string
	NewestSourceAt                                           time.Time
}

func (s *Store) Dashboard(ctx context.Context, query Query) (DashboardView, error) {
	query, err := normalizeDashboardQuery(query)
	if err != nil {
		return DashboardView{}, err
	}
	view := DashboardView{Query: query}
	start := query.AsOf.UTC().AddDate(0, 0, -query.PeriodDays)
	trades, err := s.queryTrades(ctx, query, start)
	if err != nil {
		return DashboardView{}, err
	}
	view.Aggregates, err = aggregateTrades(trades, query.MinimumSample)
	if err != nil {
		return DashboardView{}, err
	}
	for _, trade := range trades {
		if trade.NewestSourceAt.After(view.NewestSourceAt) {
			view.NewestSourceAt = trade.NewestSourceAt
		}
	}
	for _, aggregate := range view.Aggregates {
		if aggregate.Status == StatusInsufficientSample {
			view.States.InsufficientSample++
		}
	}
	predicate, args := dashboardTradePredicate("t", query, start)
	if err := s.db.QueryRowContext(ctx, `SELECT
		coalesce(sum(t.lineage_status='complete'),0), coalesce(sum(t.lineage_status='link_missing'),0)
		FROM performance_trades t WHERE `+predicate, args...).
		Scan(&view.States.Complete, &view.States.LinkMissing); err != nil {
		return DashboardView{}, fmt.Errorf("performance: counting lineage states: %w", err)
	}
	for _, trade := range trades {
		for _, key := range []string{"markout_5", "markout_15", "markout_30", "slippage", "mfe", "mae"} {
			if trade.Statuses[key] != StatusComplete {
				view.States.NotMeasured++
				break
			}
		}
	}
	return view, nil
}

func (s *Store) queryTrades(ctx context.Context, query Query, start time.Time) ([]queryTrade, error) {
	statement, args := dashboardSQL(query, start)
	rows, err := s.db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, fmt.Errorf("performance: querying dashboard: %w", err)
	}
	defer rows.Close()
	byID := make(map[string]*queryTrade)
	var order []string
	for rows.Next() {
		var trade queryTrade
		var metricKey, metricStatus, value, costAdjusted, source, sourceVersion, metricObservedAt sql.NullString
		if err := rows.Scan(&trade.ID, &trade.Market, &trade.LaneID, &trade.LaneVersion,
			&trade.PolicyID, &trade.PolicyVersion, &trade.PnL, &trade.RealizedR, &trade.ClosedAt,
			&trade.SnapshotID, &trade.Semantics, &metricKey, &metricStatus, &value, &costAdjusted,
			&metricObservedAt, &source, &sourceVersion); err != nil {
			return nil, fmt.Errorf("performance: reading dashboard row: %w", err)
		}
		newest, err := newestDashboardSourceTime(trade.ClosedAt, metricObservedAt.String)
		if err != nil {
			return nil, fmt.Errorf("performance: trade %s source timestamp: %w", trade.ID, err)
		}
		trade.NewestSourceAt = newest
		current := byID[trade.ID]
		if current == nil {
			trade.Metrics = make(map[string]string)
			trade.Statuses = make(map[string]Status)
			trade.Sources = make(map[string]string)
			current = &trade
			byID[trade.ID] = current
			order = append(order, trade.ID)
			if len(order) > MaxDashboardTrades {
				return nil, ErrDashboardRowLimit
			}
		}
		if trade.NewestSourceAt.After(current.NewestSourceAt) {
			current.NewestSourceAt = trade.NewestSourceAt
		}
		if metricKey.Valid {
			metricValue := value.String
			if strings.HasPrefix(metricKey.String, "markout_") {
				metricValue = costAdjusted.String
			}
			current.Metrics[metricKey.String] = metricValue
			current.Statuses[metricKey.String] = Status(metricStatus.String)
			if source.String != "" {
				current.Sources[metricKey.String] = source.String + "@" + sourceVersion.String
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("performance: iterating dashboard: %w", err)
	}
	out := make([]queryTrade, 0, len(order))
	for _, id := range order {
		out = append(out, *byID[id])
	}
	return out, nil
}

func newestDashboardSourceTime(values ...string) (time.Time, error) {
	var newest time.Time
	for _, raw := range values {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid persisted timestamp %q: %w", raw, err)
		}
		if parsed.After(newest) {
			newest = parsed.UTC()
		}
	}
	return newest, nil
}

func normalizeDashboardQuery(query Query) (Query, error) {
	if query.AsOf.IsZero() {
		return Query{}, fmt.Errorf("performance: query as-of is required")
	}
	if query.PeriodDays <= 0 {
		query.PeriodDays = 30
	}
	if query.PeriodDays > MaxDashboardPeriodDays {
		return Query{}, ErrDashboardPeriod
	}
	if query.MinimumSample <= 0 {
		query.MinimumSample = DefaultMinimumSample
	}
	query.Market = strings.TrimSpace(query.Market)
	if query.Market == "" {
		query.Market = AllMarkets
	}
	query.Lane = strings.TrimSpace(query.Lane)
	if query.Lane == "" {
		query.Lane = AllLanes
	}
	return query, nil
}

func dashboardTradePredicate(alias string, query Query, start time.Time) (string, []any) {
	prefix := alias + "."
	predicates := []string{prefix + "closed_at >= ?", prefix + "closed_at <= ?"}
	args := []any{timestamp(start), timestamp(query.AsOf)}
	if query.CompleteOnly {
		predicates = append(predicates, prefix+"lineage_status='complete'")
	}
	if query.Market != AllMarkets {
		predicates = append(predicates, prefix+"market=?")
		args = append(args, strings.ToLower(query.Market))
	}
	if query.Lane != AllLanes {
		predicates = append(predicates, prefix+"lane_id=?")
		args = append(args, query.Lane)
	}
	return strings.Join(predicates, " AND "), args
}

func dashboardSQL(query Query, start time.Time) (string, []any) {
	predicate, args := dashboardTradePredicate("t", query, start)
	args = append(args, MaxDashboardTrades+1)
	return `WITH filtered AS (
		SELECT t.id, t.market, t.lane_id, t.lane_version, t.policy_id, t.policy_version,
		       t.realized_pnl_after_costs, t.realized_r, t.closed_at
		  FROM performance_trades t
		 WHERE ` + predicate + `
		 ORDER BY t.closed_at, t.id
		 LIMIT ?
	), latest AS (
		SELECT f.id AS trade_id,
		       (SELECT ms.id FROM measurement_snapshots ms
		         WHERE ms.trade_id=f.id ORDER BY ms.id DESC LIMIT 1) AS snapshot_id
		  FROM filtered f
	)
	SELECT t.id, t.market, coalesce(t.lane_id,''), coalesce(t.lane_version,''),
	       coalesce(t.policy_id,''), coalesce(t.policy_version,''),
	       t.realized_pnl_after_costs, t.realized_r, t.closed_at,
	       s.id, coalesce(s.semantics_version,''), m.metric_key, m.status,
	       coalesce(m.value,''), coalesce(m.cost_adjusted_value,''),
	       m.observed_at, coalesce(m.source,''), coalesce(m.source_version,'')
	  FROM filtered t
	  LEFT JOIN latest l ON l.trade_id=t.id
	  LEFT JOIN measurement_snapshots s ON s.id=l.snapshot_id
	  LEFT JOIN metric_observations m ON m.snapshot_id=s.id
	 ORDER BY t.closed_at, t.id, m.metric_key`, args
}

func (s *Store) DashboardQueryPlan(ctx context.Context, query Query) (string, error) {
	query, err := normalizeDashboardQuery(query)
	if err != nil {
		return "", err
	}
	statement, args := dashboardSQL(query, query.AsOf.UTC().AddDate(0, 0, -query.PeriodDays))
	rows, err := s.db.QueryContext(ctx, "EXPLAIN QUERY PLAN "+statement, args...)
	if err != nil {
		return "", fmt.Errorf("performance: explaining dashboard: %w", err)
	}
	defer rows.Close()
	var lines []string
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
			return "", fmt.Errorf("performance: reading dashboard plan: %w", err)
		}
		lines = append(lines, detail)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("performance: iterating dashboard plan: %w", err)
	}
	return strings.Join(lines, "\n"), nil
}

func aggregateTrades(trades []queryTrade, minimum int) ([]Aggregate, error) {
	type bucket struct{ rows []queryTrade }
	buckets := make(map[string]*bucket)
	var keys []string
	for _, trade := range trades {
		key := strings.Join([]string{trade.Market, trade.LaneID, trade.LaneVersion, trade.PolicyID, trade.PolicyVersion}, "\x00")
		if buckets[key] == nil {
			buckets[key] = &bucket{}
			keys = append(keys, key)
		}
		buckets[key].rows = append(buckets[key].rows, trade)
	}
	sort.Strings(keys)
	out := make([]Aggregate, 0, len(keys))
	for _, key := range keys {
		rows := buckets[key].rows
		first := rows[0]
		agg := Aggregate{Market: first.Market, LaneID: first.LaneID, LaneVersion: first.LaneVersion,
			PolicyID: first.PolicyID, PolicyVersion: first.PolicyVersion, Samples: len(rows), Status: StatusComplete,
			SemanticsVersion: SemanticsVersion}
		if len(rows) < minimum {
			agg.Status = StatusInsufficientSample
		}
		profit, loss, net, sumR := new(big.Rat), new(big.Rat), new(big.Rat), new(big.Rat)
		peak, drawdown := new(big.Rat), new(big.Rat)
		wins := 0
		metricSums := make(map[string]*big.Rat)
		metricCounts := make(map[string]int)
		sources := make(map[string]bool)
		metricSources := make(map[string]map[string]bool)
		for _, row := range rows {
			pnl, err := rat(row.PnL)
			if err != nil {
				return nil, fmt.Errorf("performance: trade %s realized PnL: %w", row.ID, err)
			}
			if pnl.Sign() > 0 {
				wins++
				profit.Add(profit, pnl)
			} else if pnl.Sign() < 0 {
				loss.Sub(loss, pnl)
			}
			net.Add(net, pnl)
			if net.Cmp(peak) > 0 {
				peak.Set(net)
			}
			fall := new(big.Rat).Sub(peak, net)
			if fall.Cmp(drawdown) > 0 {
				drawdown.Set(fall)
			}
			realizedR, err := rat(row.RealizedR)
			if err != nil {
				return nil, fmt.Errorf("performance: trade %s realized R: %w", row.ID, err)
			}
			sumR.Add(sumR, realizedR)
			for _, metric := range []string{"slippage", "mfe", "mae", "markout_5", "markout_15", "markout_30"} {
				if row.Statuses[metric] == StatusComplete && row.Metrics[metric] != "" {
					value, err := rat(row.Metrics[metric])
					if err != nil {
						return nil, fmt.Errorf("performance: trade %s metric %s: %w", row.ID, metric, err)
					}
					if metricSums[metric] == nil {
						metricSums[metric] = new(big.Rat)
					}
					metricSums[metric].Add(metricSums[metric], value)
					metricCounts[metric]++
				} else if row.Statuses[metric] == StatusComplete {
					return nil, fmt.Errorf("performance: trade %s metric %s: invalid persisted decimal %q", row.ID, metric, row.Metrics[metric])
				}
				if row.Sources[metric] != "" {
					sources[row.Sources[metric]] = true
					if metricSources[metric] == nil {
						metricSources[metric] = make(map[string]bool)
					}
					metricSources[metric][row.Sources[metric]] = true
				}
			}
		}
		agg.NetPnL, agg.AverageR = ratText(net), ratText(new(big.Rat).Quo(sumR, big.NewRat(int64(len(rows)), 1)))
		agg.WinRate = ratText(new(big.Rat).SetFrac64(int64(wins), int64(len(rows))))
		if loss.Sign() > 0 {
			agg.ProfitFactor = ratText(new(big.Rat).Quo(profit, loss))
		}
		agg.MaxDrawdown = ratText(drawdown)
		average := func(key string) string {
			if metricCounts[key] == 0 {
				return ""
			}
			return ratText(new(big.Rat).Quo(metricSums[key], big.NewRat(int64(metricCounts[key]), 1)))
		}
		agg.SlippagePct, agg.MFEPct, agg.MAEPct = average("slippage"), average("mfe"), average("mae")
		agg.Markout5Pct, agg.Markout15Pct, agg.Markout30Pct = average("markout_5"), average("markout_15"), average("markout_30")
		var sourceList []string
		for source := range sources {
			sourceList = append(sourceList, source)
		}
		sort.Strings(sourceList)
		agg.ObservationProvenance = strings.Join(sourceList, ", ")
		provenance := func(key string) string {
			var values []string
			for source := range metricSources[key] {
				values = append(values, source)
			}
			sort.Strings(values)
			if len(values) == 0 {
				switch key {
				case "net_pnl", "average_r", "win_rate", "profit_factor", "max_drawdown":
					return "journal-outcome@" + SemanticsVersion
				default:
					return string(StatusNotMeasured)
				}
			}
			return strings.Join(values, ", ")
		}
		metric := func(key, label, help, unit, value string, samples int) MetricSummary {
			status := StatusComplete
			if samples == 0 || value == "" {
				status = StatusNotMeasured
			}
			return MetricSummary{
				Key: key, Label: label, Help: help, Unit: unit, Value: value,
				Samples: samples, Status: status, Provenance: provenance(key),
			}
		}
		agg.Metrics = []MetricSummary{
			metric("net_pnl", "비용 후 손익", "실현 매수·매도 비용을 모두 차감한 합계입니다.", "계좌 통화", agg.NetPnL, len(rows)),
			metric("average_r", "평균 실현 R", "비용 후 손익을 최초 위험금액으로 나눈 평균입니다.", "R", agg.AverageR, len(rows)),
			metric("win_rate", "승률", "비용 후 손익이 양수인 거래 비율입니다.", "비율", agg.WinRate, len(rows)),
			metric("profit_factor", "Profit factor", "총이익을 총손실 규모로 나눈 값이며 손실 표본이 없으면 미측정입니다.", "배", agg.ProfitFactor, len(rows)),
			metric("max_drawdown", "최대 낙폭", "종료 순서 누적 비용 후 손익의 peak-to-trough입니다.", "계좌 통화", agg.MaxDrawdown, len(rows)),
			metric("slippage", "평균 slippage", "결정 가격 대비 실제 진입 가격의 불리한 이동입니다.", "%", agg.SlippagePct, metricCounts["slippage"]),
			metric("mfe", "평균 MFE", "보유 중 기존 관측의 최대 유리 변동입니다.", "%", agg.MFEPct, metricCounts["mfe"]),
			metric("mae", "평균 MAE", "보유 중 기존 관측의 최대 불리 변동입니다.", "%", agg.MAEPct, metricCounts["mae"]),
			metric("markout_5", "5분 비용 후 markout", "진입 5분 뒤 +60초 이내 첫 기존 관측의 비용 차감 수익률입니다.", "%", agg.Markout5Pct, metricCounts["markout_5"]),
			metric("markout_15", "15분 비용 후 markout", "진입 15분 뒤 +60초 이내 첫 기존 관측의 비용 차감 수익률입니다.", "%", agg.Markout15Pct, metricCounts["markout_15"]),
			metric("markout_30", "30분 비용 후 markout", "진입 30분 뒤 +60초 이내 첫 기존 관측의 비용 차감 수익률입니다.", "%", agg.Markout30Pct, metricCounts["markout_30"]),
		}
		out = append(out, agg)
	}
	return out, nil
}
