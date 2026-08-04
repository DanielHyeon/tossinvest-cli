# Function Logic Map: `EvaluateUS`

- Source: `internal/weeklyvaluelane/evaluate.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| US weekly request | sealed EDGAR evidence/config/calendar/reservation/context | lane package | shared evaluator returns typed refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless wrapper invokes admitted shared evaluator with US/EDGAR binding | none | exact shared outcome | paired weekly tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `evaluate` | admitted US weekly path | pure, no retry | AST and paired tests |

## State mutations and fallbacks

- None; no fallback or mutation.

## Safety conclusion

- Safe edit boundary: preserve US/EDGAR binding.
- High-risk impact: yes; wrong binding could cross markets or disclosure sources.
