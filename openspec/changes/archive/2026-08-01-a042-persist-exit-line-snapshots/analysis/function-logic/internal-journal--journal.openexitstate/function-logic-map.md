# Function Logic Map: `Journal.OpenExitState`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ExitStateSeed`, projected position row | eligible position; valid decimal t0; pinned immutable policy identity | OpenSpec a041/a042, `position.ExitEligible`, exitpolicy registry | reject before write; no partial state |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B8 | validate id, kind, policy and t0 arithmetic | none | typed invalid/conflict error | existing exit-state validation tests |
| B9-B14 | begin transaction; read position identity/generation; verify eligibility and policy identity | read lock only | rollback on error | `TestOpenExitState*` and v10 seed tests |
| B15-B20 | insert state+OPENED event and commit; duplicate position | one atomic state/event write | unique conflict or rollback | atomic seed/reopen tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitpolicy.OpenRatchetState`, `seedPolicyIdentity` | validate t0 and immutable policy meaning | fail closed; no retry | CodeGraph + AST |
| SQLite transaction / `appendExitEventTx` | persist seed and history atomically | BEGIN IMMEDIATE, rollback on all errors | CodeGraph + AST |

## State mutations and fallbacks

- Before: only policy ID was stored and returned identity lived in memory.
- After boundary: v10 rows persist identity+generation as a typed `SEED`; historical NULLs are never backfilled.

## Safety conclusion

- Safe edit boundary: additive nullable columns in the existing atomic open transaction.
- High-risk impact: yes; no order submission or operational toggle is added.
