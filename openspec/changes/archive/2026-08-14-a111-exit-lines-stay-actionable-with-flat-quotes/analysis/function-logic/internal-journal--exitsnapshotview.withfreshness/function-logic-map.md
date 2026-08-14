# Function Logic Map: `ExitSnapshotView.WithFreshness`

- Source: `internal/journal/exit_snapshot.go`
- Frozen span: lines 202-217 at HEAD `3355df0fe9c82c3bb8c522e2d79abf107dd5f2c3`
- AST evidence: `ast.json` (5 branches, 4 calls, 3 returns)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| snapshot | nil or validated persisted `StoredExitSnapshot` | journal reader | nil passes through unchanged |
| `ObservedAt` | RFC3339Nano | persisted effective JSON/flattened tuple | invalid becomes stale |
| `asOf`, `maxAge` | nonzero time and positive duration | transport boundary | disabled check passes view unchanged |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | no snapshot, nonpositive max age, or zero as-of | none | unchanged view | nil/disabled tests |
| B2 | observed-at parse fails | set stale + `invalid_observed_at` | view | integrity test |
| B3 | age strictly greater than limit | set stale + `observation_older_than_limit` | view | 30s boundary test |
| B4 | observed time after as-of | set stale + `observation_in_future` | view | future test |
| Return | inside inclusive limit | none | fresh view | exact-boundary test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Parse` | validate persisted observation clock | parse failure is fail-closed | AST |
| `time.Time.Sub/After` | enforce age and future limits | pure | AST |

## State mutations and fallbacks

- Value receiver returns a decorated view without altering persisted evidence.
- CodeGraph callers prove this is the shared age authority for HTTP and an unwired console.
- A111 changes the writer timestamp semantics, not the 30-second comparison or failure vocabulary.

## Safety conclusion

- Safe edit boundary: preserve this function unchanged; make `ObservedAt` truthfully advance on eligible flat observations.
- High-risk impact: yes as shared actionable-line gate, though no mutation occurs here.
