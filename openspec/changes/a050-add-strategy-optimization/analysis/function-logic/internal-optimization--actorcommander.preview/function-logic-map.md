# Function Logic Map: `actorCommander.Preview`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| bound Store and preview request | immutable wrapper | `ForActor` | Store preview error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | preview delegation | candidate creation only | Store result/error | lifecycle tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Store.Preview` | validate preview | propagates error | AST |

## State mutations and fallbacks

- Does not mutate wrapper actor or trading authority.

## Safety conclusion

- Safe edit boundary: narrow delegation.
- High-risk impact: no.
