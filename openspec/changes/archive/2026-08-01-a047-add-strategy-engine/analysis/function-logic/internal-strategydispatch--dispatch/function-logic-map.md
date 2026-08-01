# Function Logic Map: `Dispatch`

- Source: `internal/strategydispatch/dispatch.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| opaque decision | private valid bit minted by pure lane | strategyengine | reject before dependencies/authority |
| dependencies | dormant concrete gate/manifest/Guardian/execgw adapters | approved future wiring | activation error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid opaque decision | none | `decision_invalid` | zero decision test |
| Success | opaque decision valid | delegates to private validated core | core result | lane-to-dispatch integration |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Decision.Valid/Record` | one-way sealed handoff | invalid record cannot be restored/minted | AST + purity guards |
| `dispatchValidated` | gate/plan/lease/official terminal core | no exported bypass | AST + same-package tests |

## State mutations and fallbacks

- This exported entry is the only production path into the private testable core.

## Safety conclusion

- Safe edit boundary: opaque decision validation before any authority.
- High-risk impact: yes, it precedes a potential official order call.
