# Function Logic Map: `Store.queryTrades`

- Source: `internal/performance/query.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| query/start | normalized dashboard query and inclusive start | `Store.Dashboard` | SQL errors returned; no fallback scan |
| joined rows | at most 10,001 distinct trade sentinel; valid RFC3339Nano source timestamps | persisted performance schema | row/parse/limit error returns no result |
| metric rows | optional key; markout uses cost-adjusted value; optional source/version | measurement schema | missing metric stays missing, never zero-filled |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | SQL query fails | none | wrapped error | DB error coverage |
| B2 | iterate joined rows | bounded local map/slice | continue | dashboard tests |
| B3 | row scan fails | none | wrapped error | schema/scan error coverage |
| B4 | persisted source timestamp invalid | none | trade-scoped wrapped error | corrupt timestamp test |
| B5 | first row for trade | allocate metric maps and append stable order | continue | aggregation tests |
| B6 | distinct trade count exceeds max | bounded local state | `ErrDashboardRowLimit` | row-limit test |
| B7 | row has newer source timestamp than current trade | update local maximum | continue | freshness maximum test |
| B8 | metric key present | store metric/status | continue | metric tests |
| B9 | metric is markout | use cost-adjusted value | continue | markout tests |
| B10 | source present | store `source@version` provenance | continue | provenance test |
| B11 | row iterator fails | none | wrapped error | DB error coverage |
| B12 | iterate stable trade order | append bounded values | result | deterministic order test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `dashboardSQL`/SQLite | bounded filtered trades and latest snapshot as-of query | caller context; one indexed query | EXPLAIN and dashboard tests |
| `newestDashboardSourceTime` | validates and chooses persisted freshness | corruption fails closed | freshness/corruption tests |

## State mutations and fallbacks

- All mutation is bounded local assembly. No raw price observations, current clock, or writer path is used.
- Missing metrics/provenance remain absent; corrupt timestamps never fall back to row order or wall clock.

## Safety conclusion

- Safe edit boundary: indexed read model assembly with persisted freshness.
- High-risk impact: yes for optimization evidence correctness; read-only.
