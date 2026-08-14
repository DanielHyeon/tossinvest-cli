# Function Logic Map: `httpAPIReader.projectPositions`

- Source: `cmd/tossctl/httpapi_reader.go`
- Post-edit AST evidence: `ast.json` (13 branches; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| time/evidence state | injected clock or process-local monotonic lease, persisted evidence, and one marker status read | current source or explicit frozen-base revision + approved A111 contract | wall rollback, stale, stopped, invalid, or unavailable evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at `cmd/tossctl/httpapi_reader.go:130`; finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` |
| B2 | AST `if` at `cmd/tossctl/httpapi_reader.go:133`; finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` |
| B3 | AST `if` at `cmd/tossctl/httpapi_reader.go:137`; finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` |
| B4 | AST `range` at `cmd/tossctl/httpapi_reader.go:140`; finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` |
| B5 | AST `range` at `cmd/tossctl/httpapi_reader.go:145`; finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` |
| B6 | AST `range` at `cmd/tossctl/httpapi_reader.go:156`; finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` |
| B7 | AST `if` at `cmd/tossctl/httpapi_reader.go:160`; finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` |
| B8 | AST `else` at `cmd/tossctl/httpapi_reader.go:169`; finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` |
| B9 | AST `if` at `cmd/tossctl/httpapi_reader.go:164`; finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` |
| B10 | AST `if` at `cmd/tossctl/httpapi_reader.go:171`; finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` |
| B11 | AST `range` at `cmd/tossctl/httpapi_reader.go:180`; finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` |
| B12 | AST `if` at `cmd/tossctl/httpapi_reader.go:181`; finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` |
| B13 | AST `if` at `cmd/tossctl/httpapi_reader.go:191`; finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line | limited to the stated seam; no order authority added | typed/read-only fail-closed result | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` |
| Return | all admitted paths | finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line | exact function result | `TestA111RealPositionsRouteUsesPostMarkerResponseClock` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| direct A111 time/evidence collaborators | finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line | no clock movement may extend a lease or upgrade a stopped marker | AST + named RED |

## State mutations and fallbacks

- finish journal, policy and runtime reads, then obtain one post-marker response authority for liveness and every line.
- Monotonic anchors are process-local and never persisted; persisted observation timestamps remain UTC wall evidence.
- Marker status is read once and may only be downgraded by a later response-time authority.
- Every AST branch is paired with a named test in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: A111 evidence lifetime and fail-closed response projection; no LIVE-order authority is introduced.
- High-risk impact: operator-facing fail-closed projection.
