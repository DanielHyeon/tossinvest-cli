# Function Logic Map: `TestMigrationV12CommitAndUserVersionSurviveSIGKILL`

- Source: `internal/journal/migration_v12_crash_linux_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| crash child mode | exact environment marker set by parent test | crash harness | opens current schema then SIGKILL |
| parent mode | v11 journal path in isolated temp directory | migration fixture | child result is inspected read-only |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | child mode | opens all migrations through current `SchemaVersion`; intentional SIGKILL | process death | same named subprocess test |
| B2 | parent mode | seeds v11 rows, launches child, reads raw artifacts | pass/fail assertions | `TestMigrationV12CommitAndUserVersionSurviveSIGKILL` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openCrashChildJournal` | opens production migration plan | failure exits child before kill | AST |
| `runCrashChild` | runs isolated subprocess | expects SIGKILL termination | AST |
| `countRows` | proves legacy rows survived | exact count equality | AST |

## State mutations and fallbacks

- Only test fixture state changes. a047 changes the expected terminal version from released v13 to current v14; v12/v13 artifacts and legacy rows remain required.

## Safety conclusion

- Safe edit boundary: compare against `SchemaVersion` so future additive migrations do not stale this crash contract.
- High-risk impact: test-only verification of high-risk migration durability.
