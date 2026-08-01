# Function Logic Map: `Store.PreviewRollback`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| rollback request and store actor | explicit base/target/category; actor fixed at construction | caller and `Store.actor` | delegated rollback preview fails closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless wrapper happy path | delegates once | rollback preview or typed error | rollback actor-binding tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Store.previewRollback` | binds immutable store actor | one call/no retry | actor/reopen rollback tests |

## State mutations and fallbacks

- Wrapper does not mutate history and cannot accept a caller-selected actor.

## Safety conclusion

- Safe edit boundary: actor-bound rollback preview entry point.
- High-risk impact: yes; rollback is append-only candidate generation.
