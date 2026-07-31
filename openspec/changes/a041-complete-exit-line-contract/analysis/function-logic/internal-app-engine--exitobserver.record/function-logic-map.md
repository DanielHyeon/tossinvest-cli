# Function Logic Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| authoritative snapshot/judgement | coherent policy/state/action/ratio/projected quantity tuple | exitpolicy evaluator | refuse/hold; never recalculate |
| managed position + cycle | current journal generation and held quantity | working set | no proposal count before arm |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | action non-orderable/projected zero | state judgement only | nil after record | zero-order tests |
| B2 | full exit or cancel-first | cancel conflicting orders first | error/withhold if uncleared | existing delay tests |
| B3 | orderable | mint intent and attach proposal to judgement | continue | integration tests |
| B4 | concurrent pending proposal | no second submission | nil | race test |
| B5 | judgement-only | no issuer/broker call | nil | state-only tests |
| B6 | armed proposal | increment cycle and submit projected quantity | submit result | integration tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `clearTheSymbol` | prevent oversell before full/cancel-first exit | failures withhold and alert by delay path | CodeGraph + AST |
| `Journal.RecordExitJudgement` | atomically advance state/arm | pending race is benign no-op | CodeGraph + AST |
| `submit` | Guardian reduction and official broker path | existing release/in-doubt contract | CodeGraph + AST |

## State mutations and fallbacks

- Journal mutation occurs before broker submission. Zero projection must never mint/arm/submit.

## Safety conclusion

- Safe edit boundary: replace local quantity recalculation with snapshot projection while preserving arm-before-submit and liquidation clearing.
- High-risk impact: yes — journal and sell submission path.
