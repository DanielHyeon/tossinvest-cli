# Function Logic Map: `actorCommander.Apply`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| immutable bound actor and apply request | actor cannot be changed by another wrapper | `ForActor` | internal apply failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | apply delegation | passes bound actor, never writes Store.actor | Store result/error | `TestCloseIsNilSafeAndForActorDoesNotMutateStoreActor` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Store.apply` | durable validation and audit insertion | propagates typed errors | AST |

## State mutations and fallbacks

- The wrapper is value-immutable and concurrency-safe by construction.

## Safety conclusion

- Safe edit boundary: audit attribution only.
- High-risk impact: no trading authority.
