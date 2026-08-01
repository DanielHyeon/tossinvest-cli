# Function Logic Map: `managedPolicyIdentity`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| exit state + adopted flag | full persisted identity or pinned pre-a042 compatibility | a041 identity contract | fail closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | persisted tuple, ladder legacy, ratchet legacy | none | identity/error | identity tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| policy identity validators | immutable meaning | pure | CodeGraph + AST |

## State mutations and fallbacks

- v10 semantic classification prevents partial tuples reaching this compatibility function.

## Safety conclusion

- Safe edit boundary: unchanged.
- High-risk impact: yes.
