# Function Logic Map: `RiskGuardian.IssueStrategyEntry`

- Source: `internal/execgw/riskguardian.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| opaque decision | `Decision.Valid()` minted by pure lane | strategyengine | reject before sizing/journal |
| expected policy/limits | exact Guardian-owned versions | activation manifest vs Guardian | snapshot mismatch |
| attempt/settings/activation | non-empty exact plan bindings | dispatch manifest | atomic journal refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | opaque decision invalid | none | error | invalid decision adapter tests |
| B2 | policy or limits digest differs | none | error | Guardian binding tests |
| B3 | exact sizing yields error/zero | none | wrapped sizing error | risk sizing tests |
| B4 | record encoding fails | none | error | structurally unreachable for scalar record |
| Success | complete exact lineage built | delegates to `IssueEntry` atomic strategy branch | issued decision/reservation/receipt | strategy issuance integration |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `risk.StrategyEntryQuantity` | Guardian-owned min-cap sizing | exact/fail closed | AST + risk tests |
| `strategyDecisionLineage` | bind and hash the complete opaque decision payload | no lossy projection | AST + 60-field payload test |
| `IssueEntry` | reuse chain and atomic reservation issuance | no gateway call here | AST + Guardian tests |

## State mutations and fallbacks

- This function does not submit orders or enable runtime wiring.
- The caller cannot supply quantity or a public journal plan; both are built here.

## Safety conclusion

- Safe edit boundary: opaque decision to existing Guardian issuance only.
- High-risk impact: yes, exposure-raising authority is minted atomically.
