# Function Logic Map: `dashboardSQL`

- Source: `internal/performance/query.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| normalized query/start | bounded server query and UTC start | `Store.Dashboard` | caller rejects invalid query before this helper |
| SQL arguments | predicate args followed by 10,001 trade sentinel limit | fixed query builder | placeholders and args remain ordered |
| selected freshness fields | persisted trade `closed_at` and metric `observed_at` | performance schema | snapshot calculation time is not projected as source freshness |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless happy path composes fixed bounded CTE/join and projects observation time | local strings/slice only | SQL plus args | query-plan and freshness tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `dashboardTradePredicate` | shares time/market/lane/lineage filters | pure | filter tests |

## State mutations and fallbacks

- Appends only the local sentinel argument. The query is parameterized, bounded, read-only, and deterministically ordered.
- Existing latest-snapshot-by-ID behavior and query plan are preserved; the a050 change only projects `metric_observations.observed_at` for authoritative freshness.

## Safety conclusion

- Safe edit boundary: parameterized dashboard SQL projection.
- High-risk impact: yes for evidence freshness correctness; no mutation authority.
