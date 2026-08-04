# Function Logic Map: `EvaluateUS`

- Source: `internal/reversallane/evaluate.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| US request | sealed US evidence/config/context | lane package | shared evaluator returns typed refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless wrapper computes US metric and invokes admitted shared evaluator with US lane ID | none | exact shared outcome | paired reversal tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `EvaluateUSMetric`, `evaluate` | US metric and admitted lane path | pure, no retry | AST and paired tests |

## State mutations and fallbacks

- None; no fallback or mutation.

## Safety conclusion

- Safe edit boundary: preserve US lane ID and evidence binding.
- High-risk impact: yes; wrong binding could cross markets.
