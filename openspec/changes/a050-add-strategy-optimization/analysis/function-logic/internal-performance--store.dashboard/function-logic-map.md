# Function Logic Map: `Store.Dashboard`

- Source: `internal/performance/query.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| query | explicit as-of, bounded period, fixed market/lane/filter, positive minimum sample | `normalizeDashboardQuery` | invalid query fails before SQL |
| trades | bounded to `MaxDashboardTrades`, valid persisted decimals/timestamps | `queryTrades` | read/corruption/limit error returns no view |
| source freshness | maximum persisted trade close or metric observation time across returned rows | `queryTrade.NewestSourceAt` | calculation time is deliberately excluded; no wall-clock inference |
| state counts | same normalized predicate as aggregate query | dashboard SQL | count error returns no view |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | query normalization fails | none | error | period/as-of tests |
| B2 | bounded trade query fails | read only | error | SQL/corrupt/limit tests |
| B3 | aggregation fails | local view only | error | corrupt decimal tests |
| B4-B5 | iterate trades and retain maximum persisted source timestamp | local `NewestSourceAt` | continue | freshness propagation test |
| B6-B7 | iterate aggregates and count insufficient sample statuses | local counters | continue | state-count tests |
| B8 | lineage count query fails | read only | wrapped error | DB error coverage |
| B9-B11 | inspect each trade's required observations; first incomplete key increments not-measured once | local counter | continue | first-class state tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `normalizeDashboardQuery` | enforces period/filter defaults and cap | pure; error returned | query tests |
| `queryTrades`, `aggregateTrades` | bounded indexed read and deterministic aggregation | one attempt, no fallback scan | dashboard/query-plan tests |
| filtered lineage count SQL | counts states under identical predicate | caller context, one query | state tests |

## State mutations and fallbacks

- Mutates only a local view and counters. It never collects observations, polls quotes, writes the journal, or recommends/applies a setting.
- Freshness is derived exclusively from persisted source timestamps; missing or corrupt timestamps are not replaced by `time.Now`.

## Safety conclusion

- Safe edit boundary: bounded derived analytics query and persisted freshness projection.
- High-risk impact: yes for evidence correctness; no trading mutation authority.
