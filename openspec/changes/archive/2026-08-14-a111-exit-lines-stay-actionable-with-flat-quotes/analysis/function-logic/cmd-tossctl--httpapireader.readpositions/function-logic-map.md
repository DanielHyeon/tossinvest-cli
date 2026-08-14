# Function Logic Map: `httpAPIReader.readPositions`

- Source: `cmd/tossctl/httpapi_reader.go`
- Frozen-base AST evidence: `ast.json` (16 branches; revision `base`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| A111 evidence/state | persisted evidence, request/cycle state, and injected clock/marker | current source or explicitly frozen base revision + approved A111 delta | invalid, stale, unavailable, or incomplete evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `if` at `cmd/tossctl/httpapi_reader.go:94`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| B2 | AST `if` at `cmd/tossctl/httpapi_reader.go:98`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| B3 | AST `if` at `cmd/tossctl/httpapi_reader.go:102`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| B4 | AST `if` at `cmd/tossctl/httpapi_reader.go:109`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| B5 | AST `if` at `cmd/tossctl/httpapi_reader.go:111`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| B6 | AST `if` at `cmd/tossctl/httpapi_reader.go:115`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| B7 | AST `range` at `cmd/tossctl/httpapi_reader.go:118`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| B8 | AST `range` at `cmd/tossctl/httpapi_reader.go:123`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| B9 | AST `range` at `cmd/tossctl/httpapi_reader.go:131`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| B10 | AST `if` at `cmd/tossctl/httpapi_reader.go:135`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| B11 | AST `else` at `cmd/tossctl/httpapi_reader.go:144`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| B12 | AST `if` at `cmd/tossctl/httpapi_reader.go:139`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| B13 | AST `if` at `cmd/tossctl/httpapi_reader.go:146`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| B14 | AST `range` at `cmd/tossctl/httpapi_reader.go:155`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| B15 | AST `if` at `cmd/tossctl/httpapi_reader.go:156`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| B16 | AST `if` at `cmd/tossctl/httpapi_reader.go:166`; frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |
| Return | all admitted paths | frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | exact function result | `TestA111RealPositionsRouteReadsJournalEvidenceThroughTheProductionAdapter` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| direct A111 collaborators | frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable | failures never authorize an order or fresh operator line | AST + named A111 RED |

## State mutations and fallbacks

- frozen-base monolithic reader was removed; its broker read and per-request local projection responsibilities are now separately testable.
- Local journal or broker failures remain visible; cached broker data never lends freshness to local evidence.
- Every AST branch is paired with the named A111 RED in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: A111 observation heartbeat, quote-evidence lifetime, or fail-closed operator projection only.
- High-risk impact: operator-facing fail-closed projection.
