# Function Logic Map: `systemClock.leaseElapsed`

- Source: `internal/clock/clock.go`
- Post-edit AST evidence: `ast.json` (0 branches; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| time/evidence state | injected clock or process-local monotonic lease, persisted evidence, and one marker status read | current source or explicit frozen-base revision + approved A111 contract | wall rollback, stale, stopped, invalid, or unavailable evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| none | branch-free seam; measure a system lease with time.Since against the retained monotonic reading | function-defined only | exact result | `TestSystemLeaseAnchorRetainsMonotonicReading` |
| Return | all admitted paths | measure a system lease with time.Since against the retained monotonic reading | exact function result | `TestSystemLeaseAnchorRetainsMonotonicReading` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| direct A111 time/evidence collaborators | measure a system lease with time.Since against the retained monotonic reading | no clock movement may extend a lease or upgrade a stopped marker | AST + named RED |

## State mutations and fallbacks

- measure a system lease with time.Since against the retained monotonic reading.
- Monotonic anchors are process-local and never persisted; persisted observation timestamps remain UTC wall evidence.
- Marker status is read once and may only be downgraded by a later response-time authority.
- Every AST branch is paired with a named test in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: A111 evidence lifetime and fail-closed response projection; no LIVE-order authority is introduced.
- High-risk impact: yes.
