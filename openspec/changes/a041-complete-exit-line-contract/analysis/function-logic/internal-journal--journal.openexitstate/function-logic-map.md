# Function Logic Map: `Journal.OpenExitState`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| seed | eligible position, one policy kind, exact runtime policy identity | entry/adoption source + fixed registry | invalid/ambiguous identity refused before transaction |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | missing position/invalid kind | none | invalid request | validation tests |
| B2 | legacy identity omitted | bind exact pinned compatibility identity | continue | compatibility tests |
| B3 | supplied identity differs | none | identity conflict | policy seed test |
| B4 | duplicate/ineligible position | transaction rollback | domain error | existing journal tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `LegacyPolicyIdentity` | resolve pre-a042 fixed meaning | unknown fails closed | CodeGraph + AST |
| `OpenRatchetState` | derive t0 risk/baseline | arithmetic error becomes invalid request | CodeGraph + AST |
| `appendExitEventTx` | audit opening in same transaction | write failure rolls back | CodeGraph + AST |

## State mutations and fallbacks

- Identity is required in the typed seed/read seam but deliberately not persisted until a042 owns its columns.

## Safety conclusion

- Safe edit boundary: validate immutable identity before the existing schema write.
- High-risk impact: yes — opens protection state.
