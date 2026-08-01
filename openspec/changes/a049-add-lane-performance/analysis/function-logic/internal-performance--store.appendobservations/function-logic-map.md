# Function Logic Map: `Store.AppendObservations`

- Source: `internal/performance/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| observations | validated immutable IDs and canonical fields | caller-owned existing observations | invalid rows or divergent duplicate bytes fail whole transaction |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | transaction cannot start | none | wrapped error | append tests |
| B2 | ID absent | inserts row | continue | `TestObservationCompareAndAppendReplayAndDivergence` |
| B3 | ID exists with identical bytes | none | idempotent success | replay/restart/concurrent tests |
| B4 | ID exists with divergent bytes | none committed | `ErrImmutableConflict` | replay/concurrent tests |
| B5 | commit fails/process dies before commit | pending rows rollback | error/all-or-none | `TestPerformanceMigrationAndAppendSIGKILLPhasesAreAllOrNone` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `appendObservations` | validates and compare-and-appends every row | whole transaction fails on divergence | current HEAD + AST |
| `Commit` | publishes the batch atomically | error returned, no partial fallback | tests |

## State mutations and fallbacks

- Mutates only append-only derived observations. Identical replay is a no-op; bytes are never overwritten.

## Safety conclusion

- Safe edit boundary: caller-owned observation persistence.
- High-risk impact: data integrity only; no broker or polling dependency.
