# Function Logic Map: `TestFailedV12MigrationRollsBackTablesAndUserVersion`

- Source: `internal/journal/migration_v12_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| broken v12 step | table + user_version write + invalid insert | injected plan | whole transaction rolls back |
| v11 backup | seeded predecessor and one automatic backup | migration framework | restore then production migrate to 12 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | seed/fail/open survivor/version/rows | no partial DDL or metadata | fatal on mismatch | test body |
| B6-B12 | partial table/columns absent, close, backup v11, restore to v12 | recovery flow | fatal on mismatch | test body |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Open` with broken target 12 | prove DDL and PRAGMA transactional rollback | recovery error with backup path | AST |
| `assertBackupAtVersion`/`restoreBackup` | prove backup-only recovery | v11 then clean forward migration | AST |

## State mutations and fallbacks

- Failed v12 cannot leave lifecycle columns, policy tables, or user_version=12 behind.

## Safety conclusion

- Safe edit boundary: assert both schema objects and metadata, not just Open error.
- High-risk impact: yes
