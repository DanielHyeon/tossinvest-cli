# Function Logic Map: `TestExitObserverConcurrentSameQuoteArmsAndMutatesOnce`

- Source: `internal/app/engine/exit_identity_concurrency_test.go`
- Post-edit AST evidence: `ast.json` (12 branches; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| A111 state | persisted evidence, request/cycle state, and injected clock/marker | current source + approved A111 delta | invalid, stale, or incomplete evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST branch at `internal/app/engine/exit_identity_concurrency_test.go:92`; existing concurrency assertion is adjusted only for A111 durable observation semantics; it still proves one durable/order-bearing winner | branch-local only; no authority added | function-defined fail-closed result | `TestA111SameCycleRaceLeavesExactlyOneCompleteWinner` |
| B2 | AST branch at `internal/app/engine/exit_identity_concurrency_test.go:95`; existing concurrency assertion is adjusted only for A111 durable observation semantics; it still proves one durable/order-bearing winner | branch-local only; no authority added | function-defined fail-closed result | `TestA111SameCycleRaceLeavesExactlyOneCompleteWinner` |
| B3 | AST branch at `internal/app/engine/exit_identity_concurrency_test.go:113`; existing concurrency assertion is adjusted only for A111 durable observation semantics; it still proves one durable/order-bearing winner | branch-local only; no authority added | function-defined fail-closed result | `TestA111SameCycleRaceLeavesExactlyOneCompleteWinner` |
| B4 | AST branch at `internal/app/engine/exit_identity_concurrency_test.go:117`; existing concurrency assertion is adjusted only for A111 durable observation semantics; it still proves one durable/order-bearing winner | branch-local only; no authority added | function-defined fail-closed result | `TestA111SameCycleRaceLeavesExactlyOneCompleteWinner` |
| B5 | AST branch at `internal/app/engine/exit_identity_concurrency_test.go:120`; existing concurrency assertion is adjusted only for A111 durable observation semantics; it still proves one durable/order-bearing winner | branch-local only; no authority added | function-defined fail-closed result | `TestA111SameCycleRaceLeavesExactlyOneCompleteWinner` |
| B6 | AST branch at `internal/app/engine/exit_identity_concurrency_test.go:123`; existing concurrency assertion is adjusted only for A111 durable observation semantics; it still proves one durable/order-bearing winner | branch-local only; no authority added | function-defined fail-closed result | `TestA111SameCycleRaceLeavesExactlyOneCompleteWinner` |
| B7 | AST branch at `internal/app/engine/exit_identity_concurrency_test.go:127`; existing concurrency assertion is adjusted only for A111 durable observation semantics; it still proves one durable/order-bearing winner | branch-local only; no authority added | function-defined fail-closed result | `TestA111SameCycleRaceLeavesExactlyOneCompleteWinner` |
| B8 | AST branch at `internal/app/engine/exit_identity_concurrency_test.go:131`; existing concurrency assertion is adjusted only for A111 durable observation semantics; it still proves one durable/order-bearing winner | branch-local only; no authority added | function-defined fail-closed result | `TestA111SameCycleRaceLeavesExactlyOneCompleteWinner` |
| B9 | AST branch at `internal/app/engine/exit_identity_concurrency_test.go:134`; existing concurrency assertion is adjusted only for A111 durable observation semantics; it still proves one durable/order-bearing winner | branch-local only; no authority added | function-defined fail-closed result | `TestA111SameCycleRaceLeavesExactlyOneCompleteWinner` |
| B10 | AST branch at `internal/app/engine/exit_identity_concurrency_test.go:137`; existing concurrency assertion is adjusted only for A111 durable observation semantics; it still proves one durable/order-bearing winner | branch-local only; no authority added | function-defined fail-closed result | `TestA111SameCycleRaceLeavesExactlyOneCompleteWinner` |
| B11 | AST branch at `internal/app/engine/exit_identity_concurrency_test.go:141`; existing concurrency assertion is adjusted only for A111 durable observation semantics; it still proves one durable/order-bearing winner | branch-local only; no authority added | function-defined fail-closed result | `TestA111SameCycleRaceLeavesExactlyOneCompleteWinner` |
| B12 | AST branch at `internal/app/engine/exit_identity_concurrency_test.go:145`; existing concurrency assertion is adjusted only for A111 durable observation semantics; it still proves one durable/order-bearing winner | branch-local only; no authority added | function-defined fail-closed result | `TestA111SameCycleRaceLeavesExactlyOneCompleteWinner` |
| Return | all admitted paths | preserves the A111 safety boundary | propagates typed error or read-only result | `TestA111SameCycleRaceLeavesExactlyOneCompleteWinner` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| A111 direct collaborators | existing concurrency assertion is adjusted only for A111 durable observation semantics; it still proves one durable/order-bearing winner | typed conflict/stale/invalid results do not authorize an order or fresh line | current AST + A111 RED |

## State mutations and fallbacks

- existing concurrency assertion is adjusted only for A111 durable observation semantics; it still proves one durable/order-bearing winner
- The function remains inside its existing journal, engine, or projection authority; it does not create a new LIVE-order path.
- Every AST branch is paired with the named A111 RED in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: the A111 heartbeat/evidence freshness seam only; retain full judgement and existing order ordering where applicable.
- High-risk impact: yes.
