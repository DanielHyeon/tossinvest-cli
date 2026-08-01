# Function Logic Map: `TestFailedV11MigrationRollsBackIndexAndVersion`

- Source: `internal/journal/migration_v11_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| broken v11 step | create index then invalid insert | injected migration plan | transaction rolls back |
| backup | single v10 pre-migration copy | migration framework | restorable to exactly v11 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B4 | setup/failure/survivor version | no partial commit | fatal on mismatch | test body |
| B5-B8 | no partial index, one backup, restore to exactly v11 | restore fixture | fatal on mismatch | test body |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Open` with target 11 | inject atomic failure | returns recovery-bearing error | AST |
| `openJournalAtSchema(...,11)` | prove restore without v12 | exact target | AST |

## State mutations and fallbacks

- A future schema cannot turn the a043 rollback test into a different migration.

## Safety conclusion

- Safe edit boundary: preserve index/user_version atomicity and exact v11 restore.
- High-risk impact: yes
