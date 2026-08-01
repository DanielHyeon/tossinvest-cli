# Function Logic Map: `Store.RecoverConflict`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| capability and store actor | opaque capability; actor fixed at construction | caller token plus `Store.actor` | delegated recovery rejects invalid/tampered/cross-actor token |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless wrapper happy path | read-only delegation | conflict view or typed error | conflict actor-binding tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Store.recoverConflict` | binds immutable store actor | one call/no retry | cross-actor/replay tests |

## State mutations and fallbacks

- Read-only wrapper; no candidate consumption or snapshot mutation.

## Safety conclusion

- Safe edit boundary: actor-bound conflict recovery entry point.
- High-risk impact: yes; attempted changes are disclosed only after capability validation.
