# Function Logic Map: `applyStoredPositionForRequest`

- Source: `cmd/tossctl/httpapi_reader.go`
- Post-edit AST evidence: `ast.json` (2 branches; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| time/evidence state | injected clock or process-local monotonic lease, persisted evidence, and one marker status read | current source or explicit frozen-base revision + approved A111 contract | wall rollback, stale, stopped, invalid, or unavailable evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at `cmd/tossctl/httpapi_reader.go:284`; preserve stable machine stale reasons while consuming shared response authority | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111HTTPAPIUsesSharedFreshnessAtEveryLivenessAndBoundary` |
| B2 | AST `if` at `cmd/tossctl/httpapi_reader.go:287`; preserve stable machine stale reasons while consuming shared response authority | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111HTTPAPIUsesSharedFreshnessAtEveryLivenessAndBoundary` |
| Return | all admitted paths | preserve stable machine stale reasons while consuming shared response authority | exact function result | `TestA111HTTPAPIUsesSharedFreshnessAtEveryLivenessAndBoundary` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| direct A111 time/evidence collaborators | preserve stable machine stale reasons while consuming shared response authority | no clock movement may extend a lease or upgrade a stopped marker | AST + named RED |

## State mutations and fallbacks

- preserve stable machine stale reasons while consuming shared response authority.
- Monotonic anchors are process-local and never persisted; persisted observation timestamps remain UTC wall evidence.
- Marker status is read once and may only be downgraded by a later response-time authority.
- Every AST branch is paired with a named test in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: A111 evidence lifetime and fail-closed response projection; no LIVE-order authority is introduced.
- High-risk impact: operator-facing fail-closed projection.
