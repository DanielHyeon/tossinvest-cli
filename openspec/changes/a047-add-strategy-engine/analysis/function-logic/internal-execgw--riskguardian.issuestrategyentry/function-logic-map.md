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
| B1 | `!request.Decision.Valid()` | none; collector and journal untouched | invalid-decision error | direct zero-value decision test |
| B2 | expected policy version or limits digest differs | none | Guardian snapshot mismatch | authentic direct test blocked until activation has a verified source manifest |
| B3 | Guardian-owned exact sizing returns error | none | wrapped sizing error | pure sizing tests only; direct method branch remains activation-gated |
| B4 | full `DecisionRecord` JSON encoding returns error | none | encoding error | no authentic direct call while source manifest is unavailable; current scalar schema makes the path unreachable but does not justify claiming branch execution |
| Scenario | complete exact lineage built | delegates to `IssueEntry` with private atomic strategy plan | issued decision/reservation/receipt | callee integration exists; direct authentic method success is activation-gated |

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
