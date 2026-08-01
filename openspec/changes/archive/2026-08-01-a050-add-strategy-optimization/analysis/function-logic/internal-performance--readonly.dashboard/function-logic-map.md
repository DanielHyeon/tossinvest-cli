# Function Logic Map: `ReadOnly.Dashboard`

- Source: `internal/performance/readonly.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| receiver lifecycle | non-nil, not closed | `ReadOnly` mutex-protected state | explicit not-open/closed error |
| database snapshot | fresh immutable read-only clean checkpoint | `openImmutablePerformanceDB` | typed missing/WAL/change/open error |
| query | server-owned performance query | caller capability | dashboard error propagated; no partial view |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil receiver | none | not-open error | nil lifecycle test |
| B2 | receiver closed | read lock only | closed error | post-close test |
| B3 | fresh immutable DB open fails | none | propagate error | WAL/missing-after-open test |
| B4 | dashboard query fails | closes ephemeral handle | query error | invalid/cancelled query test |
| B5 | ephemeral handle close fails | close attempt | close error | lifecycle fault review |
| B6 | DB identity/sidecars changed during query | none | typed changed/WAL error | concurrent writer/change test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openImmutablePerformanceDB` | obtains one fresh query-only snapshot | one attempt/no retry | immutable opener tests |
| `Store.Dashboard` | reuses derived read query implementation only | caller context cancellation; error propagated | dashboard result test |
| `db.Close`, `unchangedImmutablePerformanceDB` | closes snapshot and rejects drift | both must succeed | mutation/sidecar assertions |

## State mutations and fallbacks

- Holds an RW read lock across one ephemeral query. It never migrates, collects, prunes, or starts a writer transaction.
- Any query, close, WAL, or identity uncertainty returns no dashboard and has no stale fallback.

## Safety conclusion

- Safe edit boundary: one derived immutable dashboard read per call.
- High-risk impact: yes; recommendation evidence must not silently read a stale checkpoint.
