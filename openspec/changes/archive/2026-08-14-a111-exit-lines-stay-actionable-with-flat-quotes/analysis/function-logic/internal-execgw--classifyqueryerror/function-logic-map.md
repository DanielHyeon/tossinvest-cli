# Function Logic Map: `ClassifyQueryError`

- Source: `internal/execgw/retry.go`
- Post-edit AST evidence: `ast.json` (13 branches; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| A111 evidence/state | persisted evidence, request/cycle state, and injected clock/marker | current source or explicitly frozen base revision + approved A111 delta | invalid, stale, unavailable, or incomplete evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST `switch` at `internal/execgw/retry.go:70`; classify decoded invalid evidence as permanent so it cannot retry or stamp query success | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111AllInvalidHTTPBatchIsPermanentNonRetryableAndDoesNotOpenTheGate` |
| B2 | AST `case` at `internal/execgw/retry.go:71`; classify decoded invalid evidence as permanent so it cannot retry or stamp query success | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111AllInvalidHTTPBatchIsPermanentNonRetryableAndDoesNotOpenTheGate` |
| B3 | AST `case` at `internal/execgw/retry.go:73`; classify decoded invalid evidence as permanent so it cannot retry or stamp query success | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111AllInvalidHTTPBatchIsPermanentNonRetryableAndDoesNotOpenTheGate` |
| B4 | AST `case` at `internal/execgw/retry.go:75`; classify decoded invalid evidence as permanent so it cannot retry or stamp query success | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111AllInvalidHTTPBatchIsPermanentNonRetryableAndDoesNotOpenTheGate` |
| B5 | AST `case` at `internal/execgw/retry.go:77`; classify decoded invalid evidence as permanent so it cannot retry or stamp query success | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111AllInvalidHTTPBatchIsPermanentNonRetryableAndDoesNotOpenTheGate` |
| B6 | AST `case` at `internal/execgw/retry.go:79`; classify decoded invalid evidence as permanent so it cannot retry or stamp query success | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111AllInvalidHTTPBatchIsPermanentNonRetryableAndDoesNotOpenTheGate` |
| B7 | AST `case` at `internal/execgw/retry.go:81`; classify decoded invalid evidence as permanent so it cannot retry or stamp query success | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111AllInvalidHTTPBatchIsPermanentNonRetryableAndDoesNotOpenTheGate` |
| B8 | AST `if` at `internal/execgw/retry.go:85`; classify decoded invalid evidence as permanent so it cannot retry or stamp query success | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111AllInvalidHTTPBatchIsPermanentNonRetryableAndDoesNotOpenTheGate` |
| B9 | AST `switch` at `internal/execgw/retry.go:86`; classify decoded invalid evidence as permanent so it cannot retry or stamp query success | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111AllInvalidHTTPBatchIsPermanentNonRetryableAndDoesNotOpenTheGate` |
| B10 | AST `case` at `internal/execgw/retry.go:87`; classify decoded invalid evidence as permanent so it cannot retry or stamp query success | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111AllInvalidHTTPBatchIsPermanentNonRetryableAndDoesNotOpenTheGate` |
| B11 | AST `case` at `internal/execgw/retry.go:89`; classify decoded invalid evidence as permanent so it cannot retry or stamp query success | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111AllInvalidHTTPBatchIsPermanentNonRetryableAndDoesNotOpenTheGate` |
| B12 | AST `case` at `internal/execgw/retry.go:91`; classify decoded invalid evidence as permanent so it cannot retry or stamp query success | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111AllInvalidHTTPBatchIsPermanentNonRetryableAndDoesNotOpenTheGate` |
| B13 | AST `case` at `internal/execgw/retry.go:93`; classify decoded invalid evidence as permanent so it cannot retry or stamp query success | limited to the stated seam; no new order authority | typed/read-only fail-closed result | `TestA111AllInvalidHTTPBatchIsPermanentNonRetryableAndDoesNotOpenTheGate` |
| Return | all admitted paths | classify decoded invalid evidence as permanent so it cannot retry or stamp query success | exact function result | `TestA111AllInvalidHTTPBatchIsPermanentNonRetryableAndDoesNotOpenTheGate` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| direct A111 collaborators | classify decoded invalid evidence as permanent so it cannot retry or stamp query success | failures never authorize an order or fresh operator line | AST + named A111 RED |

## State mutations and fallbacks

- classify decoded invalid evidence as permanent so it cannot retry or stamp query success.
- Local journal or broker failures remain visible; cached broker data never lends freshness to local evidence.
- Every AST branch is paired with the named A111 RED in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: A111 observation heartbeat, quote-evidence lifetime, or fail-closed operator projection only.
- High-risk impact: yes.
