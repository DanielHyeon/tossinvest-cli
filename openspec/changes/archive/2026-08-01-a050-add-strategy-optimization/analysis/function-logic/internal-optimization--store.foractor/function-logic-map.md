# Function Logic Map: `Store.ForActor`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Store and server-verified actor | nonnil Store/DB and nonblank actor | server session integration | return error, never mutate Store actor |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil Store/DB or blank actor | none | actor-required error | `TestCloseIsNilSafeAndForActorDoesNotMutateStoreActor` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `actorCommander` | immutable value wrapper | no I/O/retry | AST |

## State mutations and fallbacks

- Actor attribution is held in a wrapper, not Store global state.

## Safety conclusion

- Safe edit boundary: audit attribution.
- High-risk impact: no LIVE authority.
