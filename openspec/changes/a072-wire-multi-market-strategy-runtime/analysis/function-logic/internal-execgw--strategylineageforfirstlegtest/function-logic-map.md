# Function Logic Map: `StrategyLineageForFirstLegTest`

- Source: `internal/execgw/export_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| strategy decision record and exact first-leg identity | record plus exact final quantity, q_final policy, settings version, manifest digest and attempt | strategyengine decision and frozen strategy admission inputs | delegates the production lineage constructor error unchanged |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | delegate accepts or rejects the exact lineage tuple | none | complete `journal.StrategyDecisionLineage` or validation error | paired KR/US atomic first-leg test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyDecisionLineage` | exercise the production private lineage encoder without duplicating payload/canonicalization rules | synchronous; propagate error; no fallback | CodeGraph + AST |

## State mutations and fallbacks

- Read-only test adapter. It neither persists a decision nor creates a dispatch lease.
- Exact journal validation still runs inside the atomic first-leg record operation.

## Safety conclusion

- Safe edit boundary: expose only the already-constructed lineage value to package tests.
- High-risk impact: **no** — `_test.go` only and no durable mutation.
