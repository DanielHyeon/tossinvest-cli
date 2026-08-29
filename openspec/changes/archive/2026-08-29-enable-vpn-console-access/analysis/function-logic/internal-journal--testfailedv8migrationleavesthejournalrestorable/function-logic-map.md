# Function Logic Map: `TestFailedV8MigrationLeavesTheJournalRestorable`

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
| B1 | existing if branch at line 327 | only the branch's documented state transition | existing return/error contract | `TestTestFailedV8MigrationLeavesTheJournalRestorable` |
| B2 | existing if branch at line 342 | only the branch's documented state transition | existing return/error contract | `TestTestFailedV8MigrationLeavesTheJournalRestorable` |
| B3 | existing if branch at line 346 | only the branch's documented state transition | existing return/error contract | `TestTestFailedV8MigrationLeavesTheJournalRestorable` |
| B4 | existing if branch at line 349 | only the branch's documented state transition | existing return/error contract | `TestTestFailedV8MigrationLeavesTheJournalRestorable` |
| B5 | existing if branch at line 356 | only the branch's documented state transition | existing return/error contract | `TestTestFailedV8MigrationLeavesTheJournalRestorable` |
| B6 | existing if branch at line 360 | only the branch's documented state transition | existing return/error contract | `TestTestFailedV8MigrationLeavesTheJournalRestorable` |
| B7 | existing if branch at line 365 | only the branch's documented state transition | existing return/error contract | `TestTestFailedV8MigrationLeavesTheJournalRestorable` |
| B8 | existing if branch at line 369 | only the branch's documented state transition | existing return/error contract | `TestTestFailedV8MigrationLeavesTheJournalRestorable` |
| B9 | existing if branch at line 372 | only the branch's documented state transition | existing return/error contract | `TestTestFailedV8MigrationLeavesTheJournalRestorable` |
| B10 | existing if branch at line 375 | only the branch's documented state transition | existing return/error contract | `TestTestFailedV8MigrationLeavesTheJournalRestorable` |
| B11 | existing if branch at line 384 | only the branch's documented state transition | existing return/error contract | `TestTestFailedV8MigrationLeavesTheJournalRestorable` |
| B12 | existing if branch at line 388 | only the branch's documented state transition | existing return/error contract | `TestTestFailedV8MigrationLeavesTheJournalRestorable` |
| B13 | existing if branch at line 391 | only the branch's documented state transition | existing return/error contract | `TestTestFailedV8MigrationLeavesTheJournalRestorable` |

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
