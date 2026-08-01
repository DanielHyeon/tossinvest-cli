# Function Logic Map: `TestMigrationV12ToV13IsAdditiveAndPreservesExistingRows`

- Source: `internal/journal/migration_v13_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| v12 journal fixture | isolated path with representative legacy rows | migration helpers | failure aborts test |
| production migration plan | current `SchemaVersion` | `openTestJournalAt` | opening/migration error fails test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | old journal close fails | no further migration | fatal | same test |
| B2 | current schema version read fails or is stale | none | fatal | same test |
| B3 | legacy row counts differ | none | fatal | same test |
| B4 | each v12/v13 artifact is inspected | read-only sqlite metadata | loop | same test |
| B5 | artifact missing/duplicated/query failure | none | fatal | same test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openJournalAtSchema` | constructs exact v12 baseline | isolated fixture | AST |
| `openTestJournalAt` | applies every released migration | failure is fatal | AST |
| `countRows` | verifies legacy data preservation | exact comparison | AST |

## State mutations and fallbacks

- Test-only fixture state. With v14 appended, the terminal expected version follows `SchemaVersion`; v13 artifacts remain mandatory.

## Safety conclusion

- Safe edit boundary: expectation-only update; do not weaken v13 artifact or legacy row assertions.
- High-risk impact: test-only coverage of high-risk journal migration.
