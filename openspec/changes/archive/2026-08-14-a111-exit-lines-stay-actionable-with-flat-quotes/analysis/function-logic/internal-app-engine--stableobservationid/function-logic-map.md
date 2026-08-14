# Function Logic Map: `stableObservationID`

- Source: `internal/app/engine/exitloop.go`
- Post-edit AST evidence: `ast.json` (4 branches; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| time/evidence state | injected clock or process-local monotonic lease, persisted evidence, and one marker status read | current source or explicit frozen-base revision + approved A111 contract | wall rollback, stale, stopped, invalid, or unavailable evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at `internal/app/engine/exitloop.go:1094`; bind official identity to fetched-at and fallback identity to durable cycle:N | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111FallbackObservationSourceIsDurableAndRestartMonotone` |
| B2 | AST `if` at `internal/app/engine/exitloop.go:1098`; bind official identity to fetched-at and fallback identity to durable cycle:N | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111FallbackObservationSourceIsDurableAndRestartMonotone` |
| B3 | AST `if` at `internal/app/engine/exitloop.go:1101`; bind official identity to fetched-at and fallback identity to durable cycle:N | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111FallbackObservationSourceIsDurableAndRestartMonotone` |
| B4 | AST `range` at `internal/app/engine/exitloop.go:1106`; bind official identity to fetched-at and fallback identity to durable cycle:N | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111FallbackObservationSourceIsDurableAndRestartMonotone` |
| Return | all admitted paths | bind official identity to fetched-at and fallback identity to durable cycle:N | exact function result | `TestA111FallbackObservationSourceIsDurableAndRestartMonotone` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| direct A111 time/evidence collaborators | bind official identity to fetched-at and fallback identity to durable cycle:N | no clock movement may extend a lease or upgrade a stopped marker | AST + named RED |

## State mutations and fallbacks

- bind official identity to fetched-at and fallback identity to durable cycle:N.
- Monotonic anchors are process-local and never persisted; persisted observation timestamps remain UTC wall evidence.
- Marker status is read once and may only be downgraded by a later response-time authority.
- Every AST branch is paired with a named test in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: A111 evidence lifetime and fail-closed response projection; no LIVE-order authority is introduced.
- High-risk impact: yes.
