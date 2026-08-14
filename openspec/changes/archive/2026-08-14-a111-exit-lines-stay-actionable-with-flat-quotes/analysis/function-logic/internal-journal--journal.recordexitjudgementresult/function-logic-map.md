# Function Logic Map: `Journal.RecordExitJudgementResult`

- Source: `internal/journal/exit_state.go`
- Frozen span: lines 385-390 at HEAD `3355df0fe9c82c3bb8c522e2d79abf107dd5f2c3`
- AST evidence: `ast.json` (0 branches, 1 call, 1 return)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `judgement` | legacy scalar or complete provenance-bound snapshot | engine/compatibility callers | transaction returns typed validation/storage error |
| result default | `not_requested`, no proposal | wrapper | remains conservative unless transaction commits and overwrites |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| Return | always | delegate all validation/mutation | return result plus exact transaction error | wrapper/delegation test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `recordExitJudgementTx` | one transaction and post-commit authority projection | error propagated unchanged | CodeGraph: sole callee; impact 47 symbols at public seam |

## State mutations and fallbacks

- Wrapper performs no database mutation itself.
- The result starts fail-closed, preventing callers from treating a failed transaction as an armed proposal.
- A new narrow refresh should have a separate typed API/transaction, not overload this proposal-bearing result seam.

## Safety conclusion

- Safe edit boundary: preserve wrapper and existing result semantics.
- High-risk impact: yes through its transaction, though wrapper itself is trivial.
