# Function Logic Map: `ExitObserver.observe`

- Source: `internal/app/engine/exitloop.go`
- Post-edit AST evidence: `ast.json` (15 branches; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| time/evidence state | injected clock or process-local monotonic lease, persisted evidence, and one marker status read | current source or explicit frozen-base revision + approved A111 contract | wall rollback, stale, stopped, invalid, or unavailable evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at `internal/app/engine/exitloop.go:750`; capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |
| B2 | AST `range` at `internal/app/engine/exitloop.go:759`; capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |
| B3 | AST `range` at `internal/app/engine/exitloop.go:763`; capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |
| B4 | AST `if` at `internal/app/engine/exitloop.go:765`; capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |
| B5 | AST `if` at `internal/app/engine/exitloop.go:769`; capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |
| B6 | AST `else` at `internal/app/engine/exitloop.go:772`; capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |
| B7 | AST `if` at `internal/app/engine/exitloop.go:772`; capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |
| B8 | AST `if` at `internal/app/engine/exitloop.go:779`; capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |
| B9 | AST `if` at `internal/app/engine/exitloop.go:785`; capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |
| B10 | AST `if` at `internal/app/engine/exitloop.go:788`; capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |
| B11 | AST `if` at `internal/app/engine/exitloop.go:794`; capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |
| B12 | AST `for` at `internal/app/engine/exitloop.go:797`; capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |
| B13 | AST `if` at `internal/app/engine/exitloop.go:799`; capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |
| B14 | AST `range` at `internal/app/engine/exitloop.go:805`; capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |
| B15 | AST `if` at `internal/app/engine/exitloop.go:806`; capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |
| Return | all admitted paths | capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | exact function result | `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| direct A111 time/evidence collaborators | capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote | no clock movement may extend a lease or upgrade a stopped marker | AST + named RED |

## State mutations and fallbacks

- capture clock.LeaseAnchor after the one broker read, validate official wall evidence, and attach the same process-local lease anchor to every accepted quote.
- Monotonic anchors are process-local and never persisted; persisted observation timestamps remain UTC wall evidence.
- Marker status is read once and may only be downgraded by a later response-time authority.
- Every AST branch is paired with a named test in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: A111 evidence lifetime and fail-closed response projection; no LIVE-order authority is introduced.
- High-risk impact: yes.
