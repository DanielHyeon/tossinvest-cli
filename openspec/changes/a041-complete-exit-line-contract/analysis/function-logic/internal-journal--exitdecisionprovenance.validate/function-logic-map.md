# Function Logic Map: `ExitDecisionProvenance.validate`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| provenance | all-zero legacy or complete IDs plus valid policy identity | exit snapshot | invalid request/identity conflict |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | all-zero | none | nil | existing journal tests |
| B2 | missing any ID | none | invalid request | provenance validation test |
| B3 | invalid policy | none | identity error | identity tests |
| B4 | complete | none | nil | concurrent engine test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `zero` | distinguish legacy omission | no error | AST |
| `PolicyIdentity.Validate` | validate semantic identity | error propagates | AST |

## State mutations and fallbacks

- Validation occurs before journal transaction.

## Safety conclusion

- Safe edit boundary: typed input validation.
- High-risk impact: yes — proposal provenance gate.
