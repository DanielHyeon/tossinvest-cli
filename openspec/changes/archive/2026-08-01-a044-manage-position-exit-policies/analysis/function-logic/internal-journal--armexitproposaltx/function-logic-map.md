# Function Logic Map: `armExitProposalTx`

- Source: `internal/journal/apply_hook.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| proposal and transaction | valid test/domain fixture | journal apply-point contract | typed conflict/error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1..Bn | each AST branch | arms only an empty proposal | assertion/error | branch map below |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `SQL read/update` | enforce the mapped contract | fail closed; no automatic retry | CodeGraph + AST |

## State mutations and fallbacks

- arms only an empty proposal.

## Safety conclusion

- Safe edit boundary: preserve sole-writer guarded columns.
- High-risk impact: yes.
