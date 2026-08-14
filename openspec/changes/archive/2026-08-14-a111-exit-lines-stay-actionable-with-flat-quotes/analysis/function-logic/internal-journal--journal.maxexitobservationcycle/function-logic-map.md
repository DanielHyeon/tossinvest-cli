# Function Logic Map: `Journal.MaxExitObservationCycle`

- Source: `internal/journal/exit_observation_refresh.go`
- Post-edit AST evidence: `ast.json` (5 branches; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| A111 evidence/state | persisted evidence, request/cycle state, and injected clock/marker | current source or explicitly frozen base revision + approved A111 delta | invalid, stale, unavailable, or incomplete evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at `internal/journal/exit_observation_refresh.go:243`; recover only the maximum valid cycle:N from the current account managed noncompleted working set, ignoring every out-of-scope or corrupt shape | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111MaxExitObservationCycleIgnoresEveryOutOfScopeEvidenceShape` |
| B2 | AST `for` at `internal/journal/exit_observation_refresh.go:248`; recover only the maximum valid cycle:N from the current account managed noncompleted working set, ignoring every out-of-scope or corrupt shape | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111MaxExitObservationCycleIgnoresEveryOutOfScopeEvidenceShape` |
| B3 | AST `if` at `internal/journal/exit_observation_refresh.go:250`; recover only the maximum valid cycle:N from the current account managed noncompleted working set, ignoring every out-of-scope or corrupt shape | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111MaxExitObservationCycleIgnoresEveryOutOfScopeEvidenceShape` |
| B4 | AST `if` at `internal/journal/exit_observation_refresh.go:254`; recover only the maximum valid cycle:N from the current account managed noncompleted working set, ignoring every out-of-scope or corrupt shape | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111MaxExitObservationCycleIgnoresEveryOutOfScopeEvidenceShape` |
| B5 | AST `if` at `internal/journal/exit_observation_refresh.go:257`; recover only the maximum valid cycle:N from the current account managed noncompleted working set, ignoring every out-of-scope or corrupt shape | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111MaxExitObservationCycleIgnoresEveryOutOfScopeEvidenceShape` |
| Return | all admitted paths | recover only the maximum valid cycle:N from the current account managed noncompleted working set, ignoring every out-of-scope or corrupt shape | exact function result | `TestA111MaxExitObservationCycleIgnoresEveryOutOfScopeEvidenceShape` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| direct A111 collaborators | recover only the maximum valid cycle:N from the current account managed noncompleted working set, ignoring every out-of-scope or corrupt shape | failures never authorize an order or fresh operator line | AST + named A111 RED |

## State mutations and fallbacks

- recover only the maximum valid cycle:N from the current account managed noncompleted working set, ignoring every out-of-scope or corrupt shape.
- Local journal or broker failures remain visible; cached broker data never lends freshness to local evidence.
- Every AST branch is paired with the named A111 RED in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: A111 observation heartbeat, quote-evidence lifetime, or fail-closed operator projection only.
- High-risk impact: yes.
