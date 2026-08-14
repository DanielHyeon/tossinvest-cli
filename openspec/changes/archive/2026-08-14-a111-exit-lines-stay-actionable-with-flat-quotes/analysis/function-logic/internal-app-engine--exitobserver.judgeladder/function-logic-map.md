# Function Logic Map: `ExitObserver.judgeLadder`

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
| B1 | AST `if` at `internal/app/engine/exitloop.go:971`; retain ladder identity/rung guards and recheck the monotonic lease before persistence | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B2 | AST `if` at `internal/app/engine/exitloop.go:975`; retain ladder identity/rung guards and recheck the monotonic lease before persistence | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B3 | AST `if` at `internal/app/engine/exitloop.go:980`; retain ladder identity/rung guards and recheck the monotonic lease before persistence | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B4 | AST `if` at `internal/app/engine/exitloop.go:985`; retain ladder identity/rung guards and recheck the monotonic lease before persistence | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B5 | AST `if` at `internal/app/engine/exitloop.go:986`; retain ladder identity/rung guards and recheck the monotonic lease before persistence | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B6 | AST `if` at `internal/app/engine/exitloop.go:1014`; retain ladder identity/rung guards and recheck the monotonic lease before persistence | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B7 | AST `if` at `internal/app/engine/exitloop.go:1027`; retain ladder identity/rung guards and recheck the monotonic lease before persistence | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B8 | AST `if` at `internal/app/engine/exitloop.go:1030`; retain ladder identity/rung guards and recheck the monotonic lease before persistence | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| Return | all admitted paths | retain ladder identity/rung guards and recheck the monotonic lease before persistence | exact function result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| direct A111 time/evidence collaborators | retain ladder identity/rung guards and recheck the monotonic lease before persistence | no clock movement may extend a lease or upgrade a stopped marker | AST + named RED |

## State mutations and fallbacks

- retain ladder identity/rung guards and recheck the monotonic lease before persistence.
- Monotonic anchors are process-local and never persisted; persisted observation timestamps remain UTC wall evidence.
- Marker status is read once and may only be downgraded by a later response-time authority.
- Every AST branch is paired with a named test in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: A111 evidence lifetime and fail-closed response projection; no LIVE-order authority is introduced.
- High-risk impact: yes.
