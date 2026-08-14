# Function Logic Map: `compareObservationEvidence`

- Source: `internal/journal/exit_observation_refresh.go`
- Post-edit AST evidence: `ast.json` (12 branches; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| A111 state | persisted evidence, request/cycle state, and injected clock/marker | current source + approved A111 delta | invalid, stale, or incomplete evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST branch at `internal/journal/exit_observation_refresh.go:195`; older evidence is stale; equal timestamps use official-over-cycle precedence, cycle:N total ordering, exact duplicate no-op, otherwise conflict | branch-local only; no authority added | function-defined fail-closed result | `TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence` |
| B2 | AST branch at `internal/journal/exit_observation_refresh.go:199`; older evidence is stale; equal timestamps use official-over-cycle precedence, cycle:N total ordering, exact duplicate no-op, otherwise conflict | branch-local only; no authority added | function-defined fail-closed result | `TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence` |
| B3 | AST branch at `internal/journal/exit_observation_refresh.go:202`; older evidence is stale; equal timestamps use official-over-cycle precedence, cycle:N total ordering, exact duplicate no-op, otherwise conflict | branch-local only; no authority added | function-defined fail-closed result | `TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence` |
| B4 | AST branch at `internal/journal/exit_observation_refresh.go:205`; older evidence is stale; equal timestamps use official-over-cycle precedence, cycle:N total ordering, exact duplicate no-op, otherwise conflict | branch-local only; no authority added | function-defined fail-closed result | `TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence` |
| B5 | AST branch at `internal/journal/exit_observation_refresh.go:210`; older evidence is stale; equal timestamps use official-over-cycle precedence, cycle:N total ordering, exact duplicate no-op, otherwise conflict | branch-local only; no authority added | function-defined fail-closed result | `TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence` |
| B6 | AST branch at `internal/journal/exit_observation_refresh.go:213`; older evidence is stale; equal timestamps use official-over-cycle precedence, cycle:N total ordering, exact duplicate no-op, otherwise conflict | branch-local only; no authority added | function-defined fail-closed result | `TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence` |
| B7 | AST branch at `internal/journal/exit_observation_refresh.go:216`; older evidence is stale; equal timestamps use official-over-cycle precedence, cycle:N total ordering, exact duplicate no-op, otherwise conflict | branch-local only; no authority added | function-defined fail-closed result | `TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence` |
| B8 | AST branch at `internal/journal/exit_observation_refresh.go:219`; older evidence is stale; equal timestamps use official-over-cycle precedence, cycle:N total ordering, exact duplicate no-op, otherwise conflict | branch-local only; no authority added | function-defined fail-closed result | `TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence` |
| B9 | AST branch at `internal/journal/exit_observation_refresh.go:220`; older evidence is stale; equal timestamps use official-over-cycle precedence, cycle:N total ordering, exact duplicate no-op, otherwise conflict | branch-local only; no authority added | function-defined fail-closed result | `TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence` |
| B10 | AST branch at `internal/journal/exit_observation_refresh.go:221`; older evidence is stale; equal timestamps use official-over-cycle precedence, cycle:N total ordering, exact duplicate no-op, otherwise conflict | branch-local only; no authority added | function-defined fail-closed result | `TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence` |
| B11 | AST branch at `internal/journal/exit_observation_refresh.go:223`; older evidence is stale; equal timestamps use official-over-cycle precedence, cycle:N total ordering, exact duplicate no-op, otherwise conflict | branch-local only; no authority added | function-defined fail-closed result | `TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence` |
| B12 | AST branch at `internal/journal/exit_observation_refresh.go:227`; older evidence is stale; equal timestamps use official-over-cycle precedence, cycle:N total ordering, exact duplicate no-op, otherwise conflict | branch-local only; no authority added | function-defined fail-closed result | `TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence` |
| Return | all admitted paths | preserves the A111 safety boundary | propagates typed error or read-only result | `TestA111OfficialRefreshTemporalCASIsNoWriteForOlderOrAmbiguousEvidence` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| A111 direct collaborators | older evidence is stale; equal timestamps use official-over-cycle precedence, cycle:N total ordering, exact duplicate no-op, otherwise conflict | typed conflict/stale/invalid results do not authorize an order or fresh line | current AST + A111 RED |

## State mutations and fallbacks

- older evidence is stale; equal timestamps use official-over-cycle precedence, cycle:N total ordering, exact duplicate no-op, otherwise conflict
- The function remains inside its existing journal, engine, or projection authority; it does not create a new LIVE-order path.
- Every AST branch is paired with the named A111 RED in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: the A111 heartbeat/evidence freshness seam only; retain full judgement and existing order ordering where applicable.
- High-risk impact: yes.
