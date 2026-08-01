# Function Logic Map: `TestMigrationV13CommitAndUserVersionSurviveSIGKILL`

- Source: `internal/journal/migration_v13_crash_linux_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| crash child mode | exact environment marker set by parent test | crash harness | opens current schema then SIGKILL |
| parent mode | v12 journal path in isolated temp directory | migration fixture | child result is inspected read-only |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | exact AST `if` at source line 16: `if os.Getenv(crashEnvMode) == crashModeMigrationV13AfterCommit {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 18: `if version, err := j.SchemaVersion(context.Background()); err != nil \|\| version != SchemaVersion {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `if` at source line 29: `if err := old.Close(); err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 37: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B5 | exact AST `if` at source line 42: `if err := raw.QueryRow("PRAGMA user_version").Scan(&version); err != nil \|\| version != SchemaVersion {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B6 | exact AST `range` at source line 45: `for _, name := range []string{"protection_sagas", "protection_mutation_attempts", "idx_protection_sagas_live_claim"} {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B7 | exact AST `if` at source line 47: `if err := raw.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name=?`, name).Scan(&count); err != nil \|\| count != 1 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B8 | exact AST `if` at source line 53: `if after := countRows(t, reopened.db, v8AllTables); !sameCounts(before, after) {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openCrashChildJournal` | opens production migration plan | failure exits child before kill | AST |
| `runCrashChild` | runs isolated subprocess | expects SIGKILL termination | AST |
| `countRows` | proves legacy rows survived | exact count equality | AST |

## State mutations and fallbacks

- Only test fixture state changes. a047 changes the expected terminal version from v13 to current v14; v13 protection artifacts and legacy rows remain required.

## Safety conclusion

- Safe edit boundary: compare against `SchemaVersion` so future additive migrations do not stale this crash contract.
- High-risk impact: test-only verification of high-risk migration durability.
