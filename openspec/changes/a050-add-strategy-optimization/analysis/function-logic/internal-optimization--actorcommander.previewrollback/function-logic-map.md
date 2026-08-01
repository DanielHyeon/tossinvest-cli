# Function Logic Map: `actorCommander.PreviewRollback`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| bound Store and rollback request | immutable wrapper | `ForActor` | Store preview error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | rollback preview delegation | candidate creation only | Store result/error | rollback tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Store.PreviewRollback` | validate rollback preview | propagates error | AST |

## State mutations and fallbacks

- No audit actor mutation until apply.

## Safety conclusion

- Safe edit boundary: narrow delegation.
- High-risk impact: no.
