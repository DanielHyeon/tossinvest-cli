# Function Logic Map: `Provider.ReadEvidence`

- Source: `internal/optimizationevidence/provider.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| provider dependencies | non-nil provider, dashboard reader, and UTC clock | console composition root | unavailable evidence plus typed error; never authorizes apply |
| dashboard query | server-owned 30-day/all-market/all-lane defaults with `CompleteOnly=false` | `performance.DefaultQuery` plus a050 adapter policy | callers cannot alter filters through form/query input |
| evidence digest | SHA-256 of canonical query, states, sorted aggregates, and sorted metrics | `digestDashboard` | canonicalization error returns unavailable evidence |
| completeness | every missing lineage/sample/state/metric reason is explicit | `missingEvidence` | any reason yields insufficient, not complete |
| source freshness | exact dashboard echo; newest persisted trade close/metric observation time is present, not future, and at most 72h old | `DashboardView.Query` and `NewestSourceAt` | mismatch is unavailable; missing/future/stale source is stale, never complete |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | provider, reader, or clock is nil | none | unavailable with `performance-dashboard-error` | `TestProviderRejectsNilDependenciesAndZeroClock` |
| B2 | clock returns zero time | none | unavailable with observation-time error | `TestProviderRejectsNilDependenciesAndZeroClock` |
| B3 | fixed dashboard read fails | read-only dashboard call only | unavailable with wrapped error | dashboard error test |
| B4 | dashboard does not echo exact fixed query | none | unavailable with query-mismatch reason | exact-query echo test |
| B5 | canonical digest generation fails | none | unavailable | defensive branch review plus deterministic digest test |
| B6-B8 | classify persisted source time as missing/future, stale, or current | local status only | stale for missing/future/older than 72h | persisted freshness tests |
| B9 | temporal defect exists | append unique reason | stale | stale/future/missing tests |
| B10-B11 | no temporal defect but evidence defects exist | local status only | insufficient; otherwise complete | completeness matrices |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `performance.DefaultQuery` | fixes the recommendation evidence window and dimensions | pure; no retry | adapter query contract test |
| `DashboardReader.Dashboard` | obtains derived a049 evidence through one read-only method | caller context controls cancellation; one attempt, error fails closed | dashboard error and query-capture tests |
| `digestDashboard` | creates deterministic evidence identity | one attempt; error becomes unavailable | digest stability tests |
| `missingEvidence` | derives explicit incompleteness reasons | pure, deterministic, sorted | evidence state matrix tests |

## State mutations and fallbacks

- Local mutation is limited to the server-created query and evidence result assembly. `ObservedAt` is the persisted source timestamp, never request/as-of or wall clock.
- There is no journal collector, migration, broker, order, lane, gate, or LIVE authority and no fallback that converts unavailable evidence into complete evidence.

## Safety conclusion

- Safe edit boundary: translate a049 read-only dashboard state into a050 evidence without widening authority.
- High-risk impact: yes; evidence gates strategy recommendation, so every dependency/error/incomplete state fails closed.
