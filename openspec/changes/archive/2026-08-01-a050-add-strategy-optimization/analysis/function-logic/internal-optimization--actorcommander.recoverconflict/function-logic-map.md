# Function Logic Map: `actorCommander.RecoverConflict`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| bound Store and capability | immutable wrapper | `ForActor` | Store recovery error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | recovery delegation | no mutation | Store conflict View/error | conflict tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Store.RecoverConflict` | read-only recovery | propagates error | AST |

## State mutations and fallbacks

- No actor or trading-state mutation.

## Safety conclusion

- Safe edit boundary: narrow delegation.
- High-risk impact: no.
