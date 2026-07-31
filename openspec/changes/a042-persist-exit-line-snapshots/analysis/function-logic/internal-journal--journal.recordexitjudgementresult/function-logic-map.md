# Function Logic Map: `Journal.RecordExitJudgementResult`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| judgement request | validated by transaction helper; request proposal is never direct submit authority | journal durable transaction | returns typed result plus error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | helper success/error | none beyond delegated transaction | typed post-commit result/error | durable result and rollback tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `recordExitJudgementTx` | execute one atomic write and populate result only after commit | error leaves default no-authority result | CodeGraph + AST |

## State mutations and fallbacks

- Initializes `not_requested`; no result is upgraded until the delegated transaction commits.

## Safety conclusion

- Safe edit boundary: typed wrapper around the journal transaction.
- High-risk impact: yes; callers gate broker submission on this return value.
