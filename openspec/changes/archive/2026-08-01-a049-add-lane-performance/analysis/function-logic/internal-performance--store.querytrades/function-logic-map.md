# Function Logic Map: `Store.queryTrades`

- Source: `internal/performance/query.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| query/start | validated dashboard predicates | `Dashboard` | SQL errors fail without fallback |
| row limit | 10,001 filtered trades maximum sentinel | implementation constant | caller rejects >10,000 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | SQL query fails | none | wrapped error | DB error contract |
| B2 | iterate joined rows | bounded local map/slice | continue | dashboard tests |
| B3 | row scan fails | none | wrapped error | DB error contract |
| B4 | first row for trade | allocates fixed maps and appends order | continue | dashboard tests |
| B5 | unique trade count exceeds 10,000 | bounded state only | `ErrDashboardRowLimit` | row-bound test |
| B6 | metric row exists | records value/status/provenance | continue | metric tests |
| B7 | metric key is markout | chooses cost-adjusted value | continue | markout aggregate test |
| B8 | source exists | stores source@version | continue | provenance test |
| B9 | row iterator fails | none | wrapped error | DB error contract |
| B10 | iterate deterministic trade IDs | appends bounded output | slice | dashboard tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQLite filtered CTE | uses closed_at/filter covering indexes and LIMIT | read-only | EXPLAIN test |
| `idx_measurement_snapshots_trade` | latest id per already-filtered trade | no global GROUP BY/full snapshot scan | EXPLAIN test |

## State mutations and fallbacks

- Memory is bounded by max trades times the fixed six metric rows; no raw price observations are read.

## Safety conclusion

- Safe edit boundary: indexed read model query.
- High-risk impact: read performance/correctness only.
