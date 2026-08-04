# Function Logic Map: `EvaluateUS`

- Source: `internal/continuationlane/evaluator.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| US request | sealed plan/evidence/config/context | lane package | shared evaluator returns typed refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless wrapper invokes US signal evaluation and admitted shared evaluator | none | exact shared outcome | paired continuation tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `EvaluateUSParticipation`, `evaluate` | US metric then admitted evaluation | pure, no retry | AST and paired tests |

## State mutations and fallbacks

- None; wrapper has no fallback or mutation.

## Safety conclusion

- Safe edit boundary: preserve exact US-to-US binding.
- High-risk impact: yes; wrong binding could cross markets.
