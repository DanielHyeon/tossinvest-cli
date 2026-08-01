# Function Logic Map: `Store.Apply`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Apply request and Store default actor | actor was validated at Open | immutable Store options | delegated validation fails closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | default actor apply | delegates without mutating actor | internal apply result | `TestCloseIsNilSafeAndForActorDoesNotMutateStoreActor` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Store.apply` | central durable validation and CAS | propagates typed failures | AST |

## State mutations and fallbacks

- No direct mutation and no LIVE/journal/lane/gate/order binding.

## Safety conclusion

- Safe edit boundary: default actor delegation.
- High-risk impact: no.
