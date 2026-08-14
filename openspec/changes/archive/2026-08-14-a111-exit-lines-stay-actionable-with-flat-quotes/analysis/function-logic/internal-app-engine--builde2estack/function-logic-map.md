# Function Logic Map: `buildE2EStack`

- Source: `internal/app/engine/exit_e2e_test.go`
- Post-edit AST evidence: `ast.json` (4 branches; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| A111 state | persisted evidence, request/cycle state, and injected clock/marker | current source + approved A111 delta | invalid, stale, or incomplete evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST branch at `internal/app/engine/exit_e2e_test.go:258`; test-stack setup retains the observer and journal seams required to assert a durable heartbeat without a live order | branch-local only; no authority added | function-defined fail-closed result | `TestA111FlatAndInBandQuotesRefreshProvenanceWithoutHistoryOrOrders` |
| B2 | AST branch at `internal/app/engine/exit_e2e_test.go:262`; test-stack setup retains the observer and journal seams required to assert a durable heartbeat without a live order | branch-local only; no authority added | function-defined fail-closed result | `TestA111FlatAndInBandQuotesRefreshProvenanceWithoutHistoryOrOrders` |
| B3 | AST branch at `internal/app/engine/exit_e2e_test.go:272`; test-stack setup retains the observer and journal seams required to assert a durable heartbeat without a live order | branch-local only; no authority added | function-defined fail-closed result | `TestA111FlatAndInBandQuotesRefreshProvenanceWithoutHistoryOrOrders` |
| B4 | AST branch at `internal/app/engine/exit_e2e_test.go:280`; test-stack setup retains the observer and journal seams required to assert a durable heartbeat without a live order | branch-local only; no authority added | function-defined fail-closed result | `TestA111FlatAndInBandQuotesRefreshProvenanceWithoutHistoryOrOrders` |
| Return | all admitted paths | preserves the A111 safety boundary | propagates typed error or read-only result | `TestA111FlatAndInBandQuotesRefreshProvenanceWithoutHistoryOrOrders` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| A111 direct collaborators | test-stack setup retains the observer and journal seams required to assert a durable heartbeat without a live order | typed conflict/stale/invalid results do not authorize an order or fresh line | current AST + A111 RED |

## State mutations and fallbacks

- test-stack setup retains the observer and journal seams required to assert a durable heartbeat without a live order
- The function remains inside its existing journal, engine, or projection authority; it does not create a new LIVE-order path.
- Every AST branch is paired with the named A111 RED in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: the A111 heartbeat/evidence freshness seam only; retain full judgement and existing order ordering where applicable.
- High-risk impact: yes.
