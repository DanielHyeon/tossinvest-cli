# Function Logic Map: `scanPositionPolicyScope`

- Source: `internal/journal/position_policy.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| provenance columns | nullable entry/adoption IDs | positions row | typed UNKNOWN/ENGINE_ENTRY/EXTERNAL_ADOPTION/AMBIGUOUS |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | scan/default/lifecycle/released projections | hydrate DTO only | error/state | policy projection tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| provenance classifier | derive typed lifecycle eligibility | deterministic | AST |

## State mutations and fallbacks

- Read-only projection exposes both provenance and typed eligibility to engine and UI.

## Safety conclusion

- Safe edit boundary: classify from explicit IDs; never infer external adoption from status or exit state.
- High-risk impact: yes
