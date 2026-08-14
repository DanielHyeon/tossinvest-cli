# Function Logic Map: `ExitObserver.judge`

- Source: `internal/app/engine/exitloop.go`
- Post-edit AST evidence: `ast.json` (8 branches; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| time/evidence state | injected clock or process-local monotonic lease, persisted evidence, and one marker status read | current source or explicit frozen-base revision + approved A111 contract | wall rollback, stale, stopped, invalid, or unavailable evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at `internal/app/engine/exitloop.go:859`; reject unusable lease evidence before identity, recovery and policy judgement | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111QuoteUseLeaseCannotBeExtendedByWallClockRollback` |
| B2 | AST `if` at `internal/app/engine/exitloop.go:862`; reject unusable lease evidence before identity, recovery and policy judgement | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111QuoteUseLeaseCannotBeExtendedByWallClockRollback` |
| B3 | AST `if` at `internal/app/engine/exitloop.go:866`; reject unusable lease evidence before identity, recovery and policy judgement | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111QuoteUseLeaseCannotBeExtendedByWallClockRollback` |
| B4 | AST `if` at `internal/app/engine/exitloop.go:876`; reject unusable lease evidence before identity, recovery and policy judgement | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111QuoteUseLeaseCannotBeExtendedByWallClockRollback` |
| B5 | AST `if` at `internal/app/engine/exitloop.go:882`; reject unusable lease evidence before identity, recovery and policy judgement | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111QuoteUseLeaseCannotBeExtendedByWallClockRollback` |
| B6 | AST `switch` at `internal/app/engine/exitloop.go:887`; reject unusable lease evidence before identity, recovery and policy judgement | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111QuoteUseLeaseCannotBeExtendedByWallClockRollback` |
| B7 | AST `case` at `internal/app/engine/exitloop.go:888`; reject unusable lease evidence before identity, recovery and policy judgement | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111QuoteUseLeaseCannotBeExtendedByWallClockRollback` |
| B8 | AST `case` at `internal/app/engine/exitloop.go:890`; reject unusable lease evidence before identity, recovery and policy judgement | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111QuoteUseLeaseCannotBeExtendedByWallClockRollback` |
| Return | all admitted paths | reject unusable lease evidence before identity, recovery and policy judgement | exact function result | `TestA111QuoteUseLeaseCannotBeExtendedByWallClockRollback` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| direct A111 time/evidence collaborators | reject unusable lease evidence before identity, recovery and policy judgement | no clock movement may extend a lease or upgrade a stopped marker | AST + named RED |

## State mutations and fallbacks

- reject unusable lease evidence before identity, recovery and policy judgement.
- Monotonic anchors are process-local and never persisted; persisted observation timestamps remain UTC wall evidence.
- Marker status is read once and may only be downgraded by a later response-time authority.
- Every AST branch is paired with a named test in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: A111 evidence lifetime and fail-closed response projection; no LIVE-order authority is introduced.
- High-risk impact: yes.
