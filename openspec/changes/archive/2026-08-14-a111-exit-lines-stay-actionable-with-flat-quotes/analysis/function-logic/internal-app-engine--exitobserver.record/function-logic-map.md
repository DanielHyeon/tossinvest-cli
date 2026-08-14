# Function Logic Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go`
- Post-edit AST evidence: `ast.json` (16 branches; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| time/evidence state | injected clock or process-local monotonic lease, persisted evidence, and one marker status read | current source or explicit frozen-base revision + approved A111 contract | wall rollback, stale, stopped, invalid, or unavailable evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at `internal/app/engine/exitloop.go:1180`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B2 | AST `if` at `internal/app/engine/exitloop.go:1199`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B3 | AST `if` at `internal/app/engine/exitloop.go:1200`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B4 | AST `else` at `internal/app/engine/exitloop.go:1202`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B5 | AST `if` at `internal/app/engine/exitloop.go:1223`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B6 | AST `if` at `internal/app/engine/exitloop.go:1224`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B7 | AST `else` at `internal/app/engine/exitloop.go:1246`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B8 | AST `if` at `internal/app/engine/exitloop.go:1248`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B9 | AST `if` at `internal/app/engine/exitloop.go:1251`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B10 | AST `else` at `internal/app/engine/exitloop.go:1255`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B11 | AST `if` at `internal/app/engine/exitloop.go:1262`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B12 | AST `if` at `internal/app/engine/exitloop.go:1264`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B13 | AST `if` at `internal/app/engine/exitloop.go:1276`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B14 | AST `if` at `internal/app/engine/exitloop.go:1277`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B15 | AST `if` at `internal/app/engine/exitloop.go:1283`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| B16 | AST `if` at `internal/app/engine/exitloop.go:1296`; recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |
| Return | all admitted paths | recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | exact function result | `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| direct A111 time/evidence collaborators | recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work | no clock movement may extend a lease or upgrade a stopped marker | AST + named RED |

## State mutations and fallbacks

- recheck monotonic evidence lease immediately before clear/arm/submit full-judgement work.
- Monotonic anchors are process-local and never persisted; persisted observation timestamps remain UTC wall evidence.
- Marker status is read once and may only be downgraded by a later response-time authority.
- Every AST branch is paired with a named test in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: A111 evidence lifetime and fail-closed response projection; no LIVE-order authority is introduced.
- High-risk impact: yes.
