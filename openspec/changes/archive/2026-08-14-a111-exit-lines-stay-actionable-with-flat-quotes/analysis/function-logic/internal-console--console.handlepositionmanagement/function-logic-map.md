# Function Logic Map: `Console.handlePositionManagement`

- Source: `internal/console/position_policy.go`
- Post-edit AST evidence: `ast.json` (25 branches; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| time/evidence state | injected clock or process-local monotonic lease, persisted evidence, and one marker status read | current source or explicit frozen-base revision + approved A111 contract | wall rollback, stale, stopped, invalid, or unavailable evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at `internal/console/position_policy.go:228`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B2 | AST `else` at `internal/console/position_policy.go:230`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B3 | AST `if` at `internal/console/position_policy.go:230`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B4 | AST `else` at `internal/console/position_policy.go:232`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B5 | AST `if` at `internal/console/position_policy.go:236`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B6 | AST `if` at `internal/console/position_policy.go:241`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B7 | AST `else` at `internal/console/position_policy.go:243`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B8 | AST `range` at `internal/console/position_policy.go:246`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B9 | AST `if` at `internal/console/position_policy.go:251`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B10 | AST `if` at `internal/console/position_policy.go:259`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B11 | AST `range` at `internal/console/position_policy.go:260`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B12 | AST `if` at `internal/console/position_policy.go:269`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B13 | AST `if` at `internal/console/position_policy.go:271`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B14 | AST `range` at `internal/console/position_policy.go:274`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B15 | AST `range` at `internal/console/position_policy.go:285`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B16 | AST `if` at `internal/console/position_policy.go:293`; row projection consumes downgrade-only marker liveness; a stopped read remains stopped after response-clock rollback | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementNeverResurrectsAStoppedMarkerAfterClockRollback` |
| B17 | AST `if` at `internal/console/position_policy.go:299`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B18 | AST `if` at `internal/console/position_policy.go:304`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B19 | AST `if` at `internal/console/position_policy.go:310`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B20 | AST `if` at `internal/console/position_policy.go:320`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B21 | AST `if` at `internal/console/position_policy.go:323`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B22 | AST `else` at `internal/console/position_policy.go:332`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B23 | AST `range` at `internal/console/position_policy.go:325`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B24 | AST `if` at `internal/console/position_policy.go:329`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| B25 | AST `if` at `internal/console/position_policy.go:332`; finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |
| Return | all admitted paths | finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | exact function result | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| direct A111 time/evidence collaborators | finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness | no clock movement may extend a lease or upgrade a stopped marker | AST + named RED |

## State mutations and fallbacks

- finish data/RPC reads, read marker once, capture a post-marker response clock, and project every row from downgrade-only liveness.
- Monotonic anchors are process-local and never persisted; persisted observation timestamps remain UTC wall evidence.
- Marker status is read once and may only be downgraded by a later response-time authority.
- Every AST branch is paired with a named test in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: A111 evidence lifetime and fail-closed response projection; no LIVE-order authority is introduced.
- High-risk impact: operator-facing fail-closed projection.
