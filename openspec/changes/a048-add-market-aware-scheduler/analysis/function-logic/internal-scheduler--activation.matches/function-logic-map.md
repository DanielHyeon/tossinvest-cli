# Function Logic Map: `Activation.matches`

- Source: `internal/scheduler/decision.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| activation | nil or sealed package-minted value | `Restore` | nil is false |
| desired/market | exact revision, versions, actor/time, scope/session/config | activation manifest binding | any mismatch is false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | receiver nil | none | false | decision activation matrix |
| return predicate | every binding field including desired revision matches and scope allows market | none | exact boolean | activation exact-state tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Time.Equal` | compare approved instant semantically | pure | CodeGraph + AST |
| `MarketScope.allows` | bind activation to evaluated market | unknown scope false | CodeGraph + AST |

## State mutations and fallbacks

- Pure comparison; no state mutation or fallback.

## Safety conclusion

- Safe edit boundary: sealed activation equality predicate.
- High-risk impact: yes, as an entry authority gate.
