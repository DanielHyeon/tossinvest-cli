# Function Logic Map: `DefaultStaleness`

- Source: `internal/execgw/retry.go`
- Post-edit AST evidence: `ast.json` (0 branches; revision `current`; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| A111 evidence/state | persisted evidence, request/cycle state, and injected clock/marker | current source or explicitly frozen base revision + approved A111 delta | invalid, stale, unavailable, or incomplete evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| none | branch-free seam; bind QueryPrice entry-gate staleness to the same exported duration used by source/use quote validation | function-defined only | function-defined result | `TestA111FallbackSequenceRecoveryIsLazyAndPriceEvidenceUsesTheGateDuration` |
| Return | all admitted paths | bind QueryPrice entry-gate staleness to the same exported duration used by source/use quote validation | exact function result | `TestA111FallbackSequenceRecoveryIsLazyAndPriceEvidenceUsesTheGateDuration` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| direct A111 collaborators | bind QueryPrice entry-gate staleness to the same exported duration used by source/use quote validation | failures never authorize an order or fresh operator line | AST + named A111 RED |

## State mutations and fallbacks

- bind QueryPrice entry-gate staleness to the same exported duration used by source/use quote validation.
- Local journal or broker failures remain visible; cached broker data never lends freshness to local evidence.
- Every AST branch is paired with the named A111 RED in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: A111 observation heartbeat, quote-evidence lifetime, or fail-closed operator projection only.
- High-risk impact: yes.
