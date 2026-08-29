# Function Logic Map: `TestMigrationV7ToV8PreservesEveryRow`

- Source: `internal/journal/migration_v8_test.go`
- AST evidence: `ast.json` (`47390b93f0a39f2a46256ea58f99f024192dce4c9953c39906e44aeded5ceb09`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | validated caller inputs | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | existing if branch at line 74 | only the branch's documented state transition | existing return/error contract | `TestTestMigrationV7ToV8PreservesEveryRow` |
| B2 | existing if branch at line 79 | only the branch's documented state transition | existing return/error contract | `TestTestMigrationV7ToV8PreservesEveryRow` |
| B3 | existing if branch at line 85 | only the branch's documented state transition | existing return/error contract | `TestTestMigrationV7ToV8PreservesEveryRow` |
| B4 | existing if branch at line 88 | only the branch's documented state transition | existing return/error contract | `TestTestMigrationV7ToV8PreservesEveryRow` |
| B5 | existing range branch at line 93 | only the branch's documented state transition | existing return/error contract | `TestTestMigrationV7ToV8PreservesEveryRow` |
| B6 | existing if branch at line 94 | only the branch's documented state transition | existing return/error contract | `TestTestMigrationV7ToV8PreservesEveryRow` |
| B7 | existing if branch at line 97 | only the branch's documented state transition | existing return/error contract | `TestTestMigrationV7ToV8PreservesEveryRow` |
| B8 | existing if branch at line 106 | only the branch's documented state transition | existing return/error contract | `TestTestMigrationV7ToV8PreservesEveryRow` |
| B9 | existing if branch at line 111 | only the branch's documented state transition | existing return/error contract | `TestTestMigrationV7ToV8PreservesEveryRow` |
| B10 | existing if branch at line 114 | only the branch's documented state transition | existing return/error contract | `TestTestMigrationV7ToV8PreservesEveryRow` |
| B11 | existing range branch at line 118 | only the branch's documented state transition | existing return/error contract | `TestTestMigrationV7ToV8PreservesEveryRow` |
| B12 | existing if branch at line 120 | only the branch's documented state transition | existing return/error contract | `TestTestMigrationV7ToV8PreservesEveryRow` |
| B13 | existing if branch at line 124 | only the branch's documented state transition | existing return/error contract | `TestTestMigrationV7ToV8PreservesEveryRow` |
| B14 | existing if branch at line 128 | only the branch's documented state transition | existing return/error contract | `TestTestMigrationV7ToV8PreservesEveryRow` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST-listed callees | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- preserve existing fail-closed behavior.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: preserve existing fail-closed behavior.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
