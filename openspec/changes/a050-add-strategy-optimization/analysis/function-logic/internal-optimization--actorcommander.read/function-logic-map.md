# Function Logic Map: `actorCommander.Read`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| bound Store and request context | immutable wrapper | `ForActor` | Store read error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | read delegation | no mutation | Store View/error | wrapper test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Store.Read` | read lifecycle state | propagates error | AST |

## State mutations and fallbacks

- No actor or trading-state mutation.

## Safety conclusion

- Safe edit boundary: narrow delegation.
- High-risk impact: no.
