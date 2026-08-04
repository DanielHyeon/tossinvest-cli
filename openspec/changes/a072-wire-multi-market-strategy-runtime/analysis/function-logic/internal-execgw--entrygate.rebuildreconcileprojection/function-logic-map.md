# Function Logic Map: EntryGate.RebuildReconcileProjection

- Source: internal/execgw/symbolgate.go
- AST evidence: ast.json
- Risk scan: risk-pattern-report.md

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function inputs and receiver state | Validated by implementation branches | internal/execgw/symbolgate.go and package-owned authority | Fail closed; no broker authority is synthesized |

## Branches and early returns

| Branch family | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| AST B1..Bn | Exact conditions are source-bound in ast.json | Only source-visible mutations | Typed error or safe return | Package and paired KR/US regression tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Direct callees in AST | Preserve authority ordering | Errors propagate or become fail-closed results | Current source and AST |

## State mutations and fallbacks

- State changes remain package-scoped; no fallback creates activation, owner, Gateway, or LIVE-order authority.
- KR and US retain identical safety ordering with market-local reads.

## Safety conclusion

- Safe edit boundary: preserve exact checks and no-byte-sent fences represented here.
- High-risk impact: reviewed and covered by paired KR/US tests.
