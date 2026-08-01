# Function Logic Map: `Journal.OpenExitState`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ExitStateSeed` | eligible existing position, pinned policy identity, positive risk | exit-policy + v10 snapshot contract | typed invalid/conflict refusal |
| current `exit_states` row | at most one per position | schema v6 primary key | unique violation becomes `ErrExitStateExists` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | missing/unknown kind | none | invalid request | exit state tests |
| B4-B10 | policy-kind/id normalization and identity validation | none | invalid/conflict refusal | policy snapshot tests |
| B11-B16 | arithmetic, tx, position/eligibility/identity failures | rollback | wrapped/typed refusal | exit state tests |
| B17-B20 | insert/event/commit/read failures | rollback until commit | wrapped error | durability and a044 release tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitpolicy.OpenRatchetState` | derive immutable t0 state | decimal validation, no I/O | CodeGraph + AST |
| `seedPolicyIdentity` | bind kind/id to registry meaning | fail closed on mismatch | a042 policy tests |
| `appendExitEventTx` | audit opening in same transaction | transaction-scoped | journal tests |

## State mutations and fallbacks

- Inserts one v10 exit state and OPENED event atomically; never overwrites an existing baseline.
- a044 release marks lifecycle state only and rejects active pending exit; it does not rewrite this function.

## Safety conclusion

- Safe edit boundary: preserve the position-scoped exit row and immutable saved policy identity.
- High-risk impact: yes — stop/protection state; a044 proves existing snapshots remain byte-stable.
